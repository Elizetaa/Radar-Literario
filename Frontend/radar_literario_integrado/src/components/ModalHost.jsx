// ModalHost: escolhe qual modal desenhar, de acordo com o valor `modal` do contexto.
// Só um modal fica aberto por vez; cada modal é recriado ao abrir, então seus
// campos locais (nota, texto, resultados) começam sempre limpos.
// Para criar um modal novo: faça o componente em components/modals/ e registre-o aqui.
import { useRadar } from '../context/useRadar.js'
import BookDetailModal from './modals/BookDetailModal.jsx'
import AddBookModal from './modals/AddBookModal.jsx'
import DiaryModal from './modals/DiaryModal.jsx'
import ReviewModal from './modals/ReviewModal.jsx'
import CommentsModal from './modals/CommentsModal.jsx'
import ReportModal from './modals/ReportModal.jsx'
import DeleteBookModal from './modals/DeleteBookModal.jsx'
import DeleteAccountModal from './modals/DeleteAccountModal.jsx'
import ActionsModal from './modals/ActionsModal.jsx'
import EditProfileModal from './modals/EditProfileModal.jsx'

// nome do modal (valor de `modal` no contexto) -> componente
const modals = {
  bookDetail: BookDetailModal,
  book: AddBookModal,
  diary: DiaryModal,
  review: ReviewModal,
  comments: CommentsModal,
  report: ReportModal,
  deleteBook: DeleteBookModal,
  deleteAccount: DeleteAccountModal,
  actions: ActionsModal,
  editProfile: EditProfileModal,
}

export default function ModalHost() {
  const { modal } = useRadar()
  const Current = modals[modal]
  return Current ? <Current /> : null
}
