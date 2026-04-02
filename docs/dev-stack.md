# Development stack

## Purpose

The current development stack is the first self-hosted local topology for Skillforge.

It is intentionally small:

- **Forgejo** provides Git hosting, authentication, branches, and pull-request review
- **skillforge-api** provides repository scanning, validation, indexing, draft lifecycle APIs, and optional live submission

## Compose services

`docker-compose.yml` defines:

- `forgejo` on `http://localhost:3000`
- `skillforge-api` on `http://localhost:8080`

## Canonical repository strategy

The stack separates two repository views:

1. a **sample/read-only repository mount** for basic browse/search verification
2. a **managed canonical working copy** for live draft submission

The default Compose file keeps the sample repository bind-mounted at `/data/skills-repo` so the API can start with zero submission setup.

For live submission, operators should instead point `SKILLFORGE_REPO_ROOT` at a writable checked-out clone of the canonical Forgejo repository. The API uses that managed working copy to:

- refresh the configured base branch from the remote
- create or reset the review branch
- materialize the draft change under `skills/<skill-name>/`
- commit and push the review branch

Temp draft workspaces remain separate from the managed canonical working copy and do not need `.git` metadata.

## Required live submission environment

Live submission is enabled only when the full submission config is present:

- `SKILLFORGE_FORGEJO_SERVER_URL` — base Forgejo URL, for example `http://forgejo:3000`
- `SKILLFORGE_FORGEJO_REMOTE_NAME` — Git remote name for the managed working copy, usually `origin`
- `SKILLFORGE_FORGEJO_OWNER` — Forgejo owner or organization
- `SKILLFORGE_FORGEJO_REPO` — canonical repository name
- `SKILLFORGE_FORGEJO_BASE_BRANCH` — review target branch, usually `main`
- `SKILLFORGE_FORGEJO_TOKEN` — API/push token used for Forgejo requests
- `SKILLFORGE_FORGEJO_AUTH_METHOD` — optional auth method override; defaults to token auth

If those values are incomplete, draft responses still report machine-readable submission status, but `POST /api/v1/drafts/{id}/submit` returns `submission_unavailable` instead of pretending submission is live.

## Sample content

The sample repository currently includes:

- `skills/git-pr-review/SKILL.md`
- `skills/pdf-search-helper/SKILL.md`

That is enough to verify list, get, and search behavior immediately after starting the stack.

## Local usage

Start the default stack:

```bash
docker compose up --build
```

Then verify the read path:

```bash
curl http://localhost:8080/healthz
curl http://localhost:8080/api/v1/skills
curl 'http://localhost:8080/api/v1/search?q=git'
curl http://localhost:8080/api/v1/skills/git-pr-review
curl http://localhost:8080/api/v1/index/status
```

## Local live-submission verification

A minimal local verification flow for draft-to-PR submission is:

1. start Forgejo and the API stack
2. create the canonical skills repository in Forgejo
3. create a writable local clone of that repository for the API service
4. set the live submission environment listed above
5. create a draft with `POST /api/v1/drafts`
6. inspect the draft with `GET /api/v1/drafts/{id}` and confirm `submission.enabled` is `true`
7. submit the draft with `POST /api/v1/drafts/{id}/submit`
8. confirm the response includes the published branch, base branch, commit hash, pull-request number, and pull-request URL

## Web UI development

The browser UI lives in `web/` and now supports both read flows and the first browser draft-authoring path.

Run it locally with:

```bash
cd web
npm install
npm run dev
```

The Vite dev server proxies `/api/*` to `http://localhost:8080` by default. Set `VITE_API_BASE_URL` when the UI should talk to a different API origin or path prefix.

For write-flow verification, point the UI at an API instance that has access to a writable managed working copy. Draft creation works against the current draft API; draft submission only succeeds when live submission is configured, otherwise the UI will surface the reported disabled-submission reason.

## Notes

- The sample Compose mount is intentionally read-only and is suitable only for browse/search verification.
- Live submission requires a writable managed canonical working copy and an authenticated Forgejo token.
- If the repository root is a real Git working copy, index status also exposes the current Git revision.
