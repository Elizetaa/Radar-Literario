// Confirmação para remover um livro da estante (DELETE /me/shelf/:id).
import { useRadar } from '../../context/useRadar.js'
import { libraryService } from '../../services/libraryService.js'
import Modal from '../Modal.jsx'

export default function DeleteBookModal() {
  const { selectedBook, closeModal, refresh, attempt, feedback } = useRadar()

  const remove = attempt(async () => {
    await libraryService.deleteBook(selectedBook.id)
    await refresh()
    closeModal()
    feedback('Livro removido da estante.')
  })

  return (
    <Modal title="Remover livro?" onClose={closeModal}>
      <div className="confirm-box"><span className="confirm-icon">!</span><p><strong>{selectedBook?.title}</strong> será removido da sua estante.</p></div>
      <div className="modal-actions">
        <button className="outline-button" onClick={closeModal}>Cancelar</button>
        <button className="danger-button" onClick={remove}>Remover livro</button>
      </div>
    </Modal>
  )
}
