import { useEffect, useRef, useState } from 'react'
import './BookDiscovery.css'

const coverPalettes = {
  Fantasia: ['#443254', '#947394', '#e3c48c'],
  Romance: ['#82485f', '#d496a1', '#ffe5bc'],
  Ficção: ['#294c58', '#7da0a4', '#e2d1a6'],
  Mistério: ['#253b3b', '#779385', '#e9cf9c'],
}

function DiscoveryIcon({ name, className = '' }) {
  const paths = {
    heart: <path d="M20.8 4.6a5.5 5.5 0 0 0-7.8 0L12 5.7l-1.1-1.1a5.5 5.5 0 0 0-7.8 7.8L12 21l8.8-8.6a5.5 5.5 0 0 0 0-7.8Z" />,
    close: <path d="m6 6 12 12M18 6 6 18" />,
    book: <><path d="M12 5c-3-2-6-2-9-1v15c3-1 6-1 9 1 3-2 6-2 9-1V4c-3-1-6-1-9 1Z" /><path d="M12 5v15" /></>,
    arrow: <path d="M4 12h16m-6-6 6 6-6 6" />,
    check: <path d="m5 12 4 4L19 6" />,
  }
  return <svg className={className} viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" strokeWidth="1.7" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">{paths[name]}</svg>
}

function IllustratedCover({ book }) {
  const palette = coverPalettes[book.genre] || coverPalettes.Fantasia
  return <div className="bd-cover" style={{ '--bd-cover-dark': palette[0], '--bd-cover-light': palette[1], '--bd-cover-accent': palette[2] }} aria-hidden="true">
    <div className="bd-cover-orbit bd-cover-orbit-one" />
    <div className="bd-cover-orbit bd-cover-orbit-two" />
    <div className="bd-cover-stars"><span>✦</span><span>✧</span><span>✦</span></div>
    <span className="bd-cover-caption">UMA HISTÓRIA PARA DESCOBRIR</span>
    <strong className="bd-cover-title">{book.title}</strong>
    <span className="bd-cover-divider" />
    <span className="bd-cover-author">{book.author}</span>
    <span className="bd-cover-edition">RADAR LITERÁRIO · CAPA ILUSTRATIVA</span>
  </div>
}

