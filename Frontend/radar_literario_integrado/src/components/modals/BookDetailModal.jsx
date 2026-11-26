// Detalhe de um livro (capa, autor, nota, sinopse).
// Os botões mudam conforme o livro já esteja ou não na estante:
//  - na estante: "Avaliar livro" e "Registrar leitura"
//  - fora dela: "Adicionar à estante" (avaliar/registrar exigem que o livro esteja lá; a API recusa senão)
import { useRadar } from '../../context/useRadar.js'
import { libraryService } from '../../services/libraryService.js'
import Modal from '../Modal.jsx'
import Cover from '../Cover.jsx'

export default function BookDetailModal() {
  const { selectedBook: book, books, closeModal, openForm, refresh, setPage, attempt, feedback } = useRadar()
  if (!book) return null

  const onShelf = books.some((item) => item.id === book.id)

  // Adiciona à estante (status "planejado") e leva o usuário para a página da estante.
  const addToShelf = attempt(async () => {
    await libraryService.addBook({ workId: book.id })
    await refresh()
    closeModal()
    setPage('shelf')
    feedback('Livro adicionado à sua estante.')
  })

  return (
    <Modal title={book.title} onClose={closeModal}>
      <div className="book-detail">
        <Cover book={book} />
        <div>
          <span className="eyebrow">{book.genre}</span>
          <h3>{book.author}</h3>
          <div className="gold">★ {book.rating || '—'} <span className="muted">· {book.totalPages} páginas</span></div>
          <p>{book.description}</p>
          <div className="modal-actions">
            {onShelf ? (
              <>
                <button className="outline-button" onClick={() => openForm('review', book)}>Avaliar livro</button>
                <button className="primary-button" onClick={() => openForm('diary', book)}>Registrar leitura</button>
              </>
            ) : <button className="primary-button" onClick={addToShelf}>Adicionar à estante</button>}
          </div>
        </div>
      </div>
    </Modal>
  )
}
