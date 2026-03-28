## Context

The `bootstrap-read-path` change established the core read architecture for Skillforge. The backend can now discover canonical skills, validate them, derive an index, and expose list/search/get APIs. The next missing capability is authoring.

Skillforge must support multiuser create/update/delete flows without abandoning the product’s core constraints:

- one canonical Git repository
- Git remains the source of truth
- contributions go through branch review and merge
- Forgejo is the collaboration system for users, branches, and PRs

The write path should therefore stage edits into managed branch workspaces and treat PR creation as the submission step.

## Goals / Non-Goals

**Goals:**
- Define how the backend creates and manages per-change branch workspaces.
- Define write operations for create, update, and delete skill changes against those workspaces.
- Define how staged changes are validated before publication.
- Define the Forgejo integration boundary for pushing a branch and opening a pull request.
- Define the shared contract for future CLI and web UI write flows.

**Non-Goals:**
- Fine-grained collaborative editing of the same draft by multiple users at once.
- Rich text/web editor implementation details.
- Full permission modeling beyond deferring to Forgejo/org auth and later policy layers.
- Auto-merge or advanced review automation.
- Background sync for long-lived local clones beyond the minimal workspace strategy required for the first slice.

## Decisions

### 1. Mutation occurs in managed workspaces, not directly on default branch

The backend will materialize a workspace derived from the canonical repository for a requested draft change. That workspace is associated with a branch name and becomes the target for create/update/delete operations.

This keeps the write path aligned with the existing read path:
- default branch remains canonical published state
- workspace branch contains proposed edits
- PR merge promotes the change into canonical state

### 2. One draft branch per submission flow

The first version should use one branch per draft submission flow, with deterministic but unique naming such as:

```text
skillforge/<operation>/<skill-name>/<short-id>
```

This makes branch ownership and intent easy to inspect in Forgejo.

### 3. Write operations are semantic, filesystem-backed mutations

The backend will expose semantic operations:
- create skill
- update skill
- delete skill
- validate draft
- submit draft for review

Internally, these mutate files under `skills/<skill-name>/` in the managed workspace.

This avoids exposing raw Git primitives directly to clients while still preserving Git as the backing store.

### 4. Validation runs before branch publication and PR creation

The same repository and AgentSkills validation model used by the read path must run against the draft workspace before submission.

Submission must fail if:
- the resulting draft is structurally invalid
- the edited skill name/path contract is broken
- required files are missing
- the draft cannot be committed cleanly

### 5. Forgejo integration boundary

The backend is responsible for:
- ensuring the branch exists locally
- committing staged changes with a structured commit message
- pushing the branch to Forgejo
- creating a PR against the configured base branch

The backend is **not** responsible in this slice for implementing review policy itself. Review, approvals, and merge semantics remain Forgejo-native.

### 6. Shared write contract for CLI and web UI

The CLI and future web UI should both target the same workflow-oriented API surface. The first contract should include:

- create draft
- inspect draft status and validation findings
- apply create/update/delete mutation
- publish draft
- submit PR

This keeps authoring behavior consistent across interfaces.

## Risks / Trade-offs

- Managing local branch workspaces adds operational complexity, but it provides a clean abstraction over Git for clients.
- Deterministic branch naming improves operability, but uniqueness rules must prevent collisions.
- Backend-driven commits simplify client behavior, but commit metadata and authorship decisions will need follow-up design.
- Deferring the actual CLI/web editor implementation keeps the slice smaller, but requires clear API contracts now to avoid drift later.
