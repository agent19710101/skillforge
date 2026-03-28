## 1. Draft workspace and mutation model

- [x] 1.1 Define the backend package layout for managed draft workspaces.
- [x] 1.2 Implement workspace creation from the canonical repository with deterministic branch naming.
- [x] 1.3 Implement semantic create/update/delete mutations under `skills/<skill-name>/`.
- [x] 1.4 Reuse repository/AgentSkills validation against the draft workspace.
- [x] 1.5 Add tests for valid mutations, invalid drafts, and branch/workspace isolation.

## 2. Forgejo submission flow

- [x] 2.1 Define configuration required for Forgejo remote, credentials, owner/repo, and base branch.
- [x] 2.2 Implement draft commit creation and branch publication.
- [x] 2.3 Implement pull request creation against Forgejo.
- [x] 2.4 Add tests/fakes around branch publication and PR submission boundaries.

## 3. Shared client contract

- [x] 3.1 Add API specs for draft lifecycle endpoints used by CLI and web UI.
- [x] 3.2 Document draft status, validation findings, and submission responses.
- [x] 3.3 Define the initial CLI and web UI integration points for create/update/delete/submit flows.

## 4. Verification and rollout

- [x] 4.1 Validate the OpenSpec change strictly.
- [x] 4.2 Implement the first backend slice on a feature branch and update the open PR.
- [ ] 4.3 Archive the change when the write-flow implementation slice is coherent.
