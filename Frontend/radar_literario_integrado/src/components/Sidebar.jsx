// Menu lateral (desktop): páginas principais + gêneros da estante + atalho de diário.
import { useRadar } from '../context/useRadar.js'
import { navItems } from '../utils/constants.js'

export default function Sidebar() {
  const { page, setPage, shelfGenres, shelfFilter, setShelfFilter, openForm } = useRadar()
  return (
    <aside className="sidebar">
      <div className="sidebar-caption">MENU</div>
      {navItems.map(([id, label, icon]) => (
        <button key={id} className={`nav-item ${page === id ? 'active' : ''}`} onClick={() => setPage(id)}>
          <span className="nav-icon">{icon}</span>{label}
        </button>
      ))}
      <div className="sidebar-rule" />
      <div className="sidebar-caption">SUAS LEITURAS</div>
      {/* Só mostra gêneros dos livros que estão na estante (shelfGenres vem do contexto).
          Clicar abre a estante já filtrada por aquele gênero. */}
      {shelfGenres.length ? shelfGenres.map((item, index) => (
        <button
          key={item}
          className={`nav-item ${page === 'shelf' && shelfFilter === item ? 'active' : ''}`}
          onClick={() => { setPage('shelf'); setShelfFilter(item) }}
        >
          {/* Alterna a cor da bolinha entre violeta e rosa */}
          <span className={`genre-dot ${index % 2 ? 'pink' : 'violet'}`} /> {item}
        </button>
      )) : <p className="sidebar-empty">Seus gêneros aparecem aqui quando houver livros na estante.</p>}
      <button className="nav-item" onClick={() => openForm('diary')}>＋ Registrar leitura</button>
      <div className="sidebar-bottom">
        <div className="tip-icon">✦</div>
        <strong>Leia, acompanhe,<br />compartilhe.</strong>
        <p>Um lugar para cada história.</p>
      </div>
    </aside>
  )
}
