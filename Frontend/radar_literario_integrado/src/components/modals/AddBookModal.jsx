// "Adicionar livro": busca obras do CATÁLOGO e coloca uma na estante.
// (Livros novos só entram no catálogo pela administração; o leitor escolhe entre os existentes.)
import { useEffect, useState } from 'react'
import { useRadar } from '../../context/useRadar.js'
import { libraryService } from '../../services/libraryService.js'
import Modal from '../Modal.jsx'

export default function AddBookModal() {
  const { closeModal, refresh, setPage, setShelfFilter, attempt, feedback } = useRadar()
  const [text, setText] = useState('') // termo digitado
  const [results, setResults] = useState([]) // obras encontradas

  // Mesma técnica de "debounce" da página de pesquisa: espera 300 ms após a última tecla.
  // Com o campo vazio lista as obras do catálogo logo ao abrir o modal.
  useEffect(() => {
    let active = true
    const timer = window.setTimeout(() => {
      libraryService.searchCatalog(text.trim()).then((result) => { if (active) setResults(result) }).catch(() => {})
    }, 300)
    return () => { active = false; window.clearTimeout(timer) }
  }, [text])

  // Botão "Buscar" (para quem prefere clicar/Enter): busca na hora, sem esperar o timer.
  const search = attempt(async (event) => {
    event.preventDefault()
    setResults(await libraryService.searchCatalog(text.trim()))
  })

  const add = attempt(async (work) => {
    await libraryService.addBook({ workId: work.id })
    await refresh()
    closeModal()
    setShelfFilter('Todos')
    setPage('shelf')
    feedback('Livro adicionado à sua estante.')
  })

  return (
    <Modal title="Adicionar livro" onClose={closeModal}>
      <p className="modal-lead">Busque uma obra do catálogo para colocar na sua estante.</p>
      <form className="form-stack" onSubmit={search}>
        <label>Título<input value={text} onChange={(event) => setText(event.target.value)} placeholder="Ex.: O nome do vento" autoFocus /></label>
        <button className="primary-button">Buscar</button>
      </form>
      <div className="comments-list">
        {results.map((work) => (
          <div className="comment" key={work.id}>
            <div><strong>{work.title}</strong><small>{work.author}</small></div>
            <button type="button" className="outline-button" onClick={() => add(work)}>Adicionar</button>
          </div>
        ))}
        {results.length === 0 && <p className="modal-lead">Nenhum resultado ainda. Busque pelo título; se a obra não existir no catálogo, ela precisa ser cadastrada por um administrador.</p>}
      </div>
    </Modal>
  )
}
