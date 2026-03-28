## MODIFIED Requirements

### Requirement: Provide a self-hosted local compose stack
The system SHALL provide a Docker Compose topology for local self-hosted development and demonstration, including an intentional path for enabling live draft submission.

#### Scenario: Start local stack with live submission disabled by default
- **WHEN** an operator starts the compose stack without submission credentials or a writable canonical working copy
- **THEN** Forgejo and the Skillforge API services start successfully
- **AND** the API reports submission as unavailable with a machine-readable reason

### Requirement: Make canonical repository source explicit
The compose setup SHALL make the canonical skills repository used by the API explicit and configurable for both read-only indexing and writable submission flows.

#### Scenario: Configure writable canonical repository for API
- **WHEN** an operator enables live submission in the compose stack
- **THEN** the API service configuration clearly identifies the writable canonical repository path, remote, base branch, and Forgejo connection inputs
