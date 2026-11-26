// "Fazer diário de leitura": registra até que página o usuário leu + uma opinião.
// Gera uma publicação do tipo "diário" e a API atualiza o progresso na estante.
import { useState } from 'react'
import { useRadar } from '../../context/useRadar.js'
import { libraryService } from '../../services/libraryService.js'
import Modal from '../Modal.jsx'

export default function DiaryModal() {
  const { selectedBook, books, closeModal, refresh, setPage, attempt, feedback } = useRadar()
  const [notice, setNotice] = useState('') // erro de validação exibido no formulário

  const submit = attempt(async (event) => {
    event.preventDefault()
    const form = new FormData(event.currentTarget)
    // O campo "Nome do livro" é texto livre (com sugestões); aqui achamos o livro da estante pelo título.
    const book = books.find((item) => item.title.toLowerCase() === String(form.get('title')).trim().toLowerCase())
    if (!book) { setNotice('Esse livro não está na estante. Cadastre-o antes de registrar a leitura.'); return }
    // Não deixa passar do total de páginas da edição.
    const pages = Math.min(book.totalPages, Number(form.get('pages')))
    await libraryService.addPost({ title: book.title, author: book.author, bookId: book.id, text: form.get('opinion'), pages, totalPages: book.totalPages, rating: 0 })
    await refresh()
    closeModal()
    setPage('home')
    feedback('Diário de leitura registrado.')
  })

  return (
    <Modal title="Fazer diário de leitura" onClose={closeModal}>
      <p className="modal-lead">Registre seu progresso e compartilhe suas impressões.</p>
      <form className="form-stack" onSubmit={submit}>
        <label>Nome do livro
          {/* defaultValue vem do livro escolhido antes de abrir o modal (se houver) */}
          <input name="title" required defaultValue={selectedBook?.title || ''} placeholder="Livro" list="book-options" />
          <datalist id="book-options">{books.map((book) => <option key={book.id} value={book.title} />)}</datalist>
        </label>
        <label>Quantidade de páginas lidas
          <input name="pages" type="number" required min="1" max={selectedBook?.totalPages || undefined} placeholder="Páginas lidas até agora" />
        </label>
        <label>Sua opinião<textarea name="opinion" required rows="4" maxLength="700" placeholder="O que está achando da leitura?" /></label>
        {/* As duas caixas abaixo ainda não são enviadas à API (addPost não recebe esses campos) */}
        <label className="check-label"><input type="checkbox" name="sensitive" /> Conteúdo sensível</label>
        <label className="check-label"><input type="checkbox" name="spoiler" /> Contém spoiler</label>
        {notice && <p className="form-error" role="alert">{notice}</p>}
        <div className="modal-actions">
          <button type="button" className="outline-button" onClick={closeModal}>Cancelar</button>
          <button className="primary-button">Enviar diário</button>
        </div>
      </form>
    </Modal>
  )
}
