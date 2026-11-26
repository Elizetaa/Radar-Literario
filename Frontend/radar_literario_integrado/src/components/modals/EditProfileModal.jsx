// "Editar perfil": nome, apelido e bio (PATCH /me).
// A API valida o apelido (3 a 30 caracteres: letras minúsculas, números e "_").
import { useRadar } from '../../context/useRadar.js'
import { backendApi } from '../../services/backendApi.js'
import { toProfile } from '../../utils/profile.js'
import Modal from '../Modal.jsx'

export default function EditProfileModal() {
  const { profile, setProfile, closeModal, refresh, attempt, feedback } = useRadar()

  const submit = attempt(async (event) => {
    event.preventDefault()
    const form = new FormData(event.currentTarget)
    const saved = await backendApi.patchMe({
      name: form.get('name').trim(),
      nickname: form.get('handle').trim().replace(/^@/, ''), // aceita digitar com ou sem "@"
      biography: form.get('bio').trim(),
    })
    setProfile(toProfile(saved?.data || saved))
    await refresh() // os posts mostram o apelido do autor, então recarregamos o feed
    closeModal()
    feedback('Perfil atualizado.')
  })

  return (
    <Modal title="Editar perfil" onClose={closeModal}>
      <p className="modal-lead">Atualize as informações que aparecem no seu perfil.</p>
      <form className="form-stack" onSubmit={submit}>
        <label>Nome<input name="name" defaultValue={profile.name} required maxLength="60" /></label>
        <label>Nickname<input name="handle" defaultValue={profile.handle} required maxLength="30" /></label>
        <label>Sobre você<textarea name="bio" rows="3" defaultValue={profile.bio} maxLength="240" /></label>
        <div className="modal-actions">
          <button type="button" className="outline-button" onClick={closeModal}>Cancelar</button>
          <button className="primary-button">Salvar alterações</button>
        </div>
      </form>
    </Modal>
  )
}
