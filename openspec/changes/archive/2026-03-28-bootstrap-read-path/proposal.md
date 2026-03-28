## Why

Skillforge has a clear product direction and repository model, but it still lacks the first implementation slice that turns those decisions into executable system boundaries.

The highest-leverage next step is to make the canonical single-repository data model usable through a read path before attempting mutation and PR submission flows.

This change creates that foundation by defining how Skillforge will:

- scan the canonical repository for skills
- validate skills against AgentSkills directory and `SKILL.md` rules
- derive a searchable index from Git state
- expose read-only list/search/get APIs
- ship a local self-hosted development stack with Forgejo and Skillforge services

Doing the read path first reduces risk. It proves the repository layout, validation model, and search contract before we add authoring, branch creation, PR submission, or review automation.

## What Changes

This change introduces the first development slice for Skillforge:

1. A repository-ingestion capability that scans `skills/<skill-name>/`, parses `SKILL.md`, validates core AgentSkills constraints, and produces normalized metadata records plus validation errors.
2. A catalog read API capability that exposes list/search/get behavior over the derived index.
3. A self-hosted compose stack capability that runs Forgejo and Skillforge together for local development and demos.
4. An implementation plan that keeps Git as the source of truth and treats all indexes/caches as rebuildable runtime state.

This slice is intentionally read-oriented. It does **not** include write APIs, branch creation, PR creation, or review tooling yet.

## Capabilities

### New Capabilities
- `repository-ingestion`: Scan the canonical skills repository, validate skill layout and metadata, and derive normalized catalog records from Git state.
- `catalog-read-api`: Provide read-only HTTP endpoints for listing, searching, and retrieving skills from the derived catalog.
- `compose-dev-stack`: Run Forgejo and Skillforge together in a local Docker Compose topology suitable for self-hosted development.

### Modified Capabilities
- None.

## Impact

Affected areas:

- repository layout contracts under `skills/<skill-name>/`
- Go backend architecture for scanning, validation, indexing, and HTTP APIs
- future Go CLI and TypeScript web UI consumers of the read API
- Docker Compose deployment and local development experience
- integration boundary with Forgejo as the Git/PR/auth system
