// "Fazer avaliação": nota de 1 a 5 estrelas + resenha.
// A API só aceita avaliar livros que estão "Em leitura" ou "Concluído" na estante.
import { useState } from 'react'
import { useRadar } from '../../context/useRadar.js'
import { libraryService } from '../../services/libraryService.js'
import Modal from '../Modal.jsx'
import Stars from '../Stars.jsx'

export default function ReviewModal() {
  const { selectedBook, books, closeModal, refresh, setPage, attempt, feedback } = useRadar()
  const [rating, setRating] = useState(0) // zera sozinho: o modal é recriado a cada abertura
  const [notice, setNotice] = useState('')

  const submit = attempt(async (event) => {
    event.preventDefault()
    const form = new FormData(event.currentTarget)
    const book = books.find((item) => item.title.toLowerCase() === String(form.get('title')).trim().toLowerCase())
    if (!book) { setNotice('Esse livro não está na estante. Cadastre-o antes de avaliá-lo.'); return }
    if (!rating) { setNotice('Escolha uma avaliação de uma a cinco estrelas.'); return }
    // addPost com rating > 0 cria uma resenha (POST /reviews) no libraryService.
    await libraryService.addPost({ title: book.title, author: book.author, bookId: book.id, rating, text: form.get('opinion') })
    await refresh()
    closeModal()
    setPage('home')
    feedback('Avaliação publicada.')
  })

  return (
    <Modal title="Fazer avaliação" onClose={closeModal}>
      <p className="modal-lead">Conte à comunidade o que achou da obra.</p>
      <form className="form-stack" onSubmit={submit}>
        <label>Nome do livro
          <input name="title" required defaultValue={selectedBook?.title || ''} list="review-book-options" placeholder="Livro" />
          <datalist id="review-book-options">{books.map((book) => <option key={book.id} value={book.title} />)}</datalist>
        </label>
        <div className="rating-field"><span>Sua avaliação</span><Stars rating={rating} interactive onChange={setRating} /></div>
        <label>Sua opinião<textarea name="opinion" required rows="4" maxLength="700" placeholder="Escreva uma resenha..." /></label>
        <label className="check-label"><input type="checkbox" /> Conteúdo sensível</label>
        <label className="check-label"><input type="checkbox" /> Contém spoiler</label>
        {notice && <p className="form-error" role="alert">{notice}</p>}
        <div className="modal-actions">
          <button type="button" className="outline-button" onClick={closeModal}>Cancelar</button>
          <button className="primary-button">Enviar avaliação</button>
        </div>
      </form>
    </Modal>
  )
}
