import { useCallback, useEffect, useState } from 'react'
import { Empty, Field, Heading, Json, Pages, Submit, useApi } from '../ui.jsx'
import { catalogRoles } from '../workForm.js'
import { buildReferencePayload, filterReferences, referenceResources, referenceValues } from '../referenceForm.js'

export default function References() {
  const { perform, busy, user } = useApi()
  const allowed = catalogRoles.includes(user?.role)
  const [resource, setResource] = useState('authors')
  const [items, setItems] = useState([])
  const [pagination, setPagination] = useState({ page: 1, page_size: 20, total: 0 })
  const [pageSize, setPageSize] = useState(20)
  const [query, setQuery] = useState('')
  const [activeFilter, setActiveFilter] = useState('')
  const [lookup, setLookup] = useState('')
  const [record, setRecord] = useState(null)
  const [editing, setEditing] = useState(true)
  const [values, setValues] = useState(() => referenceValues('authors'))
  const [formError, setFormError] = useState('')
  const [loading, setLoading] = useState(false)
  const [loadError, setLoadError] = useState(false)
  const config = referenceResources[resource]
  const path = `${allowed ? '/admin' : ''}/${resource}`

  const load = useCallback(async (page = 1, size = pageSize, isActive = () => true) => {
    setLoading(true)
    setLoadError(false)
    try {
      const result = await perform(`${path}?page=${page}&page_size=${size}`)
      if (!isActive()) return
      if (!result) { setLoadError(true); return }
      const lastPage = Math.max(1, Math.ceil(result.pagination.total / size))
      if (page > lastPage) {
        const previousPage = await perform(`${path}?page=${lastPage}&page_size=${size}`)
        if (!isActive()) return
        if (!previousPage) { setLoadError(true); return }
        setItems(previousPage.data)
        setPagination(previousPage.pagination)
      } else {
        setItems(result.data)
        setPagination(result.pagination)
      }
    } finally {
      if (isActive()) setLoading(false)
    }
  }, [path, pageSize, perform])

  useEffect(() => {
    let current = true
    setItems([])
    setPagination({ page: 1, page_size: pageSize, total: 0 })
    setRecord(null)
    setValues(referenceValues(resource))
    setEditing(true)
    setFormError('')
    setQuery('')
    setActiveFilter('')
    setLookup('')
    load(1, pageSize, () => current)
    return () => { current = false }
  }, [resource, user?.id, pageSize, load])

  function startNew() {
    setRecord(null)
    setValues(referenceValues(resource))
    setEditing(true)
    setFormError('')
  }

  function show(item, edit = false) {
    setRecord(item)
    setValues(referenceValues(resource, item))
    setEditing(edit)
    setFormError('')
  }

  async function open(item, edit = false) {
    // O contrato público não oferece GET /genres/:id.
    if (!allowed && resource === 'genres') { show(item); return }
    const result = await perform(`${path}/${item.id}`)
    if (result) show(result.data, edit)
  }

  async function findById(event) {
    event.preventDefault()
    if (!/^[1-9]\d*$/.test(lookup)) { setFormError('Informe um ID inteiro positivo.'); return }
    const result = await perform(`${path}/${lookup}`)
    if (result) show(result.data)
  }

  async function save(event) {
    event.preventDefault()
    setFormError('')
    let body
    try {
      body = buildReferencePayload(resource, values, record)
      if (!Object.keys(body).length) { setFormError('Nenhuma alteração para salvar.'); return }
    } catch (error) { setFormError(error.message); return }
    const result = await perform(`${path}${record ? `/${record.id}` : ''}`, record ? 'PATCH' : 'POST', body, record ? 'Cadastro atualizado.' : 'Cadastro criado.')
    if (result) {
      show(result.data)
      await load(pagination.page)
    }
  }

  async function remove(item) {
    const message = resource === 'genres'
      ? `Desativar o gênero “${item.name}” (#${item.id})? Ele será mantido no banco e poderá ser reativado.`
      : `Excluir ${config.singular} “${item.name}” (#${item.id})? A API impedirá a exclusão se houver ${resource === 'authors' ? 'obras' : 'edições'} vinculadas.`
    if (!window.confirm(message)) return
    const result = await perform(`${path}/${item.id}`, 'DELETE', undefined, resource === 'genres' ? 'Gênero desativado.' : 'Cadastro excluído.')
    if (result) {
      if (record?.id === item.id) {
        if (resource === 'genres') show({ ...item, active: false })
        else startNew()
      }
      await load(pagination.page)
    }
  }

  async function activate(item) {
    const result = await perform(`${path}/${item.id}`, 'PATCH', { active: true }, 'Gênero reativado.')
    if (result) {
      if (record?.id === item.id) show(result.data)
      await load(pagination.page)
    }
  }

  function switchResource(next) {
    setValues(referenceValues(next)); setRecord(null); setEditing(true); setResource(next)
  }
  const change = (name) => (event) => setValues((current) => ({ ...current, [name]: event.target.value }))
  const filtered = filterReferences(items, query, resource === 'genres' ? activeFilter : '')

  return <>
    <Heading title="Autores, gêneros e editoras" description="Liste os cadastros auxiliares, consulte os detalhes e teste suas operações de manutenção." />
    <div className="tabs" role="group" aria-label="Tipo de cadastro">{Object.entries(referenceResources).map(([key, value]) => <button key={key} className={resource === key ? 'active' : ''} aria-pressed={resource === key} disabled={busy || loading} onClick={() => switchResource(key)}>{value.title}</button>)}</div>
    {formError && <p className="inline-error" role="alert">{formError}</p>}
    <div className="columns wide-left">
      <section>
        <div className="card">
          <div className="section-heading"><h2>{config.title} <span className="badge">{pagination.total} registros</span></h2><button disabled={busy || loading} onClick={() => load(pagination.page)}>Atualizar lista</button></div>
          <p className="hint">{allowed ? 'Consulta administrativa' : 'Consulta pública'}{resource === 'genres' ? (allowed ? ' · Inclui gêneros inativos.' : ' · Exibe somente gêneros ativos.') : '.'}</p>
          <div className="form-row">
            <Field label="Filtrar nomes ou IDs nesta página" value={query} onChange={(event) => setQuery(event.target.value)} placeholder="Ex.: Machado ou 12" />
            <Field label="Registros por página"><select value={pageSize} disabled={busy || loading} onChange={(event) => setPageSize(Number(event.target.value))}>{[10, 20, 50, 100].map((size) => <option key={size} value={size}>{size}</option>)}</select></Field>
          </div>
          {resource === 'genres' && allowed && <Field label="Estado nesta página"><select value={activeFilter} onChange={(event) => setActiveFilter(event.target.value)}><option value="">Todos</option><option value="true">Ativos</option><option value="false">Inativos</option></select></Field>}
          <small>A API oferece paginação para estes cadastros. Os filtros acima atuam somente sobre os registros da página atual.</small>
          {loading && <p role="status">Carregando {config.title.toLowerCase()}…</p>}
          {loadError && <p className="inline-error" role="alert">Não foi possível carregar a lista. Confira o aviso da API e use “Atualizar lista” para tentar novamente.</p>}
          {!loading && !loadError && <>
            {filtered.length ? <div className="table-wrap"><table><caption className="sr-only">Lista de {config.title.toLowerCase()}</caption><thead><tr><th scope="col">ID</th><th scope="col">Nome</th><th scope="col">{resource === 'genres' ? 'Estado' : 'País de origem'}</th><th scope="col">Ações</th></tr></thead><tbody>{filtered.map((item) => <tr key={item.id}><td>#{item.id}</td><td>{item.name}</td><td>{resource === 'genres' ? <span className="badge">{item.active ? 'Ativo' : 'Inativo'}</span> : item.country_of_origin || '—'}</td><td><div className="actions"><button disabled={busy} onClick={() => open(item)}>Detalhes</button>{allowed && <><button disabled={busy} onClick={() => open(item, true)}>Editar</button>{resource === 'genres' && !item.active ? <button disabled={busy} onClick={() => activate(item)}>Reativar</button> : <button className="danger" disabled={busy} onClick={() => remove(item)}>{resource === 'genres' ? 'Desativar' : 'Excluir'}</button>}</>}</div></td></tr>)}</tbody></table></div> : <Empty>{items.length ? 'Nenhum registro corresponde aos filtros desta página.' : 'Nenhum cadastro encontrado.'}</Empty>}
            <Pages page={pagination.page} count={items.length} total={pagination.total} pageSize={pagination.page_size} onChange={(page) => load(page)} />
          </>}
        </div>
        {(allowed || resource !== 'genres') && <form className="card search" onSubmit={findById}><Field label={`Consultar ${config.singular} por ID`} value={lookup} onChange={(event) => setLookup(event.target.value)} type="number" min="1" step="1" required placeholder="ID do cadastro" /><Submit>Consultar ID</Submit></form>}
      </section>
      <section className="card">
        <div className="section-heading"><h2>{record ? `${editing ? 'Editar' : 'Detalhes'}: ${record.name}` : config.create}</h2>{allowed && <button disabled={busy} onClick={startNew}>{config.create}</button>}</div>
        {record && !editing ? <>
          <dl><dt>ID</dt><dd>#{record.id}</dd><dt>Nome</dt><dd>{record.name}</dd>{resource === 'authors' && <><dt>Biografia</dt><dd className="preserve-lines">{record.biography || 'Não informada'}</dd><dt>Nascimento</dt><dd>{record.birth_date || 'Não informado'}</dd><dt>Falecimento</dt><dd>{record.death_date || 'Não informado'}</dd><dt>Sexo</dt><dd>{record.sex || 'Não informado'}</dd></>}{resource !== 'genres' && <><dt>País</dt><dd>{record.country_of_origin || 'Não informado'}</dd></>}{resource === 'publishers' && <><dt>Site</dt><dd>{record.website_url || 'Não informado'}</dd></>}{resource === 'genres' && <><dt>Estado</dt><dd>{record.active ? 'Ativo' : 'Inativo'}</dd></>}</dl>
          {allowed && <div className="actions"><button className="primary" disabled={busy} onClick={() => setEditing(true)}>Editar cadastro</button></div>}
          <details className="hint"><summary>JSON do cadastro</summary><Json value={record} /></details>
        </> : allowed ? <form onSubmit={save}><fieldset disabled={busy || loading}>
          {record && <small>Editando {config.singular} #{record.id}. Somente os campos alterados serão enviados via PATCH.</small>}
          <Field label="Nome" value={values.name} onChange={change('name')} required maxLength={resource === 'genres' ? 100 : 160} />
          {resource === 'authors' && <>
            <Field label="Biografia (opcional)"><textarea value={values.biography} onChange={change('biography')} rows="5" maxLength="30000" /></Field>
            <div className="form-row"><Field label="Nascimento (opcional)" type="date" value={values.birth_date} onChange={change('birth_date')} /><Field label="Falecimento (opcional)" type="date" value={values.death_date} onChange={change('death_date')} /></div>
            <Field label="Sexo (opcional)" value={values.sex} onChange={change('sex')} maxLength="80" />
          </>}
          {resource !== 'genres' && <Field label="País de origem (opcional)" value={values.country_of_origin} onChange={change('country_of_origin')} maxLength="100" />}
          {resource === 'publishers' && <Field label="Site (opcional)" value={values.website_url} onChange={change('website_url')} type="url" maxLength="2048" placeholder="https://..." />}
          {resource === 'genres' && <label className="check"><input type="checkbox" checked={values.active} onChange={(event) => setValues((current) => ({ ...current, active: event.target.checked }))} />Gênero ativo</label>}
          <div className="actions"><Submit>{record ? 'Salvar alterações' : 'Cadastrar'}</Submit>{record && <button type="button" onClick={() => show(record)}>Cancelar edição</button>}</div>
          {resource === 'genres' ? <small>O DELETE desativa o gênero; use “Reativar” para torná-lo disponível novamente.</small> : <small>A API protege os vínculos: {resource === 'authors' ? 'autores com obras' : 'editoras com edições'} não podem ser excluídos.</small>}
        </fieldset></form> : <><p>Entre com uma conta de gestor de catálogo, administrador ou administrador do sistema para criar, editar e excluir cadastros. As listagens públicas estão disponíveis ao lado.</p><a className="button-link primary" href="#account">Ir para Conta</a></>}
      </section>
    </div>
  </>
}
