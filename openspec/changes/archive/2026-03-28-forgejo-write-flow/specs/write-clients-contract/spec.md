## ADDED Requirements

### Requirement: Expose a shared draft lifecycle contract
The system SHALL provide a shared backend contract for future CLI and web UI authoring flows.

#### Scenario: Draft lifecycle contract is interface-neutral
- **WHEN** a CLI client and the web UI both create and inspect drafts
- **THEN** they use the same draft identity, validation, and submission semantics

### Requirement: Report draft validation state
The system SHALL return draft validation findings in a machine-readable form.

#### Scenario: Inspect invalid draft status
- **WHEN** a client requests draft status for a draft with validation errors
- **THEN** the response includes validation findings that identify the blocking issues

### Requirement: Report submission results
The system SHALL return submission metadata after successful PR creation.

#### Scenario: Submission returns PR metadata
- **WHEN** a draft is successfully submitted for review
- **THEN** the response includes the resulting branch name, pull request identifier, and pull request URL
