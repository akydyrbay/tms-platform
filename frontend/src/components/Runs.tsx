import { useEffect, useState } from 'react'
import {
  api,
  type Bug,
  type Run,
  type RunResult,
  type RunStats,
  type Suite,
} from '../api'

const RESULT_STATUSES = ['passed', 'failed', 'skipped', 'blocked']
const RUN_STATUSES = ['pending', 'in_progress', 'completed', 'cancelled']

// BugCell shows external bug-tracker references attached to a result and lets
// you link a new one (Integration module).
function BugCell({ resultId }: { resultId: string }) {
  const [bugs, setBugs] = useState<Bug[]>([])
  const [open, setOpen] = useState(false)
  const [tracker, setTracker] = useState('jira')
  const [externalId, setExternalId] = useState('')
  const [url, setUrl] = useState('')

  useEffect(() => {
    api.listBugs(resultId).then(setBugs).catch(() => undefined)
  }, [resultId])

  async function add() {
    if (!externalId.trim()) return
    const b = await api.attachBug(resultId, tracker, externalId, url || undefined)
    setBugs([...bugs, b])
    setExternalId('')
    setUrl('')
    setOpen(false)
  }

  return (
    <div className="col">
      {bugs.map((b) =>
        b.url ? (
          <a key={b.id} href={b.url} target="_blank" rel="noreferrer">{b.tracker}:{b.external_id}</a>
        ) : (
          <span key={b.id} className="muted">{b.tracker}:{b.external_id}</span>
        ),
      )}
      {open ? (
        <div className="row">
          <input style={{ width: 70 }} value={tracker} onChange={(e) => setTracker(e.target.value)} />
          <input style={{ width: 90 }} placeholder="ID" value={externalId} onChange={(e) => setExternalId(e.target.value)} />
          <input placeholder="url" value={url} onChange={(e) => setUrl(e.target.value)} />
          <button className="link" onClick={add}>save</button>
        </div>
      ) : (
        <button className="link" onClick={() => setOpen(true)}>+ link bug</button>
      )}
    </div>
  )
}

export default function Runs({ projectId }: { projectId: string }) {
  const [runs, setRuns] = useState<Run[]>([])
  const [suites, setSuites] = useState<Suite[]>([])
  const [name, setName] = useState('')
  const [suiteId, setSuiteId] = useState('')
  const [selected, setSelected] = useState<Run | null>(null)
  const [results, setResults] = useState<RunResult[]>([])
  const [stats, setStats] = useState<RunStats | null>(null)
  const [comments, setComments] = useState<Record<string, string>>({})
  const [error, setError] = useState('')

  async function load() {
    try {
      setRuns(await api.listRuns(projectId))
      setSuites(await api.listSuites(projectId))
    } catch (err) {
      setError((err as Error).message)
    }
  }

  useEffect(() => {
    load()
    setSelected(null)
  }, [projectId])

  async function create(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    try {
      await api.createRun(suiteId, name)
      setName('')
      setSuiteId('')
      await load()
    } catch (err) {
      setError((err as Error).message)
    }
  }

  async function open(run: Run) {
    setSelected(run)
    await refresh(run.id)
  }

  async function refresh(runId: string) {
    setResults(await api.listResults(runId))
    setStats(await api.getStats(runId))
  }

  async function mark(caseId: string, status: string) {
    if (!selected) return
    try {
      await api.markResult(selected.id, caseId, status, comments[caseId] || undefined)
      await refresh(selected.id)
    } catch (err) {
      setError((err as Error).message)
    }
  }

  async function changeStatus(status: string) {
    if (!selected) return
    const updated = await api.setRunStatus(selected.id, status)
    setSelected(updated)
    await load()
  }

  return (
    <div className="panel two-col">
      <div>
        <h2>Runs</h2>
        <form className="col" onSubmit={create}>
          <select value={suiteId} onChange={(e) => setSuiteId(e.target.value)} required>
            <option value="">Suite to run…</option>
            {suites.map((s) => (
              <option key={s.id} value={s.id}>{s.name}</option>
            ))}
          </select>
          <div className="row">
            <input placeholder="Run name" value={name} onChange={(e) => setName(e.target.value)} required />
            <button type="submit">Create run</button>
          </div>
        </form>
        {error && <div className="error">{error}</div>}
        <ul className="list">
          {runs.map((r) => (
            <li key={r.id} className={r.id === selected?.id ? 'item selected' : 'item'} onClick={() => open(r)}>
              {r.name} <span className="badge">{r.status}</span>
            </li>
          ))}
          {runs.length === 0 && <li className="muted">No runs yet.</li>}
        </ul>
      </div>

      <div>
        {selected && (
          <>
            <div className="row between">
              <h3>{selected.name}</h3>
              <label>
                Status
                <select value={selected.status} onChange={(e) => changeStatus(e.target.value)}>
                  {RUN_STATUSES.map((s) => <option key={s}>{s}</option>)}
                </select>
              </label>
            </div>

            {stats && (
              <div className="stats">
                <span>Total: <strong>{stats.total}</strong></span>
                <span className="ok">Passed: {stats.passed}</span>
                <span className="bad">Failed: {stats.failed}</span>
                <span>Skipped: {stats.skipped}</span>
                <span>Blocked: {stats.blocked}</span>
                <span>Not run: {stats.not_run}</span>
              </div>
            )}

            <table className="results">
              <thead>
                <tr><th>Case</th><th>Status</th><th>Frozen version</th><th>Comment</th><th>Mark</th><th>Bugs</th></tr>
              </thead>
              <tbody>
                {results.map((res) => (
                  <tr key={res.id}>
                    <td>{res.title}</td>
                    <td><span className={`badge ${res.status}`}>{res.status}</span></td>
                    <td className="muted mono">v{res.version_number} · {res.test_case_version_id.slice(0, 8)}</td>
                    <td>
                      <input
                        placeholder="comment"
                        value={comments[res.test_case_id] ?? res.comment ?? ''}
                        onChange={(e) => setComments({ ...comments, [res.test_case_id]: e.target.value })}
                      />
                    </td>
                    <td className="row">
                      {RESULT_STATUSES.map((s) => (
                        <button key={s} className="link" onClick={() => mark(res.test_case_id, s)}>{s}</button>
                      ))}
                    </td>
                    <td><BugCell resultId={res.id} /></td>
                  </tr>
                ))}
              </tbody>
            </table>
            <p className="muted">The “frozen version” stays fixed even if the case is edited after the run is created.</p>
          </>
        )}
        {!selected && <p className="muted">Select a run to execute it.</p>}
      </div>
    </div>
  )
}
