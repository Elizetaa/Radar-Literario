// [MOBILE] App fornece os destinos, a página atual e as ações usadas também no desktop.
// Em telefones o atalho de leitura fica separado da barra para liberar espaço aos cinco destinos.
export default function BottomNav({ items, page, onNavigate, onRead, onCreate }) {
  return (
    <>
      {/* [MOBILE] aria-label identifica a navegação e aria-current anuncia o destino ativo.
          aria-hidden evita que símbolos decorativos sejam lidos antes dos rótulos. */}
      <nav className="mobile-nav" aria-label="Navegação principal">
        {items.map(([id, label, icon]) => (
          <button
            key={id}
            className={page === id ? 'active' : ''}
            aria-current={page === id ? 'page' : undefined}
            onClick={() => onNavigate(id)}
          >
            <span aria-hidden="true">{icon}</span>
            {label}
          </button>
        ))}
        <button className="mobile-add" onClick={onCreate} aria-label="Criar">＋</button>
      </nav>
      <button className="primary-button mobile-read" onClick={onRead}>＋ Registrar leitura</button>
    </>
  )
}
