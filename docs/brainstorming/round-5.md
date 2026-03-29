# Round 5 brainstorming

Date: 2026-03-29

## Context

Round 4 closed the first browser-facing Skillforge UI on `main`.

Current shipped baseline:

- canonical single-repo skill scanning and validation
- read APIs for list/get/search/index status
- draft create/status/submit lifecycle endpoints
- managed Git publication into a canonical working copy
- real Forgejo pull-request creation for draft submission
- local Docker Compose stack for Forgejo + API
- first Go CLI for browse and draft submission flows
- first read-only web UI for browse/search/detail with deep links and paginated catalog loading
- GitHub Actions verification for the web build/test/lint path

The next round should expand the product toward the full multiuser workflow without restarting architecture work that already proved out in earlier rounds.

## Direction 1 — First write-capable web UI for draft authoring and submission

Build the next browser UI slice for:

- creating draft changes for new or existing skills
- updating or deleting existing skills through the draft API
- validating and previewing authored skill content before submit
- submitting a draft and showing submission/PR status

Why it matters:

- directly advances the project brief requirement that the web UI support add/update/delete flows, not only list/search
- turns the current read-only demo into a real end-user workflow for non-CLI users
- reuses the backend draft/submit path that is already live instead of inventing a new architecture
- is the clearest next proof that Skillforge is a real multiuser product rather than a backend plus demos

Tradeoff:

- needs careful frontend state design for forms, validation feedback, and asynchronous draft status
- should stay scoped to the existing draft API rather than mixing in reviewer/admin dashboards

## Direction 2 — Repository freshness and derived-index sync automation

Add stronger repository freshness behavior such as:

- explicit refresh triggers after merge/submission events
- webhook or polling hooks for external canonical-repo changes
- stronger stale-state reporting in the API and clients
- better visibility into which Git revision the UI/CLI are reading

Why it matters:

- improves trust in browse/search results once multiple clients depend on them
- reduces operator confusion when repository state changes outside the API process
- strengthens production-readiness before broader adoption

Tradeoff:

- mostly infrastructure hardening after several backend-heavy rounds
- lower product visibility than giving the web UI a real write path

## Direction 3 — CLI/operator adoption polish

Extend the Go CLI and operator surface with:

- persisted CLI config for server defaults
- better output ergonomics for repeated usage
- packaging/release-install polish
- more explicit operator visibility into submission capability and runtime state

Why it matters:

- makes the existing CLI easier to adopt daily
- improves the terminal-native experience without major architecture change
- would harden the current deployment surface before more UI work

Tradeoff:

- mostly improves an already-usable surface
- does less to close the largest remaining gap in the project brief than write-capable web flows

## Chosen direction

**Direction 1 — First write-capable web UI for draft authoring and submission**

## Rationale

Skillforge now has the backend, CLI, and first read-only web UI in place, so the biggest remaining product gap is obvious: the browser surface still cannot create, update, or delete skills. The project brief explicitly requires both CLI and web UI support for add/update/delete/list/search flows, and the backend draft submission path already exists and was proven in earlier rounds. That makes a write-capable web UI the highest-leverage next move.

Compared with more infrastructure work, this direction creates a more complete end-user workflow immediately. Compared with more CLI polish, it unlocks a whole class of non-terminal usage instead of refining a surface that already works. The main risk is frontend scope creep, so the next change should stay tightly aligned with the existing draft API and avoid pulling in reviewer dashboards, authentication redesign, or general admin tooling.

## Ready OpenSpec prompt

```text
Create an OpenSpec change for the next Skillforge round focused on the first write-capable web UI slice.

Goal:
Extend the existing TypeScript web UI so users can create, update, delete, validate, and submit skill drafts against the current Go draft APIs.

Scope:
1. Build a write-capable web flow on top of the existing read-only UI and draft lifecycle APIs.
2. Add UI affordances for:
   - creating a new skill draft
   - updating an existing skill draft
   - deleting an existing skill through the draft flow
3. Add draft editing inputs for the canonical skill content with pragmatic validation/error feedback.
4. Add draft status and submission views so the UI can show whether submission is available, what failed, and what PR was created when submission succeeds.
5. Reuse the existing backend APIs where possible; only add narrow compatibility improvements if the current contract is insufficient for a maintainable UI.
6. Add targeted frontend tests for representative authoring, validation, failure, and submission states.
7. Update docs for local development and demoing of the write-capable UI against the Compose/dev stack.

Constraints:
- Do not redesign the single-repo Git model.
- Do not start authentication or user/role management in this change.
- Do not build reviewer/admin dashboards yet.
- Keep the frontend stack and state management pragmatic and maintainable.
- Prefer incremental API compatibility fixes over broad backend redesign.

Deliverables:
- OpenSpec proposal/design/tasks/spec deltas for the write-capable web UI slice
- TypeScript UI implementation for draft authoring/status/submission
- tests covering representative write-path behavior
- docs showing how to run and verify the write-capable UI locally
```
