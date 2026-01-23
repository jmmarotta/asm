---
id: PRD-003
title: Targets + Sync
status: done # draft | ready | in_progress | done | cancelled
projectId: PROJ-001
branchName: asm-cli-mvp
owners:
  - julian
createdAt: 2026-01-21
updatedAt: 2026-01-22
stories:
  - id: US-001
    title: Register explicit targets
    acceptanceCriteria:
      - `asm target add <name> <path>` requires a name and path
      - `asm target list` shows registered targets and paths
      - `asm target remove <name>` removes the target from config
    priority: 1
    passes: 1
    completedAt: 2026-01-22
    notes: ""
  - id: US-002
    title: Sync skills to targets via symlinks
    acceptanceCriteria:
      - `asm sync` links installed skills into each target path
      - `asm sync --scope local|global` limits the source store
      - Sync replaces stale or incorrect symlinks safely
    priority: 2
    passes: 1
    completedAt: 2026-01-22
    notes: ""
---

# PRD-003: Targets + Sync

## Summary
Provide explicit target management and syncing so skills from the central store can be symlinked into dotfiles or agent directories without auto-detection.

## Goals
- Register targets with required name + path
- Sync skills to targets using symlinks
- Support local vs global scope selection during sync

## Non-goals
- Auto-detect agent installations or paths
- Copy-based syncing (non-symlink)

## Scope Boundaries
**In scope**
- CRUD for target definitions
- Symlink creation and refresh during `asm sync`
- Removing target entries from config

**Out of scope**
- Agent-specific path discovery
- Cross-device fallback behavior

## User Stories
### US-001 — Register explicit targets (P1)
**Acceptance Criteria**
- `asm target add <name> <path>` requires a name and path
- `asm target list` shows registered targets and paths
- `asm target remove <name>` removes the target from config

**Notes**
- ...

### US-002 — Sync skills to targets via symlinks (P2)
**Acceptance Criteria**
- `asm sync` links installed skills into each target path
- `asm sync --scope local|global` limits the source store
- Sync replaces stale or incorrect symlinks safely

**Notes**
- ...

## Edge Cases
- Target path does not exist and must be created
- Target directory already contains a non-symlink folder
- Missing permissions for creating symlinks

## Verification
- Manual flow: add target, add skill, sync, inspect symlinks

## Open Questions
None.

## Implementation Notes (Append-only)
### 2026-01-21
**Decisions**
- Targets require explicit names and paths; no auto-detection.
- `asm sync` uses symlinks and can scope to local or global stores.
- Sync replaces stale/incorrect symlinks rather than erroring.

**Changes**
- Drafted target management and sync behavior.

**Verification**
- N/A (planning only).

**Follow-ups**
- [ ] Decide whether `asm target remove` should clean up existing symlinks.

### 2026-01-22
**Decisions**
- `asm sync` uses effective targets by default; `--scope` filters sources only.
- Source names with slashes sync into nested directories under the target.
- `asm target remove` cleans up expected symlinks and skips real files with warnings.
- Warnings print to stderr and do not fail the command.

**Changes**
- Updated PRD decisions and resolved open question on target cleanup.

**Verification**
- N/A (planning update only).

### 2026-01-22
**Decisions**
- `asm target remove` cleans up expected symlinks and skips real files with warnings to stderr.
- `asm sync` uses effective targets by default; `--scope` filters sources only.
- Source names with slashes sync into nested directories under each target.

**Changes**
- Added `asm target add/remove` with scope-aware config persistence and `$HOME` normalization.
- Added `asm sync` command and sync engine for safe symlink updates.
- Added deterministic cleanup on target removal and tests for sync/target behavior.

**Verification**
- `go test ./...`
