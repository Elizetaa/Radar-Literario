// Página "Início": feed da comunidade (coluna principal) + "Continuar lendo" e "Para você" (coluna lateral).
import { useRadar } from '../context/useRadar.js'
import PostCard from '../components/PostCard.jsx'
import Cover from '../components/Cover.jsx'
import MobileGreeting from '../components/MobileGreeting.jsx'

export default function HomePage() {
  const { profile, posts, books, readingBooks, setPage, openForm, openBookDetail, openComments, openReport, likePost } = useRadar()
  const firstName = profile.name.split(' ')[0]

  return (
    <>
      <MobileGreeting />
      <div className="home-grid">
        {/* ---- coluna do feed ---- */}
        <section className="feed-column">
          <div className="welcome-row">
            <div>
              {/* Data de hoje por extenso, em maiúsculas (ex.: "SEGUNDA-FEIRA, 23 DE NOVEMBRO") */}
              <span className="eyebrow">{new Date().toLocaleDateString('pt-BR', { weekday: 'long', day: 'numeric', month: 'long' }).toUpperCase()}</span>
              <h1>Bom te ver por aqui, {firstName}.</h1>
              <p>Descubra o que a comunidade está lendo.</p>
            </div>
            <button className="outline-button" onClick={() => setPage('search')}>Explorar livros <span>→</span></button>
          </div>
          <div className="section-heading">
            <div><h2>Leituras da comunidade</h2><p>Novas histórias compartilhadas por leitores.</p></div>
            <button className="text-button" onClick={() => setPage('search')}>Ver tudo →</button>
          </div>
          <div className="feed-list">
            {posts.map((post) => <PostCard key={post.id} post={post} onLike={likePost} onReport={openReport} onComments={openComments} />)}
            {/* Estado vazio: conta nova ou banco sem publicações */}
            {posts.length === 0 && <div className="empty-state">Ainda não há publicações. Registre sua próxima leitura para começar!</div>}
          </div>
        </section>

        {/* ---- coluna lateral ---- */}
        <aside className="home-aside">
          <div className="aside-heading"><h2>Continuar lendo</h2><button className="text-button" onClick={() => setPage('shelf')}>Estante →</button></div>
          {/* Até 3 livros em andamento; clicar abre o diário já com o livro escolhido */}
          {readingBooks.length ? readingBooks.slice(0, 3).map((book) => (
            <button className="continue-card" key={book.id} onClick={() => openForm('diary', book)}>
              <Cover book={book} />
              <span>
                <strong>{book.title}</strong>
                <small>{book.author}</small>
                <span className="progress-track"><span style={{ width: `${Math.round(book.pages / book.totalPages * 100)}%` }} /></span>
                <small>{book.pages} de {book.totalPages} páginas</small>
              </span>
            </button>
          )) : <p className="muted">Sua estante está pronta para a próxima história.</p>}

          <div className="aside-divider" />
          <div className="aside-heading"><h2>Para você</h2><button className="text-button" onClick={() => setPage('suggestions')}>Ver mais →</button></div>
          {/* Livros da estante ainda não iniciados (0 páginas). Clicar abre o detalhe. */}
          {books.filter((book) => !book.pages).slice(0, 3).map((book) => (
            <button className="recommend-row" key={book.id} onClick={() => openBookDetail(book)}>
              <Cover book={book} />
              <span>
                <strong>{book.title}</strong>
                <small>{book.author}</small>
                <small className="gold">★ {book.rating} <span className="muted">· {book.genre}</span></small>
              </span>
            </button>
          ))}
        </aside>
      </div>
    </>
  )
}
