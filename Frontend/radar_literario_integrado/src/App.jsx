import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { backendApi } from './services/backendApi.js'
import Auth from './features/auth/Auth.jsx'
import { libraryService } from './services/libraryService.js'
import BookDiscovery from './features/book-discovery/BookDiscovery.jsx'
// [MOBILE] Componentes de apresentação reaproveitam as ações e os dados de App.
import BottomNav from './components/BottomNav.jsx'
import Modal from './components/Modal.jsx'
import Skeleton from './components/Skeleton.jsx'
import './App.css'

const navItems = [
  ['home', 'Início', '⌂'], ['search', 'Pesquisar', '⌕'], ['shelf', 'Estante', '▤'], ['suggestions', 'Sugestões', '♧'], ['profile', 'Perfil', '●'],
]
const emptyProfile = { name: '', handle: '', bio: '', location: '', joined: '', favoriteBook: '', initials: '' }
const toProfile = (me) => ({ name: me.name, handle: `@${me.nickname}`, bio: me.biography || '', location: me.country_of_origin || '', joined: new Date(me.created_at).toLocaleDateString('pt-BR', { month: 'short', year: 'numeric' }), favoriteBook: '', initials: me.name.slice(0, 2).toUpperCase() })
// [ALTERADO] Os gêneros são lidos da estante e os estados seguem os dados retornados pelo serviço.
const shelfStatuses = ['Todos', 'Tenho interesse', 'Em leitura', 'Quero ler', 'Concluídos']

function Avatar({ size = '' }) { return <span className={`avatar ${size}`} aria-hidden="true"><svg viewBox="0 0 24 24" fill="currentColor"><circle cx="12" cy="8.5" r="4.2" /><path d="M3.5 22c0-4.9 3.8-8 8.5-8s8.5 3.1 8.5 8z" /></svg></span> }
function Cover({ book, className = '' }) {
  const palettes = book.genre === 'Romance' ? ['#ba6e83', '#523445'] : ['#697e72', '#243b36']
  return <div className={`cover ${className}`} style={{ '--cover-start': palettes[0], '--cover-end': palettes[1] }} aria-label={`Capa ilustrativa de ${book.title}`}>
    <span>{book.title}</span>
  </div>
}
function Stars({ rating, interactive = false, onChange }) {
  // [ALTERADO] aria-pressed informa qual nota foi escolhida ao tocar as estrelas.
  return <span className={`stars ${interactive ? 'interactive' : ''}`} aria-label={`Avaliação: ${rating} de 5`}>
    {[1, 2, 3, 4, 5].map((star) => <button type="button" key={star} aria-label={star === 1 ? '1 estrela' : `${star} estrelas`} aria-pressed={interactive ? star === rating : undefined} onClick={() => interactive && onChange(star)} disabled={!interactive} className={star <= rating ? 'filled' : ''}>★</button>)}
  </span>
}
function BookCard({ book, onOpen }) {
  const progress = Math.min(100, Math.round(book.pages / book.totalPages * 100))
  return <button className="book-card" onClick={() => onOpen(book)}><Cover book={book} /><span className="book-info"><strong>{book.title}</strong><small>{book.author}</small>{book.status === 'interested' && <small className="font-semibold text-purple-700">Tenho interesse</small>}<span className="book-meta">{book.rating > 0 && <><span className="gold">★</span> {book.rating}</>}<span>{book.genre}</span></span>{book.pages > 0 && <span className="progress-track"><span style={{ width: `${progress}%` }} /></span>}{book.pages > 0 && <small>{book.pages}/{book.totalPages} páginas · {progress}%</small>}</span></button>
}
function PostCard({ post, onLike, onReport, onComments }) {
  // [ALTERADO] aria-pressed comunica se a publicação já está curtida.
  return <article className="post-card"><div className="post-head"><Avatar initials={post.initials} /><div className="post-author"><strong>{post.person}</strong><small>{post.time}</small></div><button className="text-button report" onClick={() => onReport(post)}>Denunciar</button></div><div className="post-book"><div><h3>{post.title}</h3><p>{post.author}</p>{post.rating > 0 && <Stars rating={post.rating} />}</div><Cover book={{ title: post.title, genre: 'Fantasia' }} className="mini-cover" /></div>{post.pages && <div className="post-progress"><span className="progress-track"><span style={{ width: `${Math.round(post.pages / post.totalPages * 100)}%` }} /></span><small>{Math.round(post.pages / post.totalPages * 100)}%</small></div>}<p className="post-text">{post.text}</p><div className="post-actions"><button onClick={() => onLike(post.id)} aria-label="Curtir publicação" aria-pressed={post.liked} className={post.liked ? 'liked' : ''}>♡ <span>{post.likes}</span></button><button onClick={() => onComments(post)}>▣ <span>Comentar</span></button></div></article>
}

