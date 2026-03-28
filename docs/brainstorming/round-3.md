# Round 3 brainstorming

Date: 2026-03-28

## Context

Round 2 closed the live Forgejo-backed submission backend on `main`.

Current shipped baseline:

- canonical single-repo skill scanning and validation
- read APIs for list/get/search/index status
- draft create/status/submit lifecycle endpoints
- managed Git publication into a canonical working copy
- real Forgejo pull-request creation for draft submission
- local Docker Compose stack for Forgejo + API

The highest-value next slice should turn that backend into a usable product surface without widening infrastructure scope again.

## Direction 1 — First Go CLI for browse + draft flows

Build the first `skillforge` CLI against the existing HTTP API for:

- `list`
- `search`
- `get`
- `draft create`
- `draft status`
- `draft submit`

Why it matters:

- smallest usable client surface on top of the now-real backend
- aligns with the project's terminal-native bias
- exercises the machine-readable API contract end to end
- gives operators and early users an immediate way to dogfood the system before a web UI exists

## Direction 2 — First TypeScript web UI for browse/search and draft authoring

Build the first browser UI slice for:

- skill listing/search
- skill detail view
- basic draft create/update flow
- submit action wired to the existing API

Why it matters:

- moves the project toward the multiuser product shape faster
- makes demos and internal adoption easier
- validates the API against a richer client workflow

Tradeoff:

- larger scope than the CLI because it needs frontend scaffolding, API integration, and UI state design in one round

## Direction 3 — Repository sync and indexing refresh automation

Add operator-facing repo freshness and automation primitives such as:

- explicit index refresh endpoint/command
- optional webhook or polling-based refresh after merges
- clearer index/runtime health reporting
- failure visibility around stale repository state

Why it matters:

- improves production-readiness of the backend
- reduces drift between Git state and read/search results
- supports future clients with a more trustworthy backend

Tradeoff:

- useful, but mostly infrastructure polish before the project has a first real client

## Chosen direction

**Direction 1 — First Go CLI for browse + draft flows**

## Rationale

The backend is now coherent enough to be used, but there is still no practical end-user surface besides raw HTTP calls. The CLI is the smallest next step that converts the current backend into a real product workflow while staying aligned with the user's preference to start small and expand only after the core proves useful.

Compared with the web UI, the CLI has lower implementation and review overhead, is easier to verify in CI, and should expose contract gaps in the draft lifecycle API earlier. Compared with more backend automation, it creates direct user value immediately instead of polishing internals ahead of adoption.

## Ready OpenSpec prompt

```text
Create an OpenSpec change for the next Skillforge round focused on the first Go CLI client.

Goal:
Ship a practical terminal-native CLI that exercises the existing Skillforge read and draft submission APIs end to end.

Scope:
1. Introduce a Go CLI binary (for example `skillforge`) with subcommands for:
   - `list`
   - `search <query>`
   - `get <skill-name>`
   - `draft create --operation <create|update|delete> --skill <name> [--title/--description/--tags as needed]`
   - `draft status <id>`
   - `draft submit <id>`
2. Add shared API client code for the current HTTP contract, including machine-readable error handling.
3. Design human-readable default output plus a structured JSON mode where it clearly helps automation.
4. Reuse the existing submission capability/status reporting so the CLI can explain disabled submission honestly.
5. Add CLI-focused tests for request shaping, output formatting, and representative failure handling.
6. Update README/docs with local usage examples against the existing Docker Compose stack.

Constraints:
- Do not start the TypeScript web UI in this change.
- Do not redesign the backend API unless a narrow compatibility fix is necessary.
- Keep the single-repo Git + Forgejo model unchanged.
- Prefer a low-dependency Go implementation.

Deliverables:
- OpenSpec proposal/design/tasks/spec deltas for the CLI contract
- Go CLI implementation with tests
- docs showing local CLI usage against the running stack
```
