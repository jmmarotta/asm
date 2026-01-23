# Context Index

## Purpose
This directory stores specs, initiatives, projects, and PRDs for repo-level planning and execution. Keep documents comprehensive so work can proceed without follow-up questions.

## Structure
```
context/
  README.md
  specs/
  initiatives/
  projects/
  prds/
```

## How to use
- **Specs**: stable invariants (data model, verification, standards).
- **Initiatives**: outcome-focused containers; multiple active initiatives are allowed.
- **Projects**: execution containers with milestones and PRD links.
- **PRDs**: shippable slices with user stories and append-only Implementation Notes.

## Workflow highlights
- **End of planning**: update the active PRD and append Implementation Notes.
- **End of building**: update story `completedAt`/`passes` and append Implementation Notes.
- **No separate updates folder**: progress is colocated inside PRDs.

## Notes
- Remote store identity is derived from `origin + ref`; changing either creates a new store path unless explicitly migrated.
- Local path sources use `ref: worktree` and sync directly from the filesystem (no store directory).

## Entry points
- Start with PRDs in `context/prds/` for execution.
- Use `context/specs/` for cross-cutting invariants and verification guidance.
