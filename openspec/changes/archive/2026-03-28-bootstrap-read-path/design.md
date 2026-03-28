## Context

Skillforge is building a multiuser skill sharing system backed by a single Git repository. Prior slices fixed two important architectural decisions:

1. Forgejo is the collaboration backbone for authentication, branches, pull requests, and review.
2. Canonical skill data lives in `skills/<skill-name>/`, with `SKILL.md` as the source of truth and no checked-in global catalog in v1.

The immediate implementation need is a dependable read path. Before the system can safely support create/update/delete/submit flows, it must be able to:

- discover skills from repository state
- validate them consistently
- normalize metadata for search and listing
- serve stable read APIs to the web UI and CLI

## Goals / Non-Goals

**Goals:**
- Define a Go service boundary for repository scanning, validation, indexing, and read APIs.
- Treat Git repository content as canonical and all runtime indexes as disposable derived state.
- Specify the first HTTP endpoints needed by the CLI and web UI.
- Define a Docker Compose topology that is self-hostable and practical for local demos.
- Keep the slice small enough to implement without prematurely designing full authoring workflows.

**Non-Goals:**
- Write APIs for creating, updating, or deleting skills.
- Automatic branch creation or PR submission.
- Review workflow automation beyond relying on Forgejo later.
- Fine-grained permissions and org policy logic beyond what Forgejo already provides.
- Full-text search sophistication beyond metadata-oriented search on name, description, tags, and related metadata.

## Decisions

### 1. Canonical vs derived state

Canonical state is the contents of the skills Git repository at a chosen revision, normally the default branch HEAD.

Derived state includes:
- parsed skill metadata
- validation results
- search index documents
- list/search API views

Derived state must be rebuildable from repository contents at any time.

### 2. Repository scan model

The backend scans the repository root for `skills/<skill-name>/SKILL.md`.

For each discovered skill directory, it will:
- verify directory placement under `skills/`
- verify directory name validity
- parse `SKILL.md` frontmatter and body
- verify `name` matches the parent directory name
- extract normalized fields including name, description, license, compatibility, metadata, allowed-tools, and derived tags when present in metadata
- record validation findings without preventing the overall scan from completing

This means the catalog can surface partial health instead of failing globally because one skill is malformed.

### 3. Validation layers

Validation is split into two layers:

- **core repository validation**: layout checks, required file presence, duplicate names, path rules, and frontmatter parsing
- **AgentSkills compatibility validation**: constraint checks aligned with the published AgentSkills specification

Implementation should allow swapping in or wrapping a dedicated validator later, but the first slice should not depend on remote services.

### 4. Indexing strategy

The backend maintains a local derived catalog index built from the last successful scan.

The first version can use an in-process index/store, because the primary goal is correctness and clarity of contract, not distributed scale.

The index should support:
- list all skills
- exact lookup by skill name
- metadata search across name, description, and selected metadata/tag fields
- visibility into validation state for each skill

### 5. Read API surface

The first HTTP API is read-only and versioned under `/api/v1`.

Initial endpoints:
- `GET /api/v1/skills` — paginated listing, optional sorting/filtering
- `GET /api/v1/skills/:name` — fetch a single skill and its validation summary
- `GET /api/v1/search?q=...` — metadata search across indexed fields
- `GET /api/v1/index/status` — last scan revision, timestamp, counts, and error summary

The CLI and web UI should consume the same API contract.

### 6. Compose topology

The first local self-hosted stack includes:
- `forgejo` service for Git hosting and future PR/review flows
- `skillforge-api` service for scanning, validation, indexing, and HTTP APIs

A separate web UI service may be added in the next slice, but this change should reserve its place in the topology and networking model.

The compose setup must make it obvious which repository is canonical and how the API reads it.

### 7. Mutation boundary

Even after write flows exist later, canonical skill changes should still land as Git commits reviewed through Forgejo.

This slice therefore avoids direct write semantics in the API. The goal is to preserve a clean separation:
- read model now
- branch/PR mutation model next

## Risks / Trade-offs

- A local in-process index is simple but may need replacement later for scale or advanced ranking.
- Implementing partial scan success adds complexity, but it prevents one broken skill from making the whole catalog unusable.
- Deferring write flows means the system remains incomplete for authoring in the short term, but it reduces early architectural churn and de-risks the core data model.
- Depending on Forgejo for Git/PR/auth keeps the product focused, but integration details will still need careful follow-up in later slices.
