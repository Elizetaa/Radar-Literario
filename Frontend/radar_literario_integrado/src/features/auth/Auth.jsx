import { useState } from 'react'
import { backendApi } from '../../services/backendApi.js'

const today = new Date().toISOString().slice(0, 10)

export default function Auth({ onDone }) {
  const [mode, setMode] = useState('login')
  const [form, setForm] = useState({ name: '', nickname: '', email: '', password: '', birth_date: '' })
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)
  const set = (key) => (event) => setForm({ ...form, [key]: key === 'nickname' ? event.target.value.toLowerCase() : event.target.value })

  async function submit(event) {
    event.preventDefault()
    setError(''); setBusy(true)
    try {
      if (mode === 'register') await backendApi.register(form)
      await backendApi.login({ email: form.email, password: form.password })
      await onDone()
    } catch (cause) {
      setError(cause.message || 'Não foi possível concluir.')
    } finally { setBusy(false) }
  }

  // [MOBILE] As sugestões de preenchimento e o teclado seguem o tipo de cada campo.
  // O formulário usa os mesmos tamanhos de fonte e de toque dos demais fluxos.
  return <div className="app-shell auth-shell">
    <form className="form-stack" onSubmit={submit}>
      <h1>{mode === 'login' ? 'Entrar no Radar Literário' : 'Criar conta'}</h1>
      {mode === 'register' && <>
        <label>Nome<input required autoComplete="name" value={form.name} onChange={set('name')} /></label>
        <label>Apelido<input required autoComplete="nickname" autoCapitalize="none" spellCheck={false} pattern="[a-z0-9_]{3,30}" placeholder="leitor_123" value={form.nickname} onChange={set('nickname')} /></label>
        <label>Data de nascimento<input required type="date" autoComplete="bday" max={today} value={form.birth_date} onChange={set('birth_date')} /></label>
      </>}
      <label>E-mail<input required type="email" inputMode="email" autoCapitalize="none" autoComplete="email" value={form.email} onChange={set('email')} /></label>
      <label>Senha<input required type="password" minLength="8" autoComplete={mode === 'login' ? 'current-password' : 'new-password'} value={form.password} onChange={set('password')} /></label>
      {/* [ALTERADO] role=alert anuncia a falha de autenticação junto ao formulário. */}
      {error && <p className="form-error" role="alert">{error}</p>}
      <button className="primary-button" disabled={busy}>{busy ? 'Aguarde…' : mode === 'login' ? 'Entrar' : 'Cadastrar e entrar'}</button>
      <button type="button" className="text-button" onClick={() => { setError(''); setMode(mode === 'login' ? 'register' : 'login') }}>
        {mode === 'login' ? 'Ainda não tem conta? Cadastre-se' : 'Já tem conta? Entrar'}
      </button>
    </form>
  </div>
}
