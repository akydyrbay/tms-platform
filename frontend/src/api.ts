// Thin typed wrapper over the TMS REST API. The bearer token is kept in
// localStorage so a refresh keeps you logged in.

const BASE = '/api/v1'
const TOKEN_KEY = 'tms_token'

let token = localStorage.getItem(TOKEN_KEY) ?? ''

export function getToken() {
  return token
}
export function setToken(t: string) {
  token = t
  localStorage.setItem(TOKEN_KEY, t)
}
export function clearToken() {
  token = ''
  localStorage.removeItem(TOKEN_KEY)
}

async function req<T>(method: string, path: string, body?: unknown): Promise<T> {
  const headers: Record<string, string> = { 'Content-Type': 'application/json' }
  if (token) headers.Authorization = `Bearer ${token}`

  const res = await fetch(BASE + path, {
    method,
    headers,
    body: body === undefined ? undefined : JSON.stringify(body),
  })

  const text = await res.text()
  const data = text ? JSON.parse(text) : undefined
  if (!res.ok) {
    throw new Error(data?.error ?? `${res.status} ${res.statusText}`)
  }
  return data as T
}

// --- types ---

export interface Project {
  id: string
  name: string
  description?: string
  created_at: string
}

export interface Folder {
  id: string
  project_id: string
  parent_id?: string
  name: string
  children: Folder[]
}

export interface Step {
  step_number?: number
  action: string
  expected_result?: string
}

export interface CaseVersion {
  id: string
  test_case_id: string
  version_number: number
  title: string
  description?: string
  preconditions?: string
  expected_result?: string
  module?: string
  component?: string
  priority: string
  type: string
  status: string
  created_at: string
  steps: Step[]
}

export interface CaseSummary {
  test_case_id: string
  folder_id?: string
  version_number: number
  title: string
  priority: string
  status: string
}

export interface VersionSummary {
  version_number: number
  created_by: string
  created_at: string
}

export interface Suite {
  id: string
  project_id: string
  name: string
  description?: string
  case_ids?: string[]
  created_at: string
}

export interface Run {
  id: string
  project_id: string
  suite_id: string
  name: string
  status: string
  started_at?: string
  completed_at?: string
  created_at: string
}

export interface RunResult {
  id: string
  test_run_id: string
  test_case_id: string
  test_case_version_id: string
  status: string
  comment?: string
  executed_by?: string
  executed_at?: string
}

export interface Bug {
  id: string
  test_run_result_id: string
  tracker: string
  external_id: string
  url?: string
  title?: string
  created_at: string
}

export interface RunStats {
  total: number
  passed: number
  failed: number
  skipped: number
  blocked: number
  not_run: number
}

export interface CaseContent {
  title: string
  description?: string
  preconditions?: string
  expected_result?: string
  module?: string
  component?: string
  priority?: string
  type?: string
  status?: string
  steps: Step[]
}

// --- endpoints ---

export const api = {
  register: (name: string, email: string, password: string, role: string) =>
    req<unknown>('POST', '/auth/register', { name, email, password, role }),
  login: (email: string, password: string) =>
    req<{ access_token: string }>('POST', '/auth/login', { email, password }),

  listProjects: () => req<Project[]>('GET', '/projects'),
  createProject: (name: string, description?: string) =>
    req<Project>('POST', '/projects', { name, description }),

  listFolders: (projectId: string) =>
    req<Folder[]>('GET', `/folders?project_id=${projectId}`),
  createFolder: (projectId: string, name: string, parentId?: string | null) =>
    req<Folder>('POST', '/folders', {
      project_id: projectId,
      name,
      parent_id: parentId ?? null,
    }),

  listCases: (projectId: string) =>
    req<CaseSummary[]>('GET', `/test-cases?project_id=${projectId}`),
  getCase: (id: string) => req<CaseVersion>('GET', `/test-cases/${id}`),
  createCase: (projectId: string, folderId: string | null, content: CaseContent) =>
    req<CaseVersion>('POST', '/test-cases', {
      project_id: projectId,
      folder_id: folderId,
      ...content,
    }),
  editCase: (id: string, content: CaseContent) =>
    req<CaseVersion>('PUT', `/test-cases/${id}`, content),
  listVersions: (id: string) =>
    req<VersionSummary[]>('GET', `/test-cases/${id}/versions`),
  getVersion: (id: string, n: number) =>
    req<CaseVersion>('GET', `/test-cases/${id}/versions/${n}`),

  listSuites: (projectId: string) =>
    req<Suite[]>('GET', `/suites?project_id=${projectId}`),
  getSuite: (id: string) => req<Suite>('GET', `/suites/${id}`),
  createSuite: (projectId: string, name: string, description?: string) =>
    req<Suite>('POST', '/suites', { project_id: projectId, name, description }),
  addCaseToSuite: (suiteId: string, caseId: string) =>
    req<Suite>('POST', `/suites/${suiteId}/cases`, { test_case_id: caseId }),
  removeCaseFromSuite: (suiteId: string, caseId: string) =>
    req<Suite>('DELETE', `/suites/${suiteId}/cases/${caseId}`),

  listRuns: (projectId: string) =>
    req<Run[]>('GET', `/runs?project_id=${projectId}`),
  getRun: (id: string) => req<Run>('GET', `/runs/${id}`),
  createRun: (suiteId: string, name: string) =>
    req<Run>('POST', '/runs', { suite_id: suiteId, name }),
  listResults: (runId: string) =>
    req<RunResult[]>('GET', `/runs/${runId}/results`),
  markResult: (runId: string, caseId: string, status: string, comment?: string) =>
    req<RunResult>('POST', `/runs/${runId}/results/${caseId}`, { status, comment }),
  setRunStatus: (runId: string, status: string) =>
    req<Run>('PATCH', `/runs/${runId}`, { status }),
  getStats: (runId: string) => req<RunStats>('GET', `/runs/${runId}/stats`),

  listBugs: (resultId: string) => req<Bug[]>('GET', `/run-results/${resultId}/bugs`),
  attachBug: (resultId: string, tracker: string, externalId: string, url?: string) =>
    req<Bug>('POST', `/run-results/${resultId}/bugs`, {
      tracker,
      external_id: externalId,
      url: url || null,
    }),
}
