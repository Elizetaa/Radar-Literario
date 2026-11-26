// Menu "O que vamos fazer?", aberto pelo botão "＋" da barra inferior (celular).
// Cada opção fecha este modal e abre o formulário correspondente.
import { useRadar } from '../../context/useRadar.js'
import Modal from '../Modal.jsx'

export default function ActionsModal() {
  const { closeModal, openForm } = useRadar()
  return (
    <Modal title="O que vamos fazer?" onClose={closeModal}>
      <div className="action-choices">
        <button onClick={() => openForm('book')}><span>＋</span><strong>Cadastrar livro</strong><small>Adicione uma obra à estante.</small></button>
        <button onClick={() => openForm('diary')}><span>▤</span><strong>Registrar leitura</strong><small>Compartilhe seu progresso.</small></button>
        <button onClick={() => openForm('review')}><span>★</span><strong>Avaliar livro</strong><small>Deixe sua opinião.</small></button>
      </div>
    </Modal>
  )
}
