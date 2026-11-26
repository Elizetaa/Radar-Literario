// Saudação que só aparece no celular (o CSS a esconde no desktop).
// Cada página a coloca no topo; a tela de Sugestões (modo "cartões") não usa,
// porque o cartão ocupa a tela inteira.
import { useRadar } from '../context/useRadar.js'

export default function MobileGreeting() {
  const { profile } = useRadar()
  return (
    <div className="mobile-greeting">
      <strong>Olá, {profile.name.split(' ')[0]} 👋</strong>
      <span>Que história vamos descobrir hoje?</span>
    </div>
  )
}
