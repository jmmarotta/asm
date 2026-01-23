# asm

Manage skill sources and sync them into target directories via symlinks.

## Status
- MVP CLI for source management, targets, and sync.
- Planning docs and open questions live in `context/`.

## Install
```sh
go install github.com/jmmarotta/agent_skills_manager/cmd/asm@latest
```

For reproducible installs, pin a version:
```sh
go install github.com/jmmarotta/agent_skills_manager/cmd/asm@v0.1.0
```

Ensure `$(go env GOPATH)/bin` (or `GOBIN`) is on your `PATH`.

## Quickstart
```sh
# add a source (local path or git url)
asm add /path/to/skills-repo --path plugins/foo

# add a target
asm target add dotfiles ~/.config/skills

# sync all sources into all targets
asm sync
```

GitHub tree URLs are supported:
```sh
asm add https://github.com/org/repo/tree/main/plugins/foo
```

## Concepts
- **Sources**
  - `path`: references a working tree directly (`ref: worktree`), no store clone.
  - `git`: shallow clone stored in the ASM store, optional `@ref` pin.
- **Targets**
  - Named output directories for symlinked skills.

## Scopes
- Mutating commands default to **global** scope; use `--local` to target repo-local config.
- Read commands inside a repo merge local+global; local wins on conflicts.
- `asm sync --scope local|global` filters **sources** only; targets remain effective.

## Paths & Config
Global scope (default):
- Config: `~/.config/asm/config.jsonc` (fallback `config.json`)
- Store: `~/.local/share/asm/store/`
- Cache: `~/.cache/asm/`

Local scope (when `--local` is used inside a repo):
- Config: `<repo>/.asm/config.jsonc` (fallback `config.json`)
- Store: `<repo>/.asm/store/`
- Cache: `<repo>/.asm/cache/`

Local path origins and target paths are stored as `$HOME/...` when possible.

## Commands
- `asm add <path-or-url> [--path subdir] [--local|--global]`
- `asm update [name] [--local|--global]`
- `asm remove <name> [--local|--global]`
- `asm list [--scope local|global|all]`
- `asm show <name> [--scope local|global]`
- `asm target add <name> <path> [--local|--global]`
- `asm target remove <name> [--local|--global]`
- `asm target list [--scope local|global|all]`
- `asm sync [--scope local|global]`

## Sync behavior
- Skills are symlinked to `<target>/<source.name>`.
- Names with slashes (e.g. `author/skill`) create nested directories.
- If a destination exists and is not a symlink, sync skips it and prints a warning to stderr.
- Exit code is `0` for warnings; non-zero for hard errors.

## Private repositories
Auth is not explicitly configured for git fetches yet; public repos work by default.
Use local paths for private repositories until auth support is added.

## Development
See `docs/development.md` for build/test commands and contributor notes.

Project docs live in `context/`.
