# skill-write-flow Specification

## Purpose
Define the managed draft-workspace model for creating, updating, deleting, and validating canonical skills before review submission.
## Requirements
### Requirement: Create managed draft workspaces
The system SHALL create managed draft workspaces derived from the canonical repository for skill authoring flows.

#### Scenario: Create draft for new skill
- **WHEN** a client requests a draft for creating `skills/new-skill/`
- **THEN** the system creates a draft workspace bound to a unique branch name
- **AND** the draft is based on the configured canonical base branch

#### Scenario: Create draft for editing existing skill
- **WHEN** a client requests a draft for updating `existing-skill`
- **THEN** the system creates a draft workspace initialized from current canonical repository state

### Requirement: Apply semantic skill mutations
The system SHALL expose semantic create, update, and delete mutations against draft workspaces.

#### Scenario: Create skill in draft workspace
- **WHEN** a client submits valid create data for `new-skill`
- **THEN** the system writes canonical files under `skills/new-skill/`
- **AND** the draft reflects the staged change without changing the default branch

#### Scenario: Update skill in draft workspace
- **WHEN** a client submits valid updated content for `existing-skill`
- **THEN** the system updates files only inside the draft workspace

#### Scenario: Delete skill in draft workspace
- **WHEN** a client requests deletion of `obsolete-skill`
- **THEN** the system stages removal of `skills/obsolete-skill/` only inside the draft workspace

### Requirement: Validate drafts before submission
The system SHALL run repository and AgentSkills validation against draft workspaces before submission.

#### Scenario: Reject invalid draft submission
- **WHEN** a draft workspace contains a skill whose `name` does not match its directory
- **THEN** the system returns validation findings
- **AND** the system SHALL NOT allow draft submission until the findings are resolved

