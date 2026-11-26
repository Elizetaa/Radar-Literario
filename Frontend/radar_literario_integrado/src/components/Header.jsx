// Barra superior: logo, busca, atalho do perfil e botão "Novo livro".
import { useRadar } from '../context/useRadar.js'
import Avatar from './Avatar.jsx'

export default function Header() {
  const { profile, query, setQuery, setPage, openForm } = useRadar()
  return (
    <header className="topbar">
      <button className="brand" onClick={() => setPage('home')}>
        <img className="brand-mark" src="../../public/icons.svg"/>
        <span>radar<span className="brand-light">literário</span></span>
      </button>
      <label className="header-search min-w-0">
        <span aria-hidden="true">⌕</span>
        {/* Ao digitar, guarda o texto e leva o usuário para a página de pesquisa;
            a SearchPage lê o mesmo `query` do contexto e busca no catálogo. */}
        <input
          value={query}
          onChange={(event) => { setQuery(event.target.value); if (event.target.value) setPage('search') }}
          placeholder="Pesquisar livros, autores, usuários..."
          aria-label="Pesquisar livros, autores ou usuários"
        />
        <button aria-label="Pesquisar" onClick={() => setPage('search')}>⌕</button>
      </label>
      <div className="header-user">
        <button onClick={() => setPage('profile')} className="user-short">
          <Avatar size="small" /><span>{profile.name.split(' ')[0]}</span>
        </button>
        <button className="primary-button compact" onClick={() => openForm('book')}>＋ <span>Novo livro</span></button>
      </div>
    </header>
  )
}
