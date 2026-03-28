import { FormEvent, useEffect, useMemo, useState } from 'react'
import './App.css'
import { getIndexStatus, getSkill, listSkills, searchSkills, type IndexStatus, type SkillRecord } from './api'

type LoadState = 'idle' | 'loading' | 'ready' | 'error'

function App() {
  const [query, setQuery] = useState('')
  const [submittedQuery, setSubmittedQuery] = useState('')
  const [skills, setSkills] = useState<SkillRecord[]>([])
  const [skillsState, setSkillsState] = useState<LoadState>('idle')
  const [skillsError, setSkillsError] = useState('')
  const [selectedSkillName, setSelectedSkillName] = useState('')
  const [selectedSkill, setSelectedSkill] = useState<SkillRecord | null>(null)
  const [detailState, setDetailState] = useState<LoadState>('idle')
  const [detailError, setDetailError] = useState('')
  const [indexStatus, setIndexStatus] = useState<IndexStatus | null>(null)

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

    const request = submittedQuery.trim() === '' ? listSkills() : searchSkills(submittedQuery)

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

  const onSearch = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    setSubmittedQuery(query.trim())
  }

  const onReset = () => {
    setQuery('')
    setSubmittedQuery('')
  }

  return (
    <div className="app-shell">
      <header className="topbar">
        <div>
          <p className="eyebrow">Skillforge</p>
          <h1>Browse skills</h1>
          <p className="subtitle">Read-only web UI for catalog discovery on top of the current Go API.</p>
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
            <article className="detail-card">
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
        </section>
      </main>
    </div>
  )
}

export default App
