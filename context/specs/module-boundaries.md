# Module Boundaries

## Goals
- Keep CLI code shallow and predictable; push complexity behind a few deep modules.
- Enforce invariants at module boundaries to avoid duplicated checks.
- Make dependency direction obvious and hard to violate.

## Dependency direction
- `cmd/asm` -> `internal/cli` -> `internal/asm` -> `internal/manifest`, `internal/source`, `internal/gitstore`, `internal/linker`.
- Only `internal/gitstore` imports `go-git`.

## Module responsibilities
- `internal/cli`: parse flags/args, call `asm`, format report output.
- `internal/asm`: command orchestration; no direct filesystem walking or git operations.
- `internal/manifest`: manifest + sum IO, validation, repo layout paths, state load/save.
- `internal/source`: user input parsing, GitHub tree URL parsing, skill discovery.
- `internal/gitstore`: clone/fetch/checkout, ref and lockfile resolution, replace fallback handling.
- `internal/linker`: symlink creation, pruning of managed symlinks, warning generation.

## Key invariants
- Manifest enforces unique skill names and unique `(origin, subdir)` identities.
- Subdir values are normalized and must be relative (no absolute or escape paths).
- `skills-lock.json` maps `(origin, version)` to a commit; moved tags are an error in strict mode.
- Install never deletes real files/dirs; only expected symlinks are pruned.

## Reporting
- Use-case functions return report structs; CLI is responsible for output formatting.
