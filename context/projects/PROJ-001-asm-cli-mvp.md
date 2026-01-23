---
id: PROJ-001
title: ASM CLI MVP
status: active # draft | active | paused | done | cancelled
owners:
  - julian
createdAt: 2026-01-21
updatedAt: 2026-01-22
---

# PROJ-001: ASM CLI MVP

## Summary
Deliver a GNU-style CLI for managing agent skills with a central store, supporting local (repo) and global scopes, plus explicit target directories for symlink linking.

## Outcome
Users can add local or remote skills once and reliably sync them into dotfiles or agent-specific directories without auto-detection.

## Scope Boundaries
**In scope**
- Storage layout for global/local stores and config
- Source management: add/update/remove with multi-skill detection
- Target management and sync via symlinks

**Out of scope**
- Agent auto-detection or marketplace browsing
- Skill validation/packaging beyond basic structure
- GUI or web interface
- Non-symlink copy-based syncing (unless explicitly added later)
- Skill scaffolding via `asm init` (handled elsewhere)

## Milestones (Project Phases)
### M1 — Storage + source management
**Exit criteria**
- Config schema and storage paths defined
- `asm add/update/remove` implemented with multi-skill rules
- `asm add` auto-syncs after install

### M2 — Targets + sync + init
**Exit criteria**
- `asm target` add/remove/list with required names
- `asm sync` creates/upserts symlinks for all targets

## PRDs (Derived)
PRDs in this project are `context/prds/PRD-*.md` where frontmatter `projectId: PROJ-001`.

## Risks / Dependencies
- Git availability and auth for remote sources
- Symlink behavior across filesystems
- Ambiguous multi-skill repo layouts

## Verification (Project-level)
- Manual CLI flows for add/update/remove, target sync, and init
