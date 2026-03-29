# web-ui-read-slice

## Why

Skillforge now has a coherent backend and the first Go CLI, but it still lacks the browser experience called out in the project brief. That makes the product harder to demo, harder to adopt for non-CLI users, and leaves a major milestone item unfinished.

## What Changes

This change introduces the first TypeScript web UI slice for Skillforge.

It will:

1. add a pragmatic self-hosted frontend app
2. implement read-only browse/search/detail flows against the existing Go API
3. add a small shared UI API layer for the current read endpoints
4. add targeted frontend tests for representative loading, empty, and error states
5. document how to run the UI locally next to the current API stack

## Impact

- creates the first browser-based product surface for Skillforge
- exercises the existing read API in a second client surface
- advances the project toward its multiuser product shape without taking on draft-authoring UI yet
