# Title
Draft workspace copy follows symlinks and can read files outside the repository

# Type
Security Bug

# Severity
High

# Summary
`copyRepoTree` copies repository files by opening each path directly. When the source repository contains symlinks, `os.Open(src)` follows the symlink target. This can copy arbitrary host files into the draft workspace if a symlink points outside the repo.

# Impact
- Sensitive host data may be read into a draft workspace unexpectedly.
- Subsequent submit/publish flow could commit leaked data into branches/PRs.
- In multi-tenant/self-hosted scenarios this becomes a data exfiltration risk.

# Evidence
The copy flow:

```go
return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
    ...
    info, err := d.Info()
    ...
    return copyFile(path, target, info.Mode())
})
```

`copyFile` then does:

```go
in, err := os.Open(src)
...
io.Copy(out, in)
```

No symlink filtering is present.

# Reproduction
1. Add a symlink in the canonical repo: `ln -s /etc/hosts skills/leak/SKILL.md`
2. Trigger draft workspace creation.
3. Observe the resulting workspace file contains `/etc/hosts` content, not symlink metadata.

# Suggested fix
- During tree walk, detect and reject symlinks (`d.Type()&fs.ModeSymlink != 0`) or preserve symlinks without dereferencing.
- Enforce that all copied paths resolve within the repo root before opening.
- Add a regression test with an external symlink target.

# Affected files
- `internal/draft/manager.go`
