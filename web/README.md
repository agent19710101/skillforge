# Skillforge web UI

This package contains the current browser UI for Skillforge.

## Scope of this slice

The current UI supports:

- skill list browsing
- paginated catalog loading so browse mode can reach the full catalog instead of only the API's first page
- full-text search
- skill detail viewing
- URL-backed `?q=` search and `?skill=` selection for shareable deep links
- browser draft creation for `create`, `update`, and `delete` operations
- draft validation and submission-capability/status display
- browser submission of the current draft through the existing draft API
- basic loading, empty, and error states

It still does **not** include multi-draft history, reviewer/admin dashboards, or auth/user management.

## Local development

Start the Go API first, then run:

```bash
npm install
npm run dev
```

The Vite dev server proxies `/api/*` requests to `http://localhost:8080` by default.

If your API is mounted elsewhere, set `VITE_API_BASE_URL` before starting the dev server.

## Verification

```bash
npm run build
npm test
npm run lint
```
