import { useState } from 'react'
import { Empty, Field, Heading, Json, Submit, useApi } from '../ui.jsx'

export default function Account() {
  const { perform, busy, user, setUser } = useApi()
  const [mode, setMode] = useState('login')

  async function submit(event) {
    event.preventDefault()
    const form = event.currentTarget
    const body = Object.fromEntries(new FormData(form))
    if (mode === 'register') {
      const result = await perform('/readers', 'POST', body, 'Conta criada. Entre com o e-mail e a senha cadastrados.')
      if (result) { form.reset(); setMode('login') }
    } else {
      const result = await perform('/auth/login', 'POST', body, 'Sessão iniciada.')
      if (result) { setUser(result.data.user); form.reset() }
    }
  }

  return <>
    <Heading title="Conta e sessão" description="Cadastre um leitor, entre e confira os dados da sessão atual." />
    <div className="columns">
      <section className="card"><div className="tabs"><button className={mode === 'login' ? 'active' : ''} onClick={() => setMode('login')}>Entrar</button><button className={mode === 'register' ? 'active' : ''} onClick={() => setMode('register')}>Cadastrar leitor</button></div>
        <form onSubmit={submit} key={mode}><fieldset disabled={busy}>
          {mode === 'register' && <><Field label="Nome" name="name" required maxLength={120} /><Field label="Apelido" name="nickname" pattern="[a-z0-9_]{3,30}" placeholder="leitor_teste" required /><Field label="Data de nascimento" name="birth_date" type="date" max={new Date().toISOString().slice(0, 10)} required /></>}
          <Field label="E-mail" name="email" type="email" autoComplete="email" required />
          <Field label="Senha" name="password" type="password" minLength={mode === 'register' ? 8 : undefined} autoComplete={mode === 'register' ? 'new-password' : 'current-password'} required />
          {mode === 'register' && <small>A nova senha precisa ter entre 8 e 72 bytes. O apelido aceita letras minúsculas, números e sublinhado.</small>}
          <Submit>{mode === 'login' ? 'Entrar' : 'Criar leitor'}</Submit>
        </fieldset></form>
      </section>
      <section className="card"><h2>Sessão atual</h2>{user ? <><dl><dt>Nome</dt><dd>{user.name}</dd><dt>Apelido</dt><dd>@{user.nickname}</dd><dt>E-mail</dt><dd>{user.email}</dd><dt>Permissão</dt><dd>{user.role || 'Leitor'}</dd></dl><details><summary>Perfil completo</summary><Json value={user} /></details></> : <Empty>Entre para consultar seu perfil.</Empty>}
        <div className="actions"><button disabled={busy} onClick={async () => { const result = await perform('/me'); if (result) setUser(result.data) }}>Consultar sessão</button><button disabled={busy || !user} onClick={async () => { const result = await perform('/auth/logout', 'POST', undefined, 'Sessão encerrada.'); if (result) setUser(null) }}>Sair</button></div>
      </section>
    </div>
  </>
}
