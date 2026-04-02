# Design: web-ui-draft-write-slice

## Summary

Extend the current React web UI with a draft authoring panel backed by the existing Go draft lifecycle API. Keep the first write batch pragmatic: one panel for composing a create/update/delete draft, validation feedback after draft creation, and a submit action for the currently active draft.

## Goals

- ship the first write-capable browser workflow for Skillforge
- reuse the existing draft API contract without broad backend redesign
- expose validation and submission status honestly in the UI
- keep the first write batch maintainable and easy to verify

## Non-goals

- authentication or user identity
- reviewer/admin dashboards
- concurrent multi-draft session management
- rich Markdown editing or preview tooling
- backend draft API redesign beyond narrow compatibility fixes

## Architecture

### Frontend API layer

Extend `web/src/api.ts` with typed functions and payloads for:

- `POST /api/v1/drafts`
- `GET /api/v1/drafts/{id}`
- `POST /api/v1/drafts/{id}/submit`

Preserve the existing same-origin API behavior.

### UI shape

Keep the current two-panel browse/detail layout and add a draft authoring section in the detail panel.

The authoring section will:

- allow choosing `create`, `update`, or `delete`
- allow editing the target skill name
- allow editing draft content for create/update operations
- disable content editing for delete operations
- create a draft against the backend and render the returned validation and submission status
- allow submitting the current created draft when submission is enabled
- show explicit create/submit loading and error states

### State model

The first write slice uses local component state in `App.tsx` instead of introducing a new global state layer.

Track:

- draft form inputs (`operation`, `skillName`, `content`)
- create-draft request state and error
- current created draft record
- submit request state, error, and success payload

When a user selects a skill in browse mode, the form may prefill the skill name and current body for update-oriented workflows, but explicit user edits remain source of truth.

## Verification strategy

- `npm test` in `web/`
- `npm run build` in `web/`
- `npm run lint` in `web/`
- `openspec validate 2026-03-29-web-ui-draft-write-slice --strict`

## Risks and tradeoffs

- keeping authoring state in `App.tsx` is simple for the first write batch, but may need extraction as the UI grows
- surfacing only the current active draft keeps scope small, but does not yet provide a draft history view
- the UI depends on the existing server-side draft contract, so any rough edges in error payloads will show up more clearly than they did in the CLI
