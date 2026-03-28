# Skillforge project brief

## Problem

Organizations need a shared, reviewable system for managing agent skills without spreading them across many repositories or relying on a centralized database as the primary source of truth.

## Product statement

Skillforge is a self-hosted, open source skill sharing platform that stores all skills in a **single Git repository** and exposes both a **web interface** and a **CLI** for searching, browsing, editing, reviewing, and publishing skills.

## Non-negotiable requirements

1. Multiuser system
2. Web interface
3. Git repository as the database / source of truth
4. May rely on third-party services such as Forgejo, Gogs, or Soft Serve
5. Open source
6. Self-hosted via Docker container and/or Docker Compose
7. CLI client for add/update/delete/list/search flows
8. Web UI for add/update/delete/list/search flows
9. PR-like review workflow for updates
10. Must use a **single** Git repository and must not create per-skill repositories
11. Must enforce the AgentSkills template/specification: https://agentskills.io/specification
12. Search by name, description, tags, and related metadata
13. List all skills in the system

## Early design constraints

- Git is the authoritative datastore.
- Branches + PRs are the review and merge mechanism.
- Search indexing may exist, but only as a derived cache from Git state.
- Skill content must live under a deterministic repository layout.
- The system must be practical to deploy for internal organizational use.
- Current repository model decision: use canonical skill directories under `skills/<skill-name>/` with `SKILL.md` as the source of truth and no checked-in global catalog in v1.

## Candidate stack

- Backend/API: Go
- Web UI: TypeScript
- Git/PR platform: Forgejo first, with future evaluation of Gogs / Soft Serve compatibility where useful
- Packaging: Docker + Docker Compose

## Success criteria for the first milestone

- Single-repo skill layout defined
- Skill validation against AgentSkills spec defined and testable
- Searchable metadata model defined
- Basic web UI can list/search skills
- CLI can list/search/create/update skills
- PR-based update flow works with Forgejo
- Local self-hosted demo runs via Docker Compose
