import { useCallback, useEffect, useState } from 'react'
import { backendApi } from '../../services/backendApi.js'
import './backend-test.css'
import './typography.css'

const today = new Date().toISOString().slice(0, 10)
const emptyRegister = { name: '', nickname: '', email: '', password: '', birth_date: '' }

export default function BackendTest({ onBack }) {
  const [section, setSection] = useState('catalog')
  const [user, setUser] = useState(null)
  const [authMode, setAuthMode] = useState('login')
  const [register, setRegister] = useState(emptyRegister)
  const [credentials, setCredentials] = useState({ email: '', password: '' })
  const [works, setWorks] = useState([])
  const [shelf, setShelf] = useState([])
  const [entries, setEntries] = useState([])
  const [editions, setEditions] = useState({})
  const [query, setQuery] = useState('')
  const [diary, setDiary] = useState({ work_id: '', edition_id: '', current_page: 0, text: '', contains_spoiler: false })
  const [message, setMessage] = useState('')
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)

  const loadPersonal = useCallback(async () => {
    const [me, myShelf, myEntries] = await Promise.all([backendApi.me(), backendApi.shelf(), backendApi.entries()])
    setUser(me?.data || me)
    setShelf(myShelf?.data || [])
    setEntries(myEntries?.data || [])
  }, [])

  const loadWorks = useCallback(async (term = '') => {
    const result = await backendApi.works(term)
    setWorks(result?.data || [])
  }, [])

  useEffect(() => {
    loadWorks().catch((cause) => setError(cause.message))
    backendApi.me().then((result) => setUser(result?.data || result)).catch(() => {})
  }, [loadWorks])

  async function run(action, success) {
    setError('')
    setMessage('')
    setBusy(true)
    try {
      await action()
      if (success) setMessage(success)
    } catch (cause) {
      setError(cause.message || 'Não foi possível concluir a operação.')
    } finally {
      setBusy(false)
    }
  }

  async function submitAuth(event) {
    event.preventDefault()
    await run(async () => {
      if (authMode === 'register') await backendApi.register(register)
      const loginCredentials = authMode === 'register' ? { email: register.email, password: register.password } : credentials
      const result = await backendApi.login(loginCredentials)
      setUser(result?.user)
      setCredentials({ email: '', password: '' })
      await Promise.all([loadPersonal(), loadWorks(query)])
    }, authMode === 'register' ? 'Conta criada e sessão iniciada.' : 'Login realizado.')
  }

  async function chooseWork(work) {
    setError('')
    try {
      const result = await backendApi.editions(work.id)
      const list = result?.data || []
      setEditions((current) => ({ ...current, [work.id]: list }))
      if (list[0]) setDiary((current) => ({ ...current, work_id: String(work.id), edition_id: String(list[0].id) }))
      else setMessage(`“${work.title}” ainda não tem edição publicada disponível.`)
    } catch (cause) { setError(cause.message) }
  }

  async function addToShelf(work) {
    await run(async () => {
      const available = editions[work.id] || (await backendApi.editions(work.id))?.data || []
      if (!available.length) throw new Error('Esta obra não tem edição publicada disponível.')
      const edition = available[0]
      await backendApi.updateShelf(work.id, { edition_id: edition.id, current_page: 0, status: 'planned' })
      await loadPersonal()
    }, 'Livro adicionado à estante como planejado.')
  }

  async function submitDiary(event) {
    event.preventDefault()
    await run(async () => {
      await backendApi.createEntry({
        work_id: Number(diary.work_id), edition_id: Number(diary.edition_id),
        current_page: Number(diary.current_page), text: diary.text,
        contains_spoiler: diary.contains_spoiler, status: 'published',
      })
      await loadPersonal()
      setDiary((value) => ({ ...value, text: '' }))
    }, 'Registro publicado no diário.')
  }

  return <div className="backend-test">
    <header className="bt-header"><button className="bt-back" onClick={onBack}>← Voltar ao Radar</button><span className="bt-brand">radar<span>literário</span><small>AMBIENTE DE TESTE</small></span><span className={`bt-session ${user ? 'online' : ''}`}>{user ? `@${user.nickname}` : 'Sem sessão'}</span></header>
    <main className="bt-main">
      <div className="bt-intro"><span className="bt-eyebrow">PAINEL DE INTEGRAÇÃO</span><h1>Teste o backend</h1><p>Use uma conta de teste para conferir cadastro, catálogo, estante e diário com dados reais da API.</p></div>
      {(error || message) && <div className={`bt-alert ${error ? 'is-error' : 'is-success'}`} role={error ? 'alert' : 'status'}>{error || message}<button onClick={() => { setError(''); setMessage('') }} aria-label="Fechar aviso">×</button></div>}
      <nav className="bt-tabs" aria-label="Áreas de teste">{[['catalog', 'Catálogo'], ['shelf', 'Minha estante'], ['diary', 'Diário']].map(([id, label]) => <button key={id} className={section === id ? 'active' : ''} onClick={() => setSection(id)}>{label}</button>)}</nav>

      {!user && <section className="bt-card bt-auth"><div><span className="bt-eyebrow">SUA CONTA</span><h2>{authMode === 'login' ? 'Entre para testar' : 'Criar conta de teste'}</h2><p>As operações pessoais precisam de uma sessão autenticada.</p><button className="bt-link" onClick={() => setAuthMode(authMode === 'login' ? 'register' : 'login')}>{authMode === 'login' ? 'Ainda não tem conta? Cadastre-se' : 'Já tem conta? Entrar'}</button></div><form onSubmit={submitAuth} className="bt-form">
        {authMode === 'register' && <><label>Nome<input required value={register.name} onChange={(e) => setRegister({ ...register, name: e.target.value })} /></label><label>Apelido<input required pattern="[a-z0-9_]{3,30}" value={register.nickname} onChange={(e) => setRegister({ ...register, nickname: e.target.value.toLowerCase() })} placeholder="leitor_123" /></label><label>Data de nascimento<input required type="date" max={today} value={register.birth_date} onChange={(e) => setRegister({ ...register, birth_date: e.target.value })} /></label></>}
        <label>E-mail<input required type="email" autoComplete="email" value={authMode === 'login' ? credentials.email : register.email} onChange={(e) => authMode === 'login' ? setCredentials({ ...credentials, email: e.target.value }) : setRegister({ ...register, email: e.target.value })} /></label>
        <label>Senha<input required type="password" minLength="8" autoComplete={authMode === 'login' ? 'current-password' : 'new-password'} value={authMode === 'login' ? credentials.password : register.password} onChange={(e) => authMode === 'login' ? setCredentials({ ...credentials, password: e.target.value }) : setRegister({ ...register, password: e.target.value })} /></label>
        <button className="bt-primary" disabled={busy}>{busy ? 'Aguarde…' : authMode === 'login' ? 'Entrar' : 'Cadastrar e entrar'}</button>
      </form></section>}

      {section === 'catalog' && <section className="bt-section"><div className="bt-section-head"><div><span className="bt-eyebrow">OBRAS PUBLICADAS</span><h2>Catálogo</h2></div><form className="bt-search" onSubmit={(e) => { e.preventDefault(); run(() => loadWorks(query)) }}><input value={query} onChange={(e) => setQuery(e.target.value)} placeholder="Buscar pelo título" /><button className="bt-primary">Buscar</button></form></div><div className="bt-work-grid">{works.map((work) => <article className="bt-work" key={work.id}><span className="bt-book-mark">✳</span><div><h3>{work.title}</h3><p>{(work.authors || []).map((author) => author.name || author).join(', ') || 'Autor não informado'}</p><small>{work.original_language || work.language || 'Idioma não informado'} · classificação {work.minimum_age ?? 0}+</small></div><div className="bt-work-actions"><button className="bt-secondary" onClick={() => chooseWork(work)}>Usar no diário</button>{user && <button className="bt-primary" onClick={() => addToShelf(work)}>Adicionar à estante</button>}</div></article>)}{!works.length && <div className="bt-empty">Nenhuma obra encontrada. Confirme se há obras publicadas no catálogo.</div>}</div></section>}

      {section === 'shelf' && <section className="bt-section"><div className="bt-section-head"><div><span className="bt-eyebrow">DADOS DA CONTA</span><h2>Minha estante</h2></div><button className="bt-secondary" disabled={!user || busy} onClick={() => run(loadPersonal, 'Estante atualizada.')}>Atualizar</button></div>{!user ? <div className="bt-empty">Entre para consultar sua estante.</div> : <div className="bt-list">{shelf.map((item) => <article className="bt-list-row" key={item.id}><div><strong>{item.title}</strong><small>{item.edition_title} · página {item.current_page}</small></div><span className="bt-status">{item.status}</span></article>)}{!shelf.length && <div className="bt-empty">Sua estante ainda está vazia. Adicione uma obra pelo catálogo.</div>}</div>}</section>}

      {section === 'diary' && <section className="bt-section"><div className="bt-section-head"><div><span className="bt-eyebrow">PUBLICAÇÃO</span><h2>Diário de leitura</h2></div><button className="bt-secondary" disabled={!user || busy} onClick={() => run(loadPersonal, 'Diário atualizado.')}>Atualizar</button></div>{user ? <div className="bt-columns"><form className="bt-form bt-diary-form" onSubmit={submitDiary}><label>Obra<select required value={diary.work_id} onChange={(e) => { const work = works.find((item) => String(item.id) === e.target.value); setDiary({ ...diary, work_id: e.target.value, edition_id: '' }); if (work) chooseWork(work) }}><option value="">Selecione uma obra</option>{works.map((work) => <option key={work.id} value={work.id}>{work.title}</option>)}</select></label><label>Edição<select required value={diary.edition_id} onChange={(e) => setDiary({ ...diary, edition_id: e.target.value })}><option value="">Selecione uma edição</option>{(editions[diary.work_id] || []).map((edition) => <option key={edition.id} value={edition.id}>{edition.title || edition.isbn || `Edição ${edition.id}`}</option>)}</select></label><label>Página atual<input required min="0" type="number" value={diary.current_page} onChange={(e) => setDiary({ ...diary, current_page: e.target.value })} /></label><label>O que achou da leitura?<textarea required maxLength="10000" rows="4" value={diary.text} onChange={(e) => setDiary({ ...diary, text: e.target.value })} /></label><label className="bt-check"><input type="checkbox" checked={diary.contains_spoiler} onChange={(e) => setDiary({ ...diary, contains_spoiler: e.target.checked })} /> Contém spoiler</label><button className="bt-primary" disabled={busy}>Publicar diário</button></form><div className="bt-list bt-entry-list"><h3>Seus registros</h3>{entries.map((entry) => <article className="bt-entry" key={entry.id}><strong>{entry.title || entry.work_title || `Registro #${entry.id}`}</strong><p>{entry.text || 'Conteúdo protegido ou indisponível.'}</p><small>{entry.status || 'publicado'}{entry.current_page != null ? ` · página ${entry.current_page}` : ''}</small></article>)}{!entries.length && <div className="bt-empty">Seus diários publicados aparecerão aqui.</div>}</div></div> : <div className="bt-empty">Entre para publicar e consultar seus registros.</div>}</section>}

      {user && <footer className="bt-footer"><span>Sessão autenticada · cookies seguros e CSRF gerenciados pela API</span><button className="bt-link" disabled={busy} onClick={() => run(async () => { await backendApi.logout(); setUser(null); setShelf([]); setEntries([]) }, 'Sessão encerrada.')}>Sair da conta</button></footer>}
    </main>
  </div>
}
