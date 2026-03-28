# forgejo-live-submission

## Why

Skillforge round 1 shipped the read path, managed draft workspaces, validation, and the first shared draft lifecycle API. It also established the correct backend seam: draft edits live in temp workspaces while submission targets the canonical repository boundary.

What is still missing is the production path from `POST /api/v1/drafts/{id}/submit` to a real Forgejo pull request. Today that endpoint is either fake-backed in tests or disabled by default in production because the API server has no real Git publisher, no Forgejo client, and no operator-configured writable working copy.

Without this slice, Skillforge still cannot deliver its core multiuser promise: propose a skill change, publish a branch, and open a reviewable PR against the canonical repository.

## What Changes

This change adds the first production-grade submission backend for the existing draft lifecycle.

It will:

1. define runtime/operator configuration for a writable canonical working copy plus Forgejo submission settings
2. implement a real Git publisher that applies validated draft changes from temp workspaces into the managed canonical working copy, commits them, and pushes a review branch
3. implement a real Forgejo pull-request client
4. wire the default API server so submission capability becomes enabled only when configuration is complete
5. document how local Docker/Compose and standalone operators provide the required repository path, remote, base branch, and auth

## Impact

- Affects backend runtime configuration, submission internals, API server wiring, and operator docs
- Builds on the merged draft lifecycle API instead of introducing new client surfaces
- Keeps the temp draft model and single canonical repository model intact
- Unblocks later CLI and web UI slices because the shared submit contract will finally have a real backend behind it
