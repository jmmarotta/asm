# AGENTS.md

## Overview
- Repo is a Go CLI (`asm`) for managing skill sources and sync targets.
- Module name: `github.com/jmmarotta/agent_skills_manager` (Go `1.24.9`).
- CLI commands live under `internal/cli`, entrypoint in `cmd/asm`.
- Config and storage live under `~/.config/asm`, `~/.local/share/asm`, `~/.cache/asm`.
- Project planning docs live in `context/` (PRDs are append-only in Implementation Notes).

## Repo layout (high level)
- `cmd/asm/` — CLI entrypoint (`main.go`).
- `internal/cli/` — Cobra commands and flag parsing.
- `internal/config/` — config structs, IO, validation, merge helpers.
- `internal/source/` — source parsing, GitHub tree URLs, discovery.
- `internal/remote/` — go-git clone/ref helpers.
- `internal/store/` — store key/path helpers for git sources.
- `internal/syncer/` — pure sync/cleanup logic (symlink handling).
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
- `go test ./internal/syncer -run TestSyncCreatesSymlink -count=1`

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

### Config conventions
- Config file preferred: `config.jsonc`; fallback to `config.json`.
- Use `config.Load` / `config.Save` for all config IO.
- Normalize `$HOME/...` for `type:"path"` origins and `Target.Path` when saving.
- Expand `$HOME` paths immediately after loading.
- Validate configs via `Config.Validate()`; fail fast on missing fields.

### Source handling
- Local sources use `type:"path"` and `ref:"worktree"`.
- Remote sources use `type:"git"` and optional `ref` from `origin@ref`.
- Use `source.ParseInput` and `source.DiscoverSkills` to resolve inputs.
- GitHub tree URLs should be parsed via `source.ParseGitHubTreeURL`.

### Store handling
- Remote store identity is `origin + ref` (`store.RepoKey`).
- Local sources do not use the store; sync reads from the working tree.
- Store paths are created under scope-specific store roots.

### Sync behavior
- `syncer.Sync` handles symlink creation/replacement safely.
- Destinations are `<target>/<source.name>`; names may include `/`.
- Never delete real files/dirs during sync; warn and skip instead.
- Cleanup only removes expected symlinks for a removed target.

## Error handling
- Prefer `return err` for propagation; wrap with `fmt.Errorf` for context.
- Use `os.IsNotExist` for missing files/dirs.
- Treat per-item conflicts as warnings (stderr), not hard failures.
- Exit codes: non-zero only for hard failures (config errors, invalid flags, etc.).

## Testing guidelines
- Use `t.TempDir()` for filesystem fixtures.
- Use `t.Setenv("HOME", tempDir)` to avoid touching real user config.
- Use `setWorkingDir` helper for repo-local behavior in CLI tests.
- Avoid `t.Parallel()` in CLI tests (global cobra/viper state).
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
