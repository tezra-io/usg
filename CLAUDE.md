# github.com/tezra-io/usg

## Project
- **Summary:** Project overview still needs a tighter one-liner.
- **Stack:** Go
- **Manifests:** go.mod

## How to Work

### Planning
- Plan mode for any non-trivial task (3+ steps or architectural decisions)
- Detailed specs upfront — good plan = 1-shot implementation
- State assumptions explicitly before coding. If multiple interpretations exist, surface them instead of picking silently.
- If the request is ambiguous, ask. If a simpler approach exists, say so.
- For multi-step work, write a short plan in `step -> verify` form.
- If something goes sideways, STOP and re-plan

### Test-First (Mandatory)
1. Write failing tests that define correct behavior
2. Make them pass
3. Refactor while green

"Write failing tests, then make them pass" — not "implement this feature."

### Verification
1. Write failing tests
2. Implement to pass them
3. Typecheck: `go test ./...`
4. Full test suite: `go test ./...`
5. Lint: `go vet ./...`

Never mark done without proving it works.

## Code Rules (Non-Negotiable)

1. **Linear flow.** Max 2 nesting levels. Top to bottom.
2. **Bound loops.** Explicit max on retries, polls, recursion. Define cap behavior.
3. **Small functions.** 40-60 lines max. One job per function.
4. **Own resources.** Open → close on every path, including errors.
5. **Narrow state.** No module globals. Pass deps explicitly.
6. **Assert assumptions.** Guards and validation on every public function. Fail loud.
7. **Never swallow errors.** No bare `rescue`. No `{:error, _} -> :ok`. Log, raise, or return.
8. **Visible side effects.** I/O obvious at call site. Separate pure from effectful.
9. **Minimal indirection.** Readable > elegant. One layer of abstraction max.
10. **Surgical changes only.** Touch only what the request requires. Do not refactor adjacent code, comments, or formatting unless the task needs it. Remove only the dead code your change creates.
11. **Warnings = errors.** Linters, typecheckers, analyzers are hard gates. Zero warnings.

## Conventions
- Prefer small packages, explicit error returns, and standard library first.
- Keep interfaces narrow and only introduce them when multiple concrete implementations exist.

## Commands
```sh
go build ./...
go test ./...
go vet ./...
test -z "$(gofmt -l .)"
```

## Docs
- `docs/spec.md` — Product spec: features, business rules
- `docs/tech.md` — Architecture: stack, schema, decisions
- `docs/lessons.md` — Rules from past mistakes (update immediately on correction)

## Don'ts
- Don't commit without running tests
- Don't implement without failing tests first
- Don't add abstractions you weren't asked for
- Don't silently choose among ambiguous interpretations
- Don't improve adjacent code that wasn't part of the request
- Don't assume intent on ambiguous bugs — ask

## Principles
- Simplest correct solution
- If 200 lines could be 50, rewrite it
- Find root causes, no band-aids
- Minimal blast radius
- Own mistakes — write a rule to prevent repeating

## Known Pitfalls
- Update this section every time the repo teaches you the same lesson twice.

---
_Every mistake is a rule waiting to be written._
