// Capa ilustrada do livro: um gradiente com o título.
// Ainda não usa a imagem real da edição (image_url); o gradiente depende do gênero.
// className: ex. 'mini-cover' para a versão pequena dentro dos posts.
export default function Cover({ book, className = '' }) {
  const palettes = book.genre === 'Romance' ? ['#ba6e83', '#523445'] : ['#697e72', '#243b36']
  return (
    // As cores entram como variáveis CSS (--cover-start/--cover-end), lidas pelo App.css.
    <div className={`cover ${className}`} style={{ '--cover-start': palettes[0], '--cover-end': palettes[1] }} aria-label={`Capa ilustrativa de ${book.title}`}>
      <span>{book.title}</span>
    </div>
  )
}
