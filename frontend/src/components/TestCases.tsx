import { useEffect, useState } from 'react'
import {
  api,
  type CaseContent,
  type CaseSummary,
  type CaseVersion,
  type Folder,
  type Step,
  type VersionSummary,
} from '../api'

const PRIORITIES = ['low', 'medium', 'high']
const TYPES = ['functional', 'smoke', 'regression', 'integration', 'e2e']
const STATUSES = ['draft', 'active', 'deprecated']

function flatten(nodes: Folder[], depth = 0): { id: string; label: string }[] {
  return nodes.flatMap((f) => [
    { id: f.id, label: `${' '.repeat(depth * 2)}${f.name}` },
    ...flatten(f.children ?? [], depth + 1),
  ])
}

function emptyContent(): CaseContent {
  return {
    title: '',
    priority: 'medium',
    type: 'functional',
    status: 'active',
    steps: [{ action: '' }],
  }
}

function CaseForm({
  initial,
  folders,
  showFolder,
  submitLabel,
  onSubmit,
}: {
  initial: CaseContent
  folders?: { id: string; label: string }[]
  showFolder?: boolean
  submitLabel: string
  onSubmit: (content: CaseContent, folderId: string | null) => Promise<void>
}) {
  const [c, setC] = useState<CaseContent>(initial)
  const [folderId, setFolderId] = useState('')
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)

  function set<K extends keyof CaseContent>(k: K, v: CaseContent[K]) {
    setC({ ...c, [k]: v })
  }
  function setStep(i: number, patch: Partial<Step>) {
    const steps = c.steps.map((s, idx) => (idx === i ? { ...s, ...patch } : s))
    set('steps', steps)
  }

  async function submit(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    setBusy(true)
    try {
      const steps = c.steps.filter((s) => s.action.trim() !== '')
      await onSubmit({ ...c, steps }, folderId || null)
    } catch (err) {
      setError((err as Error).message)
    } finally {
      setBusy(false)
    }
  }

  return (
    <form className="caseform" onSubmit={submit}>
      <input placeholder="Title *" value={c.title} onChange={(e) => set('title', e.target.value)} required />
      <textarea placeholder="Description" value={c.description ?? ''} onChange={(e) => set('description', e.target.value)} />
      <textarea placeholder="Preconditions" value={c.preconditions ?? ''} onChange={(e) => set('preconditions', e.target.value)} />
      <textarea placeholder="Expected result" value={c.expected_result ?? ''} onChange={(e) => set('expected_result', e.target.value)} />
      <div className="row">
        <input placeholder="Module" value={c.module ?? ''} onChange={(e) => set('module', e.target.value)} />
        <input placeholder="Component" value={c.component ?? ''} onChange={(e) => set('component', e.target.value)} />
      </div>
      <div className="row">
        <label>Priority<select value={c.priority} onChange={(e) => set('priority', e.target.value)}>{PRIORITIES.map((p) => <option key={p}>{p}</option>)}</select></label>
        <label>Type<select value={c.type} onChange={(e) => set('type', e.target.value)}>{TYPES.map((p) => <option key={p}>{p}</option>)}</select></label>
        <label>Status<select value={c.status} onChange={(e) => set('status', e.target.value)}>{STATUSES.map((p) => <option key={p}>{p}</option>)}</select></label>
      </div>
      {showFolder && (
        <label>
          Folder
          <select value={folderId} onChange={(e) => setFolderId(e.target.value)}>
            <option value="">(none)</option>
            {folders?.map((o) => (
              <option key={o.id} value={o.id}>{o.label}</option>
            ))}
          </select>
        </label>
      )}

      <h4>Steps</h4>
      {c.steps.map((s, i) => (
        <div className="row step" key={i}>
          <span className="muted">{i + 1}.</span>
          <input placeholder="Action" value={s.action} onChange={(e) => setStep(i, { action: e.target.value })} />
          <input placeholder="Expected" value={s.expected_result ?? ''} onChange={(e) => setStep(i, { expected_result: e.target.value })} />
          <button type="button" className="link" onClick={() => set('steps', c.steps.filter((_, idx) => idx !== i))}>✕</button>
        </div>
      ))}
      <button type="button" className="link" onClick={() => set('steps', [...c.steps, { action: '' }])}>+ add step</button>

      {error && <div className="error">{error}</div>}
      <button type="submit" disabled={busy}>{submitLabel}</button>
    </form>
  )
}