function App() {
  const [page, setPage] = useState('home')
  const [profile, setProfile] = useState(emptyProfile)
  const [session, setSession] = useState('loading')
  const [catalogQuery, setCatalogQuery] = useState('')
  // [ALTERADO] null representa espera; uma lista vazia representa uma resposta sem obras.
  const [catalogResults, setCatalogResults] = useState(null)
  const [catalogSearch, setCatalogSearch] = useState(0)
  const [catalogError, setCatalogError] = useState(false)
  const [searchError, setSearchError] = useState(false)
  const [searchRetry, setSearchRetry] = useState(0)
  const [catalogBooks, setCatalogBooks] = useState(null)
  const [genreList, setGenreList] = useState([])
  const [books, setBooks] = useState([])
  const [recommendations, setRecommendations] = useState(null)
  const [suggestionsView, setSuggestionsView] = useState('discover')
  // [MOBILE] Estados de carga reservam espaço no feed, na estante e nos comentários.
  const [posts, setPosts] = useState([])
  const [loadingLibrary, setLoadingLibrary] = useState(true)
  const [libraryError, setLibraryError] = useState('')
  const [comments, setComments] = useState(null)
  const [commentLoad, setCommentLoad] = useState(0)
  const [commentError, setCommentError] = useState('')
  const [query, setQuery] = useState('')
  const [genre, setGenre] = useState('Todos')
  const [shelfFilter, setShelfFilter] = useState('Todos')
  const [profileTab, setProfileTab] = useState('Publicações')
  const [modal, setModal] = useState('')
  const [selectedBook, setSelectedBook] = useState(null)
  const [selectedPost, setSelectedPost] = useState(null)
  const [toast, setToast] = useState('')
  const [notice, setNotice] = useState('')
  const [commentText, setCommentText] = useState('')
  const [rating, setRating] = useState(0)
  const [noticeField, setNoticeField] = useState('')
  const [formBookId, setFormBookId] = useState('')
  // [MOBILE] A busca do topo abre sob demanda e usa a mesma consulta da página.
  const [headerSearchOpen, setHeaderSearchOpen] = useState(false)
  const headerSearch = useRef(null)
  const searchToggle = useRef(null)
  const toastTimer = useRef(null)
  const modalOpener = useRef(null)

  // [ALTERADO] Cancela o aviso anterior para que cada resultado tenha tempo de ser lido.
  const feedback = useCallback((message) => {
    window.clearTimeout(toastTimer.current)
    setToast(message)
    toastTimer.current = window.setTimeout(() => setToast(''), 4500)
  }, [])
  useEffect(() => () => window.clearTimeout(toastTimer.current), [])

  // [ALTERADO] Ao entrar ou atualizar dados, preserva as telas que carregaram
  // caso outra consulta falhe. Uma falha após o login mantém a sessão aberta.
  const refresh = useCallback(async () => {
    const results = await Promise.allSettled([
      libraryService.getBooks(), libraryService.getPosts(), libraryService.getRecommendations(),
    ])
    const [shelf, feed, suggestions] = results
    if (shelf.status === 'fulfilled') setBooks(shelf.value)
    if (feed.status === 'fulfilled') setPosts(feed.value)
    if (suggestions.status === 'fulfilled') setRecommendations(suggestions.value)
    const error = results.find((result) => result.status === 'rejected')
    setLibraryError(error ? 'Não foi possível atualizar todas as leituras.' : '')
    if (error) feedback(error.reason?.message || 'Não foi possível atualizar as leituras.')
    setLoadingLibrary(false)
  }, [feedback])
  const activateSession = useCallback(async (me) => {
    setProfile(toProfile(me?.data || me))
    setLoadingLibrary(true)
    setSession('in')
    await refresh()
  }, [refresh])
  // [ALTERADO] Na abertura consulta a sessão existente e ignora a resposta de
  // montagens encerradas. O mesmo fluxo recebe o retorno depois de autenticar.
  const startSession = () => backendApi.me().then(activateSession)
  useEffect(() => {
    let active = true
    backendApi.me()
      .then((me) => { if (active) return activateSession(me) })
      .catch(() => { if (active) setSession('out') })
    return () => { active = false }
  }, [activateSession])

  // [ALTERADO] Trocar de publicação ou fechar o painel invalida a consulta em curso.
  useEffect(() => {
    if (modal !== 'comments' || !selectedPost) return
    let active = true
    libraryService.getComments(selectedPost.id)
      .then((result) => { if (active) setComments(result) })
      .catch((cause) => { if (active) { setCommentError(cause.message); feedback(cause.message) } })
    return () => { active = false }
  }, [modal, selectedPost, commentLoad, feedback])

  // [MOBILE] Move o foco para o campo só quando a pessoa abre a busca no cabeçalho.
  useEffect(() => { if (headerSearchOpen) headerSearch.current?.focus() }, [headerSearchOpen])

  function openForm(kind, book = null) {
    setSelectedBook(book)
    setFormBookId(book ? String(book.id) : '')
    setRating(0)
    setNotice('')
    setNoticeField('')
    // [ALTERADO] Reabrir o catálogo sempre começa com a lista geral.
    if (kind === 'book') { setCatalogQuery(''); setCatalogResults(null); setCatalogError(false) }
    setModal(kind)
  }
  // [MOBILE] Trocar de destino recolhe a busca e volta ao início da página.
  function navigate(destination) {
    setHeaderSearchOpen(false)
    if (destination !== page) window.scrollTo({ top: 0 })
    if (destination === 'search' && page !== 'search') { setCatalogBooks(null); setSearchError(false) }
    setPage(destination)
  }

  function updateQuery(value) {
    if (value !== query) { setCatalogBooks(null); setSearchError(false) }
    setQuery(value)
  }
  function openComments(post) {
    setComments(null)
    setCommentError('')
    setCommentText('')
    setSelectedPost(post)
    setModal('comments')
  }

  const visibleBooks = catalogBooks ?? []
  const genreChips = useMemo(() => ['Todos', ...genreList.map((item) => item.name)], [genreList])
  const genreId = genreList.find((item) => item.name === genre)?.id || 0
  // [ALTERADO] Busca os gêneros apenas durante a sessão e ignora respostas após sair.
  useEffect(() => {
    if (session !== 'in') return
    let active = true
    libraryService.getGenres()
      .then((result) => { if (active) setGenreList(result) })
      .catch((cause) => { if (active) feedback(cause.message) })
    return () => { active = false }
  }, [session, feedback])

  // [ALTERADO] Digitar ou trocar o gênero reinicia o debounce. A limpeza impede
  // que uma resposta antiga preencha o catálogo depois da próxima consulta.
  useEffect(() => {
    if (session !== 'in' || page !== 'search') return
    let active = true
    const timer = window.setTimeout(() => {
      libraryService.searchCatalog(query.trim(), genreId)
        .then((result) => { if (active) setCatalogBooks(result) })
        .catch((cause) => { if (active) { setSearchError(true); feedback(cause.message) } })
    }, 300)
    return () => { active = false; window.clearTimeout(timer) }
  }, [session, page, query, genreId, searchRetry, feedback])

  // [ALTERADO] A abertura e o botão Buscar consultam o mesmo catálogo;
  // a limpeza descarta a busca anterior e impede respostas após fechar o painel.
  useEffect(() => {
    if (session !== 'in' || modal !== 'book') return
    let active = true
    const timer = window.setTimeout(() => {
      libraryService.searchCatalog(catalogQuery.trim())
        .then((result) => { if (active) setCatalogResults(result) })
        .catch((cause) => { if (active) { setCatalogError(true); feedback(cause.message) } })
    }, 300)
    return () => { active = false; window.clearTimeout(timer) }
  }, [session, modal, catalogQuery, catalogSearch, feedback])
  // [ALTERADO] A lista de retomada inclui apenas leituras ainda em andamento.
  const readingBooks = books.filter((book) => book.pages > 0 && book.pages < book.totalPages)
  const shelfGenres = useMemo(() => [...new Set(books.map((book) => book.genre).filter(Boolean))].sort((a, b) => a.localeCompare(b, 'pt-BR')), [books])
  const [shelfGenre, setShelfGenre] = useState('Todos')
  const shelfBooks = books.filter((book) => {
    if (shelfGenre !== 'Todos' && book.genre !== shelfGenre) return false
    if (shelfFilter === 'Tenho interesse') return book.status === 'interested'
    if (shelfFilter === 'Em leitura') return readingBooks.includes(book)
    if (shelfFilter === 'Quero ler') return book.pages === 0 && book.status !== 'interested'
    if (shelfFilter === 'Concluídos') return book.pages >= book.totalPages
    return true
  })
  // [MOBILE] O seletor por ID evita títulos ambíguos e ajusta o limite de páginas ao mudar de livro.
  const formBook = books.find((book) => String(book.id) === formBookId)
  const rankedBooks = [...books].filter((book) => book.rating).sort((a, b) => b.rating - a.rating)

  async function addFromCatalogRaw(work) { await libraryService.addBook({ workId: work.id }); await refresh(); setModal(''); setShelfFilter('Todos'); setShelfGenre('Todos'); navigate('shelf'); feedback('Livro adicionado à sua estante.') }
  async function submitReadingRaw(event) { event.preventDefault(); const form = new FormData(event.currentTarget); const book = books.find((item) => String(item.id) === form.get('bookId')); if (!book) { setNoticeField('book'); setNotice('Esse livro não está na estante. Cadastre-o antes de registrar a leitura.'); return } const pages = Math.min(book.totalPages, Number(form.get('pages'))); await libraryService.addPost({ title: book.title, author: book.author, bookId: book.id, text: form.get('opinion'), pages, totalPages: book.totalPages, rating: 0 }); await refresh(); setModal(''); navigate('home'); feedback('Diário de leitura registrado.') }
  async function submitReviewRaw(event) { event.preventDefault(); const form = new FormData(event.currentTarget); const book = books.find((item) => String(item.id) === form.get('bookId')); if (!book) { setNoticeField('book'); setNotice('Esse livro não está na estante. Cadastre-o antes de avaliá-lo.'); return } if (!rating) { setNoticeField('rating'); setNotice('Escolha uma avaliação de uma a cinco estrelas.'); return } await libraryService.updateBook(book.id, { rating }); await libraryService.addPost({ title: book.title, author: book.author, bookId: book.id, rating, text: form.get('opinion') }); await refresh(); setModal(''); navigate('home'); feedback('Avaliação publicada.') }
  async function likePostRaw(id) { await libraryService.toggleLike(id, posts.find((post) => post.id === id)?.liked); await refresh() }
  async function reportPost(post) { setSelectedPost(post); setModal('report') }
  async function confirmReportRaw() { await libraryService.reportPost(selectedPost.id); await refresh(); setModal(''); feedback('Denúncia enviada.') }
  async function submitCommentRaw(event) { event.preventDefault(); if (!commentText.trim()) return; await libraryService.addComment(selectedPost.id, commentText.trim()); setCommentText((value) => value === commentText ? '' : value); setCommentLoad((value) => value + 1); await refresh() }
  async function deleteBookRaw(id) { await libraryService.deleteBook(id); await refresh(); setModal(''); setSelectedBook(null); feedback('Livro removido da estante.') }
  async function saveInterestRaw(book) { await libraryService.saveInterest(book.id); await refresh() }
  // [ALTERADO] A falha em salvar chega ao cartão para permitir uma nova tentativa.
  function openInterestedShelf() { setShelfFilter('Tenho interesse'); setShelfGenre('Todos'); navigate('shelf') }

  // [ALTERADO] As ações chamam os serviços apenas após a interação e anunciam erros no toast.
  async function attempt(action) {
    try { return await action() }
    catch (cause) { feedback(cause.message || 'Não foi possível concluir a ação.') }
  }
  const submitReading = (...args) => attempt(() => submitReadingRaw(...args))
  const submitReview = (...args) => attempt(() => submitReviewRaw(...args))
  const likePost = (...args) => attempt(() => likePostRaw(...args))
  const confirmReport = (...args) => attempt(() => confirmReportRaw(...args))
  const submitComment = (...args) => attempt(() => submitCommentRaw(...args))
  const deleteBook = (...args) => attempt(() => deleteBookRaw(...args))
  const addFromCatalog = (...args) => attempt(() => addFromCatalogRaw(...args))
  // [ALTERADO] Limpa os dados da conta ao sair, e usa o toast se a API falhar.
  const logout = () => attempt(async () => {
    await backendApi.logout()
    setProfile(emptyProfile)
    setBooks([])
    setPosts([])
    setRecommendations(null)
    setGenreList([])
    setGenre('Todos')
    setShelfGenre('Todos')
    setShelfFilter('Todos')
    setProfileTab('Publicações')
    setQuery('')
    setToast('')
    setSession('out')
    navigate('home')
  })
  if (session === 'loading') return <div className="session-loading"><Skeleton label="Carregando sessão" /></div>
  if (session === 'out') return <Auth onDone={startSession} />
  // [MOBILE] Guarda o botão acionado fora da folha para restaurar o foco ao final do fluxo.
  return <div className="app-shell min-h-screen bg-[#faf9fb] text-[#29242c]" onClickCapture={(event) => {
    if (!modal) modalOpener.current = event.target.closest('button')
  }}>
    <header className="topbar">
      <button className="brand" onClick={() => navigate('home')}><span className="brand-mark">r.</span><span>radar<span className="brand-light">literário</span></span></button>
      {/* [MOBILE] aria-expanded comunica o estado da busca e aria-controls aponta para o formulário. */}
      <button type="button" ref={searchToggle} className="icon-button header-search-toggle" aria-label={headerSearchOpen ? 'Recolher busca' : 'Abrir busca'} aria-expanded={headerSearchOpen} aria-controls="header-search" onClick={() => setHeaderSearchOpen(!headerSearchOpen)}>⌕</button>
      <form id="header-search" className={`header-search min-w-0 ${headerSearchOpen ? 'search-open' : ''}`} onSubmit={(event) => { event.preventDefault(); navigate('search') }} onKeyDown={(event) => { if (event.key === 'Escape' && headerSearchOpen) { setHeaderSearchOpen(false); searchToggle.current?.focus() } }}>
        <span aria-hidden="true">⌕</span>
        <input ref={headerSearch} value={query} enterKeyHint="search" onChange={(event) => { updateQuery(event.target.value); if (event.target.value && page !== 'search') { window.scrollTo({ top: 0 }); setPage('search') } }} placeholder="Pesquisar livros, autores, usuários..." aria-label="Pesquisar livros ou autores" />
        <button aria-label="Pesquisar">⌕</button>
      </form>
      <div className="header-user"><button aria-label="Meu perfil" onClick={() => navigate('profile')} className="user-short"><Avatar size="small" /><span>{profile.name.split(' ')[0]}</span></button><button className="primary-button compact" onClick={() => openForm('book')}>＋ <span>Novo livro</span></button></div>
    </header>
    <div className="layout"><aside className="sidebar"><div className="sidebar-caption">MENU</div>{navItems.map(([id, label, icon]) => <button key={id} className={`nav-item ${page === id ? 'active' : ''}`} onClick={() => navigate(id)}><span className="nav-icon">{icon}</span>{label}</button>)}<div className="sidebar-rule"/><div className="sidebar-caption">SUAS LEITURAS</div>{shelfGenres.length ? shelfGenres.map((item, index) => <button key={item} className={`nav-item ${page === 'shelf' && shelfGenre === item ? 'active' : ''}`} onClick={() => { navigate('shelf'); setShelfFilter('Todos'); setShelfGenre(item) }}><span className={`genre-dot ${index % 2 ? 'pink' : 'violet'}`}/> {item}</button>) : <p className="sidebar-empty">Seus gêneros aparecem aqui quando houver livros na estante.</p>}<button className="nav-item" onClick={() => openForm('diary')}>＋ Registrar leitura</button><div className="sidebar-bottom"><div className="tip-icon">✦</div><strong>Leia, acompanhe,<br/>compartilhe.</strong><p>Um lugar para cada história.</p></div></aside>
    <main className="main-content min-w-0">
      {libraryError && <div className="empty-state load-error">{libraryError} <button className="text-button" onClick={refresh}>Tentar novamente</button></div>}
      {/* [MOBILE] A ordem do início acompanha a leitura e a sequência de foco em telas estreitas. */}
      {page === 'home' && <div className="home-grid"><div className="welcome-row"><div><span className="eyebrow">{new Date().toLocaleDateString('pt-BR', { weekday: 'long', day: 'numeric', month: 'long' }).toUpperCase()}</span><h1>Bom te ver por aqui, {profile.name.split(' ')[0]}.</h1><p>Descubra o que a comunidade está lendo.</p></div><button className="outline-button" onClick={() => navigate('search')}>Explorar livros <span>→</span></button></div><aside className="home-aside"><div className="aside-heading"><h2>Continuar lendo</h2><button className="text-button" onClick={() => navigate('shelf')}>Estante →</button></div>{readingBooks.length ? readingBooks.slice(0, 3).map((book) => <button className="continue-card" key={book.id} onClick={() => openForm('diary', book)}><Cover book={book} /><span><strong>{book.title}</strong><small>{book.author}</small><span className="progress-track"><span style={{ width: `${Math.round(book.pages / book.totalPages * 100)}%` }} /></span><small>{book.pages} de {book.totalPages} páginas</small></span></button>) : <p className="muted">Sua estante está pronta para a próxima história.</p>}<div className="aside-divider"/><div className="aside-heading"><h2>Para você</h2><button className="text-button" onClick={() => navigate('suggestions')}>Ver mais →</button></div>{/* [MOBILE] A fileira deixa o início do feed mais perto da saudação. */}
        <div className="recommend-list">
          {(recommendations ?? []).slice(0, 3).map((book) => (
            <button className="recommend-row" key={book.id} onClick={() => openForm('bookDetail', book)}>
              <Cover book={book} />
              <span>
                <strong>{book.title}</strong>
                <small>{book.author}</small>
                <small>{book.rating > 0 && <span className="gold">★ {book.rating} · </span>}{book.genre}</small>
              </span>
            </button>
          ))}
        </div></aside><section className="feed-column"><div className="section-heading"><div><h2>Leituras da comunidade</h2><p>Novas histórias compartilhadas por leitores.</p></div><button className="text-button" onClick={() => navigate('search')}>Ver tudo →</button></div><div className="feed-list">{posts.map((post) => <PostCard key={post.id} post={post} onLike={likePost} onReport={reportPost} onComments={openComments} />)}{loadingLibrary ? <Skeleton label="Carregando feed" rows={3} /> : posts.length === 0 && <div className="empty-state">Ainda não há publicações. <button className="text-button" onClick={() => openForm('diary')}>Registrar leitura</button></div>}</div></section></div>}
      {page === 'search' && <section className="page-panel"><div className="page-title"><span className="eyebrow">ENCONTRE SUA PRÓXIMA LEITURA</span><h1>Pesquisar</h1><p>Explore livros e autores recomendados pela comunidade.</p></div><label className="large-search"><span aria-hidden="true">⌕</span><input type="search" enterKeyHint="search" placeholder="Título ou nome do autor..." aria-label="Título ou nome do autor" value={query} onChange={(event) => updateQuery(event.target.value)} /><kbd>Enter</kbd></label><div className="filter-row" aria-label="Filtrar por gênero">{genreChips.map((item) => <button className={`filter-chip ${genre === item ? 'selected' : ''}`} key={item} aria-pressed={genre === item} onClick={() => { if (genre !== item) { setCatalogBooks(null); setSearchError(false) } setGenre(item) }}>{item}</button>)}</div><div className="section-heading"><div><h2>{query ? `Resultados para “${query}”` : 'Obras do catálogo'}</h2><p>{catalogBooks !== null && `${visibleBooks.length} ${visibleBooks.length === 1 ? 'livro encontrado' : 'livros encontrados'}`}</p></div><button className="outline-button" onClick={() => openForm('book')}>＋ Cadastrar livro</button></div>{searchError ? <div className="empty-state">Não foi possível carregar o catálogo. <button className="text-button" onClick={() => { setSearchError(false); setSearchRetry((value) => value + 1) }}>Tentar novamente</button></div> : catalogBooks === null ? <Skeleton label="Buscando no catálogo" /> : visibleBooks.length ? <div className="book-grid">{visibleBooks.map((book) => <BookCard key={book.id} book={book} onOpen={(item) => { setSelectedBook(item); setModal('bookDetail') }} />)}</div> : <div className="empty-state">Nenhum livro encontrado no catálogo. Tente outro termo.{(query || genre !== 'Todos') && <button className="text-button" onClick={() => { updateQuery(''); setGenre('Todos'); setCatalogBooks(null) }}>Limpar filtros</button>}</div>}</section>}
      {page === 'shelf' && <section className="page-panel"><div className="page-title shelf-title"><div><span className="eyebrow">SEU CANTINHO DE LEITURA</span><h1>Minha estante</h1><p>Acompanhe suas leituras e encontre tudo em um só lugar.</p></div><button className="primary-button" onClick={() => openForm('book')}>＋ Cadastrar livro</button></div><div className="shelf-stats"><span><strong>{books.length}</strong> livros</span><span><strong>{readingBooks.length}</strong> em leitura</span><span><strong>{books.filter((book) => book.pages >= book.totalPages).length}</strong> concluídos</span></div><div className="filter-row shelf-filters" aria-label="Filtrar por estado de leitura">{shelfStatuses.map((item) => <button key={item} className={`filter-chip ${shelfFilter === item ? 'selected' : ''}`} aria-pressed={shelfFilter === item} onClick={() => setShelfFilter(item)}>{item}</button>)}</div>
        {/* [MOBILE] aria-pressed anuncia cada filtro ativo. Os gêneros combinam com o estado escolhido. */}
        {shelfGenres.length > 0 && <div className="shelf-genres"><span className="eyebrow">SUAS LEITURAS</span><div className="filter-row shelf-filters" aria-label="Filtrar estante por gênero">{['Todos', ...shelfGenres].map((item) => <button key={item} className={`filter-chip ${shelfGenre === item ? 'selected' : ''}`} aria-pressed={shelfGenre === item} onClick={() => setShelfGenre(item)}>{item === 'Todos' ? 'Todos os gêneros' : item}</button>)}</div></div>}
        {loadingLibrary ? <Skeleton label="Carregando estante" /> : shelfBooks.length ? <div className="book-grid">{shelfBooks.map((book) => <div className="shelf-book" key={book.id}><BookCard book={book} onOpen={(item) => openForm('bookDetail', item)} /><div className="shelf-book-actions"><button onClick={() => openForm('diary', book)}>Registrar progresso</button><button onClick={() => openForm('deleteBook', book)}>Remover</button></div></div>)}</div> : <div className="empty-state">Nenhum livro nesta categoria. <button className="text-button" onClick={() => openForm('book')}>Adicionar livro</button></div>}</section>}
      {page === 'suggestions' && <section className="page-panel">
        <div className="filter-row" aria-label="Tipo de sugestões">
          <button type="button" className={`filter-chip ${suggestionsView === 'discover' ? 'selected' : ''}`} aria-pressed={suggestionsView === 'discover'} onClick={() => setSuggestionsView('discover')}>Indicações para você</button>
          <button type="button" className={`filter-chip ${suggestionsView === 'highlights' ? 'selected' : ''}`} aria-pressed={suggestionsView === 'highlights'} onClick={() => setSuggestionsView('highlights')}>Destaques da comunidade</button>
        </div>
        {suggestionsView === 'discover' ? recommendations === null ? loadingLibrary ? <Skeleton label="Carregando indicações" /> : <div className="empty-state">Indicações indisponíveis. <button className="text-button" onClick={refresh}>Tentar novamente</button></div> : <BookDiscovery books={recommendations.filter((candidate) => !books.some((book) => book.id === candidate.id))} onInterest={saveInterestRaw} onOpenShelf={openInterestedShelf} /> : <section className="page-panel"><div className="page-title"><span className="eyebrow">ESCOLHIDOS POR LEITORES</span><h1>Descubra histórias</h1><p>Recomendações e destaques da comunidade do Radar Literário.</p></div><div className="feature-banner"><div><span className="eyebrow">EM DESTAQUE</span><h2>Histórias que ficam<br/>com você.</h2><p>Obras queridas pelos leitores, reunidas em um só lugar.</p><button className="light-button" onClick={() => navigate('search')}>Explorar catálogo <span>→</span></button></div><span className="feature-spark">✳</span></div><div className="section-heading"><div><h2>Mais populares</h2><p>As obras que a comunidade está lendo.</p></div></div><div className="book-grid">{rankedBooks.map((book) => <BookCard key={book.id} book={book} onOpen={(item) => { setSelectedBook(item); setModal('bookDetail') }} />)}</div><div className="section-heading section-gap"><div><h2>Mais bem avaliados</h2><p>Favoritos dos leitores.</p></div></div><div className="book-grid">{[...rankedBooks].reverse().map((book) => <BookCard key={book.id} book={book} onOpen={(item) => { setSelectedBook(item); setModal('bookDetail') }} />)}</div></section>}
      </section>}
      {page === 'profile' && <section className="page-panel profile-page"><div className="profile-banner"><div className="profile-cover-art"/><div className="profile-main"><Avatar size="large"/><div className="profile-identity"><h1>{profile.name}</h1><span>{profile.handle}</span><p>{profile.bio}</p><small>{profile.location} · No Radar desde {profile.joined}</small></div><button className="outline-button profile-edit" onClick={() => openForm('editProfile')}>Editar perfil</button></div><div className="profile-stats"><span><strong>{posts.filter((post) => post.handle === profile.handle).length}</strong> Publicações</span></div></div><div className="favorite-book"><span className="eyebrow">LIVRO FAVORITO</span><strong>{profile.favoriteBook || '—'}</strong><button onClick={() => navigate('search')}>Encontrar sua próxima leitura →</button></div><div className="profile-tabs">{['Publicações', 'Avaliações', 'Diários', 'Curtidas'].map((tab) => <button key={tab} aria-pressed={profileTab === tab} className={profileTab === tab ? 'selected' : ''} onClick={() => setProfileTab(tab)}>{tab}</button>)}</div>{(() => { const ownPosts = posts.filter((post) => post.handle === profile.handle); const shownPosts = (profileTab === 'Curtidas' ? posts : ownPosts).filter((post) => profileTab === 'Avaliações' ? post.rating > 0 : profileTab === 'Diários' ? post.rating === 0 : profileTab === 'Curtidas' ? post.liked : true); return shownPosts.length ? <div className="profile-feed">{shownPosts.map((post) => <PostCard key={post.id} post={post} onLike={likePost} onReport={reportPost} onComments={openComments} />)}</div> : <div className="empty-state">{profileTab === 'Curtidas' ? 'As publicações que você curtir aparecerão aqui.' : `Suas ${profileTab.toLowerCase()} aparecerão aqui.`}</div> })()}<div className="account-actions"><span>Configurações da conta</span><button onClick={logout}>Sair da conta</button><button onClick={() => setModal('deleteAccount')}>Sobre a exclusão da conta</button></div></section>}
    </main></div>
    <BottomNav items={navItems} page={page} onNavigate={navigate} onRead={() => openForm('diary')} onCreate={() => openForm('actions')} />

    {modal === 'bookDetail' && selectedBook && <Modal toast={toast} returnFocus={modalOpener} title={selectedBook.title} onClose={() => { setSelectedBook(null); setModal('') }}><div className="book-detail"><Cover book={selectedBook}/><div><span className="eyebrow">{selectedBook.genre}</span><h3>{selectedBook.author}</h3><div className="gold">★ {selectedBook.rating || '—'} <span className="muted">· {selectedBook.editionId ? `${selectedBook.totalPages} páginas` : 'Total de páginas indisponível'}</span></div><p>{selectedBook.description}</p><div className="modal-actions">{books.some((item) => item.id === selectedBook.id) ? <><button className="outline-button" onClick={() => { openForm('review', selectedBook) }}>Avaliar livro</button><button className="primary-button" onClick={() => { openForm('diary', selectedBook) }}>Registrar leitura</button></> : <button className="primary-button" onClick={() => addFromCatalog(selectedBook)}>Adicionar à estante</button>}</div></div></div></Modal>}
    {modal === 'book' && <Modal toast={toast} returnFocus={modalOpener} title="Adicionar livro" onClose={() => setModal('')}><p className="modal-lead">Busque uma obra do catálogo para colocar na sua estante.</p><form className="form-stack" onSubmit={(event) => { event.preventDefault(); setCatalogResults(null); setCatalogError(false); setCatalogSearch((value) => value + 1) }}><label>Título ou autor<input type="search" enterKeyHint="search" value={catalogQuery} onChange={(event) => { setCatalogResults(null); setCatalogError(false); setCatalogQuery(event.target.value) }} placeholder="Ex.: O nome do vento" /></label><button className="primary-button">Buscar</button></form><div className="comments-list">{catalogError ? <p className="modal-lead">Não foi possível consultar o catálogo. Tente buscar novamente.</p> : catalogResults === null && <Skeleton label="Carregando obras do catálogo" />}{catalogResults?.map((work) => <div className="comment" key={work.id}><div><strong>{work.title}</strong><small>{work.author}</small></div><button type="button" className="outline-button" onClick={() => addFromCatalog(work)}>Adicionar</button></div>)}{catalogResults?.length === 0 && <p className="modal-lead">Nenhum resultado ainda. Busque pelo título; se a obra não existir no catálogo, ela precisa ser cadastrada por um administrador.</p>}</div></Modal>}
    {modal === 'diary' && <Modal toast={toast} returnFocus={modalOpener} title="Fazer diário de leitura" onClose={() => setModal('')}><p className="modal-lead">Registre seu progresso e compartilhe suas impressões.</p><form className="form-stack" onSubmit={submitReading}><label>Nome do livro<select name="bookId" required value={formBookId} onChange={(event) => { setFormBookId(event.target.value); if (noticeField === 'book') setNoticeField('') }} aria-describedby={noticeField === 'book' ? 'book-error' : undefined}><option value="" disabled>Escolha uma obra da estante</option>{books.map((book) => <option key={book.id} value={book.id}>{book.title} · {book.author}</option>)}</select></label>{noticeField === 'book' && <p id="book-error" className="form-error" role="alert">{notice}</p>}{books.length === 0 && <button type="button" className="text-button" onClick={() => openForm('book')}>Adicionar livro à estante</button>}<label>Quantidade de páginas lidas<input name="pages" type="number" inputMode="numeric" required min="1" max={formBook?.totalPages || undefined} placeholder="Páginas lidas até agora" /></label><label>Sua opinião<textarea name="opinion" required rows="4" maxLength="700" placeholder="O que está achando da leitura?" /></label><label className="check-label"><input type="checkbox" name="sensitive" /> Conteúdo sensível</label><label className="check-label"><input type="checkbox" name="spoiler" /> Contém spoiler</label><div className="modal-actions"><button type="button" className="outline-button" onClick={() => setModal('')}>Cancelar</button><button className="primary-button">Enviar diário</button></div></form></Modal>}
    {modal === 'review' && <Modal toast={toast} returnFocus={modalOpener} title="Fazer avaliação" onClose={() => setModal('')}><p className="modal-lead">Conte à comunidade o que achou da obra.</p><form className="form-stack" onSubmit={submitReview}><label>Nome do livro<select name="bookId" required value={formBookId} onChange={(event) => { setFormBookId(event.target.value); if (noticeField === 'book') setNoticeField('') }} aria-describedby={noticeField === 'book' ? 'book-error' : undefined}><option value="" disabled>Escolha uma obra da estante</option>{books.map((book) => <option key={book.id} value={book.id}>{book.title} · {book.author}</option>)}</select></label>{noticeField === 'book' && <p id="book-error" className="form-error" role="alert">{notice}</p>}{books.length === 0 && <button type="button" className="text-button" onClick={() => openForm('book')}>Adicionar livro à estante</button>}<div className="rating-field"><span>Sua avaliação</span><Stars rating={rating} interactive onChange={(value) => { setRating(value); setNoticeField('') }} /></div>{noticeField === 'rating' && <p className="form-error" role="alert">{notice}</p>}<label>Sua opinião<textarea name="opinion" required rows="4" maxLength="700" placeholder="Escreva uma resenha..." /></label><label className="check-label"><input type="checkbox" /> Conteúdo sensível</label><label className="check-label"><input type="checkbox" /> Contém spoiler</label><div className="modal-actions"><button type="button" className="outline-button" onClick={() => setModal('')}>Cancelar</button><button className="primary-button">Enviar avaliação</button></div></form></Modal>}
    {modal === 'comments' && <Modal toast={toast} returnFocus={modalOpener} title="Comentários" wide onClose={() => setModal('')}><p className="comment-context">Sobre <strong>{selectedPost?.title}</strong> · por {selectedPost?.person}</p><div className="comments-list">{commentError ? <div className="empty-state">Não foi possível carregar os comentários. <button className="text-button" onClick={() => { setCommentError(''); setCommentLoad((value) => value + 1) }}>Tentar novamente</button></div> : comments === null ? <Skeleton label="Carregando comentários" /> : comments.length === 0 && <p className="modal-lead">Seja a primeira pessoa a comentar.</p>}{comments?.map((comment) => <div className="comment" key={comment.id}><Avatar initials={comment.person === 'Pudim' ? 'PN' : comment.person.slice(0, 2).toUpperCase()} size="small"/><div><strong>{comment.person}</strong><small>{comment.time}</small><p>{comment.text}</p></div></div>)}</div><form className="comment-form" onSubmit={submitComment}><label className="comment-label" htmlFor="comment-input">Escreva um comentário</label><input id="comment-input" enterKeyHint="send" value={commentText} onChange={(event) => setCommentText(event.target.value)} placeholder="Escreva uma mensagem..." maxLength="500" required /><button className="primary-button" aria-label="Enviar comentário">Enviar <span>→</span></button></form></Modal>}
    {modal === 'report' && <Modal toast={toast} returnFocus={modalOpener} title="Denunciar publicação" onClose={() => setModal('')}><div className="confirm-box"><span className="confirm-icon">!</span><p>A publicação de <strong>{selectedPost?.person}</strong> sobre <strong>{selectedPost?.title}</strong> será enviada para análise.</p></div><div className="modal-actions"><button className="outline-button" onClick={() => setModal('')}>Cancelar</button><button className="danger-button" onClick={confirmReport}>Confirmar denúncia</button></div></Modal>}
    {modal === 'deleteBook' && <Modal toast={toast} returnFocus={modalOpener} title="Remover livro?" onClose={() => setModal('')}><div className="confirm-box"><span className="confirm-icon">!</span><p><strong>{selectedBook?.title}</strong> será removido da sua estante.</p></div><div className="modal-actions"><button className="outline-button" onClick={() => setModal('')}>Cancelar</button><button className="danger-button" onClick={() => deleteBook(selectedBook.id)}>Remover livro</button></div></Modal>}
    {/* [ALTERADO] O serviço consumido pelo frontend ainda não oferece a operação de exclusão. */}
    {modal === 'deleteAccount' && <Modal toast={toast} returnFocus={modalOpener} title="Exclusão da conta" onClose={() => setModal('')}><p className="modal-lead">A exclusão de conta está indisponível nesta versão.</p></Modal>}
    {modal === 'actions' && <Modal toast={toast} returnFocus={modalOpener} title="O que vamos fazer?" onClose={() => setModal('')}><div className="action-choices"><button onClick={() => openForm('book')}><span>＋</span><strong>Cadastrar livro</strong><small>Adicione uma obra à estante.</small></button><button onClick={() => openForm('diary')}><span>▤</span><strong>Registrar leitura</strong><small>Compartilhe seu progresso.</small></button><button onClick={() => openForm('review')}><span>★</span><strong>Avaliar livro</strong><small>Deixe sua opinião.</small></button></div></Modal>}
    {modal === 'editProfile' && <Modal toast={toast} returnFocus={modalOpener} title="Editar perfil" onClose={() => setModal('')}><p className="modal-lead">Atualize as informações que aparecem no seu perfil.</p><form className="form-stack" onSubmit={async (event) => { event.preventDefault(); const form = new FormData(event.currentTarget); try { const saved = await backendApi.patchMe({ name: form.get('name').trim(), nickname: form.get('handle').trim().replace(/^@/, ''), biography: form.get('bio').trim() }); setProfile(toProfile(saved?.data || saved)); await refresh(); setModal(''); feedback('Perfil atualizado.') } catch (cause) { feedback(cause.message) } }}><label>Nome<input name="name" autoComplete="name" defaultValue={profile.name} required maxLength="60" /></label><label>Nickname<input name="handle" autoComplete="nickname" autoCapitalize="none" defaultValue={profile.handle} required maxLength="30" /></label><label>Sobre você<textarea name="bio" rows="3" defaultValue={profile.bio} maxLength="240" /></label><div className="modal-actions"><button type="button" className="outline-button" onClick={() => setModal('')}>Cancelar</button><button className="primary-button">Salvar alterações</button></div></form></Modal>}
    {/* [MOBILE] role=status anuncia feedback sem tirar o foco. O diálogo tem sua própria área de avisos. */}
    {toast && !modal && <div className="toast" role="status">{toast}</div>}
  </div>
}

export default App