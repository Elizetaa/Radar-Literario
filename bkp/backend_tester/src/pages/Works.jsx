import { useEffect, useRef, useState } from 'react'
import { Empty, Field, Heading, Json, Pages, Submit, useApi } from '../ui.jsx'
import { buildWorkPayload, catalogRoles } from '../workForm.js'
import { catalogOptions, catalogStatuses, listPath, minimumAges } from '../catalogAdmin.js'

const initialFilters = { title: '', author: '', status: '', genre_id: '', isbn: '', language: '', order: 'newest' }

export function CatalogAccess() {
  return <section className="card"><h2>Permissão de catálogo necessária</h2><p>Entre com uma conta de gestor de catálogo, administrador ou administrador do sistema para gerenciar os registros. Leitores e visitantes podem consultar o catálogo público.</p><a className="button-link" href="#account">Ir para Conta</a></section>
}

export function StatusOptions({ all = false }) {
  return <>{all && <option value="">Todos os estados</option>}{Object.entries(catalogStatuses).map(([value, label]) => <option key={value} value={value}>{label}</option>)}</>
}

export function AgeOptions() {
  return minimumAges.map((age) => <option key={age} value={age}>{age === 0 ? 'Livre' : `${age} anos`}</option>)
}

function WorkEditor({ work, options, onSaved, onCancel }) {
  const { busy, perform } = useApi()
  const [status, setStatus] = useState(work.status)
  const [error, setError] = useState('')
  async function save(event) {
    event.preventDefault()
    setError('')
    let body
    const fields = new FormData(event.currentTarget)
    try {
      body = buildWorkPayload(fields)
      body.first_publication_date = fields.get('first_publication_date') || null
    } catch (error) { setError(error.message); return }
    const result = await perform(`/admin/works/${work.id}`, 'PATCH', body, 'Obra atualizada.')
    if (result) onSaved(result.data)
  }
  return <section className="card"><div className="section-heading"><h2>Editar obra #{work.id}</h2><button disabled={busy} onClick={onCancel}>Fechar</button></div><form onSubmit={save}><fieldset disabled={busy}>
    <Field label="Título" name="title" defaultValue={work.title} required maxLength={300} />
    <div className="form-row"><Field label="Idioma original" name="original_language" defaultValue={work.original_language} required maxLength={35} /><Field label="Primeira publicação" name="first_publication_date" type="date" defaultValue={work.first_publication_date || ''} /></div>
    <Field label="Sinopse"><textarea name="synopsis" rows={5} maxLength={30000} defaultValue={work.synopsis} required={status === 'published'} /></Field>
    <div className="form-row"><Field label="Estado"><select name="status" value={status} onChange={(event) => setStatus(event.target.value)}><StatusOptions /></select></Field><Field label="Classificação etária"><select name="minimum_age" defaultValue={work.minimum_age}><AgeOptions /></select></Field></div>
    <div className="form-row">{['authors', 'genres'].map((resource) => <fieldset key={resource}><legend className="field-label">{resource === 'authors' ? 'Autores' : 'Gêneros'}</legend><div className="choice-list">{options[resource].map((item) => <label className="check" key={item.id}><input type="checkbox" name={resource === 'authors' ? 'author_ids' : 'genre_ids'} value={item.id} defaultChecked={work[resource]?.some((selected) => selected.id === item.id)} />{item.name}{item.active === false ? ' (inativo)' : ''}</label>)}{!options[resource].length && <small>Nenhum registro cadastrado.</small>}</div></fieldset>)}</div>
    <small>Para publicar, informe sinopse, autor e gênero ativo. Limpe a data para removê-la. Edições são gerenciadas na página Edições.</small>
    {error && <p className="inline-error" role="alert">{error}</p>}<Submit>Salvar alterações</Submit>
  </fieldset></form><details className="hint"><summary>Dados completos da obra</summary><Json value={work} /></details></section>
}