function VersionView({ v }: { v: CaseVersion }) {
  return (
    <div className="version">
      <p><strong>v{v.version_number}</strong> — {v.title} <span className="badge">{v.priority}</span> <span className="badge">{v.status}</span></p>
      {v.description && <p className="muted">{v.description}</p>}
      {v.preconditions && <p><em>Preconditions:</em> {v.preconditions}</p>}
      <ol>
        {v.steps.map((s) => (
          <li key={s.step_number}>{s.action}{s.expected_result ? ` → ${s.expected_result}` : ''}</li>
        ))}
      </ol>
      {v.expected_result && <p><em>Expected result:</em> {v.expected_result}</p>}
    </div>
  )
}

export default function TestCases({ projectId }: { projectId: string }) {
  const [cases, setCases] = useState<CaseSummary[]>([])
  const [folders, setFolders] = useState<{ id: string; label: string }[]>([])
  const [selected, setSelected] = useState<string | null>(null)
  const [current, setCurrent] = useState<CaseVersion | null>(null)
  const [history, setHistory] = useState<VersionSummary[]>([])
  const [viewed, setViewed] = useState<CaseVersion | null>(null)
  const [editing, setEditing] = useState(false)
  const [creating, setCreating] = useState(false)
  const [error, setError] = useState('')

  async function loadList() {
    try {
      setCases(await api.listCases(projectId))
      setFolders(flatten(await api.listFolders(projectId)))
    } catch (err) {
      setError((err as Error).message)
    }
  }

  useEffect(() => {
    loadList()
    setSelected(null)
    setCurrent(null)
  }, [projectId])

  async function open(id: string) {
    setSelected(id)
    setEditing(false)
    setViewed(null)
    try {
      setCurrent(await api.getCase(id))
      setHistory(await api.listVersions(id))
    } catch (err) {
      setError((err as Error).message)
    }
  }

  async function viewVersion(n: number) {
    if (!selected) return
    setViewed(await api.getVersion(selected, n))
  }

  return (
    <div className="panel two-col">
      <div>
        <div className="row between">
          <h2>Test Cases</h2>
          <button onClick={() => { setCreating(true); setSelected(null); setCurrent(null) }}>+ New case</button>
        </div>
        {error && <div className="error">{error}</div>}
        <ul className="list">
          {cases.map((c) => (
            <li key={c.test_case_id} className={c.test_case_id === selected ? 'item selected' : 'item'} onClick={() => { setCreating(false); open(c.test_case_id) }}>
              {c.title} <span className="badge">v{c.version_number}</span> <span className="badge">{c.priority}</span>
            </li>
          ))}
          {cases.length === 0 && <li className="muted">No test cases yet.</li>}
        </ul>
      </div>

      <div>
        {creating && (
          <>
            <h3>Create test case</h3>
            <CaseForm
              initial={emptyContent()}
              folders={folders}
              showFolder
              submitLabel="Create"
              onSubmit={async (content, folderId) => {
                await api.createCase(projectId, folderId, content)
                setCreating(false)
                await loadList()
              }}
            />
          </>
        )}

        {current && !editing && !creating && (
          <>
            <div className="row between">
              <h3>Current version</h3>
              <button onClick={() => setEditing(true)}>Edit → new version</button>
            </div>
            <VersionView v={current} />

            <h4>History</h4>
            <ul className="list">
              {history.map((h) => (
                <li key={h.version_number} className="item" onClick={() => viewVersion(h.version_number)}>
                  v{h.version_number} — {new Date(h.created_at).toLocaleString()}
                </li>
              ))}
            </ul>
            {viewed && (
              <div className="card readonly">
                <p className="muted">Read-only view of an older version:</p>
                <VersionView v={viewed} />
              </div>
            )}
          </>
        )}

        {current && editing && (
          <>
            <h3>Edit (creates a new version)</h3>
            <CaseForm
              initial={{
                title: current.title,
                description: current.description,
                preconditions: current.preconditions,
                expected_result: current.expected_result,
                module: current.module,
                component: current.component,
                priority: current.priority,
                type: current.type,
                status: current.status,
                steps: current.steps.length ? current.steps.map((s) => ({ action: s.action, expected_result: s.expected_result })) : [{ action: '' }],
              }}
              submitLabel="Save as new version"
              onSubmit={async (content) => {
                await api.editCase(current.test_case_id, content)
                await open(current.test_case_id)
                await loadList()
              }}
            />
          </>
        )}
      </div>
    </div>
  )
}
