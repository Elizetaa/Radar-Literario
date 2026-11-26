// Página "Perfil": dados do usuário, livro favorito, abas de publicações e ações da conta.
import { useState } from 'react'
import { useRadar } from '../context/useRadar.js'
import Avatar from '../components/Avatar.jsx'
import PostCard from '../components/PostCard.jsx'
import MobileGreeting from '../components/MobileGreeting.jsx'

const tabs = ['Publicações', 'Avaliações', 'Diários', 'Curtidas']

export default function ProfilePage() {
  const { profile, posts, setPage, openForm, openModal, openComments, openReport, likePost, logout } = useRadar()
  const [tab, setTab] = useState('Publicações') // aba selecionada

  // Só as publicações do próprio usuário (compara o @apelido do autor com o do perfil).
  const ownPosts = posts.filter((post) => post.handle === profile.handle)
  // Filtra pela aba: avaliação = tem nota; diário = sem nota; curtidas = que o usuário curtiu.
  const shownPosts = ownPosts.filter((post) => {
    if (tab === 'Avaliações') return post.rating > 0
    if (tab === 'Diários') return post.rating === 0
    if (tab === 'Curtidas') return post.liked
    return true
  })

  return (
    <>
      <MobileGreeting />
      <section className="page-panel profile-page">
        <div className="profile-banner">
          <div className="profile-cover-art" />
          <div className="profile-main">
            <Avatar size="large" />
            <div className="profile-identity">
              <h1>{profile.name}</h1>
              <span>{profile.handle}</span>
              <p>{profile.bio}</p>
              <small>{profile.location} · No Radar desde {profile.joined}</small>
            </div>
            <button className="outline-button profile-edit" onClick={() => openForm('editProfile')}>Editar perfil</button>
          </div>
          <div className="profile-stats"><span><strong>{ownPosts.length}</strong> Publicações</span></div>
        </div>

        <div className="favorite-book">
          <span className="eyebrow">LIVRO FAVORITO</span>
          <strong>{profile.favoriteBook || '—'}</strong>
          <button onClick={() => setPage('search')}>Encontrar sua próxima leitura →</button>
        </div>

        <div className="profile-tabs">
          {tabs.map((item) => <button key={item} className={tab === item ? 'selected' : ''} onClick={() => setTab(item)}>{item}</button>)}
        </div>

        {shownPosts.length
          ? <div className="profile-feed">{shownPosts.map((post) => <PostCard key={post.id} post={post} onLike={likePost} onReport={openReport} onComments={openComments} />)}</div>
          : <div className="empty-state">{tab === 'Curtidas' ? 'As publicações que você curtir aparecerão aqui.' : `Suas ${tab.toLowerCase()} aparecerão aqui.`}</div>}

        <div className="account-actions">
          <span>Configurações da conta</span>
          <button onClick={logout}>Sair da conta</button>
          <button onClick={() => openModal('deleteAccount')}>Excluir conta</button>
        </div>
      </section>
    </>
  )
}
