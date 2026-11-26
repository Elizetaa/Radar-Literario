// Confirmação de denúncia de uma publicação.
// A API registra a denúncia (POST /reports) e a publicação some do feed.
import { useRadar } from '../../context/useRadar.js'
import { libraryService } from '../../services/libraryService.js'
import Modal from '../Modal.jsx'

export default function ReportModal() {
  const { selectedPost, closeModal, refresh, attempt, feedback } = useRadar()

  const confirm = attempt(async () => {
    await libraryService.reportPost(selectedPost.id)
    await refresh()
    closeModal()
    feedback('Publicação removida do seu feed.')
  })

  return (
    <Modal title="Denunciar publicação" onClose={closeModal}>
      <div className="confirm-box">
        <span className="confirm-icon">!</span>
        <p>A publicação de <strong>{selectedPost?.person}</strong> sobre <strong>{selectedPost?.title}</strong> será removida do seu feed. Esta ação é simulada.</p>
      </div>
      <div className="modal-actions">
        <button className="outline-button" onClick={closeModal}>Cancelar</button>
        <button className="danger-button" onClick={confirm}>Confirmar denúncia</button>
      </div>
    </Modal>
  )
}
