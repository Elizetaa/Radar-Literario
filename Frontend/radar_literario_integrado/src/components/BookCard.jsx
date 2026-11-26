// Cartão de livro usado em grades (pesquisa, estante, destaques).
// Mostra capa, título, autor, nota, gênero e, se já começou a ler, a barra de progresso.
// onOpen(book): chamado ao clicar (normalmente abre o detalhe do livro).
import Cover from './Cover.jsx'

export default function BookCard({ book, onOpen }) {
  // Porcentagem lida, limitada a 100 para não estourar a barra.
  const progress = Math.min(100, Math.round(book.pages / book.totalPages * 100))
  return (
    <button className="book-card" onClick={() => onOpen(book)}>
      <Cover book={book} />
      <span className="book-info">
        <strong>{book.title}</strong>
        <small>{book.author}</small>
        {book.status === 'interested' && <small className="font-semibold text-purple-700">Tenho interesse</small>}
        <span className="book-meta">
          {book.rating > 0 && <><span className="gold">★</span> {book.rating}</>}
          <span>{book.genre}</span>
        </span>
        {/* Barra e contagem só aparecem para livros já iniciados */}
        {book.pages > 0 && <span className="progress-track"><span style={{ width: `${progress}%` }} /></span>}
        {book.pages > 0 && <small>{book.pages}/{book.totalPages} páginas · {progress}%</small>}
      </span>
    </button>
  )
}
