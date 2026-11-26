// Página "Sugestões": dois modos.
//  - "Indicações para você": cartões que se arrastam (componente BookDiscovery).
//  - "Destaques da comunidade": grades de livros mais populares e mais bem avaliados.
import { useState } from 'react'
import { useRadar } from '../context/useRadar.js'
import BookDiscovery from '../features/book-discovery/BookDiscovery.jsx'
import BookCard from '../components/BookCard.jsx'
import MobileGreeting from '../components/MobileGreeting.jsx'

export default function SuggestionsPage() {
  const { books, recommendations, saveInterest, openInterestedShelf, setPage, openBookDetail } = useRadar()
  const [view, setView] = useState('discover') // 'discover' | 'highlights'

  // Livros da estante que têm nota, do melhor para o pior (base dos destaques).
  const rankedBooks = [...books].filter((book) => book.rating).sort((a, b) => b.rating - a.rating)

  return (
    <>
      {/* No modo "cartões" a saudação é omitida: o cartão ocupa a tela toda no celular */}
      {view !== 'discover' && <MobileGreeting />}
      {/* No modo "cartões" a página recebe a classe suggestions-immersive: no celular o CSS
          trava a rolagem (a tela inteira vira o cartão) para dar mais imersão. */}
      <section className={`page-panel ${view === 'discover' ? 'suggestions-immersive' : ''}`}>
        <div className="filter-row" aria-label="Tipo de sugestões">
          <button type="button" className={`filter-chip ${view === 'discover' ? 'selected' : ''}`} aria-pressed={view === 'discover'} onClick={() => setView('discover')}>Indicações para você</button>
          <button type="button" className={`filter-chip ${view === 'highlights' ? 'selected' : ''}`} aria-pressed={view === 'highlights'} onClick={() => setView('highlights')}>Destaques da comunidade</button>
        </div>

        {view === 'discover' ? (
          // recommendations === null significa "ainda carregando" (diferente de lista vazia)
          recommendations === null
            ? <div className="empty-state" role="status">Carregando indicações...</div>
            : <BookDiscovery
                // Esconde o que já está na estante, mesmo que a API ainda o indique
                books={recommendations.filter((candidate) => !books.some((book) => book.id === candidate.id))}
                onInterest={saveInterest}
                onOpenShelf={openInterestedShelf}
              />
        ) : (
          <section className="page-panel">
            <div className="page-title">
              <span className="eyebrow">ESCOLHIDOS POR LEITORES</span>
              <h1>Descubra histórias</h1>
              <p>Recomendações e destaques da comunidade do Radar Literário.</p>
            </div>
            <div className="feature-banner">
              <div>
                <span className="eyebrow">EM DESTAQUE</span>
                <h2>Histórias que ficam<br />com você.</h2>
                <p>Obras queridas pelos leitores, reunidas em um só lugar.</p>
                <button className="light-button" onClick={() => setPage('search')}>Explorar catálogo <span>→</span></button>
              </div>
              <span className="feature-spark">✳</span>
            </div>
            <div className="section-heading"><div><h2>Mais populares</h2><p>As obras que a comunidade está lendo.</p></div></div>
            <div className="book-grid">{rankedBooks.map((book) => <BookCard key={book.id} book={book} onOpen={openBookDetail} />)}</div>
            <div className="section-heading section-gap"><div><h2>Mais bem avaliados</h2><p>Favoritos dos leitores.</p></div></div>
            {/* Mesma lista, em ordem invertida */}
            <div className="book-grid">{[...rankedBooks].reverse().map((book) => <BookCard key={book.id} book={book} onOpen={openBookDetail} />)}</div>
          </section>
        )}
      </section>
    </>
  )
}