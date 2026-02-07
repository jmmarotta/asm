# asm

Manage repo-local skill dependencies with a manifest + lockfile.

A skill is a directory containing `SKILL.md`. Repos can expose a single skill at the root or multiple skills under `skills/` or `plugins/`.

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

# find skills
asm find react testing
```

GitHub tree URLs are supported:
```sh
asm add https://github.com/org/repo/tree/main/plugins/foo
```

## Init behavior
- Creates `skills.jsonc` if it doesn't exist.
- Creates `.asm/` and `skills/` directories.
- Appends `.asm/` and `skills/` to `.gitignore`.

## Skill discovery
- `asm add` scans for `SKILL.md` files.
- If the repo root is a skill directory, it adds that single skill.
- If the repo contains `skills/` or `plugins/` and every child has `SKILL.md`, it adds all of them.
- If multiple skill roots are found, use `--path` to target a subdirectory.
- `--path` can point at a single skill directory or a multi-skill root.

## Files
- `skills.jsonc` (manifest; fallback `skills.json`)
- `skills-lock.json` (resolved revisions; lockfile)
- `.asm/` (store + cache)
- `skills/` (installed symlinks; gitignored)

## Reproducible installs
- Commit `skills.jsonc` and `skills-lock.json`.
- `.asm/` and `skills/` are generated and should stay gitignored.
- `asm install` uses the lockfile.
- `asm update` advances pseudo-version skills to latest HEAD and refreshes the lockfile.
- Semver-tagged skills stay pinned by default; target them explicitly to unpin.

## Manifest
```jsonc
{
  "skills": [
    {
      "name": "author/skill",
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
- `version` is required for git sources (semver tag or pseudo-version like `v0.0.0-YYYYMMDDHHMMSS-abcdef123456`).
- Omit `version` for local path sources; `origin` is the directory (non-portable).
- `replace` is best-effort: if the path is missing, installs fall back to remote.

## Commands
- `asm init [--cwd path]`
- `asm add <path-or-url> [--path subdir]`
- `asm update [name|origin] [--path subdir]`
- `asm remove <name> [<name>...]`
- `asm install`
- `asm find <query...>`
- `asm ls`
- `asm show <name>`

Aliases: `add` = `a`, `find` = `f`, `install` = `i`, `remove` = `rm`/`uninstall`, `update` = `up`.

## Install behavior
- Skills are symlinked to `skills/<name>`.
- Names with slashes (e.g. `author/skill`) create nested directories.
- If a destination exists and is not a symlink, install skips it and prints a warning to stderr.
- `asm install` prunes unmanaged symlinks under `skills/`.

## Private repositories
asm uses go-git and picks up auth from common sources without storing credentials in `skills.jsonc`.

HTTPS:
- GitHub tokens: `ASM_GITHUB_TOKEN`, `GITHUB_TOKEN`, or `GH_TOKEN`
- Generic tokens: `ASM_GIT_TOKEN` (optionally `ASM_GIT_USERNAME`)
- `.netrc` entries for the host

SSH:
- Uses your SSH agent (`SSH_AUTH_SOCK`) and `~/.ssh/known_hosts`
- Optional escape hatch: set `ASM_SSH_INSECURE=1` to skip host key verification

Git config rewriting:
- Global `url.<base>.insteadOf` rules are honored, so `https://github.com/...` can transparently use SSH.

## Development
See `docs/development.md` for build/test commands and contributor notes.
