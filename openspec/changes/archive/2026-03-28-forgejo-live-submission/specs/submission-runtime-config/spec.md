## ADDED Requirements

### Requirement: Load explicit submission runtime configuration
The system SHALL load explicit runtime configuration for live draft submission instead of assuming submission is always available.

#### Scenario: Submission enabled only when configuration is complete
- **WHEN** the API server starts with canonical working-copy, remote, base branch, Forgejo target, and authentication settings configured
- **THEN** the server enables live draft submission

#### Scenario: Submission disabled when configuration is incomplete
- **WHEN** the API server starts without one or more required submission settings
- **THEN** the server keeps submission disabled
- **AND** reports a machine-readable reason to clients

### Requirement: Isolate writable publication from temp draft editing
The system SHALL preserve the temp draft workspace model while using a managed canonical working copy for Git publication.

#### Scenario: Submit validated temp draft through canonical working copy
- **WHEN** a client submits a validated draft created from a temp workspace
- **THEN** the backend applies the intended skill changes into the managed canonical working copy before committing and pushing
- **AND** does not require the temp workspace itself to have Git metadata
