## ADDED Requirements

### Requirement: Provide a self-hosted local compose stack
The system SHALL provide a Docker Compose topology for local self-hosted development and demonstration.

#### Scenario: Start local stack
- **WHEN** an operator starts the compose stack
- **THEN** Forgejo and the Skillforge API services start with compatible networking and configuration

### Requirement: Make canonical repository source explicit
The compose setup SHALL make the canonical skills repository used by the API explicit and configurable.

#### Scenario: Configure repository source for API
- **WHEN** the API service starts in the compose stack
- **THEN** its configured canonical repository path or clone target is visible in service configuration

### Requirement: Support local verification with sample content
The development stack SHALL support local verification with sample skill content.

#### Scenario: Verify API against sample skill repository
- **WHEN** the compose stack is started with seeded sample skills
- **THEN** an operator can successfully exercise list, search, and get flows against non-empty data
