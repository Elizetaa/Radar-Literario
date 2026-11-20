import { createContext, useContext, useEffect, useState } from 'react'

export const ApiContext = createContext(null)
export const useApi = () => useContext(ApiContext)

export function Submit({ children = 'Salvar' }) {
  const { busy } = useApi()
  return <button className="primary" disabled={busy}>{busy ? 'Aguarde…' : children}</button>
}

export function Field({ label, children, ...input }) {
  return <label>{label}{children || <input {...input} />}</label>
}

export function Empty({ children = 'Nenhum registro encontrado.' }) {
  return <p className="empty">{children}</p>
}

export function Json({ value }) {
  const [copied, setCopied] = useState(false)
  const [error, setError] = useState(false)
  useEffect(() => { setCopied(false); setError(false) }, [value])
  async function copy() {
    try { await navigator.clipboard.writeText(JSON.stringify(value, null, 2)); setCopied(true); setError(false) }
    catch { setError(true) }
  }
  return <div className="json"><div className="json-toolbar"><button type="button" onClick={copy}>{copied ? 'Copiado' : 'Copiar JSON'}</button>{error && <small role="status">Selecione o texto para copiar.</small>}</div><pre>{JSON.stringify(value, null, 2)}</pre></div>
}

export function Heading({ title, description, children }) {
  return <div className="page-heading"><div><h1>{title}</h1><p>{description}</p></div>{children}</div>
}

export function Pages({ page, count = 0, total, pageSize = 20, onChange }) {
  const { busy } = useApi()
  const knownTotal = Number.isFinite(total)
  const lastPage = knownTotal ? Math.max(1, Math.ceil(total / pageSize)) : null
  return <div className="pagination"><button disabled={busy || page <= 1} onClick={() => onChange(page - 1)}>← Anterior</button><span>Página {page}{knownTotal ? ` de ${lastPage} · ${total} registro(s)` : ''}</span><button disabled={busy || (knownTotal ? page >= lastPage : count < pageSize)} onClick={() => onChange(page + 1)}>Próxima →</button></div>
}
