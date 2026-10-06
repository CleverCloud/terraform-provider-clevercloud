# Contributing

## API → state mapping

Reading the API and filling Terraform state is the provider's most repetitive
job, and the one that has produced the most user-visible damage: an attribute a
read forgets stays `null`, the schema default materialises on the next plan, and
where the attribute carries `RequiresReplace` the plan is a destroy and recreate
of a live add-on (#404, #452). `Read()` is the only import path in this provider
— `helper.Configurer.ImportState` passes the id through and nothing else — so a
mapper is the single source of truth for every attribute after an import.

One convention, everywhere:

```go
func (s *T) From<APIType>(ctx context.Context, payload *<APIType>, diags *diag.Diagnostics) *T
```

### The rules

1. **Pointer receiver, no return value.** A mapper fills the struct it is called
   on; there is nothing to hand back:

   ```go
   state := helper.StateFrom[Elasticsearch](ctx, req.State, &resp.Diagnostics)

   state.FromAddon(ctx, addonRes.Payload(), &resp.Diagnostics)
   state.FromElasticsearch(ctx, esRes.Payload(), &resp.Diagnostics)
   ```

   Returning the receiver so the calls could chain was tried and dropped: it cost
   a `return` line per mapper and a blank line before it, bought one saved
   repetition of a short variable name, and could not apply to the two families
   reached through an interface (see below) — so the convention was uniform
   nowhere and shorter nowhere that mattered.

2. **Name the method after the API response type, never after the endpoint**:
   `FromAddon(*tmp.AddonResponse)`, `FromMySQL(*tmp.MySQL)`, `FromDrain(tmp.Drain)`.
   An endpoint name rots the day the API moves a field between views; a type name
   is checked by the compiler.

3. **`ctx` and `diags` are always present**, even where the body uses neither, so
   every signature in the repository is identical and a reader can pattern-match
   instead of reading.

4. **The guard is always these three lines:**

   ```go
   if s == nil || payload == nil {
       return s
   }
   ```

   `payload == nil` is the one that matters: a CRUD whose fetch failed passes
   `nil`, and prior state must survive untouched. A failed call must never blank
   state.

5. **Never short-circuit on `diags.HasError()`.** Each mapper maps an
   independent payload; that an earlier one failed is no reason to skip the next,
   and stopping mid-chain yields half-filled state — exactly what we are trying
   to avoid. The errors abort the CRUD at the `resp.Diagnostics.HasError()` check
   after the chain. Pass `&resp.Diagnostics` straight in; there is no reason for
   a chain-local accumulator.

6. **A mapper makes no API call and has no side effect.** That is what lets
   every one of them be unit-tested with no server. Fetching, retrying,
   `RemoveResource`, status checks and the `Sync*` calls all stay in the CRUD,
   above the chain. A mapper never removes a resource.

7. **Mappers live in `schema.go`**, after the state struct and the schema, next
   to the `ToEnv` they are the mirror of. Order: the generic add-on view first,
   then the product type, then `FromEnv`. They do not get a file of their own —
   a file per resource buys a header and an import block and little else, and
   keeping each mapper beside its opposite direction is worth more.

Two families are reached through an interface rather than on their concrete
type: `ResourceDrain[T]` is generic over `DrainAttributes`, and
`application.Read[T]` over `RuntimePlan`. That is what settled rule 1 — an
interface method cannot return the concrete type, so a chaining convention could
never have covered them. They keep their mappers where they are: the drains' next
to their seven structs in `drain/schema.go`, each runtime's beside its own
`ToEnv`.

### Which absent-value policy

The hard part of a mapper is not the assignment, it is deciding what an absent
value means. There are four answers and picking the wrong one is how #404
happened. For a named-toggle list (`tmp.MySQLFeature` and friends):

| Attribute shape | Use | Absent key means |
| --- | --- | --- |
| `Computed`, with or without a schema `Default` | `pkg.Features.Or(&s.X, "key", fallback)` | the fallback — which **must** equal the schema default where there is one |
| `Optional`, no default, not `Computed` | `pkg.Features.Keep(&s.X, "key")` | leave it alone: `null` still means "never configured" |
| emptiness is meaningful, no default | `pkg.Features.Null(&s.X, "key")` | `null` |
| gates another attribute rather than being one | `pkg.Features.Enabled("key")` | disabled |

For a boolean carried by an environment variable, the same split already exists
and is pinned by `pkg/setbool_test.go`:

- attribute **with** a schema `Default` → `pkg.SetBool` (assigns either way)
- attribute **without** one → `pkg.SetBoolIf` (assigns only on a match)

`Features.Or` is to a toggle what `SetBool` is to a variable; `Features.Keep`
what `SetBoolIf` is. These choices were made attribute by attribute in #462 —
do not flip one as a drive-by.

### Environment variables come in two flavours, deliberately

The parameter type tells them apart, so handing a mapper the wrong one is a
compile error:

- **Application runtimes — destructive.** `FromEnv(ctx, env *maps.Map[string, string], diags)`.
  Every key the method owns is `PopPtr`-ed out, and whatever remains once
  `application.Read` has run `FromEnv`, `FromEnvHooks` and `FromEnvIntegrations`
  becomes the catch-all `environment` attribute. Adding a `PopPtr` silently
  removes a key from `environment`; reading a key without popping it silently
  duplicates it into `environment`. Never change one without the other.
- **Add-ons — non-destructive.** `FromEnv(ctx, env tmp.EnvVars, diags)`, usually
  via `env.Map()`. An add-on has no catch-all attribute to receive a residue, so
  there is nothing to pop and nothing to preserve.

### What a mapper must guarantee

Every mapper carries three unit tests, named the same way everywhere, after
`pkg/resources/drain/schema_test.go`:

- `TestXFromAPI_PopulatesEveryAttribute` — zero state plus a full payload: nothing
  the API returns is left null.
- `TestXFromAPI_PreservesWhatTheAPIDoesNotReturn` — populated state plus a
  payload: sensitive attributes and attributes the API never returns keep their
  state value. The API is not the authority on a value it does not send.
- `TestXFromAPI_AbsentFeatureUsesFallback` — a sparse or empty toggle list: every
  `Computed` attribute is non-null and equal to its schema default.

`pkg/registry/mapping_contract_test.go` is the cross-resource half: every
registered resource has to name the state struct it fills, and that struct has
to carry at least one `From*` method whose signature follows the convention. So
a new resource cannot ship without someone looking at its mapping, and a mapper
cannot be renamed or deleted unnoticed. A resource that genuinely reads nothing
back goes in its `unmapped` table, with the reason.

It deliberately proves nothing about the values: read its doc comment for what
it does *not* cover, and do not let its green substitute for the three cases
above.
