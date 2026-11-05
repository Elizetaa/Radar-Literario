import { useState } from 'react'
import { Json } from './ui.jsx'

export default function RequestHistory({ entries, onClear }) {
  const [selected, setSelected] = useState(null)
  if (!entries.length) return null
  const entry = entries.find((item) => item.id === selected) || entries[0]
  return <details className="response"><summary>Última resposta · {entries[0].method} {entries[0].path} · HTTP {entries[0].status} · {entries[0].duration} ms</summary>
    <div className="section-heading history-heading"><h2>Histórico da sessão</h2><button type="button" onClick={onClear}>Limpar histórico</button></div>
    <p>Até 30 requisições nesta aba. Senhas e tokens ficam ocultos. O tempo inclui a obtenção do CSRF nas escritas.</p>
    <FieldSelector entries={entries} selected={entry.id} onChange={setSelected} />
    <dl><dt>Requisição</dt><dd><code>{entry.method} /api/v1{entry.path}</code></dd><dt>Resposta</dt><dd>HTTP {entry.status} · {entry.duration} ms · {new Date(entry.time).toLocaleTimeString('pt-BR')}</dd></dl>
    {entry.body !== undefined && <details><summary>Corpo enviado</summary><Json value={entry.body} /></details>}
    <Json value={entry.payload} />
  </details>
}

function FieldSelector({ entries, selected, onChange }) {
  return <label>Requisição registrada<select value={selected} onChange={(event) => onChange(Number(event.target.value))}>{entries.map((entry) => <option key={entry.id} value={entry.id}>{entry.method} {entry.path} · HTTP {entry.status} · {entry.duration} ms</option>)}</select></label>
}
