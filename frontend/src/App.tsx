import { useState } from 'react'
import { clearToken, getToken, type Project } from './api'
import Login from './components/Login'
import Projects from './components/Projects'
import Folders from './components/Folders'
import TestCases from './components/TestCases'
import Suites from './components/Suites'
import Runs from './components/Runs'
import './App.css'

const TABS = ['Folders', 'Test Cases', 'Suites', 'Runs'] as const
type Tab = (typeof TABS)[number]

export default function App() {
  const [authed, setAuthed] = useState(!!getToken())
  const [project, setProject] = useState<Project | null>(null)
  const [tab, setTab] = useState<Tab>('Folders')

  if (!authed) {
    return <Login onLogin={() => setAuthed(true)} />
  }

  return (
    <div className="app">
      <header>
        <h1>TMS</h1>
        {project && <span className="current">{project.name}</span>}
        <button
          className="link logout"
          onClick={() => {
            clearToken()
            setAuthed(false)
            setProject(null)
          }}
        >
          Logout
        </button>
      </header>

      <Projects selectedId={project?.id} onSelect={setProject} />

      {project && (
        <>
          <nav className="tabs">
            {TABS.map((t) => (
              <button key={t} className={t === tab ? 'tab active' : 'tab'} onClick={() => setTab(t)}>
                {t}
              </button>
            ))}
          </nav>
          {tab === 'Folders' && <Folders projectId={project.id} />}
          {tab === 'Test Cases' && <TestCases projectId={project.id} />}
          {tab === 'Suites' && <Suites projectId={project.id} />}
          {tab === 'Runs' && <Runs projectId={project.id} />}
        </>
      )}
    </div>
  )
}
