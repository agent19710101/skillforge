## ADDED Requirements

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
