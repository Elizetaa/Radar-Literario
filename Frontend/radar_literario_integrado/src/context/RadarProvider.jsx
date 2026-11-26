// RadarProvider: guarda o ESTADO GLOBAL do app e as AÇÕES compartilhadas.
//
// Por que um Provider? Antes tudo isso ficava dentro do App.jsx e era passado
// de componente em componente. Agora cada página lê só o que precisa com useRadar().
//
// O que fica aqui (é usado por várias telas):
//   - sessão e perfil do usuário
//   - estante (books), feed (posts) e indicações (recommendations)
//   - página atual, texto da busca do cabeçalho, filtro da estante
//   - qual modal está aberto e qual livro/publicação ele usa
//   - aviso temporário (toast)
// O que NÃO fica aqui (é local de cada página/modal): campos de formulário,
// resultados de busca do catálogo, comentários abertos, aba do perfil etc.
import { useCallback, useEffect, useMemo, useState } from 'react'
import { backendApi } from '../services/backendApi.js'
import { libraryService } from '../services/libraryService.js'
import { emptyProfile, toProfile } from '../utils/profile.js'
import { RadarContext } from './RadarContext.js'
import { useLocation, useNavigate } from 'react-router-dom'
import { pagePaths } from '../utils/constants.js'


export default function RadarProvider({ children }) {
  // ---- sessão e dados vindos da API ---------------------------------------
  const [session, setSession] = useState('loading') // 'loading' | 'in' | 'out'
  const [profile, setProfile] = useState(emptyProfile)
  const [books, setBooks] = useState([])
  const [posts, setPosts] = useState([])
  const [recommendations, setRecommendations] = useState(null) // null = ainda carregando

  // ---- navegação e interface ----------------------------------------------
  const navigate = useNavigate()
  const { pathname } = useLocation()

  // A página atual agora é derivada da URL
  const page = Object.keys(pagePaths).find((id) => pagePaths[id] === pathname) ?? 'home'

  // setPage continua existindo, mas agora navega (cria uma entrada no histórico)
  const setPage = (id) => navigate(pagePaths[id])
  const [query, setQuery] = useState('') // texto digitado na busca do cabeçalho
  const [shelfFilter, setShelfFilter] = useState('Todos')
  const [modal, setModal] = useState('') // '' = nenhum modal aberto
  const [selectedBook, setSelectedBook] = useState(null)
  const [selectedPost, setSelectedPost] = useState(null)
  const [toast, setToast] = useState('')

  // Aviso que some sozinho depois de 2,8 s.
  const feedback = useCallback((message) => {
    setToast(message)
    window.setTimeout(() => setToast(''), 2800)
  }, [])

  // Envolve uma ação assíncrona: se a API falhar, mostra o erro no toast
  // em vez de deixar a promessa rejeitada "morrer" no console.
  const attempt = useCallback((fn) => async (...args) => {
    try {
      return await fn(...args)
    } catch (cause) {
      feedback(cause.message || 'Não foi possível concluir a ação.')
    }
  }, [feedback])

  // Recarrega estante, feed e indicações de uma vez (em paralelo, para ser mais rápido).
  // Chamado depois de qualquer ação que mude dados no servidor.
  const refresh = useCallback(async () => {
    const [nextBooks, nextPosts, nextRecommendations] = await Promise.all([
      libraryService.getBooks(),
      libraryService.getPosts(),
      libraryService.getRecommendations(),
    ])
    setBooks(nextBooks)
    setPosts(nextPosts)
    setRecommendations(nextRecommendations)
  }, [])

  // Guarda o usuário logado e carrega os dados do app.
  const applySession = useCallback(async (me) => {
    setProfile(toProfile(me?.data || me))
    setSession('in')
    await refresh()
  }, [refresh])

  // Usado pela tela de login depois de entrar com sucesso.
  const startSession = useCallback(() => backendApi.me().then(applySession), [applySession])

  // Ao abrir o app: se já existe cookie de sessão válido, entra direto; senão mostra o login.
  // (A chamada fica dentro de .then para o setState acontecer de forma assíncrona.)
  useEffect(() => {
    backendApi.me().then(applySession).catch(() => setSession('out'))
  }, [applySession])

  const logout = useCallback(async () => {
    await backendApi.logout()
    setProfile(emptyProfile)
    setBooks([])
    setPosts([])
    setSession('out')
    setPage('home')
  }, [])

  // ---- abrir/fechar modais -------------------------------------------------
  const openModal = useCallback((name) => setModal(name), [])
  const closeModal = useCallback(() => setModal(''), [])
  // Abre um formulário (diário, avaliação, adicionar livro...) já com um livro escolhido, se houver.
  const openForm = useCallback((kind, book = null) => { setSelectedBook(book); setModal(kind) }, [])
  const openBookDetail = useCallback((book) => { setSelectedBook(book); setModal('bookDetail') }, [])
  const openComments = useCallback((post) => { setSelectedPost(post); setModal('comments') }, [])
  const openReport = useCallback((post) => { setSelectedPost(post); setModal('report') }, [])
  // Atalho usado em "Sugestões": vai para a estante já filtrada por "Tenho interesse".
  const openInterestedShelf = useCallback(() => { setShelfFilter('Tenho interesse'); setPage('shelf') }, [])

  // ---- ações de dados compartilhadas --------------------------------------
  // Curtir: o servidor alterna (curte/descurte); depois recarregamos para atualizar o contador.
  const likePost = attempt(async (id) => {
    await libraryService.toggleLike(id, posts.find((post) => post.id === id)?.liked)
    await refresh()
  })
  // "Tenho interesse": coloca o livro na estante com status "planejado".
  const saveInterest = attempt(async (book) => {
    await libraryService.saveInterest(book.id)
    await refresh()
  })

  // ---- valores derivados da estante ---------------------------------------
  // Livros que já têm alguma página lida.
  const readingBooks = useMemo(() => books.filter((book) => book.pages > 0), [books])
  // Gêneros dos livros da estante, sem repetição e em ordem alfabética (menu "Suas leituras").
  const shelfGenres = useMemo(
    () => [...new Set(books.map((book) => book.genre).filter(Boolean))].sort((a, b) => a.localeCompare(b, 'pt-BR')),
    [books],
  )

  const value = {
    // dados
    session, profile, setProfile, books, posts, recommendations, readingBooks, shelfGenres,
    // navegação / interface
    page, setPage, query, setQuery, shelfFilter, setShelfFilter,
    modal, openModal, closeModal, openForm, openBookDetail, openComments, openReport, openInterestedShelf,
    selectedBook, selectedPost, toast,
    // ações
    refresh, startSession, logout, feedback, attempt, likePost, saveInterest,
  }
  return <RadarContext.Provider value={value}>{children}</RadarContext.Provider>
}
