# Contributing

## API → state mapping

Reading the API and filling Terraform state is the provider's most repetitive
job, and the one that has produced the most user-visible damage: an attribute a
read forgets stays `null`, the schema default materialises on the next plan, and
where the attribute carries `RequiresReplace` the plan is a destroy and recreate
of a live resource (#404, #452). `Read()` is the only import path in this
provider — `helper.Configurer.ImportState` passes the id through and nothing else
— so a mapper is the single source of truth for every attribute after an import.

It is written nine different ways across the repository today. This is the one
convention it is converging on, starting with the application runtimes:

```go
func (s *T) From<APIType>(ctx context.Context, payload *<APIType>, diags *diag.Diagnostics)
```

### The rules

1. **Pointer receiver, no return value.** A mapper fills the struct it is called
   on; there is nothing to hand back.

   ```go
   state := helper.StateFrom[Elasticsearch](ctx, req.State, &resp.Diagnostics)

   state.FromAddon(ctx, addonRes.Payload(), &resp.Diagnostics)
   state.FromElasticsearch(ctx, esRes.Payload(), &resp.Diagnostics)
   ```

   Returning the receiver so the calls could chain was tried and dropped: it cost
   a `return` line per mapper, bought one saved repetition of a short variable
   name, and could not apply to the mappers reached through an interface —
   `application.Read[T]` goes through `RuntimePlan` and `ResourceDrain[T]`
   through `DrainAttributes`, and an interface method cannot return the concrete
   type.

2. **Name the method after the API response type, never after the endpoint**:
   `FromApp(*tmp.AppResponse)`, `FromAddon(*tmp.AddonResponse)`,
   `FromDrain(tmp.Drain)`. An endpoint name rots the day the API moves a field
   between views; a type name is checked by the compiler.

3. **`ctx` and `diags` are always present**, even where the body uses neither, so
   every signature in the repository is identical and a reader can pattern-match
   instead of reading.

4. **The guard is always these three lines:**

   ```go
   if s == nil || payload == nil {
       return
   }
   ```

   `payload == nil` is the one that matters: a CRUD whose fetch failed passes
   `nil`, and prior state must survive untouched. A failed call must never blank
   state.

5. **Never short-circuit on `diags.HasError()`.** Each mapper maps an
   independent payload; that an earlier one failed is no reason to skip the next,
   and stopping mid-chain yields half-filled state — exactly what we are trying
   to avoid. The errors abort the CRUD at the `resp.Diagnostics.HasError()` check
   that follows. Pass `&resp.Diagnostics` straight in; there is no reason for a
   chain-local accumulator.

6. **A mapper makes no API call and has no side effect.** That is what lets
   every one of them be unit-tested with no server. Fetching, retrying,
   `RemoveResource`, status checks and the `Sync*` calls all stay in the CRUD,
   above the mappers. A mapper never removes a resource.

7. **Mappers live in `schema.go`**, after the state struct and the schema, next
   to the `ToEnv` they are the mirror of. They do not get a file of their own —
   a file per resource buys a header and an import block and little else, and
   keeping each mapper beside its opposite direction is worth more.

### Which absent-value policy

The hard part of a mapper is not the assignment, it is deciding what an absent
value means. For a boolean carried by an environment variable the split is
pinned by `pkg/setbool_test.go`:

- attribute **with** a schema `Default` → `pkg.SetBool` (assigns either way)
- attribute **without** one → `pkg.SetBoolIf` (assigns only on a match, so
  `null` keeps meaning "never configured")

These were decided attribute by attribute in #462 — do not flip one as a
drive-by.

### The runtimes' `FromEnv` is destructive, deliberately

`FromEnv(ctx, env *maps.Map[string, string], diags)` **pops** every key it owns
out of `env`, and whatever remains once `application.Read` has run `FromEnv`,
`FromEnvHooks` and `FromEnvIntegrations` becomes the catch-all `environment`
attribute. Adding a `PopPtr` silently removes a key from `environment`; reading
a key without popping it silently duplicates it into `environment`. Never change
one without the other.

### What a mapper must guarantee

Every mapper carries unit tests, named the same way everywhere, after
`pkg/resources/drain/schema_test.go`:

- `TestXFromAPI_PopulatesEveryAttribute` — zero state plus a full payload:
  nothing the API returns is left null.
- `TestXFromAPI_PreservesWhatTheAPIDoesNotReturn` — populated state plus a
  payload: sensitive attributes and attributes the API never returns keep their
  state value. The API is not the authority on a value it does not send.
- `TestXFromAPI_NilPayloadIsANoOp` — a failed fetch must leave prior state alone.

Resources whose API answers a sparse named-toggle list carry a fourth,
`TestXFromAPI_AbsentFeatureUsesFallback`: every `Computed` attribute is non-null
and equal to its schema default. That is the #404 regression test.
