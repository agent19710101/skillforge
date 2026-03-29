import { fireEvent, render, screen } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import App from './App'

beforeEach(() => {
  window.history.replaceState({}, '', '/')
})

afterEach(() => {
  vi.restoreAllMocks()
})

describe('App', () => {
  it('loads the skill list and selected skill detail', async () => {
    const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
      const url = parseUrl(input)
      if (url.pathname === '/api/v1/index/status') {
        return response({ ready: true, source: 'git', scannedAt: '2026-03-28T21:00:00Z', skillCount: 2 })
      }
      if (url.pathname === '/api/v1/skills') {
        expect(url.searchParams.get('limit')).toBe('200')
        return response({
          skills: [
            { name: 'git-pr-review', description: 'Review PRs', path: 'skills/git-pr-review/SKILL.md', valid: true },
            { name: 'pdf-search-helper', description: 'Find text in PDFs', path: 'skills/pdf-search-helper/SKILL.md', valid: true },
          ],
          total: 2,
          offset: 0,
          limit: 200,
        })
      }
      if (url.pathname === '/api/v1/skills/git-pr-review') {
        return response({
          name: 'git-pr-review',
          description: 'Review PRs',
          path: 'skills/git-pr-review/SKILL.md',
          tags: ['git', 'review'],
          body: 'Use gh pr review.',
          valid: true,
        })
      }
      throw new Error(`unexpected URL ${url}`)
    })
    vi.stubGlobal('fetch', fetchMock)

    render(<App />)

    expect(await screen.findByText('git-pr-review')).toBeInTheDocument()
    expect(await screen.findByText('Use gh pr review.')).toBeInTheDocument()
    expect(await screen.findByText('2 indexed skill(s)')).toBeInTheDocument()
    expect(await screen.findByText('Create a browser draft')).toBeInTheDocument()
  })

  it('hydrates search and selected skill from the URL', async () => {
    window.history.replaceState({}, '', '/?q=pdf&skill=pdf-search-helper')

    const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
      const url = parseUrl(input)
      if (url.pathname === '/api/v1/index/status') {
        return response({ ready: true, source: 'git', scannedAt: '2026-03-28T21:00:00Z', skillCount: 1 })
      }
      if (url.pathname === '/api/v1/search') {
        expect(url.searchParams.get('q')).toBe('pdf')
        return response({
          query: 'pdf',
          skills: [
            { name: 'pdf-search-helper', description: 'Find text in PDFs', path: 'skills/pdf-search-helper/SKILL.md', valid: true },
          ],
          total: 1,
        })
      }
      if (url.pathname === '/api/v1/skills/pdf-search-helper') {
        return response({
          name: 'pdf-search-helper',
          description: 'Find text in PDFs',
          path: 'skills/pdf-search-helper/SKILL.md',
          tags: ['pdf', 'search'],
          body: 'Use pdftotext first.',
          valid: true,
        })
      }
      throw new Error(`unexpected URL ${url}`)
    })
    vi.stubGlobal('fetch', fetchMock)

    render(<App />)

    expect(await screen.findByDisplayValue('pdf')).toBeInTheDocument()
    expect(await screen.findByText('Use pdftotext first.')).toBeInTheDocument()
  })

  it('loads every catalog page before resolving a deep-linked browse selection', async () => {
    window.history.replaceState({}, '', '/?skill=skill-201')

    const firstPageSkills = Array.from({ length: 200 }, (_, index) => ({
      name: `skill-${index + 1}`,
      description: `Skill ${index + 1}`,
      path: `skills/skill-${index + 1}/SKILL.md`,
      valid: true,
    }))

    const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
      const url = parseUrl(input)
      if (url.pathname === '/api/v1/index/status') {
        return response({ ready: true, source: 'git', scannedAt: '2026-03-28T21:00:00Z', skillCount: 201 })
      }
      if (url.pathname === '/api/v1/skills') {
        const offset = Number(url.searchParams.get('offset') ?? '0')
        const limit = Number(url.searchParams.get('limit') ?? '0')
        expect(limit).toBe(200)

        if (offset === 0) {
          return response({
            skills: firstPageSkills,
            total: 201,
            offset: 0,
            limit,
          })
        }

        if (offset === 200) {
          return response({
            skills: [
              {
                name: 'skill-201',
                description: 'Skill 201',
                path: 'skills/skill-201/SKILL.md',
                valid: true,
              },
            ],
            total: 201,
            offset: 200,
            limit,
          })
        }
      }
      if (url.pathname === '/api/v1/skills/skill-201') {
        return response({
          name: 'skill-201',
          description: 'Skill 201',
          path: 'skills/skill-201/SKILL.md',
          body: 'Loaded from the second browse page.',
          valid: true,
        })
      }
      if (url.pathname.startsWith('/api/v1/skills/skill-')) {
        const name = decodeURIComponent(url.pathname.replace('/api/v1/skills/', ''))
        return response({
          name,
          description: `${name} description`,
          path: `skills/${name}/SKILL.md`,
          valid: true,
        })
      }
      throw new Error(`unexpected URL ${url}`)
    })
    vi.stubGlobal('fetch', fetchMock)

    render(<App />)

    expect(await screen.findByText('Loaded from the second browse page.')).toBeInTheDocument()
    expect(fetchMock).toHaveBeenCalledWith(expect.stringContaining('/api/v1/skills?limit=200'), expect.any(Object))
    expect(fetchMock).toHaveBeenCalledWith(expect.stringContaining('/api/v1/skills?offset=200&limit=200'), expect.any(Object))
  })

  it('creates and submits a draft from the browser form', async () => {
    const fetchMock = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
      const url = parseUrl(input)
      if (url.pathname === '/api/v1/index/status') {
        return response({ ready: true, source: 'git', scannedAt: '2026-03-28T21:00:00Z', skillCount: 1 })
      }
      if (url.pathname === '/api/v1/skills') {
        return response({
          skills: [{ name: 'git-pr-review', description: 'Review PRs', path: 'skills/git-pr-review/SKILL.md', valid: true }],
          total: 1,
          offset: 0,
          limit: 200,
        })
      }
      if (url.pathname === '/api/v1/skills/git-pr-review') {
        return response({
          name: 'git-pr-review',
          description: 'Review PRs',
          path: 'skills/git-pr-review/SKILL.md',
          body: '---\nname: git-pr-review\ndescription: Review PRs\n---\n# git-pr-review\n',
          valid: true,
        })
      }
      if (url.pathname === '/api/v1/drafts') {
        expect(init?.method).toBe('POST')
        expect(parseJsonBody(init)).toEqual({
          operation: 'update',
          skillName: 'git-pr-review',
          content: '---\nname: git-pr-review\ndescription: Review PRs\n---\n# git-pr-review\n',
        })
        return response({
          id: 'draft01',
          operation: 'update',
          skillName: 'git-pr-review',
          branchName: 'skillforge/update/git-pr-review/draft01',
          createdAt: '2026-03-29T05:10:00Z',
          validation: { valid: true },
          submission: { enabled: true, baseBranch: 'main' },
        })
      }
      if (url.pathname === '/api/v1/drafts/draft01/submit') {
        expect(init?.method).toBe('POST')
        return response({
          id: 'draft01',
          operation: 'update',
          skillName: 'git-pr-review',
          branchName: 'skillforge/update/git-pr-review/draft01',
          baseBranch: 'main',
          commitHash: 'abc123',
          pullRequest: { number: 17, url: 'https://forgejo.example/pr/17' },
          validation: { valid: true },
        })
      }
      throw new Error(`unexpected URL ${url}`)
    })
    vi.stubGlobal('fetch', fetchMock)

    render(<App />)

    expect(await screen.findByText('Body preview')).toBeInTheDocument()

    fireEvent.click(screen.getByRole('button', { name: 'Prefill update from selected skill' }))
    fireEvent.click(screen.getByRole('button', { name: 'Create draft' }))

    expect(await screen.findByText('Current draft')).toBeInTheDocument()
    expect(await screen.findByText('draft01')).toBeInTheDocument()
    expect(await screen.findByText('Submission is enabled. Base branch: main.')).toBeInTheDocument()

    fireEvent.click(screen.getByRole('button', { name: 'Submit draft' }))

    expect(await screen.findByText('Submission result')).toBeInTheDocument()
    expect(await screen.findByText('abc123')).toBeInTheDocument()
    expect(screen.getByRole('link', { name: '#17' })).toHaveAttribute('href', 'https://forgejo.example/pr/17')
  })

  it('creates delete drafts without sending content', async () => {
    const fetchMock = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
      const url = parseUrl(input)
      if (url.pathname === '/api/v1/index/status') {
        return response({ ready: true, source: 'git', scannedAt: '2026-03-28T21:00:00Z', skillCount: 1 })
      }
      if (url.pathname === '/api/v1/skills') {
        return response({
          skills: [{ name: 'git-pr-review', description: 'Review PRs', path: 'skills/git-pr-review/SKILL.md', valid: true }],
          total: 1,
          offset: 0,
          limit: 200,
        })
      }
      if (url.pathname === '/api/v1/skills/git-pr-review') {
        return response({
          name: 'git-pr-review',
          description: 'Review PRs',
          path: 'skills/git-pr-review/SKILL.md',
          body: 'Use gh pr review.',
          valid: true,
        })
      }
      if (url.pathname === '/api/v1/drafts') {
        expect(parseJsonBody(init)).toEqual({
          operation: 'delete',
          skillName: 'git-pr-review',
        })
        return response({
          id: 'draft-delete-01',
          operation: 'delete',
          skillName: 'git-pr-review',
          branchName: 'skillforge/delete/git-pr-review/draft-delete-01',
          createdAt: '2026-03-29T05:20:00Z',
          validation: { valid: true },
          submission: { enabled: false, reason: 'submission backend is not configured' },
        })
      }
      throw new Error(`unexpected URL ${url}`)
    })
    vi.stubGlobal('fetch', fetchMock)

    render(<App />)

    expect(await screen.findByText('Use gh pr review.')).toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: 'Prepare delete for selected skill' }))

    expect(screen.queryByLabelText('Draft content')).not.toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: 'Create draft' }))

    expect(await screen.findByText('draft-delete-01')).toBeInTheDocument()
    expect(await screen.findByText(/Submission is disabled\. submission backend is not configured/)).toBeInTheDocument()
  })

  it('shows a draft creation error when the write request fails', async () => {
    const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
      const url = parseUrl(input)
      if (url.pathname === '/api/v1/index/status') {
        return response({ ready: true, source: 'git', scannedAt: '2026-03-28T21:00:00Z', skillCount: 0 })
      }
      if (url.pathname === '/api/v1/skills') {
        return response({ skills: [], total: 0, offset: 0, limit: 200 })
      }
      if (url.pathname === '/api/v1/drafts') {
        return response({ error: 'invalid_request', message: 'content is required for create' }, 400)
      }
      throw new Error(`unexpected URL ${url}`)
    })
    vi.stubGlobal('fetch', fetchMock)

    render(<App />)

    const skillInput = await screen.findByLabelText('Skill name')
    fireEvent.change(skillInput, { target: { value: 'new-skill' } })
    fireEvent.change(screen.getByLabelText('Draft content'), { target: { value: '' } })
    fireEvent.click(screen.getByRole('button', { name: 'Create draft' }))

    expect(await screen.findByText('Could not create draft: content is required for create')).toBeInTheDocument()
  })
})

function parseUrl(input: RequestInfo | URL): URL {
  return new URL(String(input), 'http://localhost')
}

function parseJsonBody(init?: RequestInit): unknown {
  const body = init?.body
  if (typeof body !== 'string') {
    throw new Error(`expected JSON string body, got ${typeof body}`)
  }
  return JSON.parse(body)
}

function response(payload: unknown, status = 200): Response {
  return new Response(JSON.stringify(payload), {
    status,
    headers: {
      'Content-Type': 'application/json',
    },
  })
}