export default function Works() {
  const { perform, busy, user } = useApi()
  const allowed = catalogRoles.includes(user?.role)
  const [filters, setFilters] = useState(initialFilters)
  const [applied, setApplied] = useState(initialFilters)
  const [pageSize, setPageSize] = useState(20)
  const [list, setList] = useState(null)
  const [loading, setLoading] = useState(false)
  const [failed, setFailed] = useState(false)
  const [options, setOptions] = useState(null)
  const [optionError, setOptionError] = useState('')
  const [optionsVersion, setOptionsVersion] = useState(0)
  const [editing, setEditing] = useState(null)
  const [detail, setDetail] = useState(null)
  const panel = useRef(null)
  useEffect(() => { if (detail) panel.current?.scrollIntoView({ block: 'start' }) }, [detail, editing])

  async function load(page = 1, query = applied, size = list?.pagination.page_size || pageSize) {
    setLoading(true); setFailed(false)
    let result = await perform(listPath('/admin/works', query, page, size))
    if (result && page > Math.max(1, Math.ceil(result.pagination.total / size))) result = await perform(listPath('/admin/works', query, Math.max(1, Math.ceil(result.pagination.total / size)), size))
    if (result) setList(result)
    else setFailed(true)
    setLoading(false)
  }
  useEffect(() => {
    let active = true
    if (!allowed) return
    setLoading(true); setFailed(false); setList(null)
    perform(listPath('/admin/works', initialFilters)).then((result) => {
      if (active) { setList(result || null); setFailed(!result); setLoading(false) }
    })
    return () => { active = false }
  }, [allowed, user?.id, perform])
  useEffect(() => {
    let active = true
    if (!allowed) return
    setOptionError('')
    catalogOptions(perform, ['authors', 'genres'], () => active).then((result) => { if (active) setOptions(result) }).catch((error) => { if (active) setOptionError(error.message) })
    return () => { active = false }
  }, [allowed, user?.id, perform, optionsVersion])
  function change(event) { setFilters((current) => ({ ...current, [event.target.name]: event.target.value })) }
  async function open(work, edit = false) {
    const result = await perform(`/admin/works/${work.id}`)
    if (result) { setDetail(result.data); setEditing(edit ? result.data : null) }
  }
  async function archive(work) {
    if (!window.confirm(`Arquivar a obra “${work.title}” (#${work.id})? Ela sairá do catálogo público e permanecerá na administração.`)) return
    const result = await perform(`/admin/works/${work.id}`, 'DELETE', undefined, 'Obra arquivada.')
    if (result) { setDetail(null); setEditing(null); await load(list.pagination.page) }
  }
  return <>
    <Heading title="Gerenciar obras" description="Liste todos os estados do catálogo, consulte detalhes, edite e arquive obras."><a className="button-link primary" href="#create-work">+ Adicionar obra</a></Heading>
    {!allowed ? <CatalogAccess /> : <>
      <form className="card filter-grid" onSubmit={(event) => { event.preventDefault(); setApplied(filters); load(1, filters, pageSize) }}><fieldset disabled={busy}>
        <div className="form-row"><Field label="Título" name="title" value={filters.title} onChange={change} maxLength={300} /><Field label="Autor" name="author" value={filters.author} onChange={change} maxLength={160} /></div>
        <div className="form-row"><Field label="Estado"><select name="status" value={filters.status} onChange={change}><StatusOptions all /></select></Field><Field label="Gênero"><select name="genre_id" value={filters.genre_id} onChange={change}><option value="">Todos os gêneros</option>{options?.genres.map((genre) => <option key={genre.id} value={genre.id}>{genre.name}{genre.active ? '' : ' (inativo)'}</option>)}</select></Field></div>
        <details><summary>Mais filtros e ordenação</summary><div className="form-row edition-fields"><Field label="ISBN da edição" name="isbn" value={filters.isbn} onChange={change} /><Field label="Idioma da edição publicada" name="language" value={filters.language} onChange={change} placeholder="pt-br" /><Field label="Ordenar"><select name="order" value={filters.order} onChange={change}><option value="newest">Mais recentes</option><option value="title">Título</option><option value="popular">Popularidade</option><option value="top-rated">Melhor avaliação</option></select></Field><Field label="Itens por página"><select value={pageSize} onChange={(event) => setPageSize(Number(event.target.value))}>{[10, 20, 50, 100].map((size) => <option key={size}>{size}</option>)}</select></Field></div></details>
        <div className="actions"><Submit>Aplicar filtros</Submit><button type="button" onClick={() => { setFilters(initialFilters); setApplied(initialFilters); load(1, initialFilters) }}>Limpar filtros</button><button type="button" onClick={() => { load(list?.pagination.page || 1); setOptionsVersion((value) => value + 1) }}>Atualizar</button></div>
      </fieldset></form>
      {optionError && <p className="inline-error" role="alert">{optionError} Use Atualizar para tentar novamente.</p>}
      {loading && <p role="status">Carregando obras…</p>}
      {failed && <p className="inline-error" role="alert">Não foi possível listar as obras. Use Atualizar para tentar novamente.</p>}
      {list && <><div className="card table-wrap"><table><caption>{list.pagination.total} obra(s) encontradas</caption><thead><tr><th>Obra</th><th>Autores / gêneros</th><th>Estado</th><th>Ações</th></tr></thead><tbody>{list.data.map((work) => <tr key={work.id}><td><strong>{work.title}</strong><small>#{work.id} · {work.original_language} · {work.minimum_age}+</small></td><td>{work.authors?.map((author) => author.name).join(', ') || 'Sem autor'}<small>{work.genres?.map((genre) => genre.name).join(', ') || 'Sem gênero'}</small></td><td><span className="badge">{catalogStatuses[work.status] || work.status}</span></td><td><div className="actions"><button disabled={busy} onClick={() => open(work)}>Detalhes</button><button disabled={busy || !options} onClick={() => open(work, true)}>Editar</button><a className="button-link" href={`#editions?work_id=${work.id}`}>Edições</a><button className="danger" disabled={busy || work.status === 'archived'} onClick={() => archive(work)}>Arquivar</button></div></td></tr>)}</tbody></table>{!list.data.length && <Empty>Nenhuma obra corresponde aos filtros. Cadastre uma obra ou ajuste a busca.</Empty>}</div><Pages page={list.pagination.page} total={list.pagination.total} pageSize={list.pagination.page_size} onChange={load} /></>}
      <div ref={panel}>{editing && options ? <WorkEditor key={`${editing.id}-${editing.version}`} work={editing} options={options} onCancel={() => setEditing(null)} onSaved={async (work) => { setEditing(null); setDetail(work); await load(list?.pagination.page || 1) }} /> : detail && <section className="card"><div className="section-heading"><h2>{detail.title} · #{detail.id}</h2><button onClick={() => setDetail(null)}>Fechar</button></div><p className="synopsis">{detail.synopsis || 'Sem sinopse.'}</p><Json value={detail} /></section>}</div>
    </>}
  </>
}
