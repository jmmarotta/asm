# asm

Manage skill dependencies in a repository (Go/Bun-style).

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
# initialize a repo
asm init
asm init --cwd /path/to/repo

# add a skill (local path or git url)
asm add /path/to/skills-repo
asm add https://github.com/org/repo.git@v1.2.3 --path plugins/foo

# install to ./skills
asm install

# inspect configured skills
asm ls
asm show foo
```

GitHub tree URLs are supported:
```sh
asm add https://github.com/org/repo/tree/main/plugins/foo
```

## Files
- `skills.jsonc` (manifest; fallback `skills.json`)
- `skills-lock.json` (resolved revisions; go.sum analogue)
- `.asm/` (store + cache)
- `skills/` (installed symlinks; gitignored)

## Manifest
```jsonc
{
  "skills": [
    {
      "name": "author/skill",
      "type": "git",
      "origin": "https://github.com/org/repo",
      "subdir": "plugins/foo",
      "version": "v1.2.3"
    }
  ],
  "replace": {
    "https://github.com/org/repo": "../local-repo"
  }
}
```

Notes:
- `type:"git"` requires `version` (semver tag or Go pseudo-version).
- `type:"path"` uses `origin` as a local directory (non-portable).
- `replace` is best-effort: if the path is missing, installs fall back to remote.

## Commands
- `asm init [--cwd path]`
- `asm add <path-or-url> [--path subdir]`
- `asm update [name]`
- `asm remove <name> [<name>...]`
- `asm install`
- `asm ls`
- `asm show <name>`

Aliases: `add` = `a`, `install` = `i`, `remove` = `rm`/`uninstall`, `update` = `up`.

## Install behavior
- Skills are symlinked to `skills/<name>`.
- Names with slashes (e.g. `author/skill`) create nested directories.
- If a destination exists and is not a symlink, install skips it and prints a warning to stderr.
- `asm install` prunes unmanaged symlinks under `skills/`.

## Private repositories
Auth is not explicitly configured for git fetches yet; public repos work by default.
Use local paths or `replace` overrides for private repositories until auth support is added.

## Development
See `docs/development.md` for build/test commands and contributor notes.
