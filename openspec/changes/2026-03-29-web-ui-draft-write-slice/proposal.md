# web-ui-draft-write-slice

## Why

Skillforge now has a proven backend draft lifecycle, a Go CLI for draft submission, and a first read-only web UI. The browser surface is still missing the most important remaining workflow from the project brief: create, update, and delete flows for skills.

Without a write-capable browser path, the web UI is still only a demo surface for discovery. The next round should reuse the existing draft APIs to turn it into a real authoring surface for non-CLI users.

## What Changes

This change introduces the first write-capable web UI slice for Skillforge.

It will:

1. add a shared frontend API layer for draft create/status/submit flows
2. add a draft authoring panel to the existing web UI
3. support create, update, and delete draft operations from the browser
4. surface validation findings and submission capability/status in the UI
5. support submitting a created draft from the browser
6. add targeted UI tests for representative authoring, validation, and submission states
7. document the write-capable UI flow and its local-development expectations

## Impact

- turns the web UI into the first end-to-end browser authoring surface for Skillforge
- exercises the existing draft lifecycle API in a second client surface beyond the CLI
- advances the project brief toward add/update/delete/list/search support in both CLI and web UI
- keeps scope aligned with the current draft API instead of reopening architecture work
