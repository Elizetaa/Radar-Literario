import { useEffect, useRef, useState } from 'react'
import { Empty, Field, Heading, Json, Pages, Submit, useApi } from '../ui.jsx'
import { catalogRoles } from '../workForm.js'
import { buildEditionPayload, catalogOptions, catalogStatuses, listPath } from '../catalogAdmin.js'
import { AgeOptions, CatalogAccess, StatusOptions } from './Works.jsx'

const initialWorkId = () => new URLSearchParams(location.hash.split('?')[1] || '').get('work_id') || ''

function EditionEditor({ edition, options, onSaved, onCancel }) {
  const { perform, busy } = useApi()
  const [error, setError] = useState('')
  async function save(event) {
    event.preventDefault(); setError('')
    let body
    try { body = buildEditionPayload(new FormData(event.currentTarget), Boolean(edition.id)) }
    catch (error) { setError(error.message); return }
    const result = await perform(`/admin/editions${edition.id ? `/${edition.id}` : ''}`, edition.id ? 'PATCH' : 'POST', body, edition.id ? 'Edição atualizada.' : 'Edição cadastrada.')
    if (result) onSaved(result.data)
  }
  return <section className="card"><div className="section-heading"><h2>{edition.id ? `Editar edição #${edition.id}` : 'Nova edição'}</h2><button disabled={busy} onClick={onCancel}>Fechar</button></div><form onSubmit={save}><fieldset disabled={busy}>
    <Field label="Obra"><select name="work_id" defaultValue={edition.work_id || ''} required disabled={Boolean(edition.id)}><option value="">Selecione uma obra</option>{options.works.map((work) => <option key={work.id} value={work.id}>{work.title} · #{work.id} ({catalogStatuses[work.status]})</option>)}</select></Field>
    {edition.id && <small>A obra de uma edição existente não pode ser trocada.</small>}
    <Field label="Título da edição" name="title" defaultValue={edition.title || ''} required maxLength={300} />
    <div className="form-row"><Field label="Editora"><select name="publisher_id" defaultValue={edition.publisher_id || ''} required><option value="">Selecione uma editora</option>{options.publishers.map((publisher) => <option key={publisher.id} value={publisher.id}>{publisher.name} · #{publisher.id}</option>)}</select></Field><Field label="Idioma" name="language" defaultValue={edition.language || 'pt-br'} required maxLength={35} /></div>
    <div className="form-row"><Field label="Número da edição" name="edition_number" type="number" min={1} max={100000} defaultValue={edition.edition_number || 1} required /><Field label="Número de páginas" name="page_count" type="number" min={1} max={1000000} defaultValue={edition.page_count || ''} required /></div>
    <div className="form-row"><Field label="ISBN" name="isbn" defaultValue={edition.isbn || ''} maxLength={30} placeholder="ISBN-10 ou ISBN-13" /><Field label="Motivo da ausência de ISBN" name="isbn_absence_reason" defaultValue={edition.isbn_absence_reason || ''} maxLength={300} /></div>
    <div className="form-row"><Field label="Publicação" name="publication_date" type="date" defaultValue={edition.publication_date || ''} /><Field label="URL da capa" name="image_url" type="url" defaultValue={edition.image_url || ''} maxLength={2048} placeholder="https://…" /></div>
    <div className="form-row"><Field label="Estado"><select name="status" defaultValue={edition.status || 'draft'}><StatusOptions /></select></Field><Field label="Classificação etária"><select name="minimum_age" defaultValue={edition.minimum_age ?? 0}><AgeOptions /></select></Field></div>
    <small>Publicar uma edição exige obra publicada. Informe ISBN ou justificativa. A API valida ISBN, datas e alterações de páginas que afetem leituras existentes.</small>
    {error && <p className="inline-error" role="alert">{error}</p>}<Submit>{edition.id ? 'Salvar alterações' : 'Cadastrar edição'}</Submit>
  </fieldset></form></section>
}

export default function Editions() {
  const { perform, busy, user } = useApi()
  const allowed = catalogRoles.includes(user?.role)
  const [filters, setFilters] = useState(() => ({ work_id: initialWorkId(), status: '' }))
  const [applied, setApplied] = useState(() => ({ work_id: initialWorkId(), status: '' }))
  const [pageSize, setPageSize] = useState(20)
  const [list, setList] = useState(null)
  const [loading, setLoading] = useState(false)
  const [failed, setFailed] = useState(false)
  const [options, setOptions] = useState(null)
  const [optionError, setOptionError] = useState('')
  const [reload, setReload] = useState(0)
  const [editor, setEditor] = useState(null)
  const [editorVersion, setEditorVersion] = useState(0)
  const [detail, setDetail] = useState(null)
  const panel = useRef(null)
  useEffect(() => { if (detail || editor) panel.current?.scrollIntoView({ block: 'start' }) }, [detail, editor, editorVersion])
  async function load(page = 1, query = applied, size = list?.pagination.page_size || pageSize) {
    setLoading(true); setFailed(false)
    let result = await perform(listPath('/admin/editions', query, page, size))
    if (result && page > Math.max(1, Math.ceil(result.pagination.total / size))) result = await perform(listPath('/admin/editions', query, Math.max(1, Math.ceil(result.pagination.total / size)), size))
    if (result) setList(result)
    else setFailed(true)
    setLoading(false)
  }
  useEffect(() => {
    let active = true
    if (!allowed) return
    setList(null); setLoading(true); setFailed(false)
    perform(listPath('/admin/editions', { work_id: initialWorkId() })).then((result) => { if (active) { setList(result || null); setFailed(!result); setLoading(false) } })
    return () => { active = false }
  }, [allowed, user?.id, perform])
  useEffect(() => {
    let active = true
    if (!allowed) return
    setOptionError('')
    catalogOptions(perform, ['works', 'publishers'], () => active).then((result) => { if (active) setOptions(result) }).catch((error) => { if (active) setOptionError(error.message) })
    return () => { active = false }
  }, [allowed, user?.id, perform, reload])
  async function open(edition, editing = false) {
    const result = await perform(`/admin/editions/${edition.id}`)
    if (result) { setDetail(result.data); setEditor(editing ? result.data : null); setEditorVersion((value) => value + 1) }
  }
  async function archive(edition) {
    if (!window.confirm(`Arquivar a edição “${edition.title}” (#${edition.id})? Ela deixará de aparecer no catálogo público.`)) return
    const result = await perform(`/admin/editions/${edition.id}`, 'DELETE', undefined, 'Edição arquivada.')
    if (result) { setEditor(null); setDetail(null); await load(list.pagination.page) }
  }
  function change(event) { setFilters((current) => ({ ...current, [event.target.name]: event.target.value })) }
  return <>
    <Heading title="Gerenciar edições" description="Cadastre versões de uma obra, confira ISBN e editora, edite ou arquive edições.">{allowed && <button className="primary" disabled={busy || !options} onClick={() => { setDetail(null); setEditor({ work_id: applied.work_id, status: 'draft' }); setEditorVersion((value) => value + 1) }}>+ Nova edição</button>}</Heading>
    {!allowed ? <CatalogAccess /> : <>
      <form className="card" onSubmit={(event) => { event.preventDefault(); setApplied(filters); load(1, filters, pageSize) }}><fieldset disabled={busy}>
        <div className="form-row"><Field label="Obra"><select name="work_id" value={filters.work_id} onChange={change}><option value="">Todas as obras</option>{filters.work_id && !options?.works.some((work) => String(work.id) === filters.work_id) && <option value={filters.work_id}>Obra #{filters.work_id}</option>}{options?.works.map((work) => <option key={work.id} value={work.id}>{work.title} · #{work.id}</option>)}</select></Field><Field label="Estado"><select name="status" value={filters.status} onChange={change}><StatusOptions all /></select></Field></div>
        <Field label="Itens por página"><select value={pageSize} onChange={(event) => setPageSize(Number(event.target.value))}>{[10, 20, 50, 100].map((size) => <option key={size}>{size}</option>)}</select></Field>
        <div className="actions"><Submit>Aplicar filtros</Submit><button type="button" onClick={() => { const cleared = { work_id: '', status: '' }; setFilters(cleared); setApplied(cleared); load(1, cleared, pageSize) }}>Limpar filtros</button><button type="button" onClick={() => { load(list?.pagination.page || 1); setReload((value) => value + 1) }}>Atualizar</button></div>
      </fieldset></form>
      {!options && !optionError && <p role="status">Carregando obras e editoras…</p>}{optionError && <p className="inline-error" role="alert">{optionError} Use Atualizar para tentar novamente.</p>}
      {loading && <p role="status">Carregando edições…</p>}{failed && <p className="inline-error" role="alert">Não foi possível listar as edições. Use Atualizar para tentar novamente.</p>}
      {list && <><div className="card table-wrap"><table><caption>{list.pagination.total} edição(ões) encontradas</caption><thead><tr><th>Edição</th><th>Obra / editora</th><th>ISBN</th><th>Estado</th><th>Ações</th></tr></thead><tbody>{list.data.map((edition) => <tr key={edition.id}><td><strong>{edition.title}</strong><small>#{edition.id} · {edition.edition_number}ª edição · {edition.language} · {edition.page_count} páginas</small></td><td>{options?.works.find((work) => work.id === edition.work_id)?.title || `Obra #${edition.work_id}`}<small>{options?.publishers.find((publisher) => publisher.id === edition.publisher_id)?.name || `Editora #${edition.publisher_id}`}</small></td><td>{edition.isbn || 'Sem ISBN'}</td><td><span className="badge">{catalogStatuses[edition.status] || edition.status}</span></td><td><div className="actions"><button disabled={busy} onClick={() => open(edition)}>Detalhes</button><button disabled={busy || !options} onClick={() => open(edition, true)}>Editar</button><button className="danger" disabled={busy || edition.status === 'archived'} onClick={() => archive(edition)}>Arquivar</button></div></td></tr>)}</tbody></table>{!list.data.length && <Empty>Nenhuma edição encontrada. Cadastre uma edição ou ajuste os filtros.</Empty>}</div><Pages page={list.pagination.page} total={list.pagination.total} pageSize={list.pagination.page_size} onChange={load} /></>}
      <div ref={panel}>{editor && options ? <EditionEditor key={editorVersion} edition={editor} options={options} onCancel={() => setEditor(null)} onSaved={async (edition) => { setEditor(null); setDetail(edition); await load(1) }} /> : detail && <section className="card"><div className="section-heading"><h2>{detail.title} · #{detail.id}</h2><button onClick={() => setDetail(null)}>Fechar</button></div><Json value={detail} /></section>}</div>
    </>}
  </>
}
