# go-cli-client

## Why

Round 2 made the Skillforge backend real enough to use: the API can scan the canonical repository, expose read/search endpoints, create managed drafts, and submit those drafts through Git + Forgejo into reviewable pull requests.

What is still missing is a practical user surface. Today the only way to exercise that backend is raw HTTP. That blocks dogfooding, hides API contract gaps, and leaves the project without the terminal-native workflow it explicitly wants.

## What Changes

This change introduces the first Go CLI for Skillforge.

It will:

1. add a Go CLI binary for the existing HTTP API
2. implement the first bounded command set for browse/read flows and draft lifecycle actions
3. add shared API client code with machine-readable error handling
4. define human-readable default output plus optional JSON output for automation
5. document local CLI usage against the existing Compose stack

## Impact

- Adds the first end-user client surface for Skillforge
- Exercises the existing read and draft lifecycle contracts instead of redesigning backend APIs
- Keeps the project aligned with its terminal-native product bias
- Creates a smaller, lower-risk path to product validation than starting the web UI first
