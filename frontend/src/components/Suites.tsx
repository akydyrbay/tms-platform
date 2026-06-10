import { useEffect, useState } from 'react'
import { api, type CaseSummary, type Suite } from '../api'

export default function Suites({ projectId }: { projectId: string }) {
  const [suites, setSuites] = useState<Suite[]>([])
  const [cases, setCases] = useState<CaseSummary[]>([])
  const [name, setName] = useState('')
  const [selected, setSelected] = useState<Suite | null>(null)
  const [addCaseId, setAddCaseId] = useState('')
  const [error, setError] = useState('')

  async function load() {
    try {
      setSuites(await api.listSuites(projectId))
      setCases(await api.listCases(projectId))
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
      await api.createSuite(projectId, name)
      setName('')
      await load()
    } catch (err) {
      setError((err as Error).message)
    }
  }

  async function open(id: string) {
    setSelected(await api.getSuite(id))
  }

  const titleOf = (id: string) => cases.find((c) => c.test_case_id === id)?.title ?? id

  return (
    <div className="panel two-col">
      <div>
        <h2>Suites</h2>
        <form className="row" onSubmit={create}>
          <input placeholder="Suite name" value={name} onChange={(e) => setName(e.target.value)} required />
          <button type="submit">Create</button>
        </form>
        {error && <div className="error">{error}</div>}
        <ul className="list">
          {suites.map((s) => (
            <li key={s.id} className={s.id === selected?.id ? 'item selected' : 'item'} onClick={() => open(s.id)}>
              {s.name}
            </li>
          ))}
          {suites.length === 0 && <li className="muted">No suites yet.</li>}
        </ul>
      </div>

      <div>
        {selected && (
          <>
            <h3>{selected.name}</h3>
            <div className="row">
              <select value={addCaseId} onChange={(e) => setAddCaseId(e.target.value)}>
                <option value="">Select a case…</option>
                {cases.map((c) => (
                  <option key={c.test_case_id} value={c.test_case_id}>{c.title}</option>
                ))}
              </select>
              <button
                disabled={!addCaseId}
                onClick={async () => {
                  setSelected(await api.addCaseToSuite(selected.id, addCaseId))
                  setAddCaseId('')
                }}
              >
                Add case
              </button>
            </div>
            <ul className="list">
              {(selected.case_ids ?? []).map((id) => (
                <li key={id} className="item">
                  {titleOf(id)}
                  <button
                    className="link"
                    onClick={async () => setSelected(await api.removeCaseFromSuite(selected.id, id))}
                  >
                    remove
                  </button>
                </li>
              ))}
              {(selected.case_ids ?? []).length === 0 && <li className="muted">No cases in this suite.</li>}
            </ul>
          </>
        )}
        {!selected && <p className="muted">Select a suite to manage its cases.</p>}
      </div>
    </div>
  )
}
