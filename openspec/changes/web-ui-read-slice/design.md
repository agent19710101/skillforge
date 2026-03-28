# Design: web-ui-read-slice

## Summary

Add a small TypeScript web app under `web/` using Vite + React. Keep the first UI slice read-only and focused on list/search/detail flows, with a lightweight API layer and representative tests.

## Goals

- ship the first browser UI for Skillforge
- keep the implementation small, maintainable, and self-hostable
- reuse the existing read API contract without backend redesign
- provide clear loading, empty, and error states

## Non-goals

- draft authoring or submission UI
- authentication or multiuser permissions work
- backend API redesign
- full design-system or routing framework work

## Architecture

### Web app location

Place the UI in `web/` as an isolated TypeScript frontend package.

### Frontend stack

Use:

- Vite
- React
- TypeScript
- Vitest + Testing Library for targeted UI tests

This keeps setup straightforward and suitable for a self-hosted OSS project.

### API layer

Add a small `web/src/api.ts` module that talks to:

- `GET /api/v1/skills`
- `GET /api/v1/search?q=...`
- `GET /api/v1/skills/{name}`
- `GET /api/v1/index/status`

Use same-origin `/api/...` paths by default and a Vite dev proxy to the Go API for local development.

### UI shape

The first batch uses a two-panel layout:

- left panel: search plus list/results
- right panel: selected skill detail

Behavior:

- initial load fetches skills and auto-selects the first item when present
- search reuses the existing search endpoint
- clicking a skill loads its full detail
- the page shows explicit loading, empty, and error states

## Verification strategy

- `npm test` in `web/`
- `npm run build` in `web/`
- repo-level OpenSpec validation for the change

## Risks and tradeoffs

- avoiding a router keeps the first slice small, but limits deep-linking in the initial version
- using a separate `web/` package adds frontend tooling to the repo, but isolates it cleanly from the Go backend
- same-origin API defaults plus a Vite proxy keep local development simple, but deployment docs must be clear about how the UI reaches the API
