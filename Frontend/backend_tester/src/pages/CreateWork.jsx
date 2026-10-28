import { useCallback, useEffect, useState } from 'react'
import { Field, Heading, Submit, useApi } from '../ui.jsx'
import { buildWorkPayload, catalogRoles } from '../workForm.js'

const resources = [['authors', 'Autor'], ['genres', 'Gênero'], ['publishers', 'Editora']]

export default function CreateWork() {
  const { perform, busy, user } = useApi()
  const allowed = catalogRoles.includes(user?.role)
  const [options, setOptions] = useState({ authors: [], genres: [], publishers: [] })
  const [loading, setLoading] = useState(false)
  const [loadError, setLoadError] = useState('')
  const [formError, setFormError] = useState('')
  const [authors, setAuthors] = useState([])
  const [genres, setGenres] = useState([])
  const [status, setStatus] = useState('published')
  const [withEdition, setWithEdition] = useState(false)
  const [created, setCreated] = useState(null)

  const loadOptions = useCallback(async (isActive = () => true) => {
    setLoading(true)
    setLoadError('')
    try {
      const next = {}
      for (const [resource] of resources) {
        next[resource] = []
        let page = 1
        while (isActive()) {
          const result = await perform(`/admin/${resource}?page=${page}&page_size=100`)
          if (!isActive()) return
          if (!result) throw new Error('Não foi possível carregar as opções. Confira sua sessão e tente atualizar.')
          next[resource].push(...result.data)
          if (result.data.length < 100) break
          page++
        }
        if (!isActive()) return
      }
      setOptions(next)
    } catch (error) {
      if (isActive()) setLoadError(error.message)
    } finally {
      if (isActive()) setLoading(false)
    }
  }, [perform])

  useEffect(() => {
    let active = true
    if (allowed) loadOptions(() => active)
    return () => { active = false }
  }, [allowed, user?.id, loadOptions])

  function toggle(id, values, setter) {
    setter(values.includes(id) ? values.filter((value) => value !== id) : [...values, id])
  }

  async function addReference(event, resource) {
    event.preventDefault()
    const form = event.currentTarget
    const name = String(new FormData(form).get('name') || '').trim()
    if (!name) { setFormError('Informe o nome antes de cadastrar.'); return }
    const result = await perform(`/admin/${resource}`, 'POST', { name, ...(resource === 'genres' ? { active: true } : {}) }, 'Cadastro auxiliar criado.')
    if (result) {
      setOptions((current) => ({ ...current, [resource]: [...current[resource], result.data] }))
      if (resource === 'authors') setAuthors((current) => [...current, result.data.id])
      if (resource === 'genres') setGenres((current) => [...current, result.data.id])
      form.reset()
      setFormError('')
    }
  }

  async function save(event) {
    event.preventDefault()
    setFormError('')
    let body
    try { body = buildWorkPayload(new FormData(event.currentTarget)) } catch (error) { setFormError(error.message); return }
    const result = await perform('/admin/works', 'POST', body, 'Obra cadastrada com sucesso.')
    if (result) setCreated({ ...result.data, hasEdition: Boolean(body.initial_edition) })
  }

  function startAgain() {
    setCreated(null)
    setAuthors([])
    setGenres([])
    setStatus('published')
    setWithEdition(false)
    setFormError('')
  }

  return <>
    <Heading title="Adicionar obra" description="Cadastre uma obra e, se desejar, sua primeira edição." />
    {!allowed ? <section className="card"><h2>{user ? 'Permissão de catálogo necessária' : 'Entre para cadastrar obras'}</h2><p>O cadastro de obras exige uma conta com permissão de catálogo: gestor de catálogo, administrador ou administrador do sistema.</p><a className="button-link primary" href="#account">Ir para Conta</a></section> : <>
      {created ? <section className="card created-work" role="status"><span className="badge">Obra #{created.id}</span><h2>{created.title}</h2><p>{created.status === 'published' ? 'A obra foi publicada e já pode ser consultada no catálogo.' : 'A obra foi salva como rascunho. Ela ainda não aparece no catálogo público.'}{created.hasEdition ? ' A primeira edição também foi criada.' : ' Para registrar leituras, cadastre também uma edição.'}</p><div className="actions"><a className="button-link primary" href="#catalog">Ver catálogo</a><button onClick={startAgain}>Cadastrar outra obra</button></div></section> : <div className="columns wide-left">
        <section className="card"><h2>Dados da obra</h2>
          <form onSubmit={save}><fieldset disabled={busy || loading || Boolean(loadError)}>
            <Field label="Título" name="title" maxLength="300" required autoFocus />
            <div className="form-row"><Field label="Idioma original" name="original_language" defaultValue="pt-br" maxLength="35" pattern="[a-zA-Z]{2,3}(-[a-zA-Z0-9]{2,8})*" required /><Field label="Primeira publicação (opcional)" name="first_publication_date" type="date" max={status === 'published' ? new Date().toISOString().slice(0, 10) : undefined} /></div>
            <Field label="Sinopse"><textarea name="synopsis" rows="4" required={status === 'published'} /></Field>
            <div className="form-row"><Field label="Estado"><select name="status" value={status} onChange={(event) => setStatus(event.target.value)}><option value="published">Publicar no catálogo</option><option value="draft">Salvar como rascunho</option></select></Field><Field label="Classificação etária"><select name="minimum_age">{[0, 10, 12, 14, 16, 18].map((age) => <option key={age} value={age}>{age === 0 ? 'Livre' : `${age} anos`}</option>)}</select></Field></div>
            <div className="form-row"><div className="choice-group"><p className="field-label">Autores</p><div className="choice-list">{options.authors.map((author) => <label className="check" key={author.id}><input type="checkbox" name="author_ids" value={author.id} checked={authors.includes(author.id)} onChange={() => toggle(author.id, authors, setAuthors)} />{author.name}</label>)}{!options.authors.length && <small>Nenhum autor cadastrado. Use o cadastro rápido ao lado.</small>}</div></div><div className="choice-group"><p className="field-label">Gêneros ativos</p><div className="choice-list">{options.genres.filter((genre) => genre.active).map((genre) => <label className="check" key={genre.id}><input type="checkbox" name="genre_ids" value={genre.id} checked={genres.includes(genre.id)} onChange={() => toggle(genre.id, genres, setGenres)} />{genre.name}</label>)}{!options.genres.some((genre) => genre.active) && <small>Nenhum gênero ativo. Use o cadastro rápido ao lado.</small>}</div></div></div>
            <small>Para publicar, preencha a sinopse e selecione pelo menos um autor e um gênero ativo.</small>
            <div className="edition-form"><label className="check"><input type="checkbox" name="with_edition" checked={withEdition} onChange={(event) => setWithEdition(event.target.checked)} /> Cadastrar a primeira edição junto com a obra</label>
              {withEdition && <div className="edition-fields">
                <small>A edição terá o mesmo estado e classificação etária da obra. Obra e edição são salvas na mesma operação.</small>
                <Field label="Título da edição (opcional)" name="edition_title" maxLength="300" placeholder="Usar o título da obra" />
                <div className="form-row"><Field label="Editora"><select name="publisher_id" required defaultValue=""><option value="">Selecione uma editora</option>{options.publishers.map((publisher) => <option key={publisher.id} value={publisher.id}>{publisher.name}</option>)}</select></Field><Field label="Número de páginas" name="page_count" type="number" min="1" max="1000000" required /></div>
                <div className="form-row"><Field label="Idioma da edição (opcional)" name="edition_language" placeholder="Usar o idioma da obra" maxLength="35" pattern="[a-zA-Z]{2,3}(-[a-zA-Z0-9]{2,8})*" /><Field label="Publicação da edição (opcional)" name="edition_publication_date" type="date" max={status === 'published' ? new Date().toISOString().slice(0, 10) : undefined} /></div>
                <Field label="ISBN (opcional quando houver justificativa)" name="isbn" maxLength="30" placeholder="ISBN-10 ou ISBN-13" /><Field label="Motivo da ausência de ISBN" name="isbn_absence_reason" maxLength="300" placeholder="Preencha se a edição não possui ISBN" /><Field label="URL da capa (opcional)" name="image_url" type="url" maxLength="2048" placeholder="https://..." />
              </div>}
            </div>
            {formError && <p className="inline-error" role="alert">{formError}</p>}
            <Submit>{status === 'published' ? 'Publicar obra' : 'Salvar rascunho'}</Submit>
          </fieldset></form>
        </section>
        <section className="card"><div className="section-heading"><h2>Cadastros rápidos</h2><button disabled={busy || loading} onClick={() => loadOptions()}>Atualizar</button></div><p>Cadastre aqui autores, gêneros e editoras que ainda não aparecem no formulário.</p>
          {loading && <p role="status">Carregando opções…</p>}{loadError && <p className="inline-error" role="alert">{loadError}</p>}
          {resources.map(([resource, title]) => <details className="reference-form" key={resource}><summary>{resource === 'publishers' ? 'Nova editora' : `Novo ${title.toLowerCase()}`}</summary><form onSubmit={(event) => addReference(event, resource)}><fieldset disabled={busy || loading || Boolean(loadError)}><Field label={`Nome ${resource === 'publishers' ? 'da editora' : resource === 'authors' ? 'do autor' : 'do gênero'}`} name="name" required maxLength="160" /><Submit>Cadastrar {title.toLowerCase()}</Submit></fieldset></form></details>)}
          <div className="hint"><strong>Obra e edição</strong><p>A obra reúne título, autores e sinopse. A edição identifica a versão que será usada na estante, com editora e número de páginas.</p></div>
        </section>
      </div>}
    </>}
  </>
}
