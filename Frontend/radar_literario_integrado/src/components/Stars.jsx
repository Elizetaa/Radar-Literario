// Estrelas de avaliação (1 a 5).
// - Somente leitura (padrão): só mostra a nota.
// - interactive: cada estrela vira botão e chama onChange(número da estrela).
export default function Stars({ rating, interactive = false, onChange }) {
  return (
    <span className={`stars ${interactive ? 'interactive' : ''}`} aria-label={`Avaliação: ${rating} de 5`}>
      {[1, 2, 3, 4, 5].map((star) => (
        <button
          type="button" // evita enviar o formulário ao clicar numa estrela
          key={star}
          aria-label={`${star} estrelas`}
          onClick={() => interactive && onChange(star)}
          disabled={!interactive}
          className={star <= rating ? 'filled' : ''}
        >★</button>
      ))}
    </span>
  )
}
