# skillforge

Self-hosted multiuser skill sharing for organizations, backed by a **single Git repository**.

## Goal

Build an open source system for publishing, reviewing, searching, and curating agent skills across an organization.

Core constraints:

- multiuser
- web UI
- CLI client
- single-repo Git storage model
- PR-based contribution flow
- self-hostable via Docker / Docker Compose
- skill template enforcement against the AgentSkills specification
- searchable by name, description, tags, and related metadata

## Working direction

Current preferred architecture:

- **Forgejo** for authentication, branching, pull requests, reviews, and Git hosting
- **Skillforge API** (Go) for skill validation, indexing, repo orchestration, and search
- **Skillforge Web UI** (TypeScript) for browsing, editing, submission, and review workflows
- **Skillforge CLI** (Go) for listing, searching, creating, updating, deleting, and submitting skills
- **One canonical Git repository** as the source of truth for all skill data

Git remains the source of truth. Any search index/cache must be derived from repository state and rebuildable.

## Repository status

This repository now contains the early Go backend and local self-hosted stack for:

- canonical skill repository scanning and validation
- list/get/search read APIs
- draft create/status/submit lifecycle endpoints
- managed Git publication into a canonical working copy
- Forgejo pull request creation for live draft submission
- OpenSpec baseline specs for the current backend slices

## Live draft submission

The API can run in read-only mode by default, or publish review branches and open Forgejo pull requests when live submission is configured.

See [`docs/dev-stack.md`](docs/dev-stack.md) for the required submission environment variables, writable working-copy expectations, and a local verification flow.

## CLI preview

The first Go CLI slice now supports read and draft lifecycle commands against a running Skillforge API:

```bash
go run ./cmd/skillforge list
go run ./cmd/skillforge search git
go run ./cmd/skillforge get git-pr-review
go run ./cmd/skillforge list --server http://localhost:8080 --json

go run ./cmd/skillforge draft create --operation create --skill new-skill --file ./skills/new-skill/SKILL.md
go run ./cmd/skillforge draft status draft01
go run ./cmd/skillforge draft submit draft01
```

## Web UI preview

The first browser UI slice now lives under [`web/`](web/) and currently focuses on read-only catalog discovery.

Run it locally next to the API with:

```bash
cd web
npm install
npm run dev
```

By default the Vite dev server proxies `/api/*` requests to `http://localhost:8080`. Set `VITE_API_BASE_URL` when the API is mounted elsewhere.

## License

MIT
