// Cartão de uma publicação do feed (diário de leitura ou resenha).
// Props: post + três ações — onLike(id), onReport(post), onComments(post).
// As ações vêm de fora: este componente só desenha, não conversa com a API.
import Avatar from './Avatar.jsx'
import Cover from './Cover.jsx'
import Stars from './Stars.jsx'

export default function PostCard({ post, onLike, onReport, onComments }) {
  // Progresso em % (só existe em diários, que têm página atual).
  const percent = post.pages ? Math.round(post.pages / post.totalPages * 100) : 0
  return (
    <article className="post-card">
      <div className="post-head">
        <Avatar />
        <div className="post-author"><strong>{post.person}</strong><small>{post.time}</small></div>
        <button className="text-button report" onClick={() => onReport(post)}>Denunciar</button>
      </div>
      <div className="post-book">
        <div>
          <h3>{post.title}</h3>
          <p>{post.author}</p>
          {post.rating > 0 && <Stars rating={post.rating} />}
        </div>
        <Cover book={{ title: post.title, genre: 'Fantasia' }} className="mini-cover" />
      </div>
      {post.pages > 0 && (
        <div className="post-progress">
          <span className="progress-track"><span style={{ width: `${percent}%` }} /></span>
          <small>{percent}%</small>
        </div>
      )}
      <p className="post-text">{post.text}</p>
      <div className="post-actions">
        {/* A classe "liked" pinta o coração quando o usuário já curtiu */}
        <button onClick={() => onLike(post.id)} aria-label="Curtir publicação" className={post.liked ? 'liked' : ''}>♡ <span>{post.likes}</span></button>
        <button onClick={() => onComments(post)}>▣ <span>Comentar</span></button>
      </div>
    </article>
  )
}
