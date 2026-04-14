# github.com/tezra-io/usg

## Project
- **Summary:** Project overview still needs a tighter one-liner.
- **Stack:** Go
- **Manifests:** go.mod

## Behavioral Guidance
- The approved design is the plan. Implement against it, do not quietly re-design the task mid-flight.
- Don't assume. State assumptions explicitly before coding. If multiple interpretations exist, surface them instead of picking silently.
- If the request or design is unclear, stop and ask. If repo reality conflicts with the design, surface the mismatch before coding.
- Prefer the simplest correct solution. No speculative abstractions, no extra flexibility, no "while I'm here" cleverness.
- Make surgical changes. Touch only what the request requires. Mention unrelated issues, don't fix them unless asked.
- For multi-step work, define success in `step -> verify` form and keep going until the checks pass.
- If 200 lines could be 50, rewrite it.

## Execution Contract
- If changing behavior, write or update a failing test first.
- Implement the smallest change that satisfies the design.
- Run the relevant repo commands below before calling the work done. Default expectation: typecheck or build, tests, and lint.
- For docs, config, or scaffolding changes, run the relevant checks and say what is not applicable.
- Never mark work done without proof.

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

## Known Pitfalls
- Update this section every time the repo teaches you the same lesson twice.

---
_Every mistake is a rule waiting to be written._
