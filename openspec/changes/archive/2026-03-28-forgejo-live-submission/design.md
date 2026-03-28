# Design: forgejo-live-submission

## Summary

Keep draft workspaces as temporary editable copies, but add a production submission backend that materializes validated draft changes into a managed canonical working copy, publishes a branch to Forgejo, and opens a pull request.

## Goals

- Make draft submission real in production
- Preserve the canonical single-repository model
- Reuse the existing draft validation and submission contract
- Keep client responses machine-readable for future CLI and web UI consumers

## Non-goals

- building the first CLI client
- building the first web UI
- redesigning draft workspaces into persistent Git clones or worktrees
- solving multi-node coordination or distributed locking in this slice

## Architecture

### Runtime configuration

Add explicit runtime configuration for:

- canonical writable repository root
- Forgejo remote name
- Forgejo owner/repo
- Forgejo base branch
- Forgejo server URL
- authentication token or equivalent auth method

The API server should expose submission as enabled only when this configuration is complete and the required dependencies can be constructed.

### Managed canonical working copy

The writable submission path should operate on a canonical working copy that the backend owns. That working copy is a real Git repository with the configured remote and base branch available.

Draft workspaces remain temp copies used for:

- semantic create/update/delete mutations
- repository and AgentSkills validation
- isolated inspection before submit

When a draft is submitted, the Git publisher applies the relevant changes from the draft workspace into the canonical working copy, checks out or creates the draft branch from the configured base branch, commits the materialized changes, and pushes the branch.

### Git publisher

The existing abstract Git boundary already receives:

- canonical repo root
- branch name
- base branch
- structured commit message
- draft source metadata (`root`, `operation`, `skillName`)

This slice will add a concrete Git publisher implementation that uses those inputs to:

1. ensure the canonical working copy is on the configured base branch
2. create/reset the target branch from the base branch
3. materialize the draft delta from the temp draft source into the canonical working copy
4. stage and commit only the intended skill changes
5. push the branch to the configured remote

The materialization logic should be explicitly scoped to `skills/<skill-name>/` so the backend does not accidentally copy unrelated workspace files into the canonical repository.

### Forgejo PR client

Add a small Forgejo HTTP client that can create pull requests on the configured server.

The client only needs the operations already required by the submission contract:

- create pull request from head branch to base branch
- return machine-readable PR identifier and URL

This client should remain narrow and isolated from the rest of the API layer.

### API server wiring

`cmd/skillforge-api` should:

- load submission configuration from environment variables
- build the real Git publisher and Forgejo client when configuration is valid
- report submission availability accurately when configuration is partial or absent

This keeps the existing API contract honest:

- disabled with a reason when submission is not configured
- enabled when the real backend is available

## Verification strategy

- unit tests for configuration loading/validation
- unit tests for Git publisher materialization logic using temp Git repositories
- unit tests for Forgejo client request/response handling against a local test server
- API/server tests that verify submission capability becomes enabled only when runtime configuration is complete

## Risks and tradeoffs

- A managed writable working copy is simpler than redesigning drafts as Git worktrees, but it means careful branch/reset hygiene matters
- The first release can target single-node/local operation and defer distributed locking
- Restricting materialization to the mutated skill directory reduces risk and keeps the first live submission slice bounded
