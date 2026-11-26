// Comentários de uma publicação (lista + campo para comentar).
// Os comentários são carregados ao abrir, para a publicação escolhida (selectedPost).
import { useEffect, useState } from 'react'
import { useRadar } from '../../context/useRadar.js'
import { libraryService } from '../../services/libraryService.js'
import Modal from '../Modal.jsx'
import Avatar from '../Avatar.jsx'

export default function CommentsModal() {
  const { selectedPost, closeModal, refresh, attempt } = useRadar()
  const [comments, setComments] = useState([])
  const [text, setText] = useState('')

  // Busca os comentários da publicação ao abrir (e se a publicação mudar).
  useEffect(() => {
    if (!selectedPost) return
    libraryService.getComments(selectedPost.id).then(setComments).catch(() => setComments([]))
  }, [selectedPost])

  const submit = attempt(async (event) => {
    event.preventDefault()
    if (!text.trim()) return
    await libraryService.addComment(selectedPost.id, text.trim())
    setText('')
    // Recarrega a lista do modal e o feed (o contador de comentários muda).
    setComments(await libraryService.getComments(selectedPost.id))
    await refresh()
  })

  return (
    <Modal title="Comentários" wide onClose={closeModal}>
      <p className="comment-context">Sobre <strong>{selectedPost?.title}</strong> · por {selectedPost?.person}</p>
      <div className="comments-list">
        {comments.map((comment) => (
          <div className="comment" key={comment.id}>
            <Avatar size="small" />
            <div><strong>{comment.person}</strong><small>{comment.time}</small><p>{comment.text}</p></div>
          </div>
        ))}
      </div>
      <form className="comment-form" onSubmit={submit}>
        {/* Rótulo visível só para leitores de tela (classe sr-only) */}
        <label className="sr-only" htmlFor="comment-input">Escreva um comentário</label>
        <input id="comment-input" value={text} onChange={(event) => setText(event.target.value)} placeholder="Escreva uma mensagem..." maxLength="500" required />
        <button className="primary-button" aria-label="Enviar comentário">Enviar <span>→</span></button>
      </form>
    </Modal>
  )
}
