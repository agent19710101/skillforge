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

This repository currently contains:

- project brief
- first brainstorming slice
- OpenSpec bootstrap

## License

MIT
