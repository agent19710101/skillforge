## ADDED Requirements

### Requirement: Publish draft branches to Forgejo
The system SHALL publish validated draft branches to the configured Forgejo repository.

#### Scenario: Push validated draft branch
- **WHEN** a draft passes validation and a client requests publication
- **THEN** the system commits staged changes and pushes the draft branch to Forgejo

### Requirement: Create pull requests from published drafts
The system SHALL create Forgejo pull requests from published draft branches to the configured base branch.

#### Scenario: Open pull request for validated draft
- **WHEN** a published draft branch exists in Forgejo
- **THEN** the system creates a pull request targeting the configured base branch
- **AND** returns the pull request identifier and URL

### Requirement: Preserve review-based publication
The system SHALL use pull requests as the publication path for canonical skill changes.

#### Scenario: Default branch remains unchanged before merge
- **WHEN** a draft branch has been pushed and a pull request opened
- **THEN** the default branch content remains unchanged until the pull request is merged in Forgejo
