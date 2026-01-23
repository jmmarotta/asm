---
id: PRD-005
title: Module Path + Release Install Workflow
status: done # draft | ready | in_progress | done | cancelled
projectId: PROJ-001
branchName: asm-cli-mvp
owners:
  - julian
createdAt: 2026-01-22
updatedAt: 2026-01-22
stories:
  - id: US-001
    title: Update module path for go install
    acceptanceCriteria:
      - go.mod module path is github.com/jmmarotta/agent_skills_manager
      - All Go imports updated from asm/... to github.com/jmmarotta/agent_skills_manager/...
      - go build and go test succeed after the change
    priority: 1
    passes: 1
    completedAt: 2026-01-22
    notes: ""
  - id: US-002
    title: Update docs for go install and development
    acceptanceCriteria:
      - README uses go install with semver tags (latest + pinned example)
      - README examples use asm on PATH (not ./asm)
      - docs/development.md documents Makefile targets and test helpers
      - README points to docs/development.md for contributor setup
    priority: 2
    passes: 1
    completedAt: 2026-01-22
    notes: ""
  - id: US-003
    title: Document semver tagging workflow
    acceptanceCriteria:
      - Clear steps for creating annotated tags (v0.x.y)
      - Recommend initial tag v0.1.0
      - Tagging steps reference make check before tagging
    priority: 3
    passes: 1
    completedAt: 2026-01-22
    notes: ""
---

# PRD-005: Module Path + Release Install Workflow

## Summary
Align the Go module path, docs, and release guidance so users can install `asm` with `go install ...@latest` and the project follows SemVer tagging conventions.

## Goals
- Enable `go install` from the GitHub repo with SemVer tags
- Update user-facing README to reflect installation and usage from PATH
- Provide a developer doc for build/test flows (Makefile-driven)
- Establish consistent release tagging guidance (v0.x.y)

## Non-goals
- Automate releases or publish binaries
- Add package distribution beyond `go install`
- Change runtime behavior of the CLI

## Scope Boundaries
**In scope**
- Update Go module path and imports
- Update README to use `go install` (latest + pinned version example)
- Add docs/development.md with Makefile usage
- Document SemVer tagging workflow and initial tag recommendation

**Out of scope**
- CI/CD pipelines for releases
- GPG signing or checksum distribution
- Release artifacts beyond tags

## Plan
1. Update module path and all imports to github.com/jmmarotta/agent_skills_manager.
2. Update README to recommend `go install` and reference docs/development.md.
3. Add docs/development.md with Makefile targets and testing guidance.
4. Add release tagging guidance (annotated tags, v0.1.0 as first tag).
5. Verify with `make check` and update PRD notes when complete.

## Edge Cases
- Unversioned install attempts before tags exist
- Stale imports or build cache after module path change
- GOPATH/GOBIN not on PATH for go install

## Verification
- `make check`

## Open Questions
None.

## Implementation Notes (Append-only)
### 2026-01-22
**Decisions**
- Module path will be github.com/jmmarotta/agent_skills_manager.
- README will recommend go install @latest with a pinned example.
- Initial release tag should be v0.1.0 (SemVer, annotated).

**Changes**
- Drafted plan for module path updates, docs, and tagging workflow.

**Verification**
- N/A (planning only).

### 2026-01-22
**Changes**
- Updated module path and imports to github.com/jmmarotta/agent_skills_manager.
- Added docs/development.md and updated README to recommend go install.
- Added Makefile-driven developer guidance and SemVer tagging notes.

**Verification**
- `make check`
