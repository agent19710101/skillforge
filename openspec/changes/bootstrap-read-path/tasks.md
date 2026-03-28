## 1. Repository ingestion foundation

- [x] 1.1 Create the Go module and backend service layout for repository scanning and HTTP serving.
- [x] 1.2 Implement repository discovery for `skills/<skill-name>/SKILL.md` entries.
- [x] 1.3 Implement `SKILL.md` frontmatter parsing and normalized metadata extraction.
- [x] 1.4 Implement validation for directory naming, required files, and `name`/directory consistency.
- [x] 1.5 Produce a scan result model that includes both valid records and validation findings.

## 2. Derived catalog and read API

- [x] 2.1 Implement an in-process derived index for list/search/get operations.
- [x] 2.2 Add `GET /api/v1/skills` with pagination and basic filtering.
- [x] 2.3 Add `GET /api/v1/skills/:name` for skill detail retrieval.
- [x] 2.4 Add `GET /api/v1/search` for metadata-oriented search.
- [x] 2.5 Add `GET /api/v1/index/status` for scan/index health visibility.

## 3. Self-hosted development stack

- [x] 3.1 Add Dockerfiles and a Docker Compose topology for Forgejo and Skillforge API.
- [x] 3.2 Document the canonical repository mount/clone strategy used by the API.
- [x] 3.3 Seed the dev stack with a small sample skills repository for local verification.

## 4. Verification and developer workflow

- [x] 4.1 Add automated tests for repository scanning, parsing, validation, and search behavior.
- [x] 4.2 Run formatting, tests, and linting for the backend slice.
- [ ] 4.3 Prepare the follow-up change for write flows, branch creation, and PR submission.
