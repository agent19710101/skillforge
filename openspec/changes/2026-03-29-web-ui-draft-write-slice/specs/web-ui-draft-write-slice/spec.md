## ADDED Requirements

### Requirement: Provide a browser draft authoring flow
The Skillforge web UI SHALL support creating browser-driven drafts against the existing draft lifecycle API.

#### Scenario: Create a draft from the web UI
- **WHEN** a user fills the draft form and submits it from the web UI
- **THEN** the UI calls `POST /api/v1/drafts`
- **AND** shows the created draft metadata, validation result, and submission capability

### Requirement: Support multiple draft operations in the web UI
The first write-capable UI slice SHALL support create, update, and delete draft operations.

#### Scenario: Create or update draft with content
- **WHEN** a user chooses `create` or `update`
- **THEN** the UI includes draft content in the create request

#### Scenario: Create delete draft without content
- **WHEN** a user chooses `delete`
- **THEN** the UI creates the draft without requiring content
- **AND** clearly communicates that the delete operation removes the current skill from the draft workspace

### Requirement: Surface draft submission readiness honestly
The write-capable web UI SHALL expose whether the created draft can be submitted and why not when submission is unavailable.

#### Scenario: Draft submission is unavailable
- **WHEN** the draft create or status response reports `submission.enabled = false`
- **THEN** the UI shows that submission is unavailable
- **AND** includes the reported reason when one exists

#### Scenario: Submit the current draft from the web UI
- **WHEN** a user submits a created draft from the web UI
- **THEN** the UI calls `POST /api/v1/drafts/{id}/submit`
- **AND** shows the returned branch/base-branch metadata and pull request information when available

### Requirement: Show explicit write-path failures
The write-capable web UI SHALL surface create and submit failures instead of silently failing.

#### Scenario: Draft creation fails
- **WHEN** draft creation returns a non-success response
- **THEN** the UI shows an explicit draft creation error state

#### Scenario: Draft submission fails
- **WHEN** draft submission returns a non-success response
- **THEN** the UI shows an explicit draft submission error state
- **AND** preserves any returned validation or submission status details when available
