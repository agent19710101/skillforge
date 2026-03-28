# Sample Skillforge repository

This directory is a small sample **canonical skills repository** for local Skillforge development.

It is mounted into the `skillforge-api` container at:

```text
/data/skills-repo
```

The API treats that path as the canonical repository root via `SKILLFORGE_REPO_ROOT`.

For local verification, the sample content is enough to exercise:

- `GET /api/v1/skills`
- `GET /api/v1/skills/<name>`
- `GET /api/v1/search?q=...`
- `GET /api/v1/index/status`

If you want the API to report a real Git revision instead of an empty revision, initialize this directory as a Git repository before starting the stack:

```bash
cd dev/sample-skills-repo
git init
git add .
git commit -m "seed sample skills repo"
```
