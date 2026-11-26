// Hook para qualquer página ou componente ler o estado global e as ações do app.
// Uso:  const { books, setPage, feedback } = useRadar()
import { useContext } from 'react'
import { RadarContext } from './RadarContext.js'

export function useRadar() {
  const value = useContext(RadarContext)
  // Falha cedo e com mensagem clara se alguém esquecer de envolver o app no Provider.
  if (!value) throw new Error('useRadar deve ser usado dentro de <RadarProvider>.')
  return value
}
