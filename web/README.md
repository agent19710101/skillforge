# Skillforge web UI

This package contains the first read-only browser UI for Skillforge.

## Scope of this slice

The current UI supports:

- skill list browsing
- paginated catalog loading so browse mode can reach the full catalog instead of only the API's first page
- full-text search
- skill detail viewing
- URL-backed `?q=` search and `?skill=` selection for shareable deep links
- basic loading, empty, and error states

It intentionally does **not** include draft authoring or submission yet.

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
