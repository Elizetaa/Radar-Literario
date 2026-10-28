import { useState } from 'react'
import { Field, Heading, Submit, useApi } from '../ui.jsx'
import { buildConsolePath, consolePresets, consoleResources, consoleVariables, parseConsoleBody, splitConsolePath } from '../consoleRequests.js'

export default function Console() {
  const { perform, busy } = useApi()
  const [path, setPath] = useState('/works')
  const [query, setQuery] = useState(splitConsolePath('/works?page=1&page_size=20').query)
  const [variables, setVariables] = useState({})
  const [method, setMethod] = useState('GET')
  const [body, setBody] = useState('')
  const [error, setError] = useState('')
  const [resource, setResource] = useState('works')
  const [description, setDescription] = useState('')
  const variableNames = consoleVariables(`${path} ${query.filter((item) => item.enabled).map((item) => item.value).join(' ')}`)
  let preview = ''
  try { preview = buildConsolePath(path, query, variables) } catch { /* O formulário mostra o erro no envio. */ }

  function applyPreset(preset) {
    const parsed = splitConsolePath(preset.path)
    setPath(parsed.path)
    setQuery(parsed.query)
    setMethod(preset.method)
    setBody(preset.body === undefined ? '' : JSON.stringify(preset.body, null, 2))
    setVariables({})
    setDescription(preset.description)
    setError('')
  }

  function updateQuery(index, changes) {
    setQuery((items) => items.map((item, position) => position === index ? { ...item, ...changes } : item))
  }

  function formatBody() {
    setError('')
    if (!body.trim()) return
    try { setBody(JSON.stringify(JSON.parse(body), null, 2)) } catch { setError('O corpo precisa conter JSON válido.') }
  }

  async function submit(event) {
    event.preventDefault()
    setError('')
    let target, input
    try {
      target = buildConsolePath(path, query, variables)
      input = parseConsoleBody(method, body)
    } catch (failure) { setError(failure.message); return }
    if (method === 'DELETE' && !window.confirm(`Confirmar DELETE ${target}? Essa ação altera os dados reais do backend.`)) return
    await perform(target, method, input)
  }

  return <>
    <Heading title="Console da API" description="Monte requisições, escolha exemplos de CRUD e confira a resposta HTTP. A API verifica as permissões da conta." />
    <div className="columns wide-left">
      <section className="card">
        <h2>Requisição</h2>
        <form onSubmit={submit}><fieldset disabled={busy}>
          <div className="request-line">
            <Field label="Método"><select value={method} onChange={(event) => setMethod(event.target.value)}>{['GET', 'POST', 'PUT', 'PATCH', 'DELETE'].map((value) => <option key={value}>{value}</option>)}</select></Field>
            <Field label="Caminho após /api/v1" value={path} onChange={(event) => { setPath(event.target.value); setDescription('') }} required placeholder="/admin/works/{id}" />
          </div>
          {variableNames.length > 0 && <div className="form-row">{variableNames.map((name) => <Field key={name} label={`Parâmetro ${name}`} value={variables[name] || ''} onChange={(event) => setVariables((current) => ({ ...current, [name]: event.target.value }))} placeholder={name === 'isbn' ? 'ISBN da edição' : 'ID real do registro'} required />)}</div>}
          <details className="console-parameters" open={query.length > 0}>
            <summary>Parâmetros de consulta ({query.filter((item) => item.enabled && item.name.trim()).length})</summary>
            <div className="console-query-list">{query.map((item, index) => <div className="console-query-row" key={index}>
              <label className="check"><input type="checkbox" aria-label={`Incluir parâmetro ${index + 1}`} checked={item.enabled} onChange={(event) => updateQuery(index, { enabled: event.target.checked })} /></label>
              <Field label={`Nome ${index + 1}`} value={item.name} onChange={(event) => updateQuery(index, { name: event.target.value })} placeholder="page_size" />
              <Field label={`Valor ${index + 1}`} value={item.value} onChange={(event) => updateQuery(index, { value: event.target.value })} placeholder="20" />
              <button type="button" aria-label={`Remover parâmetro ${index + 1}`} onClick={() => setQuery((items) => items.filter((_, position) => index !== position))}>×</button>
            </div>)}</div>
            <button type="button" onClick={() => setQuery((items) => [...items, { name: '', value: '', enabled: true }])}>+ Parâmetro</button>
          </details>
          {method !== 'GET' && <>
            <Field label="Corpo JSON (deixe vazio para enviar sem corpo)"><textarea spellCheck="false" className="code-input" rows="13" value={body} onChange={(event) => setBody(event.target.value)} /></Field>
            <div><button type="button" onClick={formatBody} disabled={!body.trim()}>Formatar JSON</button></div>
          </>}
          {description && <p className="hint">{description}</p>}
          {preview && <code className="console-preview">{method} /api/v1{preview}</code>}
          {error && <p role="alert" className="inline-error">{error}</p>}
          <Submit>Enviar requisição</Submit>
        </fieldset></form>
        <p className="hint">Os exemplos apenas preenchem o formulário. Revise os IDs e o corpo antes de enviar. Cookies e CSRF são enviados automaticamente; escritas alteram o banco local.</p>
      </section>
      <section className="card">
        <h2>Exemplos de requisição</h2>
        <Field label="Recurso"><select value={resource} disabled={busy} onChange={(event) => setResource(event.target.value)}>{consoleResources.map(([value, label]) => <option key={value} value={value}>{label}</option>)}</select></Field>
        <div className="quick-links console-presets">{consolePresets.filter((preset) => preset.resource === resource).map((preset) => <button type="button" disabled={busy} key={`${preset.id}-${preset.label}`} onClick={() => applyPreset(preset)}>
          <span className={`method method-${preset.method.toLowerCase()}`}>{preset.method}</span> {preset.label}<small>{preset.path}</small>
        </button>)}</div>
      </section>
    </div>
  </>
}
