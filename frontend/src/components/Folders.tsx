import { useEffect, useState } from 'react'
import { api, type Folder } from '../api'

function Tree({ nodes, depth = 0 }: { nodes: Folder[]; depth?: number }) {
  return (
    <ul className="tree">
      {nodes.map((f) => (
        <li key={f.id}>
          <span style={{ paddingLeft: depth * 16 }}>📁 {f.name}</span>
          {f.children?.length > 0 && <Tree nodes={f.children} depth={depth + 1} />}
        </li>
      ))}
    </ul>
  )
}

// flatten the tree so the parent <select> can offer every folder.
function flatten(nodes: Folder[], depth = 0): { id: string; label: string }[] {
  return nodes.flatMap((f) => [
    { id: f.id, label: `${' '.repeat(depth * 2)}${f.name}` },
    ...flatten(f.children ?? [], depth + 1),
  ])
}

export default function Folders({ projectId }: { projectId: string }) {
  const [tree, setTree] = useState<Folder[]>([])
  const [name, setName] = useState('')
  const [parentId, setParentId] = useState('')
  const [error, setError] = useState('')

  async function load() {
    try {
      setTree(await api.listFolders(projectId))
    } catch (err) {
      setError((err as Error).message)
    }
  }

  useEffect(() => {
    load()
  }, [projectId])

  async function create(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    try {
      await api.createFolder(projectId, name, parentId || null)
      setName('')
      setParentId('')
      await load()
    } catch (err) {
      setError((err as Error).message)
    }
  }

  const options = flatten(tree)

  return (
    <div className="panel">
      <h2>Folders</h2>
      <form className="row" onSubmit={create}>
        <input placeholder="Folder name" value={name} onChange={(e) => setName(e.target.value)} required />
        <select value={parentId} onChange={(e) => setParentId(e.target.value)}>
          <option value="">(root)</option>
          {options.map((o) => (
            <option key={o.id} value={o.id}>
              {o.label}
            </option>
          ))}
        </select>
        <button type="submit">Add folder</button>
      </form>
      {error && <div className="error">{error}</div>}
      {tree.length > 0 ? <Tree nodes={tree} /> : <p className="muted">No folders yet.</p>}
    </div>
  )
}
