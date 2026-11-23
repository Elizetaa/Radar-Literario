// [MOBILE] Os detalhes e formulários de App usam o mesmo diálogo no desktop e no celular.
// title nomeia o painel, onClose encerra o fluxo, returnFocus aponta para o acionador e toast anuncia erros.
import { useEffect, useId, useLayoutEffect, useRef } from 'react'

export default function Modal({ title, children, onClose, toast, returnFocus, wide = false }) {
  const dialog = useRef(null)
  const panel = useRef(null)
  const drag = useRef(null)
  const titleId = useId()

  // [MOBILE] showModal prende o foco e torna o fundo inerte. A referência
  // do botão de origem funciona também em navegadores que o toque não focaliza.
  useLayoutEffect(() => {
    const element = dialog.current
    const opener = returnFocus.current
    element.showModal()
    element.querySelector('.icon-button').focus()
    return () => {
      element.close()
      // Aguarda a troca de painel. Só restaura o foco se toda a sequência acabou.
      queueMicrotask(() => {
        if (opener?.isConnected && !document.querySelector('dialog[open]')) {
          opener.focus({ preventScroll: true })
        }
      })
    }
  }, [returnFocus])

  // [MOBILE] O teclado pode encolher apenas a área visível no Safari. Acompanha
  // essa área enquanto o diálogo estiver aberto e remove os listeners ao fechar.
  useEffect(() => {
    const viewport = window.visualViewport
    if (!viewport) return
    const element = dialog.current
    function syncViewport() {
      if (viewport.scale !== 1) {
        element.style.removeProperty('--dialog-height')
        element.style.removeProperty('--dialog-top')
        return
      }
      element.style.setProperty('--dialog-height', `${viewport.height}px`)
      element.style.setProperty('--dialog-top', `${viewport.offsetTop}px`)
    }
    syncViewport()
    viewport.addEventListener('resize', syncViewport)
    viewport.addEventListener('scroll', syncViewport)
    return () => {
      viewport.removeEventListener('resize', syncViewport)
      viewport.removeEventListener('scroll', syncViewport)
    }
  }, [])

  // [MOBILE] Só a alça captura o gesto. O conteúdo preserva a rolagem vertical
  // e um deslocamento curto ou cancelado devolve a folha à posição inicial.
  function startDrag(event) {
    if (!event.isPrimary || event.button !== 0) return
    drag.current = { id: event.pointerId, y: event.clientY }
    event.currentTarget.setPointerCapture(event.pointerId)
  }

  function moveDrag(event) {
    if (drag.current?.id !== event.pointerId) return
    const distance = Math.max(0, event.clientY - drag.current.y)
    panel.current.style.setProperty('--sheet-drag', `${distance}px`)
  }

  function finishDrag(event, cancelled = false) {
    if (drag.current?.id !== event.pointerId) return
    const distance = event.clientY - drag.current.y
    drag.current = null
    panel.current.style.removeProperty('--sheet-drag')
    if (!cancelled && distance >= 80) onClose()
  }

  // [MOBILE] aria-labelledby dá ao diálogo nativo o nome do título visível.
  // aria-hidden tira a alça da leitura assistiva porque o botão Fechar faz a mesma ação.
  return (
    <dialog
      ref={dialog}
      className="modal-backdrop"
      aria-labelledby={titleId}
      onCancel={(event) => { event.preventDefault(); onClose() }}
      onClick={(event) => { if (event.target === event.currentTarget) onClose() }}
    >
      <section ref={panel} className={`modal ${wide ? 'modal-wide' : ''}`}>
        <div
          className="sheet-handle"
          aria-hidden="true"
          onPointerDown={startDrag}
          onPointerMove={moveDrag}
          onPointerUp={finishDrag}
          onPointerCancel={(event) => finishDrag(event, true)}
          onLostPointerCapture={(event) => finishDrag(event, true)}
        />
        <div className="modal-head">
          <h2 id={titleId}>{title}</h2>
          {/* [MOBILE] O botão recebe o foco inicial sem abrir o teclado virtual. */}
          <button type="button" className="icon-button" onClick={onClose} aria-label="Fechar">×</button>
        </div>
        {children}
      </section>
      {/* [MOBILE] role=status anuncia a mensagem dentro da camada do diálogo. */}
      {toast && <div className="toast" role="status">{toast}</div>}
    </dialog>
  )
}
