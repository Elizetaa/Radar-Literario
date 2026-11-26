// Página "Pesquisar": busca obras no CATÁLOGO do banco (não só na estante do usuário).
// O texto digitado vem do contexto (`query`), pois o campo do cabeçalho também o altera.
import { useEffect, useMemo, useState } from 'react'
import { useRadar } from '../context/useRadar.js'
import { libraryService } from '../services/libraryService.js'
import BookCard from '../components/BookCard.jsx'
import MobileGreeting from '../components/MobileGreeting.jsx'

export default function SearchPage() {
  const { query, setQuery, openForm, openBookDetail, feedback } = useRadar()
  const [genre, setGenre] = useState('Todos') // gênero selecionado (pelo nome)
  const [genreList, setGenreList] = useState([]) // gêneros da API: [{ id, name }]
  const [catalogBooks, setCatalogBooks] = useState(null) // null = ainda buscando

  // Carrega os gêneros uma vez, ao abrir a página, para montar os chips de filtro.
  useEffect(() => {
    libraryService.getGenres().then(setGenreList).catch(() => {})
  }, [])

  // Chips: "Todos" + nomes dos gêneros. useMemo evita recriar o array a cada renderização.
  const genreChips = useMemo(() => ['Todos', ...genreList.map((item) => item.name)], [genreList])
  // A API filtra por id, mas a tela guarda o nome; aqui converte (0 = sem filtro).
  const genreId = genreList.find((item) => item.name === genre)?.id || 0

  // Busca com "debounce": espera 300 ms depois da última tecla antes de consultar a API.
  // - A função de limpeza cancela o timer anterior quando o usuário continua digitando.
  // - A flag `active` ignora respostas antigas que chegarem depois de uma busca mais nova.
  useEffect(() => {
    let active = true
    const timer = window.setTimeout(() => {
      libraryService.searchCatalog(query.trim(), genreId)
        .then((result) => { if (active) setCatalogBooks(result) })
        .catch((cause) => { if (active) { setCatalogBooks([]); feedback(cause.message) } })
    }, 300)
    return () => { active = false; window.clearTimeout(timer) }
  }, [query, genreId, feedback])

  const results = catalogBooks ?? []

  return (
    <>
      <MobileGreeting />
      <section className="page-panel">
        <div className="page-title">
          <span className="eyebrow">ENCONTRE SUA PRÓXIMA LEITURA</span>
          <h1>Pesquisar</h1>
          <p>Explore livros e autores recomendados pela comunidade.</p>
        </div>
        <label className="large-search">
          <span>⌕</span>
          <input autoFocus placeholder="Título ou nome do autor..." value={query} onChange={(event) => setQuery(event.target.value)} />
          <kbd>Enter</kbd>
        </label>
        <div className="filter-row" aria-label="Filtrar por gênero">
          {genreChips.map((item) => (
            <button className={`filter-chip ${genre === item ? 'selected' : ''}`} key={item} onClick={() => setGenre(item)}>{item}</button>
          ))}
        </div>
        <div className="section-heading">
          <div>
            <h2>{query ? `Resultados para “${query}”` : 'Obras do catálogo'}</h2>
            <p>{catalogBooks === null ? 'Buscando…' : `${results.length} ${results.length === 1 ? 'livro encontrado' : 'livros encontrados'}`}</p>
          </div>
          <button className="outline-button" onClick={() => openForm('book')}>＋ Cadastrar livro</button>
        </div>
        {results.length
          ? <div className="book-grid">{results.map((book) => <BookCard key={book.id} book={book} onOpen={openBookDetail} />)}</div>
          : <div className="empty-state">Nenhum livro encontrado no catálogo. Tente outro termo.</div>}
      </section>
    </>
  )
}
