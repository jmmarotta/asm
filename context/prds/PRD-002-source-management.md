---
id: PRD-002
title: Source Management (add/update/remove)
status: done # draft | ready | in_progress | done | cancelled
projectId: PROJ-001
branchName: asm-cli-mvp
owners:
  - julian
createdAt: 2026-01-21
updatedAt: 2026-01-22
stories:
  - id: US-001
    title: Add skills from local or remote sources
    acceptanceCriteria:
      - `asm add` accepts local paths and remote git URLs
      - `--path` allows selecting a subdirectory within a repo
      - `asm add` auto-syncs after successful install
    priority: 1
    passes: 1
    completedAt: 2026-01-22
    notes: ""
  - id: US-002
    title: Update or remove installed skills
    acceptanceCriteria:
      - `asm update` with no args updates all known sources
      - `asm update <name>` refreshes a specific source
      - `asm remove <name>` removes the source from store and config
    priority: 2
    passes: 1
    completedAt: 2026-01-22
    notes: ""
  - id: US-003
    title: Handle multi-skill roots deterministically
    acceptanceCriteria:
      - A path is treated as multi-skill when every direct child has `SKILL.md`
      - If multiple candidate roots exist, require `--path` to disambiguate
      - Name collisions default to `author/skill-name`
    priority: 2
    passes: 1
    completedAt: 2026-01-22
    notes: ""
---

# PRD-002: Source Management (add/update/remove)

## Summary
Implement `asm add`, `asm update`, and `asm remove` for local and remote sources, including multi-skill root detection and auto-sync after add.

## Goals
- Add skills from local paths or remote git URLs
- Support subdirectory installs via `--path`
- Refresh and remove skills by name
- Auto-sync after `asm add`

## Non-goals
- Target discovery or auto-detection
- Skill validation or packaging workflows

## Scope Boundaries
**In scope**
- Source registration and storage updates
- Multi-skill root detection rules
- Default naming on collisions

**Out of scope**
- Target linking logic (handled in PRD-003)

## User Stories
### US-001 — Add skills from local or remote sources (P1)
**Acceptance Criteria**
- `asm add` accepts local paths and remote git URLs
- `--path` allows selecting a subdirectory within a repo
- `asm add` logs that sync is not implemented yet

**Notes**
- ...

### US-002 — Update or remove installed skills (P2)
**Acceptance Criteria**
- `asm update` with no args updates all known sources
- `asm update <name>` refreshes a specific source
- `asm remove <name>` removes the source from config and cleans remote store if unused

**Notes**
- ...

### US-003 — Handle multi-skill roots deterministically (P2)
**Acceptance Criteria**
- A path is treated as multi-skill when every direct child has `SKILL.md`
- If multiple candidate roots exist, require `--path` to disambiguate
- Name collisions default to `author/skill-name`

**Notes**
- ...

## Edge Cases
- Repo contains both `skills/` and `plugins/` with valid layouts
- Remote URL points to a repo without any `SKILL.md`
- Local path contains nested skills but not all children are skills

## Verification
- Manual flow: add local path, add remote URL, update, remove

## Open Questions
None.

## Implementation Notes (Append-only)
### 2026-01-21
**Decisions**
- Use `--path` for subdirectory selection within repos.
- Treat a directory as a multi-skill root only when every direct child has `SKILL.md`.
- Prefer `skills/` or `plugins/` roots when they meet the multi-skill rule; if multiple roots qualify, require `--path`.
- `asm add` auto-syncs after install.
- When names collide, default new skill name to `author/skill-name`.

**Changes**
- Drafted source management command behavior and multi-skill rules.

**Verification**
- N/A (planning only).

**Follow-ups**
- [ ] Decide how to derive `author` for non-git local paths.

### 2026-01-22
**Decisions**
- Remote cloning uses `go-git` (no git CLI dependency), with shared store per `origin+ref`.
- Local sources reference the working tree directly with `ref: "worktree"`.
- Local origins are stored as absolute paths, using `$HOME` when applicable.
- GitHub `.../tree/...` URLs are supported; ambiguous refs error with guidance to use `origin@ref --path`.
- When a local path is a subdir of a repo, infer repo root as `origin` and store relative `subdir`.

**Changes**
- Aligned source parsing and store identity rules with the updated scope and ref decisions.

**Verification**
- N/A (planning update only).

### 2026-01-22
**Decisions**
- Local author defaults to repo name or parent directory for non-git paths.
- Local sources use `ref: "worktree"` and reference the working tree directly.
- Remote store directories are shared per `origin+ref`.

**Changes**
- Added `asm add`, `asm update`, and `asm remove` commands with scope flags.
- Implemented source parsing, multi-skill discovery, and GitHub tree URL handling.
- Added go-git remote clone/update helpers and `$HOME` path normalization for local origins.

**Verification**
- `go build ./...`

### 2026-01-22
**Changes**
- Added unit tests for source parsing, discovery, config IO, store keys, CLI flows, and go-git helpers.
- Adjusted CLI flag handling for test isolation and fixed go-git branch checkout.

**Verification**
- `go test ./...`

### 2026-01-22
**Changes**
- Prior follow-ups in Implementation Notes have been resolved; items are retained for history.

**Verification**
- N/A (documentation cleanup only).

### 2026-01-27
**Changes**
- `asm remove` accepts multiple skill names and reports each removal.
- Missing skills during removal emit warnings and do not fail the command.
- Removal reporting and CLI tests updated to cover multi-skill and no-op flows.

**Verification**
- `go test ./...`
