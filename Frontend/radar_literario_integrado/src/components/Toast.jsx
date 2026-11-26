// Aviso temporário no canto da tela (ex.: "Avaliação publicada.").
// O texto vem do contexto; vazio = nada é desenhado.
import { useRadar } from '../context/useRadar.js'

export default function Toast() {
  const { toast } = useRadar()
  if (!toast) return null
  // role="status": leitores de tela anunciam a mensagem sem roubar o foco.
  return <div className="toast" role="status">✓ <span>{toast}</span></div>
}
