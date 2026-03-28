## Why

Skillforge now has a coherent read path:

- canonical repository scanning
- validation and derived indexing
- read-only list/search/get APIs
- a local Forgejo + API development stack

What is still missing from the product is the actual authoring and review workflow. The user requirement is not just to browse skills; people must be able to create, update, and delete skills through a branch-based flow that ends in a reviewed pull request.

The next development slice should therefore add the mutation path on top of the proven read model.

## What Changes

This change introduces the write-flow architecture for Skillforge:

1. a managed skill workspace model for preparing changes against the single canonical repository
2. write APIs for create, update, and delete operations that stage changes into a branch workspace rather than mutating the default branch directly
3. Forgejo integration for branch publication and pull request creation
4. follow-on CLI and web UI contracts for submitting edits through the same backend workflow

The branch/PR model preserves Git as the source of truth and aligns skill mutation with the organization’s review process.

## Capabilities

### New Capabilities
- `skill-write-flow`: Stage create/update/delete changes for skills into a managed branch workspace backed by the canonical repository.
- `forgejo-pr-submission`: Publish staged branch changes to Forgejo and open a reviewable pull request.
- `write-clients-contract`: Define the backend contract expected by both the CLI and web UI for authoring flows.

### Modified Capabilities
- `catalog-read-api`: Remains the read surface that clients use to verify current state before and after staged changes are proposed.

## Impact

Affected areas:

- backend repository workspace management
- branch naming and conflict handling
- Forgejo API integration and authentication strategy
- future CLI create/update/delete/submit commands
- future web UI authoring and submission flows
- review lifecycle from staged edits to merged PRs
