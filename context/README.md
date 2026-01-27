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
- Remote store identity is derived from `origin`; changing origin creates a new store path.
- Local path sources use `type:"path"` and install directly from the filesystem (no store directory).

## Module map
- `internal/cli` — Cobra commands and report formatting.
- `internal/asm` — use-case orchestration; returns report structs.
- `internal/manifest` — manifest + sum IO, validation, repo layout paths.
- `internal/source` — input parsing, GitHub tree URLs, discovery.
- `internal/gitstore` — go-git clone/ref/lockfile resolution, store paths.
- `internal/linker` — symlink creation and pruning.

## Entry points
- Start with PRDs in `context/prds/` for execution.
- Use `context/specs/` for cross-cutting invariants and verification guidance.
- Module responsibilities and invariants: `context/specs/module-boundaries.md`.
