// Página "Minha estante": só os livros do usuário, com estatísticas e filtros.
// Os filtros são de dois tipos: situação (Em leitura, Concluídos...) ou gênero
// (escolhido no menu lateral "Suas leituras").
import { useRadar } from '../context/useRadar.js'
import { shelfStatusFilters } from '../utils/constants.js'
import BookCard from '../components/BookCard.jsx'
import MobileGreeting from '../components/MobileGreeting.jsx'

// Decide se um livro aparece para o filtro escolhido.
function matchesFilter(book, filter) {
  if (filter === 'Todos') return true
  if (filter === 'Tenho interesse') return book.status === 'interested'
  if (filter === 'Em leitura') return book.pages > 0 && book.pages < book.totalPages
  if (filter === 'Quero ler') return book.pages === 0 && book.status !== 'interested'
  if (filter === 'Concluídos') return book.pages >= book.totalPages
  return book.genre === filter // qualquer outro valor é o nome de um gênero
}

export default function ShelfPage() {
  const { books, readingBooks, shelfFilter, setShelfFilter, openForm, openBookDetail } = useRadar()
  const shelfBooks = books.filter((book) => matchesFilter(book, shelfFilter))

  return (
    <>
      <MobileGreeting />
      <section className="page-panel">
        <div className="page-title shelf-title">
          <div>
            <span className="eyebrow">SEU CANTINHO DE LEITURA</span>
            <h1>Minha estante</h1>
            <p>Acompanhe suas leituras e encontre tudo em um só lugar.</p>
          </div>
          <button className="primary-button" onClick={() => openForm('book')}>＋ Cadastrar livro</button>
        </div>
        <div className="shelf-stats">
          <span><strong>{books.length}</strong> livros</span>
          <span><strong>{readingBooks.length}</strong> em leitura</span>
          <span><strong>{books.filter((book) => book.pages >= book.totalPages).length}</strong> concluídos</span>
        </div>
        <div className="filter-row">
          {shelfStatusFilters.map((item) => (
            <button key={item} className={`filter-chip ${shelfFilter === item ? 'selected' : ''}`} onClick={() => setShelfFilter(item)}>{item}</button>
          ))}
        </div>
        {shelfBooks.length ? (
          <div className="book-grid">
            {shelfBooks.map((book) => (
              <div className="shelf-book" key={book.id}>
                <BookCard book={book} onOpen={openBookDetail} />
                <div className="shelf-book-actions">
                  <button onClick={() => openForm('diary', book)}>Registrar progresso</button>
                  {/* openForm guarda o livro escolhido e abre o modal de confirmação, que lê esse livro */}
                  <button onClick={() => openForm('deleteBook', book)}>Remover</button>
                </div>
              </div>
            ))}
          </div>
        ) : <div className="empty-state">Nenhum livro nesta categoria ainda.</div>}
      </section>
    </>
  )
}
