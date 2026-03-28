# Round 4 brainstorming

Date: 2026-03-28

## Context

Round 3 closed the first end-user CLI on `main`.

Current shipped baseline:

- canonical single-repo skill scanning and validation
- read APIs for list/get/search/index status
- draft create/status/submit lifecycle endpoints
- managed Git publication into a canonical working copy
- real Forgejo pull-request creation for draft submission
- local Docker Compose stack for Forgejo + API
- first Go CLI for browse and draft submission flows

The highest-value next slice should expand the product from backend+CLI utility into a more complete multiuser experience without undoing the current small, maintainable scope.

## Direction 1 — First TypeScript web UI for browse/search and skill detail

Build the first browser UI slice for:

- skill listing
- full-text search
- skill detail view
- basic runtime/status visibility where helpful

Why it matters:

- directly advances the project toward its stated multiuser product shape
- satisfies a major remaining milestone item from the project brief (`basic web UI can list/search skills`)
- makes the current backend demoable to non-CLI users
- exercises the read API in a second client surface without immediately taking on draft-editing complexity

Tradeoff:

- requires frontend scaffolding and a TypeScript app structure that does not exist yet
- should stay intentionally read-focused in the first batch to avoid mixing UI bootstrap with draft authoring complexity

## Direction 2 — CLI/operator polish for real-world usage

Extend the Go CLI and deployment surface with:

- persisted CLI config for server URL/auth defaults
- install/build docs and release packaging shape
- richer CLI filtering/output ergonomics
- a small operator command or endpoint for index/runtime visibility

Why it matters:

- makes the newly shipped CLI easier to adopt repeatedly
- improves operator quality-of-life without major architecture change
- would harden the current toolchain before more product surfaces arrive

Tradeoff:

- mostly polish on top of a surface that already exists
- lower product leverage than starting the web UI now

## Direction 3 — Repository freshness/index automation

Add repo synchronization and derived-index freshness primitives such as:

- explicit refresh endpoint/command
- webhook- or polling-based refresh after merges
- stronger health/status reporting for stale repository state
- clearer behavior when the canonical repo changes outside the API process

Why it matters:

- improves production-readiness of the backend
- reduces drift risk between Git state and browse/search results
- supports future UI/CLI clients with a more trustworthy read model

Tradeoff:

- valuable infrastructure work, but still backend-centric after two backend-heavy rounds
- less immediately visible than shipping the first web UI slice

## Chosen direction

**Direction 1 — First TypeScript web UI for browse/search and skill detail**

## Rationale

Skillforge now has a coherent backend and a practical CLI, but it still lacks the most visible part of its intended product shape: a browser interface for shared discovery and browsing. The project brief explicitly calls out a web UI and lists `basic web UI can list/search skills` as a first-milestone success criterion, so this is the clearest remaining gap.

Compared with more CLI/operator polish, the web UI creates a new class of value rather than improving an existing surface. Compared with more backend freshness automation, it better matches the user’s preference to start small but prove the core product quickly. Keeping the first UI slice read-focused avoids overloading the round with authoring flows before the frontend foundation exists.

## Ready OpenSpec prompt

```text
Create an OpenSpec change for the next Skillforge round focused on the first TypeScript web UI slice.

Goal:
Ship the first browser-based Skillforge interface for browsing and searching skills against the existing Go API.

Scope:
1. Introduce a TypeScript web UI app suitable for the current self-hosted stack.
2. Implement a first read-focused UI slice with:
   - a list view for skills
   - search input/results wired to the existing search API
   - a skill detail view
   - basic empty/loading/error states
3. Add a small shared API layer in the UI for the existing read endpoints.
4. Keep the first UI slice read-only; do not implement draft authoring or submission yet.
5. Add local development/build instructions and wire the UI into the existing dev-stack documentation.
6. Add targeted tests for the UI data-loading and representative render states where practical.

Constraints:
- Do not redesign the backend API unless a narrow compatibility fix is clearly necessary.
- Do not start authentication or multiuser permissions work in this change.
- Do not mix in draft editing/submission UI yet.
- Keep the frontend stack pragmatic and maintainable for a self-hosted OSS project.

Deliverables:
- OpenSpec proposal/design/tasks/spec deltas for the first web UI slice
- TypeScript UI implementation for browse/search/detail
- docs showing how to run the UI against the local stack
```
