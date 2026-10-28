import { useEffect, useRef, useState } from 'react'
import { Empty, Field, Heading, Json, Pages, Submit, useApi } from '../ui.jsx'
import { listPath } from '../catalogAdmin.js'
import { catalogRoles } from '../workForm.js'

const initialFilters = { title: '', author: '', genre_id: '', isbn: '', language: '', order: 'title' }
const editionPageSize = 10

export default function Catalog({ onSelect }) {
  const { perform, busy, user } = useApi()
  const allowed = catalogRoles.includes(user?.role)
  const mounted = useRef(true)
  const [filters, setFilters] = useState(initialFilters)
  const [pageSize, setPageSize] = useState(20)
  const [applied, setApplied] = useState({ filters: initialFilters, size: 20 })
  const [list, setList] = useState(null)
  const [loading, setLoading] = useState(false)
  const [failed, setFailed] = useState(false)
  const [genres, setGenres] = useState([])
  const [genresLoading, setGenresLoading] = useState(false)
  const [genresError, setGenresError] = useState(false)
  const [genresVersion, setGenresVersion] = useState(0)
  const [selectedId, setSelectedId] = useState(null)
  const [detail, setDetail] = useState(null)
  const [detailLoading, setDetailLoading] = useState(false)
  const [detailFailed, setDetailFailed] = useState(false)
  const [editions, setEditions] = useState(null)
  const [editionsLoading, setEditionsLoading] = useState(false)
  const [editionsFailed, setEditionsFailed] = useState(false)

  useEffect(() => {
    mounted.current = true
    return () => { mounted.current = false }
  }, [])

  useEffect(() => {
    let active = true
    setLoading(true)
    setFailed(false)
    setList(null)
    setDetail(null)
    setSelectedId(null)
    setEditions(null)
    setFilters(initialFilters)
    setPageSize(20)
    setApplied({ filters: initialFilters, size: 20 })
    perform(listPath('/works', initialFilters)).then((result) => {
      if (active) { setList(result || null); setFailed(!result); setLoading(false) }
    })
    return () => { active = false }
  }, [perform, user?.id])

  useEffect(() => {
    let active = true
    async function loadGenres() {
      setGenresLoading(true)
      setGenresError(false)
      const all = []
      for (let page = 1; active; page++) {
        const result = await perform(listPath('/genres', {}, page, 100))
        if (!active) return
        if (!result) { setGenresError(true); setGenresLoading(false); return }
        all.push(...result.data)
        if (!result.data.length || all.length >= result.pagination.total) break
      }
      if (active) { setGenres(all); setGenresLoading(false) }
    }
    loadGenres()
    return () => { active = false }
  }, [perform, genresVersion])

  async function load(page = 1, query = applied.filters, size = applied.size) {
    setLoading(true)
    setFailed(false)
    const result = await perform(listPath('/works', query, page, size))
    if (!mounted.current) return
    setList(result || null)
    setFailed(!result)
    setLoading(false)
  }

  async function loadEditions(workId, page = 1) {
    setEditionsLoading(true)
    setEditionsFailed(false)
    const result = await perform(listPath(`/works/${workId}/editions`, {}, page, editionPageSize))
    if (!mounted.current) return
    setEditions(result || null)
    setEditionsFailed(!result)
    setEditionsLoading(false)
  }

  async function open(workId) {
    setSelectedId(workId)
    setDetail(null)
    setEditions(null)
    setDetailLoading(true)
    setDetailFailed(false)
    setEditionsFailed(false)
    const result = await perform(`/works/${workId}`)
    if (!mounted.current) return
    setDetailLoading(false)
    if (!result) { setDetailFailed(true); return }
    setDetail(result.data)
    await loadEditions(workId)
  }

  function change(event) { setFilters((current) => ({ ...current, [event.target.name]: event.target.value })) }

  function search(event) {
    event.preventDefault()
    setApplied({ filters, size: pageSize })
    load(1, filters, pageSize)
  }

  function clear() {
    setFilters(initialFilters)
    setPageSize(20)
    setApplied({ filters: initialFilters, size: 20 })
    load(1, initialFilters, 20)
  }

  return <>
    <Heading title="Catálogo público" description="Busque obras publicadas e escolha uma edição para registrar sua leitura.">{allowed && <div className="actions"><a className="button-link" href="#works">Gerenciar obras</a><a className="button-link primary" href="#create-work">+ Adicionar obra</a></div>}</Heading>
    {!user && <p className="hint">Visitantes consultam somente obras e edições com classificação livre. Entre na conta para consultar o catálogo compatível com sua idade.</p>}
    <form className="card" onSubmit={search}><fieldset disabled={busy}>
      <div className="form-row"><Field label="Título da obra" name="title" value={filters.title} onChange={change} maxLength={300} placeholder="Digite parte do título" /><Field label="Nome do autor" name="author" value={filters.author} onChange={change} maxLength={160} placeholder="Digite parte do nome" /></div>
      <div className="form-row"><Field label="Gênero"><select name="genre_id" value={filters.genre_id} onChange={change} disabled={genresLoading}><option value="">Todos os gêneros ativos</option>{genres.map((genre) => <option key={genre.id} value={genre.id}>{genre.name}</option>)}</select></Field><Field label="ISBN da edição" name="isbn" value={filters.isbn} onChange={change} maxLength={30} placeholder="ISBN-10 ou ISBN-13" /></div>
      <div className="form-row"><Field label="Idioma da edição" name="language" value={filters.language} onChange={change} maxLength={35} placeholder="Ex.: pt-br ou en" /><Field label="Ordenar"><select name="order" value={filters.order} onChange={change}><option value="title">Título</option><option value="newest">Mais recentes</option><option value="popular">Popularidade</option><option value="top-rated">Melhor avaliação</option></select></Field></div>
      <Field label="Obras por página"><select value={pageSize} onChange={(event) => setPageSize(Number(event.target.value))}>{[10, 20, 50, 100].map((size) => <option key={size} value={size}>{size}</option>)}</select></Field>
      <div className="actions"><Submit>Pesquisar</Submit><button type="button" onClick={clear}>Limpar busca</button><button type="button" onClick={() => { load(list?.pagination.page || 1); setGenresVersion((version) => version + 1) }}>Atualizar lista</button></div>
    </fieldset></form>
    {genresLoading && <p role="status">Carregando gêneros…</p>}
    {genresError && <p className="inline-error" role="alert">Não foi possível carregar os gêneros. Os demais filtros continuam disponíveis; use “Atualizar lista” para tentar novamente.</p>}
    <div className="columns wide-left">
      <section aria-label="Obras encontradas">
        {loading && <p role="status">Carregando catálogo…</p>}
        {failed && <p className="inline-error" role="alert">Não foi possível carregar o catálogo. Confira o aviso da API e use “Atualizar lista” para tentar novamente.</p>}
        {list && !loading && <>
          <p>{list.pagination.total} obra(s) encontrada(s).</p>
          <div className="book-grid">{list.data.map((work) => <article className="card book" key={work.id}>
            <span className="badge">Obra #{work.id} · {work.minimum_age === 0 ? 'Livre' : `${work.minimum_age}+`}</span>
            <h2>{work.title}</h2><p>{work.authors?.map((author) => author.name).join(', ') || 'Sem autor informado'}</p>
            <small>{work.genres?.map((genre) => genre.name).join(' · ') || 'Sem gênero informado'}</small>
            <small>{work.original_language} · {work.reader_count || 0} leitor(es) · {work.review_count || 0} avaliação(ões)</small>
            <button disabled={busy} onClick={() => open(work.id)}>Ver detalhes e edições →</button>
          </article>)}</div>
          {!list.data.length && <Empty>Nenhuma obra publicada corresponde à busca. Ajuste os filtros ou cadastre e publique uma obra na administração.</Empty>}
          <Pages page={list.pagination.page} total={list.pagination.total} pageSize={list.pagination.page_size} onChange={(page) => load(page)} />
        </>}
      </section>
      <section className="card" aria-label="Detalhes e edições da obra">
        <div className="section-heading"><h2>{detail?.title || 'Detalhes e edições'}</h2>{selectedId && <button disabled={busy} onClick={() => open(selectedId)}>Atualizar</button>}</div>
        {detailLoading && <p role="status">Carregando detalhes da obra…</p>}
        {detailFailed && <p className="inline-error" role="alert">Não foi possível consultar a obra. Use “Atualizar” para tentar novamente.</p>}
        {detail ? <>
          <dl><dt>Obra</dt><dd>#{detail.id}</dd><dt>Autores</dt><dd>{detail.authors?.map((author) => author.name).join(', ') || 'Não informados'}</dd><dt>Gêneros</dt><dd>{detail.genres?.map((genre) => genre.name).join(', ') || 'Não informados'}</dd><dt>Idioma original</dt><dd>{detail.original_language}</dd><dt>Primeira publicação</dt><dd>{detail.first_publication_date || 'Não informada'}</dd><dt>Classificação</dt><dd>{detail.minimum_age === 0 ? 'Livre' : `${detail.minimum_age} anos`}</dd><dt>Avaliação média</dt><dd>{detail.average_rating == null ? 'Sem avaliações' : `${detail.average_rating} / 5`}</dd></dl>
          <p className="preserve-lines">{detail.synopsis || 'Sem sinopse.'}</p>
          <details><summary>JSON da obra</summary><Json value={detail} /></details>
          <h2 className="hint">Edições publicadas{editions ? ` (${editions.pagination.total})` : ''}</h2>
          {editionsLoading && <p role="status">Carregando edições…</p>}
          {editionsFailed && <div><p className="inline-error" role="alert">Não foi possível carregar as edições.</p><button disabled={busy} onClick={() => loadEditions(detail.id)}>Tentar novamente</button></div>}
          {editions && !editionsLoading && <>
            {editions.data.map((edition) => <article className="edition" key={edition.id}>
              <strong>{edition.title}</strong>
              <small>Edição #{edition.id} · {edition.edition_number}ª edição · {edition.language} · {edition.page_count} páginas</small>
              <small>Editora #{edition.publisher_id} · Publicação: {edition.publication_date || 'Não informada'}</small>
              <small>ISBN: {edition.isbn || 'Não informado'}{!edition.isbn && edition.isbn_absence_reason ? ` · ${edition.isbn_absence_reason}` : ''}</small>
              <details><summary>JSON da edição</summary><Json value={edition} /></details>
              <button disabled={busy} onClick={() => onSelect({ work_id: detail.id, edition_id: edition.id, title: detail.title })}>Usar na estante</button>
            </article>)}
            {!editions.data.length && <Empty>Não há edições publicadas disponíveis para sua conta.</Empty>}
            <Pages page={editions.pagination.page} total={editions.pagination.total} pageSize={editions.pagination.page_size} onChange={(page) => loadEditions(detail.id, page)} />
          </>}
        </> : !detailLoading && !detailFailed && <Empty>Selecione uma obra para consultar seus dados e suas edições.</Empty>}
      </section>
    </div>
  </>
}
