# Tasks: go-cli-client

## 1. OpenSpec and scaffolding

- [x] 1.1 Write proposal/design/tasks for the first Go CLI change.
- [x] 1.2 Add CLI-focused spec deltas for command coverage and output modes.

## 2. Shared client foundation

- [x] 2.1 Add a reusable HTTP client for the Skillforge API with typed error handling.
- [x] 2.2 Define common CLI config/flags for API base URL and JSON output.

## 3. First read commands

- [x] 3.1 Implement `skillforge list` against `GET /api/v1/skills`.
- [x] 3.2 Implement `skillforge search <query>` against `GET /api/v1/search`.
- [x] 3.3 Implement `skillforge get <skill-name>` against `GET /api/v1/skills/{name}`.
- [x] 3.4 Add tests for read-command request shaping and output.

## 4. Draft lifecycle commands

- [x] 4.1 Implement `skillforge draft create`.
- [x] 4.2 Implement `skillforge draft status`.
- [x] 4.3 Implement `skillforge draft submit`.
- [x] 4.4 Add tests for submission-disabled and validation-error handling.

## 5. Docs and rollout

- [x] 5.1 Document local CLI usage in the README/dev docs.
- [x] 5.2 Open or update the round-3 PR with the first CLI batch.
- [x] 5.3 Validate the change strictly.
