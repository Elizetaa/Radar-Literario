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

  return <div className="app-shell min-h-screen" style={{ display: 'grid', placeItems: 'center', padding: 16 }}>
    <form className="form-stack" onSubmit={submit} style={{ width: 'min(420px, 100%)' }}>
      <h1>{mode === 'login' ? 'Entrar no Radar Literário' : 'Criar conta'}</h1>
      {mode === 'register' && <>
        <label>Nome<input required value={form.name} onChange={set('name')} /></label>
        <label>Apelido<input required pattern="[a-z0-9_]{3,30}" placeholder="leitor_123" value={form.nickname} onChange={set('nickname')} /></label>
        <label>Data de nascimento<input required type="date" max={today} value={form.birth_date} onChange={set('birth_date')} /></label>
      </>}
      <label>E-mail<input required type="email" autoComplete="email" value={form.email} onChange={set('email')} /></label>
      <label>Senha<input required type="password" minLength="8" autoComplete={mode === 'login' ? 'current-password' : 'new-password'} value={form.password} onChange={set('password')} /></label>
      {error && <p className="form-error" role="alert">{error}</p>}
      <button className="primary-button" disabled={busy}>{busy ? 'Aguarde…' : mode === 'login' ? 'Entrar' : 'Cadastrar e entrar'}</button>
      <button type="button" className="text-button" onClick={() => { setError(''); setMode(mode === 'login' ? 'register' : 'login') }}>
        {mode === 'login' ? 'Ainda não tem conta? Cadastre-se' : 'Já tem conta? Entrar'}
      </button>
    </form>
  </div>
}
