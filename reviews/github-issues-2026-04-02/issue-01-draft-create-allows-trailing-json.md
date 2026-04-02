# Title
`POST /api/v1/drafts` accepts trailing JSON payloads after the first object

# Type
Bug

# Severity
Medium

# Summary
The request decoder in `decodeDraftCreateRequest` only performs a single `Decode` call and does not verify end-of-input. This allows extra trailing tokens (including a second JSON object) to be silently ignored.

# Impact
- Client/server parsing can become ambiguous when intermediaries or client libraries treat the trailing bytes differently.
- Attackers can hide unexpected payloads in logs and audit streams while still getting a successful response for the first object.
- API behavior is inconsistent with strict JSON request contracts expected by many clients.

# Evidence
`decodeDraftCreateRequest` uses:

```go
dec := json.NewDecoder(r.Body)
dec.DisallowUnknownFields()
if err := dec.Decode(&req); err != nil {
    return draftCreateRequest{}, fmt.Errorf("decode request body: %w", err)
}
return req, nil
```

There is no second decode / EOF check.

# Reproduction
```bash
curl -i -X POST http://localhost:8080/api/v1/drafts \
  -H 'content-type: application/json' \
  --data '{"operation":"create","skillName":"demo-skill","content":"# Skill"}{"ignored":true}'
```

Expected: HTTP 400 invalid request.
Observed: request can be accepted based on the first object.

# Suggested fix
After the first successful decode, decode once more into `struct{}` and require `io.EOF`:

```go
if err := dec.Decode(&struct{}{}); err != io.EOF {
    return draftCreateRequest{}, fmt.Errorf("decode request body: trailing content")
}
```

# Affected files
- `internal/api/drafts.go`
