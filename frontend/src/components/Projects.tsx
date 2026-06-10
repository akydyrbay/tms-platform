import { useEffect, useState } from 'react'
import { api, type Project } from '../api'

export default function Projects({
  selectedId,
  onSelect,
}: {
  selectedId?: string
  onSelect: (p: Project) => void
}) {
  const [projects, setProjects] = useState<Project[]>([])
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [error, setError] = useState('')

  async function load() {
    try {
      setProjects(await api.listProjects())
    } catch (err) {
      setError((err as Error).message)
    }
  }

  useEffect(() => {
    load()
  }, [])

  async function create(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    try {
      const p = await api.createProject(name, description || undefined)
      setName('')
      setDescription('')
      await load()
      onSelect(p)
    } catch (err) {
      setError((err as Error).message)
    }
  }

  return (
    <div className="panel">
      <h2>Projects</h2>
      <form className="row" onSubmit={create}>
        <input placeholder="Name" value={name} onChange={(e) => setName(e.target.value)} required />
        <input
          placeholder="Description"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
        />
        <button type="submit">Create</button>
      </form>
      {error && <div className="error">{error}</div>}
      <ul className="list">
        {projects.map((p) => (
          <li
            key={p.id}
            className={p.id === selectedId ? 'item selected' : 'item'}
            onClick={() => onSelect(p)}
          >
            <strong>{p.name}</strong>
            {p.description && <span className="muted"> — {p.description}</span>}
          </li>
        ))}
        {projects.length === 0 && <li className="muted">No projects yet.</li>}
      </ul>
    </div>
  )
}
