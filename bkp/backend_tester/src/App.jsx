import { useCallback, useEffect, useRef, useState } from 'react'
import { request } from './api.js'
import { ApiContext } from './ui.jsx'
import { redact, redactPath } from './requestLog.js'
import RequestHistory from './RequestHistory.jsx'
import Account from './pages/Account.jsx'
import Catalog from './pages/Catalog.jsx'
import CreateWork from './pages/CreateWork.jsx'
import Shelf from './pages/Shelf.jsx'
import Publications from './pages/Publications.jsx'
import Console from './pages/Console.jsx'
import Works from './pages/Works.jsx'
import Editions from './pages/Editions.jsx'
import References from './pages/References.jsx'

const pages = [['account', 'Conta'], ['catalog', 'Catálogo público'], ['works', 'Gerenciar obras'], ['create-work', 'Adicionar obra'], ['editions', 'Edições'], ['references', 'Autores, gêneros e editoras'], ['shelf', 'Estante'], ['publications', 'Publicações'], ['console', 'Console da API']]
const currentPage = (hash) => pages.some(([id]) => id === hash.slice(1).split('?')[0]) ? hash.slice(1).split('?')[0] : 'account'

export default function App() {
  const [route, setRoute] = useState(() => location.hash)
  const page = currentPage(route)
  const [user, setUser] = useState(null)
  const [busy, setBusy] = useState(false)
  const running = useRef(0)
  const requestId = useRef(0)
  const [notice, setNotice] = useState(null)
  const [history, setHistory] = useState([])
  const [selection, setSelection] = useState(null)

  useEffect(() => {
    const update = () => setRoute(location.hash)
    window.addEventListener('hashchange', update)
    let active = true
    request('/me').then(({ payload }) => { if (active) setUser(payload.data) }).catch(() => {})
    return () => { active = false; window.removeEventListener('hashchange', update) }
  }, [])

  const perform = useCallback(async (path, method = 'GET', body, success) => {
    running.current++
    setBusy(true)
    setNotice((current) => method !== 'GET' || current?.kind === 'error' ? null : current)
    const started = performance.now()
    const record = (status, payload) => {
      const entry = { id: ++requestId.current, method, path: redactPath(path), status, body: redact(body), payload: redact(payload), time: Date.now(), duration: Math.round(performance.now() - started) }
      setHistory((current) => [entry, ...current].slice(0, 30))
    }
    try {
      const result = await request(path, method, body)
      record(result.status, result.payload)
      if (success) setNotice({ kind: 'success', message: success })
      return result.payload ?? { data: null }
    } catch (error) {
      record(error.status || 'rede', error.payload || { message: error.message })
      setNotice({ kind: 'error', message: error.message })
      if (error.status === 401) setUser(null)
    } finally {
      running.current--
      setBusy(running.current > 0)
    }
  }, [])

  function selectEdition(value) {
    setSelection(value)
    location.hash = 'shelf'
  }

  return <ApiContext.Provider value={{ perform, busy, user, setUser }}>
    <header className="topbar"><a href="#account" className="brand">Radar Literário <span>teste do backend</span></a><span className="session">{user ? `@${user.nickname} · ${user.role || 'leitor'}` : 'Visitante'}</span></header>
    <div className="layout"><aside><p className="caption">NAVEGAÇÃO</p><nav aria-label="Páginas">{pages.map(([id, title], index) => <a key={id} href={`#${id}`} aria-current={id === page ? 'page' : undefined} aria-disabled={busy || undefined} onClick={(event) => { if (busy) event.preventDefault() }}><span>0{index + 1}</span>{title}</a>)}</nav><p className="sidebar-note">API local<br /><code>127.0.0.1:8080</code></p></aside>
      <main>
        {notice && <div className={`notice ${notice.kind}`} role={notice.kind === 'error' ? 'alert' : 'status'}><span>{notice.message}</span><button aria-label="Fechar aviso" onClick={() => setNotice(null)}>×</button></div>}
        {page === 'account' && <Account />}
        {page === 'catalog' && <Catalog onSelect={selectEdition} />}
        {page === 'create-work' && <CreateWork />}
        {page === 'works' && <Works />}
        {page === 'editions' && <Editions key={route} />}
        {page === 'references' && <References />}
        {page === 'shelf' && <Shelf selection={selection} />}
        {page === 'publications' && <Publications />}
        {page === 'console' && <Console />}
        <RequestHistory entries={history} onClear={() => setHistory([])} />
      </main>
    </div>
  </ApiContext.Provider>
}
