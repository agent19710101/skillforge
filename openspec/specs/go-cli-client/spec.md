# go-cli-client Specification

## Purpose
Define the first terminal-native Skillforge CLI and its contract for browse, draft inspection, and draft submission workflows against the existing HTTP API.
## Requirements
### Requirement: Provide a terminal-native Skillforge CLI
The system SHALL provide a Go CLI that talks to the existing Skillforge HTTP API for browse and draft workflows.

#### Scenario: Read command targets configured API base URL
- **WHEN** a user runs a CLI read command with an explicit server URL
- **THEN** the command sends requests to that Skillforge API base URL

### Requirement: Support human-readable and JSON output
The CLI SHALL support concise human-readable output by default and structured JSON output for automation.

#### Scenario: Read command uses JSON output
- **WHEN** a user runs a supported CLI command with `--json`
- **THEN** the command emits the structured API response as JSON

#### Scenario: Read command uses default output
- **WHEN** a user runs a supported CLI command without `--json`
- **THEN** the command emits a concise human-readable summary

### Requirement: Expose read-path commands
The CLI SHALL expose commands for the shipped read API surface.

#### Scenario: List skills from the CLI
- **WHEN** a user runs `skillforge list`
- **THEN** the CLI calls `GET /api/v1/skills` and prints the returned skills

#### Scenario: Search skills from the CLI
- **WHEN** a user runs `skillforge search git`
- **THEN** the CLI calls `GET /api/v1/search?q=git` and prints matching skills

#### Scenario: Fetch a skill from the CLI
- **WHEN** a user runs `skillforge get git-pr-review`
- **THEN** the CLI calls `GET /api/v1/skills/git-pr-review` and prints the indexed skill details

### Requirement: Expose draft lifecycle commands
The CLI SHALL expose commands for the shipped draft lifecycle API surface.

#### Scenario: Create a draft from the CLI
- **WHEN** a user runs `skillforge draft create --operation create --skill new-skill --content ...`
- **THEN** the CLI calls `POST /api/v1/drafts`
- **AND** prints the resulting draft status and submission capability

#### Scenario: Inspect draft status from the CLI
- **WHEN** a user runs `skillforge draft status draft01`
- **THEN** the CLI calls `GET /api/v1/drafts/draft01`
- **AND** prints the current validation and submission state

#### Scenario: Submit a draft from the CLI
- **WHEN** a user runs `skillforge draft submit draft01`
- **THEN** the CLI calls `POST /api/v1/drafts/draft01/submit`
- **AND** prints the resulting branch, commit, and pull request metadata on success

#### Scenario: Submission unavailable is explained honestly
- **WHEN** the backend returns `submission_unavailable` for `skillforge draft submit draft01`
- **THEN** the CLI exits non-zero
- **AND** explains that live submission is unavailable and why

#### Scenario: Validation failures are surfaced clearly
- **WHEN** the backend returns `draft_invalid` for `skillforge draft submit draft01`
- **THEN** the CLI exits non-zero
- **AND** prints the blocking validation findings for the draft

