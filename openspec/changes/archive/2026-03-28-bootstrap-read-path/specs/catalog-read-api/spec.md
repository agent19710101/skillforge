## ADDED Requirements

### Requirement: List indexed skills
The system SHALL provide a read-only API endpoint for listing indexed skills derived from the canonical repository state.

#### Scenario: List all indexed skills
- **WHEN** a client requests `GET /api/v1/skills`
- **THEN** the system returns a paginated list of indexed skills

#### Scenario: List only valid skills
- **WHEN** a client requests `GET /api/v1/skills?validation=valid`
- **THEN** the system returns only skills whose latest scan has no validation findings

### Requirement: Retrieve skill details by name
The system SHALL provide a read-only API endpoint for fetching a skill by canonical name.

#### Scenario: Fetch existing skill
- **WHEN** a client requests `GET /api/v1/skills/example-skill`
- **THEN** the system returns the indexed metadata and validation summary for `example-skill`

#### Scenario: Fetch unknown skill
- **WHEN** a client requests `GET /api/v1/skills/missing-skill`
- **THEN** the system returns a not-found response

### Requirement: Search indexed skills by metadata
The system SHALL provide a read-only API endpoint for metadata-oriented search across indexed skills.

#### Scenario: Search by name
- **WHEN** a client requests `GET /api/v1/search?q=example`
- **THEN** matching skills by name are returned

#### Scenario: Search by description or tags
- **WHEN** a client requests `GET /api/v1/search?q=pdf`
- **THEN** matching skills by description and indexed tag-like metadata are returned

### Requirement: Expose index health and provenance
The system SHALL provide a read-only API endpoint that reports the current derived index status.

#### Scenario: Read current index status
- **WHEN** a client requests `GET /api/v1/index/status`
- **THEN** the system returns the indexed repository revision, last successful scan timestamp, indexed skill count, and validation error count
