# AGENTS.md

## Overview
- Repo is a Go CLI (`asm`) for managing repo-local skill sources and installs.
- Module name: `github.com/jmmarotta/agent_skills_manager` (Go `1.24.9`).
- CLI commands live under `internal/cli`, entrypoint in `cmd/asm`.
- Manifest lives at `skills.jsonc` (+ `skills-lock.json`), store/cache under `.asm/`, install under `./skills`.
- Project planning docs live in `context/` (PRDs are append-only in Implementation Notes).

## Repo layout (high level)
- `cmd/asm/` — CLI entrypoint (`main.go`).
- `internal/cli/` — Cobra commands and flag parsing (prints reports).
- `internal/asm/` — use-case layer (init/add/update/remove/install/list/show).
- `internal/manifest/` — manifest + sum IO, validation, repo layout paths.
- `internal/source/` — source parsing, GitHub tree URLs, discovery.
- `internal/gitstore/` — go-git clone/ref/lockfile resolution, store key/path helpers.
- `internal/linker/` — symlink link + prune logic.
- `context/` — PRDs, project docs, outstanding items.

## Build / lint / test commands
### Make targets
- `make` (fmt, vet, test, build)
- `make check` (fmt-check, vet, test, build)
- `make test-one PKG=./internal/cli RUN=TestName`
- `make test-one PKG=./internal/cli`

### Build
- `go build ./...`
- `go build -o asm ./cmd/asm`

### Test (all)
- `go test ./...`

### Test (single package)
- `go test ./internal/cli -run TestName -count=1`
- `go test ./internal/linker -run TestSyncCreatesSymlink -count=1`

### Format / lint
- `gofmt -w cmd internal`
- `go vet ./...` (optional, no dedicated lint config present)
- No `golangci-lint` configuration in this repo.

## Code style guidelines
### Formatting
- Always run `gofmt` on Go files.
- Use tabs for indentation (gofmt default).
- Keep line lengths readable; wrap long error messages if needed.

### Imports
- Group imports as: standard library, blank line, third-party, blank line, internal (`asm/...`).
- Use explicit imports (no dot/underscore imports).

### Naming
- Types use `CamelCase` (e.g., `ScopedSource`, `Result`).
- Functions use `camelCase` (`resolveSyncTargets`, `parseAddInput`).
- Constants for flag names are lowercase strings (`addLocalFlag`, `syncScopeFlag`).
- Errors use lowercase messages without punctuation.

### CLI patterns
- Use Cobra for commands; register in `internal/cli/root.go`.
- Read flags directly from `cmd.Flags()` (tests rely on this).
- Use `cmd.OutOrStdout()` for normal output.
- Use `cmd.ErrOrStderr()` for warnings.
- Avoid interactive prompts; keep commands scriptable.

### Manifest conventions
- Manifest file preferred: `skills.jsonc`; fallback to `skills.json`.
- Use `manifest.Load` / `manifest.Save` / `manifest.LoadState` for IO.
- `skills-lock.json` stores origin/version -> commit resolution.
- Validate manifests via `Config.Validate()`; fail fast on missing fields.

### Source handling
- Local sources use `type:"path"` and `ref:"worktree"`.
- Remote sources use `type:"git"` and optional `ref` from `origin@ref`.
- Use `source.ParseInput` and `source.DiscoverSkills` to resolve inputs.
- GitHub tree URLs should be parsed via `source.ParseGitHubTreeURL`.

### Store handling
- Remote store identity is `origin` (`gitstore.RepoKey`).
- Local sources do not use the store; installs read from the working tree.
- Store paths live under `.asm/store`.

### Sync behavior
- `linker.Sync` handles symlink creation/replacement safely.
- Destinations are `<target>/<source.name>`; names may include `/`.
- Never delete real files/dirs during sync; warn and skip instead.
- `linker.Prune` only removes expected symlinks for removed skills.

## Error handling
- Prefer `return err` for propagation; wrap with `fmt.Errorf` for context.
- Use `os.IsNotExist` for missing files/dirs.
- Treat per-item conflicts as warnings (stderr), not hard failures.
- Exit codes: non-zero only for hard failures (config errors, invalid flags, etc.).

## Testing guidelines
- Use `t.TempDir()` for filesystem fixtures.
- Use `t.Setenv("HOME", tempDir)` to avoid touching real user config.
- Use `setWorkingDir` helper for repo-local behavior in CLI tests.
- Avoid `t.Parallel()` in CLI tests (global cobra state).
- For go-git tests, create local repos with go-git (no network).

## Dependency guidelines
- Use `go-git` for git operations; do not shell out to `git`.
- Keep dependencies minimal; avoid adding new ones unless needed.

## Cursor/Copilot rules
- No `.cursor/rules`, `.cursorrules`, or `.github/copilot-instructions.md` found.
- If added later, update this file to reflect them.

## Context docs
- PRD Implementation Notes are append-only.
- Outstanding items live in `context/OUSTANDING.md`.
