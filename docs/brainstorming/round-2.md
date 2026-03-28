# Brainstorming slice 2

## Focus

Refine the **single-repository data model** for Skillforge so implementation can start without ambiguity.

The previous slice chose the platform direction: **Forgejo-centered**.
This slice decides how skills are represented inside the one canonical Git repository.

## External constraint check

The AgentSkills specification requires each skill to be a directory containing at minimum a `SKILL.md` file with YAML frontmatter and Markdown instructions. The `name` field must match the parent directory name.

This creates a strong natural unit for storage and validation:

```text
skills/<skill-name>/SKILL.md
```

## Direction 1 — Pure directory model, no checked-in global catalog

Store every skill under a canonical path such as:

```text
skills/<skill-name>/
  SKILL.md
  scripts/
  references/
  assets/
```

Behavior:

- Git repo is the only source of truth
- listing/search index is derived by scanning the repo and parsing skill metadata
- Skillforge API maintains a rebuildable index/cache outside Git for fast search
- no duplicated checked-in skill manifest exists

### Pros

- strongest alignment with AgentSkills spec
- no metadata duplication
- simplest merge model
- minimal risk of catalog drift
- clean PR diffs: contributors edit the actual skill files only

### Cons

- cold-start indexing must scan repository content
- list/search depends on a derived index service or repeated parsing work

## Direction 2 — Checked-in repository catalog as first-class file

Store per-skill directories plus a repo-level file such as `catalog/skills.json` containing metadata for all skills.

Behavior:

- PRs update both the skill directory and the shared catalog file
- list/search can use the checked-in catalog directly

### Pros

- easy list/search bootstrap
- repository contains a human-inspectable global inventory
- simpler integrations for downstream consumers

### Cons

- duplicates metadata already present in `SKILL.md`
- high drift risk unless strictly enforced
- shared-file merge conflicts become common
- contributors must touch unrelated global state for local skill edits

## Direction 3 — Canonical skill directories plus generated catalog artifact

Store canonical per-skill directories and additionally generate a catalog artifact like `generated/catalog.json` on merge.

Behavior:

- contributors primarily edit skill directories
- CI or merge automation refreshes a generated catalog
- runtime search may still use an external derived index

### Pros

- cleaner contributor workflow than a manual catalog
- easier static export and integration story
- still keeps canonical content in skill directories

### Cons

- adds automation complexity early
- generated artifacts still create churn in Git history
- unclear whether committed generated files add enough value for the first milestone

## Chosen direction

**Direction 1 — Pure directory model, no checked-in global catalog**

## Rationale

This is the cleanest fit for the product constraints:

- one Git repo
- Git as the database
- strong compatibility with AgentSkills
- PR-based review on real source files

It avoids inventing a second canonical metadata layer too early.

For the first milestone, the right tradeoff is:

- canonical data in `skills/<skill-name>/...`
- derived search/index outside the repo
- optional exported catalog later if real integrations require it

## Concrete repository hypothesis

```text
skills/
  <skill-name>/
    SKILL.md
    scripts/
    references/
    assets/
```

Rules:

- directory name equals `SKILL.md` frontmatter `name`
- `SKILL.md` is mandatory
- optional directories mirror AgentSkills guidance
- no per-skill repositories
- no checked-in top-level canonical catalog for v1
- search/list APIs are backed by an index derived from the repo HEAD

## Implementation implication

This gives a sufficiently clear implementation direction to start development.

Development should begin with:

1. repo scanner + metadata extractor
2. AgentSkills validator wrapper
3. derived search index
4. read-only API for list/search/get
5. CLI and web UI against that read path first
6. write/PR flows after read path is stable

## Ready OpenSpec prompt

```text
Create the first implementation change for Skillforge based on these fixed decisions:

Platform decisions:
- Forgejo is the collaboration/auth/PR layer
- Skillforge is a separate app with Go backend/API, Go CLI, and TypeScript web UI

Repository model decisions:
- all skills live in one canonical Git repository
- store skills under `skills/<skill-name>/`
- each skill must satisfy the AgentSkills specification
- `SKILL.md` is the canonical metadata/instruction file
- no checked-in global skill catalog in v1
- list/search use a derived index rebuilt from repository state

For the first implementation slice, produce:
1. proposal
2. design
3. tasks
4. spec deltas

Scope the slice to:
- repository scanner
- metadata extraction from `SKILL.md`
- validation pipeline against AgentSkills rules
- derived search/list index
- read-only HTTP API for list/search/get
- local Docker Compose topology with Forgejo + Skillforge services

Explicitly describe the boundaries between canonical Git data and derived runtime state.
```
