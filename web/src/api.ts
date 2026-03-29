export interface ValidationFinding {
  code: string
  message: string
  path?: string
}

export interface SkillRecord {
  name: string
  description?: string
  path: string
  valid: boolean
  tags?: string[]
  body?: string
  findings?: ValidationFinding[]
}

export interface ListSkillsResponse {
  skills: SkillRecord[]
  total: number
  offset: number
  limit: number
}

export interface SearchSkillsResponse {
  query: string
  skills: SkillRecord[]
  total: number
}

export interface IndexStatus {
  ready: boolean
  source: string
  scannedAt: string
  skillCount: number
  git?: {
    commit?: string
    branch?: string
    dirty?: boolean
  }
}

export interface ApiErrorPayload {
  error?: string
  message?: string
}

const configuredBase = (import.meta.env.VITE_API_BASE_URL ?? '').trim()
const catalogPageSize = 200

function normalizedBase(): string {
  if (configuredBase === '') {
    return ''
  }
  return configuredBase.replace(/\/+$/, '')
}

function buildUrl(path: string, query?: URLSearchParams): string {
  const normalizedPath = path.startsWith('/') ? path : `/${path}`
  const base = normalizedBase()
  const url = `${base}${normalizedPath}`
  return query && query.size > 0 ? `${url}?${query.toString()}` : url
}

async function getJson<T>(path: string, query?: URLSearchParams): Promise<T> {
  const response = await fetch(buildUrl(path, query), {
    headers: {
      Accept: 'application/json',
    },
  })

  if (!response.ok) {
    let payload: ApiErrorPayload | undefined
    try {
      payload = (await response.json()) as ApiErrorPayload
    } catch {
      payload = undefined
    }
    throw new Error(payload?.message ?? `request failed with status ${response.status}`)
  }

  return (await response.json()) as T
}

export function listSkills(offset = 0, limit = catalogPageSize): Promise<ListSkillsResponse> {
  const query = new URLSearchParams()
  if (offset > 0) {
    query.set('offset', String(offset))
  }
  query.set('limit', String(limit))
  return getJson<ListSkillsResponse>('/api/v1/skills', query)
}

export async function listAllSkills(): Promise<ListSkillsResponse> {
  const skills: SkillRecord[] = []
  let offset = 0
  let total = 0
  let pageLimit = catalogPageSize

  while (true) {
    const page = await listSkills(offset, catalogPageSize)
    skills.push(...page.skills)
    total = page.total
    pageLimit = page.limit

    if (skills.length >= total || page.skills.length === 0) {
      break
    }

    offset = page.offset + page.skills.length
  }

  return {
    skills,
    total,
    offset: 0,
    limit: pageLimit,
  }
}

export function searchSkills(queryText: string): Promise<SearchSkillsResponse> {
  const query = new URLSearchParams({ q: queryText })
  return getJson<SearchSkillsResponse>('/api/v1/search', query)
}

export function getSkill(name: string): Promise<SkillRecord> {
  return getJson<SkillRecord>(`/api/v1/skills/${encodeURIComponent(name)}`)
}

export function getIndexStatus(): Promise<IndexStatus> {
  return getJson<IndexStatus>('/api/v1/index/status')
}
