## ADDED Requirements

### Requirement: Discover canonical skill directories
The system SHALL discover skills only from canonical directories under `skills/<skill-name>/` in the configured Git repository revision.

#### Scenario: Valid canonical skill directory is discovered
- **WHEN** the repository contains `skills/example-skill/SKILL.md`
- **THEN** the scan includes `example-skill` as a discovered skill candidate

#### Scenario: Non-canonical directory is ignored
- **WHEN** a `SKILL.md` file exists outside `skills/<skill-name>/`
- **THEN** the system SHALL NOT treat it as a canonical skill entry

### Requirement: Validate skill identity and required files
The system SHALL validate each discovered skill directory for required files and identity consistency.

#### Scenario: Skill name matches directory name
- **WHEN** `skills/example-skill/SKILL.md` declares `name: example-skill`
- **THEN** the skill passes identity consistency validation

#### Scenario: Skill name mismatches directory name
- **WHEN** `skills/example-skill/SKILL.md` declares `name: other-name`
- **THEN** the system records a validation finding for that skill

#### Scenario: SKILL.md is missing
- **WHEN** `skills/example-skill/` exists without `SKILL.md`
- **THEN** the system records the directory as invalid and excludes it from valid catalog records

### Requirement: Produce normalized metadata records
The system SHALL parse `SKILL.md` frontmatter and produce normalized metadata records for downstream indexing.

#### Scenario: Valid frontmatter is parsed
- **WHEN** a skill contains valid YAML frontmatter with supported AgentSkills fields
- **THEN** the system extracts normalized metadata including `name` and `description`

#### Scenario: Invalid frontmatter is reported
- **WHEN** a skill contains malformed YAML frontmatter
- **THEN** the system records a validation finding without aborting the entire repository scan

### Requirement: Preserve partial scan success
The system SHALL continue scanning remaining skills even when one or more skills are invalid.

#### Scenario: One invalid and one valid skill coexist
- **WHEN** the repository contains one malformed skill and one valid skill
- **THEN** the valid skill remains available in scan results and the malformed skill includes validation findings
