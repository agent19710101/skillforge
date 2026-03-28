## MODIFIED Requirements

### Requirement: Publish draft branches to Forgejo
The system SHALL publish validated draft branches to the configured Forgejo repository from a managed canonical working copy owned by the backend.

#### Scenario: Push validated draft branch from managed working copy
- **WHEN** a draft passes validation and a client requests publication
- **THEN** the system materializes the intended draft skill changes into the managed canonical working copy
- **AND** commits and pushes the draft branch to Forgejo from that working copy
- **AND** does not require the temp draft workspace itself to be a Git repository

### Requirement: Create pull requests from published drafts
The system SHALL create Forgejo pull requests from published drafts using operator-configured Forgejo connection details.

#### Scenario: Open pull request with configured Forgejo target
- **WHEN** a published draft branch exists in Forgejo
- **THEN** the system creates a pull request targeting the configured base branch on the configured Forgejo repository
- **AND** returns the pull request identifier and URL
