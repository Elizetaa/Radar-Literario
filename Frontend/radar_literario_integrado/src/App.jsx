// App.jsx: apenas a "casca" do aplicativo.
//
// Estrutura do frontend:
//   context/     estado global (RadarProvider) e o hook useRadar()
//   pages/       uma página por tela: Home, Search, Shelf, Suggestions, Profile
//   components/  peças reutilizáveis (Header, Sidebar, cards, Modal...)
//   components/modals/   um arquivo por janela modal (ModalHost escolhe qual abrir)
//   features/    telas maiores e independentes: auth (login) e book-discovery (cartões)
//   services/    comunicação com a API Go (backendApi) e adaptador (libraryService)
//   utils/       constantes e conversões simples
//
// Navegação: não usa biblioteca de rotas; o contexto guarda `page` e o mapa abaixo
// escolhe qual página desenhar. Para criar uma página: crie em pages/, adicione aqui
// e em utils/constants.js (navItems).
import RadarProvider from './context/RadarProvider.jsx'
import { useRadar } from './context/useRadar.js'
import Auth from './features/auth/Auth.jsx'
import Header from './components/Header.jsx'
import Sidebar from './components/Sidebar.jsx'
import MobileNav from './components/MobileNav.jsx'
import ModalHost from './components/ModalHost.jsx'
import Toast from './components/Toast.jsx'
import HomePage from './pages/HomePage.jsx'
import SearchPage from './pages/SearchPage.jsx'
import ShelfPage from './pages/ShelfPage.jsx'
import SuggestionsPage from './pages/SuggestionsPage.jsx'
import ProfilePage from './pages/ProfilePage.jsx'
import './App.css'
import { Routes, Route, Navigate } from 'react-router-dom'

<main className="main-content min-w-0">
  <Routes>
    <Route path="/" element={<HomePage />} />
    <Route path="/pesquisar" element={<SearchPage />} />
    <Route path="/estante" element={<ShelfPage />} />
    <Route path="/sugestoes" element={<SuggestionsPage />} />
    <Route path="/perfil" element={<ProfilePage />} />
    {/* URL desconhecida volta para a home, sem poluir o histórico */}
    <Route path="*" element={<Navigate to="/" replace />} />
  </Routes>
</main>
// id da página (o mesmo usado em navItems) -> componente da página
const pages = {
  home: HomePage,
  search: SearchPage,
  shelf: ShelfPage,
  suggestions: SuggestionsPage,
  profile: ProfilePage,
}

function AppContent() {
  const { session, startSession, page } = useRadar()
  const CurrentPage = pages[page] ?? HomePage

  // Enquanto confirma o cookie de sessão com a API.
  if (session === 'loading') return <div className="empty-state" role="status">Carregando…</div>
  // Sem sessão: tela de login/cadastro. Ao entrar, startSession carrega perfil e dados.
  if (session === 'out') return <Auth onDone={startSession} />

  return (
    <div className="app-shell min-h-screen bg-[#faf9fb] text-[#29242c]">
      <Header />
      <div className="layout">
        <Sidebar />
        <main className="main-content min-w-0"><CurrentPage /></main>
      </div>
      <MobileNav />
      <ModalHost />
      <Toast />
    </div>
  )
}

export default function App() {
  // O Provider envolve tudo para que qualquer componente use useRadar().
  return <RadarProvider><AppContent /></RadarProvider>
}
