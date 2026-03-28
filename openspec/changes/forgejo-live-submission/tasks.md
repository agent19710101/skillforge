# Tasks: forgejo-live-submission

## 1. Runtime configuration and server wiring

- [x] 1.1 Define environment-driven submission runtime configuration for the API server.
- [x] 1.2 Build real submission dependencies only when the configuration is complete.
- [x] 1.3 Keep submission capability/status responses accurate for configured vs. unconfigured servers.
- [x] 1.4 Add tests for configuration loading and capability signaling.

## 2. Real Git publisher

- [x] 2.1 Implement a concrete Git publisher that operates on a managed canonical working copy.
- [x] 2.2 Materialize only the intended `skills/<skill-name>/` draft changes from the temp workspace into the canonical working copy.
- [x] 2.3 Create/reset the submission branch from the configured base branch, commit the changes, and push to the configured remote.
- [x] 2.4 Add tests using temp Git repositories for publish success and failure paths.

## 3. Real Forgejo PR client

- [ ] 3.1 Implement a narrow Forgejo client for pull-request creation.
- [ ] 3.2 Return machine-readable PR identifier and URL from the live client.
- [ ] 3.3 Add tests for request shaping and response handling.

## 4. Dev stack and operator docs

- [ ] 4.1 Document the required submission environment variables and working-copy expectations.
- [ ] 4.2 Update Compose/dev documentation so operators can enable live submission intentionally.
- [ ] 4.3 Add a local verification flow for creating a draft and opening a PR against Forgejo.

## 5. Verification and rollout

- [x] 5.1 Validate the OpenSpec change strictly.
- [x] 5.2 Implement the first live-submission backend slice on a feature branch and open/update a PR.
- [ ] 5.3 Archive the change when the live submission path is coherent end to end.
