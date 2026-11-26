// Barra de navegação inferior (aparece só em tablet/celular; o CSS a esconde no desktop).
//
// Diferenças em relação ao menu lateral:
//  - "Sugestões" fica no CENTRO da barra e é destacada em roxo (classe mobile-featured).
//  - Um botão flutuante "Diário" (canto inferior direito, acima da barra) abre o
//    formulário de diário de leitura de qualquer página.
import { useRadar } from '../context/useRadar.js'
import { navItems } from '../utils/constants.js'

// Página que ganha destaque e posição central na barra.
const FEATURED_ID = 'suggestions'

export default function MobileNav() {
  const { page, setPage, openForm } = useRadar()

  // Monta a ordem da barra: separa o item de destaque e o encaixa no meio dos outros.
  // Com 5 itens: [Início, Pesquisar, SUGESTÕES, Estante, Perfil]. O 3º de 5 fica no centro.
  const featured = navItems.find(([id]) => id === FEATURED_ID)
  const others = navItems.filter(([id]) => id !== FEATURED_ID)
  const middle = Math.floor(others.length / 2)
  const items = [...others.slice(0, middle), featured, ...others.slice(middle)]

  return (
    <>
      <nav className="mobile-nav" aria-label="Navegação principal">
        {items.map(([id, label, icon]) => {
          // 'active' = página atual; 'mobile-featured' = botão roxo de destaque
          const classes = [page === id ? 'active' : '', id === FEATURED_ID ? 'mobile-featured' : ''].filter(Boolean).join(' ')
          return (
            <button key={id} className={classes} aria-current={page === id ? 'page' : undefined} onClick={() => setPage(id)}>
              <span aria-hidden="true">{icon}</span>{label}
            </button>
          )
        })}
      </nav>

      {/* Botão flutuante do diário de leitura. Em "Sugestões" fica só com o ícone
          (classe mobile-read-compact) para não cobrir os botões do cartão. */}
      <button
        type="button"
        className={`primary-button mobile-read ${page === FEATURED_ID ? 'mobile-read-compact' : ''}`}
        onClick={() => openForm('diary')}
        aria-label="Fazer diário de leitura"
      >
        <span aria-hidden="true">+</span>
        <span className="mobile-read-label">Diário</span>
      </button>
    </>
  )
}