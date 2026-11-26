// Confirmação de exclusão de conta. ATENÇÃO: ainda é SIMULADA — nada é apagado no servidor.
// (A API exige a senha em DELETE /me; falta criar o campo de senha neste modal.)
import { useRadar } from '../../context/useRadar.js'
import Modal from '../Modal.jsx'

export default function DeleteAccountModal() {
  const { closeModal, setPage, feedback } = useRadar()

  return (
    <Modal title="Excluir conta" onClose={closeModal}>
      <div className="confirm-box">
        <span className="confirm-icon">!</span>
        <p>Tem certeza que deseja excluir sua conta? Esta ação é irreversível. No protótipo, a confirmação encerra apenas esta sessão simulada.</p>
      </div>
      <div className="modal-actions">
        <button className="outline-button" onClick={closeModal}>Cancelar</button>
        <button className="danger-button" onClick={() => { closeModal(); setPage('home'); feedback('Exclusão simulada. Nenhum dado real foi apagado.') }}>Sim, excluir conta</button>
      </div>
    </Modal>
  )
}
