// Foto de perfil padrão: silhueta de usuário em SVG.
// A API ainda não tem upload de foto; quando tiver, basta receber uma URL
// e mostrar um <img> no lugar do SVG.
// size: '' (padrão) | 'small' | 'large' (classes do App.css).
export default function Avatar({ size = '' }) {
  return (
    // aria-hidden: é decorativo, o leitor de tela não precisa anunciar.
    <span className={`avatar ${size}`} aria-hidden="true">
      <svg viewBox="0 0 24 24" fill="currentColor">
        <circle cx="12" cy="8.5" r="4.2" />
        <path d="M3.5 22c0-4.9 3.8-8 8.5-8s8.5 3.1 8.5 8z" />
      </svg>
    </span>
  )
}
