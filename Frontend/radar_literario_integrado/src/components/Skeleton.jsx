// [MOBILE] Representa o conteúdo enquanto App espera a sessão, catálogo, feed ou comentários.
// label informa o que está carregando e rows ajusta a quantidade de espaços reservados.
export default function Skeleton({ label, rows = 2 }) {
  // [MOBILE] role=status anuncia a espera e aria-hidden omite os blocos decorativos.
  return (
    <div className="skeleton" role="status">
      <span className="sr-only">{label}</span>
      <div className="skeleton-rows" aria-hidden="true">
        {Array.from({ length: rows }, (_, row) => <span key={row} />)}
      </div>
    </div>
  )
}
