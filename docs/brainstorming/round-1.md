# Brainstorming slice 1

## Project

Skillforge — single-repo, self-hosted, multiuser skill sharing for organizations.

## Direction 1 — Forgejo-centered platform

Use Forgejo as the collaboration backbone:

- single Git repo hosted in Forgejo
- built-in users, auth, branches, PRs, reviews, permissions
- Skillforge API manages validation, indexing, repo mutations, and search
- Skillforge Web UI focuses on the domain UX for skills
- Skillforge CLI talks to the API and/or performs local repo workflows

### Pros

- strongest PR/review model
- mature web UI and multiuser support already solved
- least custom work for auth and code review flows
- clean fit for single-repo branch-based contribution model

### Cons

- more moving parts than a single-purpose app
- requires integration around Forgejo APIs and repository events

## Direction 2 — Soft Serve + custom application layer

Use Soft Serve for Git serving, then build most product behavior in Skillforge.

### Pros

- lightweight and Git-native
- aligns well with terminal-first workflows

### Cons

- weaker native PR/review story
- likely requires substantial custom review/workflow implementation
- higher product risk for multiuser web collaboration

## Direction 3 — Gogs-backed simpler MVP

Use Gogs as the Git/PR/auth base with a thinner Skillforge layer.

### Pros

- simpler than a full custom stack
- lighter than Forgejo

### Cons

- generally less attractive than Forgejo on ecosystem and capability depth
- lower confidence for long-term extensibility

## Chosen direction

**Direction 1: Forgejo-centered platform**

## Rationale

The hardest parts of this product are not raw Git storage. They are:

- safe multiuser collaboration
- branch/PR/review workflow
- permissions and auth
- usable web review flows

Forgejo solves those well today. That lets Skillforge stay focused on the domain-specific parts:

- single-repo skill layout
- AgentSkills validation
- indexing/search
- CLI ergonomics
- tailored web UX

## Initial architectural hypothesis

- One Forgejo repository stores all skills.
- Skills live under a stable path such as `skills/<skill-id>/`.
- Each skill has metadata + files that conform to the AgentSkills specification.
- Skillforge API validates changes, indexes metadata, and exposes search/list APIs.
- Skillforge Web UI provides skill-centric browse/edit/search UX.
- Skillforge CLI provides scripting and operator workflows.
- Search index is disposable and rebuildable from repository state.

## Ready OpenSpec prompt

```text
Design the first implementation slice for Skillforge, a self-hosted multiuser skill sharing system for organizations.

Constraints:
- all skill data lives in one Git repository
- Git is the source of truth
- no per-skill repositories
- Forgejo provides auth, Git hosting, branches, PRs, and reviews
- system includes a Go backend/API, a TypeScript web UI, and a Go CLI
- skills must be validated against https://agentskills.io/specification
- search must support name, description, tags, and listing all skills
- deploy locally with Docker Compose

Produce:
1. proposal
2. design
3. task breakdown
4. spec deltas

Focus first on:
- repository layout for skills
- metadata extraction and validation pipeline
- search/indexing strategy derived from Git state
- API surface for list/search/get/create/update/delete/submit-for-review
- Forgejo integration boundaries
- deployment topology for local self-hosting
```