export default function BookDiscovery({ books = [], onInterest, onOpenShelf }) {
  const [seenIds, setSeenIds] = useState(() => new Set())
  const [savedCount, setSavedCount] = useState(0)
  const [saving, setSaving] = useState(false)
  const [pendingBook, setPendingBook] = useState(null)
  const [dragX, setDragX] = useState(0)
  const [dragging, setDragging] = useState(false)
  const [message, setMessage] = useState({ type: '', text: '' })
  const actionLocked = useRef(false)
  const pointer = useRef(null)
  const mounted = useRef(true)
  const endHeading = useRef(null)
  const remainingBooks = books.filter((book) => !seenIds.has(book.id))
  const currentBook = pendingBook || remainingBooks[0]
  const totalBooks = seenIds.size + remainingBooks.length + (pendingBook && !remainingBooks.some((book) => book.id === pendingBook.id) ? 1 : 0)
  const currentPosition = Math.min(seenIds.size + 1, totalBooks)
  const dragDirection = dragX > 26 ? 'right' : dragX < -26 ? 'left' : ''

  useEffect(() => {
    mounted.current = true
    return () => { mounted.current = false }
  }, [])

  useEffect(() => {
    if (!currentBook && seenIds.size > 0) endHeading.current?.focus()
  }, [currentBook, seenIds.size])

  function clearDrag(element) {
    const pointerId = pointer.current?.id
    pointer.current = null
    if (pointerId !== undefined && element?.hasPointerCapture(pointerId)) element.releasePointerCapture(pointerId)
    setDragX(0)
    setDragging(false)
  }

  async function choose(direction) {
    if (!currentBook || actionLocked.current) return
    actionLocked.current = true
    clearDrag(pointer.current?.element)

    if (direction === 'left') {
      setSeenIds((previous) => new Set(previous).add(currentBook.id))
      setMessage({ type: 'info', text: remainingBooks[1] ? `Próxima indicação: ${remainingBooks[1].title}.` : 'Você viu todas as indicações desta seleção.' })
      actionLocked.current = false
      return
    }

    setSaving(true)
    setPendingBook(currentBook)
    setMessage({ type: 'info', text: `Salvando “${currentBook.title}” na sua estante…` })
    try {
      await onInterest(currentBook)
      if (!mounted.current) return
      setSeenIds((previous) => new Set(previous).add(currentBook.id))
      setSavedCount((previous) => previous + 1)
      setMessage({ type: 'success', text: `“${currentBook.title}” salvo como “Tenho interesse” na sua estante.` })
    } catch {
      if (!mounted.current) return
      setMessage({ type: 'error', text: 'Não foi possível salvar este livro. Tente “Quero ler” novamente.' })
    } finally {
      actionLocked.current = false
      if (mounted.current) {
        setSaving(false)
        setPendingBook(null)
      }
    }
  }

  function startDrag(event) {
    if (actionLocked.current || pointer.current || (event.pointerType === 'mouse' && event.button !== 0) || event.isPrimary === false) return
    pointer.current = { id: event.pointerId, x: event.clientX, y: event.clientY, element: event.currentTarget }
    event.currentTarget.setPointerCapture(event.pointerId)
    setDragging(true)
  }

  function moveDrag(event) {
    if (pointer.current?.id !== event.pointerId || actionLocked.current) return
    const offsetX = event.clientX - pointer.current.x
    const offsetY = event.clientY - pointer.current.y
    if (Math.abs(offsetX) > Math.abs(offsetY)) setDragX(Math.max(-190, Math.min(190, offsetX)))
    else setDragX(0)
  }

  function finishDrag(event) {
    if (pointer.current?.id !== event.pointerId) return
    const offsetX = event.clientX - pointer.current.x
    const offsetY = event.clientY - pointer.current.y
    clearDrag(event.currentTarget)
    if (Math.abs(offsetX) >= 85 && Math.abs(offsetX) > Math.abs(offsetY) * 1.2) void choose(offsetX > 0 ? 'right' : 'left')
  }

  function cancelDrag(event) {
    if (pointer.current?.id === event.pointerId) clearDrag(event.currentTarget)
  }

  function handleKeyDown(event) {
    if (event.target !== event.currentTarget || event.altKey || event.ctrlKey || event.metaKey || event.shiftKey || event.repeat) return
    if (event.key !== 'ArrowLeft' && event.key !== 'ArrowRight') return
    event.preventDefault()
    if (pointer.current) clearDrag(event.currentTarget)
    void choose(event.key === 'ArrowRight' ? 'right' : 'left')
  }

  return <section className="book-discovery" aria-labelledby="bd-page-title">
    <div className="bd-heading">
      <div><span className="bd-eyebrow">UM ENCONTRO COM A SUA PRÓXIMA HISTÓRIA</span><h1 id="bd-page-title">Sua próxima leitura</h1><p>Às vezes, só falta um livro para despertar a vontade de ler.</p></div>
      <button type="button" className="bd-shelf-link" onClick={onOpenShelf}><DiscoveryIcon name="book" /> Minha estante <DiscoveryIcon name="arrow" /></button>
    </div>

    <div className="bd-layout">
      <div className="bd-main">
        {currentBook ? <>
          <div className="bd-progress-heading"><span><span className="bd-live-dot" /> INDICAÇÕES PARA VOCÊ</span><span>{currentPosition} <span className="bd-progress-total">de {totalBooks}</span></span></div>
          <div className="bd-progress" role="progressbar" aria-label="Livro apresentado nesta seleção" aria-valuemin={0} aria-valuemax={totalBooks} aria-valuenow={currentPosition}><span style={{ width: `${currentPosition / totalBooks * 100}%` }} /></div>
          <div className="bd-card-stage">
            {remainingBooks.length > 1 && <div className="bd-card-underlay" aria-hidden="true" />}
            <article className={`bd-card ${dragging ? 'bd-card-dragging' : ''} ${saving ? 'bd-card-saving' : ''}`} style={{ transform: `translateX(${dragX}px) rotate(${dragX / 24}deg)` }} tabIndex={0} role="group" aria-labelledby="bd-book-title" aria-describedby="bd-keyboard-hint" aria-keyshortcuts="ArrowLeft ArrowRight" aria-busy={saving} onPointerDown={startDrag} onPointerMove={moveDrag} onPointerUp={finishDrag} onPointerCancel={cancelDrag} onLostPointerCapture={cancelDrag} onKeyDown={handleKeyDown}>
              <div className="bd-card-face" key={currentBook.id}>
                <div className="bd-artwork"><IllustratedCover book={currentBook} /><span className="bd-genre">{currentBook.genre}</span></div>
                <div className="bd-book-details"><div className="bd-book-topline"><span className="bd-book-category">SUA PRÓXIMA DESCOBERTA</span>{Number(currentBook.rating) > 0 && <span className="bd-rating" aria-label={`Avaliação: ${currentBook.rating} de 5`}><span aria-hidden="true">★</span> {Number(currentBook.rating).toLocaleString('pt-BR')}</span>}</div><h2 id="bd-book-title">{currentBook.title}</h2><p className="bd-author">por {currentBook.author}</p><p className="bd-description">{currentBook.description || 'Uma nova história espera por você. Que tal dar uma chance a esta leitura?'}</p><span className="bd-page-count"><DiscoveryIcon name="book" /> {currentBook.totalPages} páginas</span></div>
              </div>
              {dragDirection && <span className={`bd-drag-stamp bd-drag-stamp-${dragDirection}`} aria-hidden="true">{dragDirection === 'right' ? 'QUERO LER' : 'NÃO QUERO LER'}</span>}
              {saving && <span className="bd-saving-indicator" aria-hidden="true"><span className="bd-spinner" /> Salvando na estante…</span>}
            </article>
          </div>
          <div className="bd-actions"><button type="button" className="bd-choice bd-choice-skip" onClick={() => void choose('left')} disabled={saving}><span className="bd-choice-icon"><DiscoveryIcon name="close" /></span><span>Não quero ler</span></button><button type="button" className="bd-choice bd-choice-save" onClick={() => void choose('right')} disabled={saving}><span className="bd-choice-icon">{saving ? <span className="bd-spinner" /> : <DiscoveryIcon name="heart" />}</span><span>{saving ? 'Salvando…' : 'Quero ler'}</span></button></div>
          <p className="bd-keyboard-hint" id="bd-keyboard-hint">Arraste o livro ou use os botões.<span> Com o cartão em foco, use as setas <kbd>←</kbd> <kbd>→</kbd>.</span></p>
        </> : <div className="bd-complete"><span className="bd-complete-icon"><DiscoveryIcon name={totalBooks ? 'check' : 'book'} /></span><span className="bd-eyebrow">CADA HISTÓRIA, UMA POSSIBILIDADE</span><h2 ref={endHeading} tabIndex={-1}>{totalBooks ? 'Por hoje, é isso!' : 'Novas histórias vêm por aí'}</h2><p>{totalBooks ? 'Você chegou ao fim desta seleção. Suas próximas leituras estão esperando na estante.' : 'Ainda não há livros para indicar nesta seleção. Explore a sua estante enquanto isso.'}</p>{savedCount > 0 && <span className="bd-complete-count"><DiscoveryIcon name="heart" /> {savedCount} {savedCount === 1 ? 'livro salvo' : 'livros salvos'} como “Tenho interesse”</span>}<button type="button" className="bd-complete-button" onClick={onOpenShelf}>Ir para minha estante <DiscoveryIcon name="arrow" /></button></div>}
        <div className={`bd-status ${message.type ? `bd-status-${message.type}` : ''}`} role="status" aria-live="polite" aria-atomic="true">{message.type === 'success' && <DiscoveryIcon name="check" />}<span>{message.text}</span></div>
      </div>

      <aside className="bd-guide" aria-labelledby="bd-guide-title"><span className="bd-guide-spark" aria-hidden="true">✦</span><span className="bd-eyebrow">DEIXE A CURIOSIDADE GUIAR</span><h2 id="bd-guide-title">Deu vontade<br />de ler?</h2><p>Encontre histórias que combinam com o seu momento, um livro de cada vez.</p><div className="bd-guide-rule" /><div className="bd-guide-step"><span className="bd-guide-direction bd-guide-direction-left" aria-hidden="true">←</span><div><strong>Para a esquerda</strong><p>Não quero ler. É só passar para o próximo livro.</p></div></div><div className="bd-guide-step"><span className="bd-guide-direction bd-guide-direction-right" aria-hidden="true">→</span><div><strong>Para a direita</strong><p>Quero ler! O livro fica salvo na sua estante como <b>Tenho interesse</b>.</p></div></div><div className="bd-guide-footnote"><DiscoveryIcon name="heart" /><span>O próximo favorito pode estar a um gesto de distância.</span></div></aside>
    </div>
  </section>
}
