# AGENTS.md

Rules for anyone, human or agent, changing this provider.

## Create saves the state as soon as the object exists

Once the API accepted a creation, the object exists and is billed. `Create` has to
save it in state, at least its ID, **right after the creation call** and before
any other call that can fail (reading the environment of an add-on, waiting for a
deployment, syncing tags, etc.).

Otherwise a failure of one of those later calls loses the object: Terraform has no
trace of it, the next apply creates a second one and `destroy` leaves the first
behind. With the early save, the same failure only taints the resource.

```go
res := resources.CreateAddon(ctx, r.Client(), r.Organization(), addonReq)
if res.HasError() {
	resp.Diagnostics.AddError("failed to create add-on", res.Error().Error())
	return
}
plan.ID = pkg.FromStr(pkg.Deref(res.Payload().RealID))

// the add-on exists from here on: save it before anything else can fail
resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)

// ... later calls, then a second State.Set with the complete state
```

- The only `return` allowed between the creation call and the first `State.Set`
  is the check of the creation's own error.
- A later failure adds an error diagnostic and still reaches a `State.Set`, or
  returns after one.
- `TestCreateSavesStateRightAfterCreation` in `pkg/resources` reads the `Create`
  of every resource and enforces this. When a new resource creates its object
  through a call the test does not recognise, add its name to `creationCall`
  there.
