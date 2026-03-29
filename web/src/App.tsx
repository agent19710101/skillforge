import { useEffect, useMemo, useState, type FormEvent } from 'react'
import './App.css'
import {
  ApiRequestError,
  createDraft,
  getDraft,
  getIndexStatus,
  getSkill,
  listAllSkills,
  searchSkills,
  submitDraft,
  type DraftOperation,
  type DraftResponse,
  type DraftSubmissionResponse,
  type IndexStatus,
  type SkillRecord,
} from './api'

type LoadState = 'idle' | 'loading' | 'ready' | 'error'
type DraftState = 'idle' | 'creating' | 'refreshing' | 'ready' | 'error'
type SubmitState = 'idle' | 'submitting' | 'ready' | 'error'

type UiLocationState = {
  query: string
  skill: string
}

const defaultDraftContent = `---
name: new-skill
description: Describe the skill
---
# new-skill
`

function App() {
  const initialLocationState = readLocationState()
  const [query, setQuery] = useState(initialLocationState.query)
  const [submittedQuery, setSubmittedQuery] = useState(initialLocationState.query)
  const [skills, setSkills] = useState<SkillRecord[]>([])
  const [skillsState, setSkillsState] = useState<LoadState>('idle')
  const [skillsError, setSkillsError] = useState('')
  const [selectedSkillName, setSelectedSkillName] = useState(initialLocationState.skill)
  const [selectedSkill, setSelectedSkill] = useState<SkillRecord | null>(null)
  const [detailState, setDetailState] = useState<LoadState>('idle')
  const [detailError, setDetailError] = useState('')
  const [indexStatus, setIndexStatus] = useState<IndexStatus | null>(null)

  const [draftOperation, setDraftOperation] = useState<DraftOperation>('create')
  const [draftSkillName, setDraftSkillName] = useState('')
  const [draftContent, setDraftContent] = useState(defaultDraftContent)
  const [draftState, setDraftState] = useState<DraftState>('idle')
  const [draftError, setDraftError] = useState('')
  const [currentDraft, setCurrentDraft] = useState<DraftResponse | null>(null)
  const [submitState, setSubmitState] = useState<SubmitState>('idle')
  const [submitError, setSubmitError] = useState('')
  const [submissionResult, setSubmissionResult] = useState<DraftSubmissionResponse | null>(null)

  useEffect(() => {
    syncLocationState({ query: submittedQuery, skill: selectedSkillName })
  }, [submittedQuery, selectedSkillName])

  useEffect(() => {
    let cancelled = false
    getIndexStatus()
      .then((status) => {
        if (!cancelled) {
          setIndexStatus(status)
        }
      })
      .catch(() => {
        if (!cancelled) {
          setIndexStatus(null)
        }
      })
    return () => {
      cancelled = true
    }
  }, [])

  useEffect(() => {
    let cancelled = false
    setSkillsState('loading')
    setSkillsError('')

    const request = submittedQuery.trim() === '' ? listAllSkills() : searchSkills(submittedQuery)

    request
      .then((response) => {
        if (cancelled) {
          return
        }
        const nextSkills = response.skills
        setSkills(nextSkills)
        setSkillsState('ready')
        setSelectedSkillName((current) => {
          if (current && nextSkills.some((skill) => skill.name === current)) {
            return current
          }
          return nextSkills[0]?.name ?? ''
        })
      })
      .catch((error: Error) => {
        if (cancelled) {
          return
        }
        setSkills([])
        setSkillsState('error')
        setSkillsError(error.message)
        setSelectedSkillName('')
      })

    return () => {
      cancelled = true
    }
  }, [submittedQuery])

  useEffect(() => {
    if (!selectedSkillName) {
      setSelectedSkill(null)
      setDetailState(skillsState === 'error' ? 'error' : 'idle')
      return
    }

    let cancelled = false
    setDetailState('loading')
    setDetailError('')

    getSkill(selectedSkillName)
      .then((skill) => {
        if (cancelled) {
          return
        }
        setSelectedSkill(skill)
        setDetailState('ready')
      })
      .catch((error: Error) => {
        if (cancelled) {
          return
        }
        setSelectedSkill(null)
        setDetailState('error')
        setDetailError(error.message)
      })

    return () => {
      cancelled = true
    }
  }, [selectedSkillName, skillsState])

  const statusLabel = useMemo(() => {
    if (!indexStatus) {
      return 'API status unavailable'
    }
    return `${indexStatus.skillCount} indexed skill(s)`
  }, [indexStatus])

  const createButtonLabel = draftState === 'creating' ? 'Creating draft…' : 'Create draft'
  const submitButtonLabel = submitState === 'submitting' ? 'Submitting…' : 'Submit draft'

  const onSearch = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    setSubmittedQuery(query.trim())
  }

  const onReset = () => {
    setQuery('')
    setSubmittedQuery('')
    setSelectedSkillName('')
  }

  const onCreateDraft = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    setDraftState('creating')
    setDraftError('')
    setSubmitState('idle')
    setSubmitError('')
    setSubmissionResult(null)

    try {
      const created = await createDraft({
        operation: draftOperation,
        skillName: draftSkillName.trim(),
        ...(draftOperation === 'delete' ? {} : { content: draftContent }),
      })
      setCurrentDraft(created)
      setDraftState('ready')
    } catch (error) {
      setCurrentDraft(null)
      setDraftState('error')
      setDraftError(error instanceof Error ? error.message : 'request failed')
    }
  }

  const onRefreshDraft = async () => {
    if (!currentDraft) {
      return
    }

    setDraftState('refreshing')
    setDraftError('')
    try {
      const refreshed = await getDraft(currentDraft.id)
      setCurrentDraft(refreshed)
      setDraftState('ready')
    } catch (error) {
      setDraftState('error')
      setDraftError(error instanceof Error ? error.message : 'request failed')
    }
  }

  const onSubmitDraft = async () => {
    if (!currentDraft) {
      return
    }

    setSubmitState('submitting')
    setSubmitError('')
    try {
      const result = await submitDraft(currentDraft.id)
      setSubmissionResult(result)
      setSubmitState('ready')
    } catch (error) {
      if (error instanceof ApiRequestError) {
        setCurrentDraft((previous) => {
          if (!previous) {
            return previous
          }
          return {
            ...previous,
            validation: error.validation ?? previous.validation,
            submission: error.submission ?? previous.submission,
          }
        })
      }
      setSubmissionResult(null)
      setSubmitState('error')
      setSubmitError(error instanceof Error ? error.message : 'request failed')
    }
  }

  const onPrefillFromSelection = () => {
    if (!selectedSkill) {
      return
    }
    setDraftOperation('update')
    setDraftSkillName(selectedSkill.name)
    setDraftContent(selectedSkill.body ?? '')
  }

  const onPrepareDeleteFromSelection = () => {
    if (!selectedSkill) {
      return
    }
    setDraftOperation('delete')
    setDraftSkillName(selectedSkill.name)
  }

  return (
    <div className="app-shell">
      <header className="topbar">
        <div>
          <p className="eyebrow">Skillforge</p>
          <h1>Browse and draft skills</h1>
          <p className="subtitle">Read the catalog, then create and submit draft mutations through the current Go API.</p>
        </div>
        <div className="status-card" aria-label="index status">
          <span className="status-label">Catalog status</span>
          <strong>{statusLabel}</strong>
          {indexStatus?.git?.commit ? <code>{indexStatus.git.commit}</code> : null}
        </div>
      </header>

      <main className="layout">
        <section className="panel sidebar">
          <form className="search-form" onSubmit={onSearch}>
            <label htmlFor="search-input">Search skills</label>
            <div className="search-row">
              <input
                id="search-input"
                name="q"
                value={query}
                onChange={(event) => setQuery(event.target.value)}
                placeholder="Try git, pdf, review…"
              />
              <button type="submit">Search</button>
              <button type="button" className="button-secondary" onClick={onReset}>
                Reset
              </button>
            </div>
          </form>

          {skillsState === 'loading' ? <p className="state-message">Loading skills…</p> : null}
          {skillsState === 'error' ? <p className="state-message error">Could not load skills: {skillsError}</p> : null}
          {skillsState === 'ready' && skills.length === 0 ? (
            <p className="state-message">No skills matched this query.</p>
          ) : null}

          <ul className="skill-list" aria-label="skills">
            {skills.map((skill) => (
              <li key={skill.name}>
                <button
                  type="button"
                  className={skill.name === selectedSkillName ? 'skill-list-item selected' : 'skill-list-item'}
                  onClick={() => setSelectedSkillName(skill.name)}
                >
                  <span className="skill-list-header">
                    <strong>{skill.name}</strong>
                    <span className={skill.valid ? 'pill valid' : 'pill invalid'}>{skill.valid ? 'valid' : 'invalid'}</span>
                  </span>
                  <span className="muted">{skill.description || 'No description available.'}</span>
                </button>
              </li>
            ))}
          </ul>
        </section>

        <section className="panel detail-panel">
          {detailState === 'idle' ? <p className="state-message">Select a skill to inspect its details.</p> : null}
          {detailState === 'loading' ? <p className="state-message">Loading skill details…</p> : null}
          {detailState === 'error' ? <p className="state-message error">Could not load skill details: {detailError}</p> : null}

          {detailState === 'ready' && selectedSkill ? (
            <article className="detail-card section-block">
              <div className="detail-header">
                <div>
                  <p className="eyebrow">Skill detail</p>
                  <h2>{selectedSkill.name}</h2>
                </div>
                <span className={selectedSkill.valid ? 'pill valid' : 'pill invalid'}>
                  {selectedSkill.valid ? 'valid' : 'invalid'}
                </span>
              </div>

              <dl className="meta-grid">
                <div>
                  <dt>Path</dt>
                  <dd>{selectedSkill.path}</dd>
                </div>
                <div>
                  <dt>Description</dt>
                  <dd>{selectedSkill.description || 'No description provided.'}</dd>
                </div>
                <div>
                  <dt>Tags</dt>
                  <dd>{selectedSkill.tags && selectedSkill.tags.length > 0 ? selectedSkill.tags.join(', ') : 'None'}</dd>
                </div>
              </dl>

              {selectedSkill.findings && selectedSkill.findings.length > 0 ? (
                <section>
                  <h3>Validation findings</h3>
                  <ul className="findings-list">
                    {selectedSkill.findings.map((finding) => (
                      <li key={`${finding.code}-${finding.path ?? finding.message}`}>
                        <strong>{finding.code}</strong>: {finding.message}
                        {finding.path ? <span className="muted"> ({finding.path})</span> : null}
                      </li>
                    ))}
                  </ul>
                </section>
              ) : null}

              {selectedSkill.body ? (
                <section>
                  <h3>Body preview</h3>
                  <pre className="body-preview">{selectedSkill.body}</pre>
                </section>
              ) : null}
            </article>
          ) : null}

          <article className="draft-card section-block">
            <div className="detail-header">
              <div>
                <p className="eyebrow">Draft authoring</p>
                <h2>Create a browser draft</h2>
              </div>
              {currentDraft ? (
                <span className={currentDraft.validation.valid ? 'pill valid' : 'pill invalid'}>
                  {currentDraft.validation.valid ? 'draft valid' : 'draft invalid'}
                </span>
              ) : null}
            </div>

            <p className="muted">
              This first write slice creates a draft workspace through the existing draft API, shows validation/submission status,
              and can submit the current draft when the backend allows it.
            </p>

            {selectedSkill ? (
              <div className="inline-actions">
                <button type="button" className="button-secondary" onClick={onPrefillFromSelection}>
                  Prefill update from selected skill
                </button>
                <button type="button" className="button-secondary" onClick={onPrepareDeleteFromSelection}>
                  Prepare delete for selected skill
                </button>
              </div>
            ) : null}

            <form className="draft-form" onSubmit={onCreateDraft}>
              <div className="draft-grid">
                <label>
                  Operation
                  <select value={draftOperation} onChange={(event) => setDraftOperation(event.target.value as DraftOperation)}>
                    <option value="create">create</option>
                    <option value="update">update</option>
                    <option value="delete">delete</option>
                  </select>
                </label>
                <label>
                  Skill name
                  <input
                    value={draftSkillName}
                    onChange={(event) => setDraftSkillName(event.target.value)}
                    placeholder="example-skill"
                  />
                </label>
              </div>

              {draftOperation !== 'delete' ? (
                <label>
                  Draft content
                  <textarea
                    value={draftContent}
                    onChange={(event) => setDraftContent(event.target.value)}
                    placeholder="---\nname: example-skill\ndescription: ...\n---\n# example-skill"
                    rows={12}
                  />
                </label>
              ) : (
                <p className="state-message">Delete drafts reuse the current skill path and do not require draft content.</p>
              )}

              <div className="inline-actions">
                <button type="submit" disabled={draftState === 'creating'}>
                  {createButtonLabel}
                </button>
                <button type="button" className="button-secondary" onClick={onRefreshDraft} disabled={!currentDraft || draftState === 'refreshing'}>
                  {draftState === 'refreshing' ? 'Refreshing…' : 'Refresh current draft'}
                </button>
              </div>
            </form>

            {draftState === 'error' ? <p className="state-message error">Could not create draft: {draftError}</p> : null}

            {currentDraft ? (
              <section className="draft-result" aria-label="current draft">
                <h3>Current draft</h3>
                <dl className="meta-grid">
                  <div>
                    <dt>Draft ID</dt>
                    <dd>{currentDraft.id}</dd>
                  </div>
                  <div>
                    <dt>Operation</dt>
                    <dd>{currentDraft.operation}</dd>
                  </div>
                  <div>
                    <dt>Skill</dt>
                    <dd>{currentDraft.skillName}</dd>
                  </div>
                  <div>
                    <dt>Branch</dt>
                    <dd>{currentDraft.branchName}</dd>
                  </div>
                  <div>
                    <dt>Created</dt>
                    <dd>{formatTimestamp(currentDraft.createdAt)}</dd>
                  </div>
                  <div>
                    <dt>Submission</dt>
                    <dd>{currentDraft.submission.enabled ? 'enabled' : 'disabled'}</dd>
                  </div>
                </dl>

                <section>
                  <h3>Validation status</h3>
                  <p className="state-message">{currentDraft.validation.valid ? 'Draft validation passed.' : 'Draft validation reported findings.'}</p>
                  {currentDraft.validation.findings && currentDraft.validation.findings.length > 0 ? (
                    <ul className="findings-list">
                      {currentDraft.validation.findings.map((finding) => (
                        <li key={`${finding.code}-${finding.path ?? finding.message}`}>
                          <strong>{finding.code}</strong>: {finding.message}
                          {finding.path ? <span className="muted"> ({finding.path})</span> : null}
                        </li>
                      ))}
                    </ul>
                  ) : null}
                </section>

                <section>
                  <h3>Submission status</h3>
                  <p className="state-message">
                    Submission is {currentDraft.submission.enabled ? 'enabled' : 'disabled'}.
                    {currentDraft.submission.baseBranch ? ` Base branch: ${currentDraft.submission.baseBranch}.` : ''}
                    {currentDraft.submission.reason ? ` ${currentDraft.submission.reason}` : ''}
                  </p>
                </section>

                <div className="inline-actions">
                  <button type="button" onClick={onSubmitDraft} disabled={!currentDraft.submission.enabled || submitState === 'submitting'}>
                    {submitButtonLabel}
                  </button>
                </div>
              </section>
            ) : null}

            {submitState === 'error' ? <p className="state-message error">Could not submit draft: {submitError}</p> : null}

            {submissionResult ? (
              <section className="submission-result" aria-label="submission result">
                <h3>Submission result</h3>
                <dl className="meta-grid">
                  <div>
                    <dt>Branch</dt>
                    <dd>{submissionResult.branchName}</dd>
                  </div>
                  <div>
                    <dt>Base branch</dt>
                    <dd>{submissionResult.baseBranch}</dd>
                  </div>
                  <div>
                    <dt>Commit</dt>
                    <dd>{submissionResult.commitHash ?? 'Not reported'}</dd>
                  </div>
                  <div>
                    <dt>Pull request</dt>
                    <dd>
                      {submissionResult.pullRequest?.url ? (
                        <a href={submissionResult.pullRequest.url} target="_blank" rel="noreferrer">
                          {submissionResult.pullRequest.number ? `#${submissionResult.pullRequest.number}` : submissionResult.pullRequest.url}
                        </a>
                      ) : submissionResult.pullRequest?.number ? (
                        `#${submissionResult.pullRequest.number}`
                      ) : (
                        'Not reported'
                      )}
                    </dd>
                  </div>
                </dl>
              </section>
            ) : null}
          </article>
        </section>
      </main>
    </div>
  )
}

function readLocationState(): UiLocationState {
  const params = new URLSearchParams(window.location.search)
  return {
    query: params.get('q')?.trim() ?? '',
    skill: params.get('skill')?.trim() ?? '',
  }
}

function syncLocationState(state: UiLocationState): void {
  const params = new URLSearchParams(window.location.search)

  if (state.query) {
    params.set('q', state.query)
  } else {
    params.delete('q')
  }

  if (state.skill) {
    params.set('skill', state.skill)
  } else {
    params.delete('skill')
  }

  const search = params.toString()
  const nextUrl = search === '' ? window.location.pathname : `${window.location.pathname}?${search}`
  window.history.replaceState({}, '', nextUrl)
}

function formatTimestamp(value: string): string {
  if (!value) {
    return 'Unknown'
  }

  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return value
  }

  return date.toISOString()
}

export default App
