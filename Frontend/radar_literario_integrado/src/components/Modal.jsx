// Janela modal genérica (título, botão de fechar, conteúdo).
// Fecha de 3 formas: botão ×, tecla Esc ou clique fora da janela.
import { useEffect } from 'react'

export default function Modal({ title, children, onClose, wide = false }) {
  // Escuta a tecla Esc enquanto o modal existe.
  // A função de limpeza remove o ouvinte ao fechar, para não acumular ouvintes.
  useEffect(() => {
    const onKey = (event) => event.key === 'Escape' && onClose()
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [onClose])

  return (
    // Só fecha se o clique foi no fundo escuro (target === currentTarget), não dentro da janela.
    <div className="modal-backdrop" onMouseDown={(event) => event.target === event.currentTarget && onClose()}>
      {/* role="dialog" + aria-modal: leitores de tela tratam como diálogo */}
      <section className={`modal ${wide ? 'modal-wide' : ''}`} role="dialog" aria-modal="true" aria-labelledby="modal-title">
        <div className="modal-head">
          <h2 id="modal-title">{title}</h2>
          <button className="icon-button" onClick={onClose} aria-label="Fechar">×</button>
        </div>
        {children}
      </section>
    </div>
  )
}
