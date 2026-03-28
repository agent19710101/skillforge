# Design: go-cli-client

## Summary

Add a thin Go CLI binary that talks directly to the existing Skillforge HTTP API and presents both human-readable and JSON output. Start with read commands first, then extend into draft lifecycle commands within the same change.

## Goals

- Ship the first practical user-facing interface for the current backend
- Reuse the existing API contract with minimal new backend coupling
- Keep implementation low-dependency and easy to test
- Support both interactive human use and shell automation

## Non-goals

- building the first TypeScript web UI
- redesigning the existing HTTP API surface
- introducing local offline skill mutation without the backend
- packaging/distribution work beyond repo-local build/run examples

## Architecture

### CLI binary

Add `cmd/skillforge` as the first user-facing CLI entrypoint.

Use standard-library argument parsing with explicit subcommands to keep dependencies small.

Planned commands for this change:

- `list`
- `search <query>`
- `get <skill-name>`
- `draft create ...`
- `draft status <id>`
- `draft submit <id>`

### Shared HTTP client

Add a small internal client package that:

- stores API base URL and underlying `http.Client`
- shapes requests to the existing API endpoints
- decodes success payloads into shared Go structs where practical
- decodes machine-readable backend errors (`error`, `message`) into a typed error for CLI handling

### Output model

Each command should support:

- a default human-readable format optimized for terminal use
- `--json` for structured output when shell automation matters

Human-readable output should stay concise and stable enough for users, while `--json` should preserve the full backend response shape where possible.

### Initial implementation batch

This first development slice will implement:

- shared client setup
- `list`
- `search`
- `get`
- output helpers and CLI tests for those commands

The remaining draft subcommands stay in-scope for the change but can land in later development slices.

## Verification strategy

- unit tests for the HTTP client against `httptest`
- command-level tests for argument parsing and representative output
- full repo checks with `gofmt`, `go test`, `go vet`, and `staticcheck`

## Risks and tradeoffs

- Standard-library flag parsing keeps the dependency surface small, but subcommand UX will be slightly more manual than a Cobra-style CLI
- Reusing backend response types reduces duplication, but the CLI must avoid leaking awkward raw JSON into the default human-readable path
- Landing read commands first delays full draft usability slightly, but produces a smaller, reviewable first implementation batch
