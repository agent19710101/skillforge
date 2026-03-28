## MODIFIED Requirements

### Requirement: Expose a shared draft lifecycle contract
The system SHALL provide a shared backend contract for future CLI and web UI authoring flows, including machine-readable submission capability reporting.

#### Scenario: Draft lifecycle contract reports submission capability
- **WHEN** a client creates or inspects a draft
- **THEN** the response includes whether live submission is enabled
- **AND** includes a reason when submission is unavailable because the server is not configured

### Requirement: Report submission results
The system SHALL return submission metadata after successful PR creation from the live backend.

#### Scenario: Submission returns live publication metadata
- **WHEN** a draft is successfully submitted for review against a configured Forgejo target
- **THEN** the response includes the resulting branch name, pull request identifier, and pull request URL
- **AND** identifies the configured base branch used for the submission
