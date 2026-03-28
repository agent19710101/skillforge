## ADDED Requirements

### Requirement: Provide a read-focused Skillforge web UI
The system SHALL provide a browser-based web UI for the existing Skillforge read API surface.

#### Scenario: Browse skills in the web UI
- **WHEN** a user opens the Skillforge web UI
- **THEN** the UI loads the skill list from `GET /api/v1/skills`
- **AND** shows a skill list view

### Requirement: Support search and detail flows in the first UI slice
The first web UI slice SHALL support search and skill detail inspection without draft editing.

#### Scenario: Search skills in the web UI
- **WHEN** a user submits a search query
- **THEN** the UI calls `GET /api/v1/search?q=...`
- **AND** shows matching skills or an explicit empty state

#### Scenario: View skill details in the web UI
- **WHEN** a user selects a skill from the list or search results
- **THEN** the UI calls `GET /api/v1/skills/{name}`
- **AND** shows the selected skill details

### Requirement: Surface representative runtime states honestly
The first web UI slice SHALL expose loading and error states instead of silently failing.

#### Scenario: Read request is loading
- **WHEN** the web UI is waiting on a read API request
- **THEN** it shows a loading state for that region

#### Scenario: Read request fails
- **WHEN** the web UI receives a failed read API response
- **THEN** it shows an explicit error state for that region
