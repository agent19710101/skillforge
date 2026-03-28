# Architecture decision log

## ADR-0001 — Forgejo-centered collaboration model

Status: accepted

- Forgejo handles users, auth, Git hosting, branches, PRs, and reviews.
- Skillforge focuses on domain logic: validation, indexing, API, CLI, and skill-centric web UX.

## ADR-0002 — Single-repo canonical directory model

Status: accepted

- Canonical skill data lives under `skills/<skill-name>/`.
- Each skill directory must satisfy the AgentSkills specification and contain `SKILL.md`.
- `SKILL.md` is the canonical metadata/instruction source.
- There is no checked-in global catalog in v1.
- Search/list functionality is powered by runtime-derived indexing from Git repository state.
