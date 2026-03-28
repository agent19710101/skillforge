# Development stack

## Purpose

The current development stack is the first self-hosted local topology for Skillforge.

It is intentionally small:

- **Forgejo** provides the future multiuser Git, auth, branch, and PR/review backbone
- **skillforge-api** provides repository scanning, validation, indexing, and read-only APIs

## Compose services

`docker-compose.yml` defines:

- `forgejo` on `http://localhost:3000`
- `skillforge-api` on `http://localhost:8080`

## Canonical repository strategy

For this bootstrap slice, the canonical repository source is made explicit with a bind mount:

- host path: `./dev/sample-skills-repo`
- container path: `/data/skills-repo`
- API configuration: `SKILLFORGE_REPO_ROOT=/data/skills-repo`

This keeps the Git-backed source of truth visible and replaceable.

In later slices, the same API contract can point at:

- a local checked-out Forgejo repository clone
- a synchronized working copy managed by the backend
- another explicit repository path in operator-managed deployments

## Sample content

The sample repository currently includes:

- `skills/git-pr-review/SKILL.md`
- `skills/pdf-search-helper/SKILL.md`

That is enough to verify list, get, and search behavior immediately after starting the stack.

## Local usage

```bash
docker compose up --build
```

Then verify:

```bash
curl http://localhost:8080/healthz
curl http://localhost:8080/api/v1/skills
curl 'http://localhost:8080/api/v1/search?q=git'
curl http://localhost:8080/api/v1/skills/git-pr-review
curl http://localhost:8080/api/v1/index/status
```

## Notes

- The sample repository is mounted read-only into the API container.
- If you initialize `dev/sample-skills-repo` as a real Git repository, index status will also expose a concrete Git revision.
- This slice does not yet wire the web UI, CLI mutation flows, branch creation, or PR submission.
