import { fireEvent, render, screen, waitFor } from '@testing-library/react'
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
      const url = String(input)
      if (url.endsWith('/api/v1/index/status')) {
        return response({ ready: true, source: 'git', scannedAt: '2026-03-28T21:00:00Z', skillCount: 2 })
      }
      if (url.endsWith('/api/v1/skills')) {
        return response({
          skills: [
            { name: 'git-pr-review', description: 'Review PRs', path: 'skills/git-pr-review/SKILL.md', valid: true },
            { name: 'pdf-search-helper', description: 'Find text in PDFs', path: 'skills/pdf-search-helper/SKILL.md', valid: true },
          ],
          total: 2,
          offset: 0,
          limit: 50,
        })
      }
      if (url.endsWith('/api/v1/skills/git-pr-review')) {
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
  })

  it('hydrates search and selected skill from the URL', async () => {
    window.history.replaceState({}, '', '/?q=pdf&skill=pdf-search-helper')

    const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input)
      if (url.endsWith('/api/v1/index/status')) {
        return response({ ready: true, source: 'git', scannedAt: '2026-03-28T21:00:00Z', skillCount: 1 })
      }
      if (url.includes('/api/v1/search?q=pdf')) {
        return response({
          query: 'pdf',
          skills: [
            { name: 'pdf-search-helper', description: 'Find text in PDFs', path: 'skills/pdf-search-helper/SKILL.md', valid: true },
          ],
          total: 1,
        })
      }
      if (url.endsWith('/api/v1/skills/pdf-search-helper')) {
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

  it('updates the URL when the selected skill changes', async () => {
    const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input)
      if (url.endsWith('/api/v1/index/status')) {
        return response({ ready: true, source: 'git', scannedAt: '2026-03-28T21:00:00Z', skillCount: 2 })
      }
      if (url.endsWith('/api/v1/skills')) {
        return response({
          skills: [
            { name: 'git-pr-review', description: 'Review PRs', path: 'skills/git-pr-review/SKILL.md', valid: true },
            { name: 'pdf-search-helper', description: 'Find text in PDFs', path: 'skills/pdf-search-helper/SKILL.md', valid: true },
          ],
          total: 2,
          offset: 0,
          limit: 50,
        })
      }
      if (url.endsWith('/api/v1/skills/git-pr-review')) {
        return response({
          name: 'git-pr-review',
          description: 'Review PRs',
          path: 'skills/git-pr-review/SKILL.md',
          valid: true,
        })
      }
      if (url.endsWith('/api/v1/skills/pdf-search-helper')) {
        return response({
          name: 'pdf-search-helper',
          description: 'Find text in PDFs',
          path: 'skills/pdf-search-helper/SKILL.md',
          body: 'Use pdftotext first.',
          valid: true,
        })
      }
      throw new Error(`unexpected URL ${url}`)
    })
    vi.stubGlobal('fetch', fetchMock)

    render(<App />)

    expect(await screen.findByText('git-pr-review')).toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: /pdf-search-helper/i }))

    await waitFor(() => {
      const params = new URLSearchParams(window.location.search)
      expect(params.get('skill')).toBe('pdf-search-helper')
    })
  })

  it('shows an empty search state and syncs the query into the URL', async () => {
    const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input)
      if (url.endsWith('/api/v1/index/status')) {
        return response({ ready: true, source: 'git', scannedAt: '2026-03-28T21:00:00Z', skillCount: 0 })
      }
      if (url.includes('/api/v1/search?q=nomatch')) {
        return response({ query: 'nomatch', skills: [], total: 0 })
      }
      if (url.endsWith('/api/v1/skills')) {
        return response({ skills: [], total: 0, offset: 0, limit: 50 })
      }
      throw new Error(`unexpected URL ${url}`)
    })
    vi.stubGlobal('fetch', fetchMock)

    render(<App />)

    const input = await screen.findByLabelText('Search skills')
    fireEvent.change(input, { target: { value: 'nomatch' } })
    fireEvent.click(screen.getByText('Search'))

    expect(await screen.findByText('No skills matched this query.')).toBeInTheDocument()
    expect(new URLSearchParams(window.location.search).get('q')).toBe('nomatch')
  })

  it('shows a loading error when the catalog request fails', async () => {
    const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input)
      if (url.endsWith('/api/v1/index/status')) {
        return response({ ready: true, source: 'git', scannedAt: '2026-03-28T21:00:00Z', skillCount: 0 })
      }
      if (url.endsWith('/api/v1/skills')) {
        return response({ error: 'unavailable', message: 'catalog unavailable' }, 503)
      }
      throw new Error(`unexpected URL ${url}`)
    })
    vi.stubGlobal('fetch', fetchMock)

    render(<App />)

    expect(await screen.findByText('Could not load skills: catalog unavailable')).toBeInTheDocument()
  })
})

function response(payload: unknown, status = 200): Response {
  return new Response(JSON.stringify(payload), {
    status,
    headers: {
      'Content-Type': 'application/json',
    },
  })
}
