import { useEffect, useState } from 'react'
import { Empty, Field, Heading, Json, Pages, useApi } from '../ui.jsx'
import { editionLabel, loadCatalogOptions } from '../catalogOptions.js'

const statuses = { planned: 'Planejado', in_progress: 'Em leitura', paused: 'Pausado', completed: 'Concluído', abandoned: 'Abandonado' }
const empty = { work_id: '', edition_id: '', current_page: 0, status: 'planned', change_kind: 'advance' }

export default function Shelf({ selection }) {
  const { perform, busy, user } = useApi()
  const [items, setItems] = useState([])
  const [page, setPage] = useState(1)
  const [pagination, setPagination] = useState({ total: 0, page_size: 20 })
  const [form, setForm] = useState(() => ({ ...empty, ...(selection || {}) }))
  const [statistics, setStatistics] = useState(null)
  const [works, setWorks] = useState({ items: [], loading: false, error: '' })
  const [editions, setEditions] = useState({ workId: '', items: [], loading: false, error: '' })
  const [reload, setReload] = useState(0)
  const [formError, setFormError] = useState('')
  const workId = String(form.work_id)
  const currentEditions = editions.workId === workId ? editions.items : []
  const selectedEdition = currentEditions.find((edition) => String(edition.id) === String(form.edition_id))
  const selectedWork = works.items.find((work) => String(work.id) === workId)
  const optionsLoading = works.loading || editions.loading || (Boolean(workId) && editions.workId !== workId)
  const canSave = Boolean(selectedWork && selectedEdition && !optionsLoading && !works.error && !editions.error)

  async function load(next = 1) {
    if (!user) return
    const result = await perform(`/me/shelf?page=${next}&page_size=20`)
    if (result) {
      const lastPage = Math.max(1, Math.ceil(result.pagination.total / result.pagination.page_size))
      if (next > lastPage) return load(lastPage)
      setItems(result.data); setPage(result.pagination.page); setPagination(result.pagination)
    }
  }

  useEffect(() => {
    let active = true
    setItems([]); setPage(1); setPagination({ total: 0, page_size: 20 }); setStatistics(null)
    if (!user) { setForm({ ...empty, ...(selection || {}) }); setFormError(''); return }
    perform('/me/shelf?page_size=20').then((result) => {
      if (active && result) { setItems(result.data); setPage(result.pagination.page); setPagination(result.pagination) }
    })
    return () => { active = false }
  }, [perform, user?.id])

  useEffect(() => {
    let active = true
    if (!user) { setWorks({ items: [], loading: false, error: '' }); return }
    setWorks({ items: [], loading: true, error: '' })
    loadCatalogOptions('/works').then((items) => {
      if (active) setWorks({ items, loading: false, error: '' })
    }).catch((error) => {
      if (active) setWorks({ items: [], loading: false, error: error.message })
    })
    return () => { active = false }
  }, [user?.id, reload])

  useEffect(() => {
    let active = true
    if (!user || !workId) {
      setEditions({ workId: '', items: [], loading: false, error: '' })
      return
    }
    setEditions({ workId, items: [], loading: true, error: '' })
    loadCatalogOptions(`/works/${workId}/editions`).then((items) => {
      if (!active) return
      setEditions({ workId, items, loading: false, error: '' })
      // Uma edição única dispensa outra escolha, preservando a edição ao editar.
      if (items.length === 1) setForm((current) => String(current.work_id) === workId && !current.edition_id ? { ...current, edition_id: items[0].id } : current)
    }).catch((error) => {
      if (active) setEditions({ workId, items: [], loading: false, error: error.message })
    })
    return () => { active = false }
  }, [user?.id, workId, reload])

  function chooseWork(event) {
    const id = event.target.value
    const existing = items.find((item) => String(item.work_id) === id)
    setForm({ ...empty, ...(existing || {}), work_id: id })
    setFormError('')
  }

  function change(event) { setForm((current) => ({ ...current, [event.target.name]: event.target.value })) }
  async function save(event) {
    event.preventDefault()
    setFormError('')
    if (!canSave) { setFormError('Selecione uma obra e uma edição disponíveis antes de salvar.'); return }
    const result = await perform(`/me/shelf/${form.work_id}`, 'PUT', {
      edition_id: Number(form.edition_id), current_page: Number(form.current_page), status: form.status, change_kind: form.change_kind,
    }, 'Estante salva.')
    if (result) await load(page)
  }

  return <>
    <Heading title="Minha estante" description="Escolha uma obra e sua edição, atualize o progresso e confira as estatísticas."><button disabled={busy || !user || optionsLoading} onClick={() => { load(page); setReload((value) => value + 1) }}>Atualizar</button></Heading>
    {!user && <Empty>Entre na página Conta para usar sua estante.</Empty>}
    <div className="columns wide-left"><section><div className="card table-wrap"><table><thead><tr><th>Obra / edição</th><th>Progresso</th><th>Ações</th></tr></thead><tbody>{items.map((item) => <tr key={item.id}><td><strong>{item.title}</strong><small>{[item.edition_title, item.language].filter(Boolean).join(' · ')}</small></td><td>{item.current_page} páginas<small>{statuses[item.status] || item.status}</small></td><td><div className="actions"><button disabled={busy} onClick={() => { setForm({ ...empty, ...item }); setFormError('') }}>Editar</button><button className="danger" disabled={busy} onClick={async () => { if (window.confirm(`Remover “${item.title}” da estante?`)) { const result = await perform(`/me/shelf/${item.work_id}`, 'DELETE', undefined, 'Item removido.'); if (result) load(page) } }}>Remover</button></div></td></tr>)}</tbody></table>{!items.length && <Empty>Sua estante está vazia.</Empty>}</div><Pages page={page} count={items.length} total={pagination.total} pageSize={pagination.page_size} onChange={load} />
      <section className="card"><div className="section-heading"><h2>Estatísticas</h2><button disabled={busy || !user} onClick={async () => { const result = await perform('/me/statistics'); if (result) setStatistics(result.data) }}>Consultar</button></div>{statistics ? <Json value={statistics} /> : <p>Consulte os totais calculados pelo backend.</p>}</section>
    </section>
    <section className="card"><h2>Adicionar ou atualizar</h2>{form.title && <p>{form.title}</p>}<form onSubmit={save}><fieldset disabled={busy || !user}>
      <Field label="Obra"><select name="work_id" required value={workId} onChange={chooseWork} disabled={works.loading || Boolean(works.error)}>
        <option value="">{works.loading ? 'Carregando obras…' : 'Selecione uma obra'}</option>
        {workId && !selectedWork && <option value={workId} disabled>{form.title || 'Obra selecionada'} (indisponível)</option>}
        {works.items.map((work) => <option key={work.id} value={work.id}>{work.title}{work.authors?.length ? ` — ${work.authors.map((author) => author.name).join(', ')}` : ''}</option>)}
      </select></Field>
      {works.error && <p className="inline-error" role="alert">{works.error} Use “Atualizar” para tentar novamente.</p>}
      {user && !works.loading && !works.error && !works.items.length && <small>Não há obras publicadas disponíveis. <a href="#create-work">Cadastrar obra</a></small>}
      <Field label="Edição"><select name="edition_id" required value={String(form.edition_id)} onChange={change} disabled={!workId || optionsLoading || Boolean(editions.error) || !currentEditions.length}>
        <option value="">{editions.loading ? 'Carregando edições…' : !workId ? 'Escolha uma obra primeiro' : 'Selecione uma edição'}</option>
        {form.edition_id && !selectedEdition && <option value={form.edition_id} disabled>{form.edition_title || 'Edição selecionada'} (indisponível)</option>}
        {currentEditions.map((edition) => <option key={edition.id} value={edition.id}>{editionLabel(edition)}</option>)}
      </select></Field>
      {editions.error && <p className="inline-error" role="alert">{editions.error} Use “Atualizar” para tentar novamente.</p>}
      {workId && !optionsLoading && !editions.error && !currentEditions.length && <small>Esta obra não possui edição publicada disponível para leitura.</small>}
      <Field label="Página atual" name="current_page" type="number" min="0" max={selectedEdition?.page_count} required value={form.current_page} onChange={change} />
      <Field label="Estado"><select name="status" value={form.status} onChange={change}>{Object.entries(statuses).map(([value, label]) => <option value={value} key={value}>{label}</option>)}</select></Field>
      <Field label="Tipo de alteração"><select name="change_kind" value={form.change_kind} onChange={change}><option value="advance">Avançar leitura</option><option value="correction">Corrigir progresso</option><option value="edition_change">Trocar edição</option></select></Field>
      <small>Planejado exige página zero. Use “Corrigir progresso” para reduzir páginas e “Trocar edição” ao mudar a edição.</small>
      {formError && <p className="inline-error" role="alert">{formError}</p>}
      <button className="primary" disabled={busy || !user || !canSave}>{busy ? 'Aguarde…' : 'Salvar na estante'}</button>
    </fieldset></form></section></div>
  </>
}
