// Objeto de contexto do app. É criado num arquivo próprio (e não no Provider)
// para o Vite/React Refresh conseguir recarregar os componentes sem perder estado.
// Quem fornece o valor: RadarProvider.jsx. Quem consome: o hook useRadar.js.
import { createContext } from 'react'

export const RadarContext = createContext(null)
