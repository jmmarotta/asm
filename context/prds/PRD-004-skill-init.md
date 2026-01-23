---
id: PRD-004
title: Skill Scaffold (asm init)
status: cancelled # draft | ready | in_progress | done | cancelled
projectId: PROJ-001
branchName: asm-cli-mvp
owners:
  - julian
createdAt: 2026-01-21
updatedAt: 2026-01-22
stories:
  - id: US-001
    title: Scaffold a new skill in the current directory
    acceptanceCriteria:
      - `asm init <skill-name>` creates `SKILL.md` in the current directory
      - The `name` field in `SKILL.md` matches the provided skill name
      - Command fails safely if `SKILL.md` already exists
    priority: 1
    passes: 0
    completedAt: null
    notes: ""
---

# PRD-004: Skill Scaffold (asm init)

## Summary
Provide a minimal `asm init` command that scaffolds a new skill in the current directory with a valid `SKILL.md` file.

## Goals
- Create `SKILL.md` in the current directory
- Fill required frontmatter fields with the provided skill name
- Avoid overwriting existing files

## Non-goals
- Creating full multi-file skill templates or examples
- Auto-adding the skill to the store

## Scope Boundaries
**In scope**
- `asm init <skill-name>` command behavior
- Minimal `SKILL.md` scaffold

**Out of scope**
- Resource folder scaffolding (`scripts/`, `references/`)

## User Stories
### US-001 — Scaffold a new skill in the current directory (P1)
**Acceptance Criteria**
- `asm init <skill-name>` creates `SKILL.md` in the current directory
- The `name` field in `SKILL.md` matches the provided skill name
- Command fails safely if `SKILL.md` already exists

**Notes**
- ...

## Edge Cases
- Current directory is not writable
- Invalid skill name characters

## Verification
- Manual flow: run `asm init foo`, inspect `SKILL.md`

## Open Questions
None.

## Implementation Notes (Append-only)
### 2026-01-21
**Decisions**
- `asm init` writes `SKILL.md` in the current directory.
- The command fails if `SKILL.md` already exists to avoid overwrites.

**Changes**
- Drafted minimal scaffold requirements.

**Verification**
- N/A (planning only).

**Follow-ups**
- [ ] Decide whether to create optional directories later.

### 2026-01-22
**Decisions**
- `asm init` will not be implemented in the CLI; skill scaffolding is handled elsewhere.

**Changes**
- Marked PRD-004 as cancelled.

**Verification**
- N/A (cancelled).
