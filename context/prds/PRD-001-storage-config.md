---
id: PRD-001
title: Storage + Config Layout
status: done # draft | ready | in_progress | done | cancelled
projectId: PROJ-001
branchName: asm-cli-mvp
owners:
  - julian
createdAt: 2026-01-21
updatedAt: 2026-01-22
stories:
  - id: US-001
    title: Resolve scope and storage paths
    acceptanceCriteria:
      - Default scope is global; `--local` uses repo-local scope when inside a git repo
      - Global paths follow XDG-style locations under ~/.local/share/asm, ~/.config/asm, ~/.cache/asm
      - Local scope uses .asm/ in repo root with store/, config.jsonc (or config.json), and cache/
    priority: 1
    passes: 1
    completedAt: 2026-01-22
    notes: ""
  - id: US-002
    title: Persist source and target metadata
    acceptanceCriteria:
      - Config records sources with type, origin, name, optional subdir, and optional ref
      - Local path sources set ref to worktree
      - Config records targets with explicit name and path
      - Name collisions default to author/skill-name
    priority: 2
    passes: 1
    completedAt: 2026-01-22
    notes: ""
---

# PRD-001: Storage + Config Layout

## Summary
Define the storage layout and configuration schema for the ASM CLI, including global (XDG-style) and local (repo) scopes plus naming collision rules.

## Goals
- Specify global and local paths for store, config, and cache
- Define config schema for sources and targets
- Establish deterministic naming rules for collisions

## Non-goals
- Implement git cloning or skill syncing
- Define target auto-detection behavior

## Scope Boundaries
**In scope**
- Path resolution for global vs local scopes
- Config persistence for sources and targets
- Naming collision defaults

**Out of scope**
- Network fetching logic
- Symlink creation or target syncing

## User Stories
### US-001 — Resolve scope and storage paths (P1)
**Acceptance Criteria**
- Default scope is global; `--local` uses repo-local scope when inside a git repo
- Global paths follow XDG-style locations under `~/.local/share/asm`, `~/.config/asm`, `~/.cache/asm`
- Local scope uses `.asm/` in repo root with `store/`, `config.jsonc` (or `config.json`), and `cache/`

**Notes**
- ...

### US-002 — Persist source and target metadata (P2)
**Acceptance Criteria**
- Config records sources with type, origin, name, optional subdir, and optional ref
- Local path sources set `ref` to `worktree`
- Config records targets with explicit name and path
- Name collisions default to `author/skill-name`

**Notes**
- ...

## Edge Cases
- Repo detection fails in nested worktrees or submodules
- Local config or store path is not writable
- Migrating between global and local scopes

## Verification
- Manual inspection of resolved paths and config contents

## Open Questions
None.

## Implementation Notes (Append-only)
### 2026-01-21
**Decisions**
- Default scope is local when inside a git repo; `--global` overrides.
- Global paths use XDG-style roots: `~/.local/share/asm`, `~/.config/asm`, `~/.cache/asm`.
- Local scope uses `.asm/` with `store/`, `config.json`, and `cache/`.
- When names collide, default new skill name to `author/skill-name`.

**Changes**
- Drafted storage/config requirements and naming rules.

**Verification**
- N/A (planning only).

**Follow-ups**
- [ ] Finalize config schema fields for source metadata.

### 2026-01-21
**Decisions**
- `asm add` defaults to global scope; `--local` opt-in and `--global` still supported.
- Config prefers `config.jsonc` but falls back to `config.json` if it exists alone.
- Source `ref` is parsed from `origin@ref` for remote URLs; local paths omit `ref`.
- Store identity is derived from `type + origin + subdir` (no persisted `sourceId`).

**Changes**
- Defined config schema fields (`type`, `origin`, optional `subdir`, optional `ref`, resolved `name`).
- Clarified scope behavior and ref parsing rules.

**Verification**
- N/A (planning update only).

### 2026-01-21
**Decisions**
- Implementation uses Go with Cobra + Viper; JSONC parsing via `github.com/tidwall/jsonc`.
- Local path sources set `ref` to HEAD SHA when in a git repo.
- Store keys use full SHA-256 hex for `type|origin|subdir`.

**Changes**
- Added Go module scaffolding and CLI entry points for list/show/target list.
- Implemented config loading/validation, scope resolution, path helpers, source parsing, and store key derivation.

**Verification**
- `go build ./...`

**Follow-ups**
- [ ] Implement config read/write with JSONC preference and fallback.
- [ ] Implement scope resolution and store path derivation.

### 2026-01-22
**Decisions**
- Local path sources use `ref: "worktree"` and reference the working tree directly.
- Store directories are shared per `origin+ref`.
- Local origins are stored as absolute paths using `$HOME` when applicable.

**Changes**
- Implemented config persistence and scope-resolved add/update/remove flows.
- Added origin normalization and store key derivation updates.

**Verification**
- `go build ./...`
- `go run ./cmd/asm list`

### 2026-01-21
**Decisions**
- Read commands (`list`, `show`, `target list`) merge local+global when in a repo; local entries win on name collisions.
- Mutating commands default to global; `--local` is required to modify repo-local config/store.
- `asm show` resolves to the effective source (local if present), with `--scope` to override.
- CLI output includes `scope: local|global` for each entry.

**Changes**
- Clarified scope precedence and default behaviors for read vs mutating commands.

**Verification**
- N/A (planning update only).

### 2026-01-22
**Changes**
- Prior follow-ups in Implementation Notes have been resolved; items are retained for history.

**Verification**
- N/A (documentation cleanup only).
