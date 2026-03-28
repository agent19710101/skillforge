## 1. Draft workspace and mutation model

- [ ] 1.1 Define the backend package layout for managed draft workspaces.
- [ ] 1.2 Implement workspace creation from the canonical repository with deterministic branch naming.
- [ ] 1.3 Implement semantic create/update/delete mutations under `skills/<skill-name>/`.
- [ ] 1.4 Reuse repository/AgentSkills validation against the draft workspace.
- [ ] 1.5 Add tests for valid mutations, invalid drafts, and branch/workspace isolation.

## 2. Forgejo submission flow

- [ ] 2.1 Define configuration required for Forgejo remote, credentials, owner/repo, and base branch.
- [ ] 2.2 Implement draft commit creation and branch publication.
- [ ] 2.3 Implement pull request creation against Forgejo.
- [ ] 2.4 Add tests/fakes around branch publication and PR submission boundaries.

## 3. Shared client contract

- [ ] 3.1 Add API specs for draft lifecycle endpoints used by CLI and web UI.
- [ ] 3.2 Document draft status, validation findings, and submission responses.
- [ ] 3.3 Define the initial CLI and web UI integration points for create/update/delete/submit flows.

## 4. Verification and rollout

- [ ] 4.1 Validate the OpenSpec change strictly.
- [ ] 4.2 Implement the first backend slice on a feature branch and update the open PR.
- [ ] 4.3 Archive the change when the write-flow implementation slice is coherent.
