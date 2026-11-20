import { useEffect, useState } from 'react'
import { Empty, Field, Heading, Pages, Submit, useApi } from '../ui.jsx'

export default function Publications() {
  const { perform, busy, user } = useApi()
  const [kind, setKind] = useState('reading-entries')
  const [posts, setPosts] = useState([])
  const [page, setPage] = useState(1)
  const [pagination, setPagination] = useState({ total: 0, page_size: 20 })
  const [reveal, setReveal] = useState({ spoilers: false, sensitive: false })
  const [loaded, setLoaded] = useState(kind)
  const diary = kind === 'reading-entries'

  async function load(next = 1) {
    if (!user) return
    const result = await perform(`/me/${kind}?${new URLSearchParams({ page: next, page_size: 20, reveal_spoilers: reveal.spoilers, reveal_sensitive: reveal.sensitive })}`)
    if (result) {
      const lastPage = Math.max(1, Math.ceil(result.pagination.total / result.pagination.page_size))
      if (next > lastPage) return load(lastPage)
      setPosts(result.data); setPage(result.pagination.page); setPagination(result.pagination); setLoaded(kind)
    }
  }

  useEffect(() => {
    let active = true
    setPosts([]); setPage(1); setPagination({ total: 0, page_size: 20 })
    if (!user) { setReveal({ spoilers: false, sensitive: false }); return }
    perform(`/me/${kind}?page_size=20`).then((result) => {
      if (active && result) { setPosts(result.data); setPage(result.pagination.page); setPagination(result.pagination); setLoaded(kind) }
    })
    return () => { active = false }
  }, [kind, perform, user?.id])

  async function publish(event) {
    event.preventDefault()
    const form = event.currentTarget
    const fields = new FormData(form)
    const body = {
      work_id: Number(fields.get('work_id')), title: fields.get('title'), text: fields.get('text'), status: fields.get('status'),
      contains_spoiler: fields.has('contains_spoiler'), contains_sensitive_content: fields.has('contains_sensitive_content'),
      minimum_age: Number(fields.get('minimum_age')),
      ...(diary ? { edition_id: Number(fields.get('edition_id')), current_page: Number(fields.get('current_page')) } : { rating: Number(fields.get('rating')) }),
    }
    const result = await perform(`/${kind}`, 'POST', body, 'Registro salvo.')
    if (result) { form.reset(); await load() }
  }

  return <>
    <Heading title="Publicações" description="Crie diários e avaliações, como rascunho ou publicação." />
    <div className="tabs"><button disabled={busy} className={diary ? 'active' : ''} onClick={() => { setKind('reading-entries'); setPosts([]); setReveal({ spoilers: false, sensitive: false }) }}>Diários</button><button disabled={busy} className={!diary ? 'active' : ''} onClick={() => { setKind('reviews'); setPosts([]); setReveal({ spoilers: false, sensitive: false }) }}>Avaliações</button></div>
    {!user && <Empty>Entre para criar e consultar suas publicações.</Empty>}
    <div className="columns"><section className="card"><h2>{diary ? 'Novo diário' : 'Nova avaliação'}</h2><form key={`${kind}-${user?.id || 'visitor'}`} onSubmit={publish}><fieldset disabled={busy || !user}>
      <div className="form-row"><Field label="ID da obra" name="work_id" type="number" min="1" required />{diary ? <Field label="ID da edição" name="edition_id" type="number" min="1" required /> : <Field label="Nota (1 a 5)" name="rating" type="number" min="1" max="5" required />}</div>
      {diary && <Field label="Página atual" name="current_page" type="number" min="0" defaultValue="0" required />}
      <Field label="Título (opcional)" name="title" maxLength="200" /><Field label="Texto"><textarea name="text" rows="5" maxLength="10000" required /></Field>
      <div className="form-row"><Field label="Estado"><select name="status"><option value="draft">Rascunho</option><option value="published">Publicado</option></select></Field><Field label="Classificação etária"><select name="minimum_age">{[0, 10, 12, 14, 16, 18].map((age) => <option key={age} value={age}>{age === 0 ? 'Livre' : `${age} anos`}</option>)}</select></Field></div>
      <label className="check"><input type="checkbox" name="contains_spoiler" /> Contém spoiler</label><label className="check"><input type="checkbox" name="contains_sensitive_content" /> Contém conteúdo sensível</label>
      {!diary && <small>Avaliações exigem leitura iniciada ou concluída. Só pode existir uma avaliação por leitor e obra.</small>}<Submit>Salvar {diary ? 'diário' : 'avaliação'}</Submit>
    </fieldset></form></section>
    <section className="card"><div className="section-heading"><h2>Meus registros</h2><button disabled={busy || !user} onClick={() => load(page)}>Atualizar</button></div>
      <div className="reveal"><label className="check"><input type="checkbox" checked={reveal.spoilers} onChange={(e) => setReveal({ ...reveal, spoilers: e.target.checked })} /> Revelar spoilers</label><label className="check"><input type="checkbox" checked={reveal.sensitive} onChange={(e) => setReveal({ ...reveal, sensitive: e.target.checked })} /> Revelar conteúdo sensível</label><small>Use “Atualizar” para aplicar. A classificação etária continua valendo.</small></div>
      {loaded === kind && posts.map((post) => <article className="post" key={post.id}><div className="section-heading"><strong>{post.title || `Registro #${post.id}`}</strong><span className="badge">{post.status}</span></div><small>Obra #{post.work_id}{post.edition_id ? ` · edição #${post.edition_id}` : ''}{post.rating ? ` · nota ${post.rating}/5` : ''}{post.current_page != null ? ` · página ${post.current_page}` : ''}</small><p>{post.content_hidden ? 'Conteúdo protegido. Confira as opções de revelação.' : post.text}</p></article>)}
      {!posts.length && <Empty>Nenhum registro encontrado.</Empty>}<Pages page={page} count={posts.length} total={pagination.total} pageSize={pagination.page_size} onChange={load} />
    </section></div>
  </>
}
