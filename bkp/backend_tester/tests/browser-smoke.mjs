// Run with: node tests/browser-smoke.mjs
// Uses an isolated Edge/Chromium profile and a local in-memory API; no database is used.
import assert from 'node:assert/strict'
import { spawn } from 'node:child_process'
import { existsSync } from 'node:fs'
import { mkdir, mkdtemp, rm, writeFile } from 'node:fs/promises'
import { createServer as createHttpServer } from 'node:http'
import { tmpdir } from 'node:os'
import { join, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { setTimeout as delay } from 'node:timers/promises'
import { createServer as createViteServer } from 'vite'

const projectRoot = fileURLToPath(new URL('..', import.meta.url))
const frontendPort = Number(process.env.SMOKE_FRONTEND_PORT || 5177)
const apiPort = Number(process.env.SMOKE_API_PORT || 18087)
const debugPort = Number(process.env.SMOKE_DEBUG_PORT || 19227)
const browserPath = [
  process.env.SMOKE_BROWSER_PATH,
  'C:/Program Files (x86)/Microsoft/Edge/Application/msedge.exe',
  'C:/Program Files/Microsoft/Edge/Application/msedge.exe',
  'C:/Program Files/Google/Chrome/Application/chrome.exe',
  '/usr/bin/chromium', '/usr/bin/chromium-browser', '/usr/bin/google-chrome',
].find((candidate) => candidate && existsSync(candidate))
assert(browserPath, 'Set SMOKE_BROWSER_PATH to an installed Edge/Chromium executable.')

const now = '2026-10-27T12:00:00Z'
const state = {
  authors: Array.from({ length: 23 }, (_, index) => ({ id: index + 1, name: `Autor ${String(index + 1).padStart(2, '0')}`, biography: '', birth_date: null, death_date: null, country_of_origin: '', sex: null })),
  genres: [{ id: 1, name: 'Ficção', active: true }, { id: 2, name: 'Inativo', active: false }],
  publishers: [{ id: 1, name: 'Editora de teste', website_url: '', country_of_origin: 'Brasil' }],
  works: [{ id: 1, title: 'Obra publicada', synopsis: 'Sinopse de teste.', original_language: 'pt-br', first_publication_date: null, status: 'published', minimum_age: 0, author_ids: [1], genre_ids: [1], version: 1, average_rating: null, review_count: 0, reader_count: 0, affinity: 0, created_at: now, updated_at: now }],
  editions: Array.from({ length: 12 }, (_, index) => ({ id: index + 1, work_id: 1, publisher_id: 1, title: `Edição de teste ${index + 1}`, isbn: null, isbn_absence_reason: 'Publicação independente', edition_number: index + 1, page_count: 120 + index, publication_date: null, language: 'pt-br', image_url: '', status: 'published', minimum_age: 0, version: 1, created_at: now, updated_at: now })),
}
const requests = []
const user = { id: 9, name: 'Pessoa de teste', nickname: 'teste', email: 'teste@example.test', role: 'system_administrator', birth_date: '1990-01-01' }
const expand = (resource, value) => resource === 'works'
  ? { ...value, authors: state.authors.filter((a) => value.author_ids.includes(a.id)), genres: state.genres.filter((g) => value.genre_ids.includes(g.id)) }
  : value

const api = createHttpServer(async (request, response) => {
  try {
    const url = new URL(request.url, `http://127.0.0.1:${apiPort}`)
    const method = request.method
    const chunks = []
    for await (const chunk of request) chunks.push(chunk)
    const raw = Buffer.concat(chunks).toString()
    const body = raw ? JSON.parse(raw) : undefined
    requests.push({ method, path: url.pathname, query: Object.fromEntries(url.searchParams), body })
    const reply = (status, payload) => { response.writeHead(status, { 'Content-Type': 'application/json' }); response.end(status === 204 ? '' : JSON.stringify(payload)) }
    if (url.pathname === '/api/v1/auth/csrf') return reply(200, { data: { csrf_token: 'isolated-test-token' } })
    if (url.pathname === '/api/v1/me') return reply(200, { data: user })
    if (url.pathname === '/api/v1/auth/logout') return reply(204)
    if (url.pathname === '/api/v1/me/shelf' && method === 'GET') return reply(200, { data: [], pagination: { page: Number(url.searchParams.get('page') || 1), page_size: Number(url.searchParams.get('page_size') || 20), total: 0 } })
    const match = url.pathname.match(/^\/api\/v1\/(admin\/)?(works|editions|authors|genres|publishers)(?:\/(\d+))?(?:\/(editions))?$/)
    if (!match) return reply(404, { error: { code: 'NOT_FOUND', message: 'Endpoint indisponível no teste isolado.' } })
    let [, admin, resource, rawId, children] = match
    const id = Number(rawId)
    const parentId = children ? id : 0
    if (children) resource = children
    const rows = state[resource]
    if (method === 'GET') {
      if (id && !children) {
        const row = rows.find((value) => value.id === id)
        return row ? reply(200, { data: expand(resource, row) }) : reply(404, { error: { code: 'NOT_FOUND', message: 'Recurso não encontrado.' } })
      }
      const page = Number(url.searchParams.get('page') || 1)
      const pageSize = Number(url.searchParams.get('page_size') || 20)
      let filtered = rows.filter((row) => (admin || !['works', 'editions'].includes(resource) || row.status === 'published')
        && (resource !== 'genres' || admin || row.active)
        && (!url.searchParams.get('status') || row.status === url.searchParams.get('status'))
        && (!url.searchParams.get('work_id') || row.work_id === Number(url.searchParams.get('work_id')))
        && (!parentId || row.work_id === parentId)
        && (resource !== 'works' || !url.searchParams.get('author') || state.authors.some((author) => row.author_ids.includes(author.id) && author.name.toLowerCase().includes(url.searchParams.get('author').toLowerCase())))
        && (resource !== 'works' || !url.searchParams.get('genre_id') || row.genre_ids.includes(Number(url.searchParams.get('genre_id'))))
        && (resource !== 'works' || !url.searchParams.get('language') || state.editions.some((edition) => edition.work_id === row.id && edition.status === 'published' && edition.language === url.searchParams.get('language')))
        && (!url.searchParams.get('title') || row.title?.toLowerCase().includes(url.searchParams.get('title').toLowerCase())))
      return reply(200, { data: filtered.slice((page - 1) * pageSize, page * pageSize).map((row) => expand(resource, row)), pagination: { page, page_size: pageSize, total: filtered.length } })
    }
    if (!admin) return reply(403, { error: { code: 'FORBIDDEN', message: 'Operação administrativa.' } })
    if (request.headers['x-csrf-token'] !== 'isolated-test-token') return reply(403, { error: { code: 'CSRF', message: 'Token CSRF ausente.' } })
    if (method === 'POST') {
      if (resource === 'genres' && rows.some((row) => row.name.toLowerCase() === body.name.toLowerCase())) return reply(409, { error: { code: 'CONFLICT', message: 'A operação conflita com outro registro ou alteração. Atualize e tente novamente.' } })
      const row = { ...(resource === 'works' ? { author_ids: [], genre_ids: [], synopsis: '', first_publication_date: null } : {}), ...body, id: Math.max(0, ...rows.map((value) => value.id)) + 1 }
      if (resource === 'works' || resource === 'editions') Object.assign(row, { version: 1, created_at: now, updated_at: now })
      rows.push(row)
      if (resource === 'works' && body.initial_edition) {
        state.editions.push({ ...body.initial_edition, id: Math.max(...state.editions.map((value) => value.id)) + 1, work_id: row.id, version: 1, created_at: now, updated_at: now })
      }
      return reply(201, { data: expand(resource, row) })
    }
    const row = rows.find((value) => value.id === id)
    if (!row) return reply(404, { error: { code: 'NOT_FOUND', message: 'Recurso não encontrado.' } })
    if (method === 'PATCH') { Object.assign(row, body); if (row.version) row.version++; return reply(200, { data: expand(resource, row) }) }
    if (method === 'DELETE') {
      if (resource === 'works' || resource === 'editions') row.status = 'archived'
      else if (resource === 'genres') row.active = false
      else rows.splice(rows.indexOf(row), 1)
      return reply(204)
    }
    return reply(405, { error: { code: 'METHOD_NOT_ALLOWED', message: 'Método inválido.' } })
  } catch (error) { response.writeHead(500, { 'Content-Type': 'application/json' }); response.end(JSON.stringify({ error: { message: error.message } })) }
})

let socket
let browser
let vite
let profile
let nextId = 0
const pending = new Map()
const errors = []
async function cdp(method, params = {}) {
  const id = ++nextId
  return new Promise((resolve, reject) => { pending.set(id, { resolve, reject }); socket.send(JSON.stringify({ id, method, params })) })
}
async function evaluate(expression) {
  const result = await cdp('Runtime.evaluate', { expression, returnByValue: true, awaitPromise: true, userGesture: true })
  assert(!result.exceptionDetails, JSON.stringify(result.exceptionDetails))
  return result.result.value
}
async function until(expression, description = expression, timeout = 10000) {
  const start = Date.now()
  while (Date.now() - start < timeout) { if (await evaluate(`Boolean(${expression})`)) return; await delay(75) }
  throw new Error(`Timed out: ${description}\n${await evaluate('document.body.innerText')}`)
}
async function navigate(hash) {
  await evaluate(`location.hash = ${JSON.stringify(hash)}`)
  await until(`location.hash === ${JSON.stringify(`#${hash}`)} && document.querySelector('h1')`)
}
async function clickText(text, selector = 'button') {
  const source = `Array.from(document.querySelectorAll(${JSON.stringify(selector)})).find(e => e.textContent.trim() === ${JSON.stringify(text)} && e.getClientRects().length && !e.disabled)`
  await until(`Boolean(${source})`, `click ${text}`)
  await evaluate(`${source}.click()`)
}
async function fill(name, value) {
  const selector = `[name="${name}"]`
  await until(`Boolean(document.querySelector(${JSON.stringify(selector)}))`, `field ${name}`)
  await setControl(`Array.from(document.querySelectorAll(${JSON.stringify(selector)})).at(-1)`, value)
}
async function setControl(source, value) {
  await evaluate(`(() => { const e = ${source}; const prototype = e.tagName === 'TEXTAREA' ? HTMLTextAreaElement.prototype : e.tagName === 'SELECT' ? HTMLSelectElement.prototype : HTMLInputElement.prototype; Object.getOwnPropertyDescriptor(prototype, 'value').set.call(e, ${JSON.stringify(String(value))}); e.dispatchEvent(new Event(e.tagName === 'SELECT' ? 'change' : 'input', { bubbles: true })); })()`)
}
async function fillLabel(label, value) {
  const source = `Array.from(document.querySelectorAll('label')).find(e => e.firstChild?.textContent.trim() === ${JSON.stringify(label)})?.querySelector('input,select,textarea')`
  await until(`Boolean(${source})`, `field ${label}`)
  await setControl(source, value)
}
async function rowButton(rowText, buttonText) {
  const source = `Array.from(document.querySelectorAll('tbody tr')).find(e => e.textContent.includes(${JSON.stringify(rowText)}))?.querySelectorAll('button')`
  const button = `Array.from(${source} || []).find(e => e.textContent.trim() === ${JSON.stringify(buttonText)} && !e.disabled)`
  await until(`Boolean(${button})`, `${buttonText} ${rowText}`)
  await evaluate(`${button}.click()`)
}
async function stateMatches(predicate, description) {
  const start = Date.now()
  while (Date.now() - start < 10000) { if (predicate()) return; await delay(75) }
  throw new Error(`Timed out: ${description}`)
}
async function ready() { await until(`!document.querySelector('fieldset:disabled') && !document.querySelector('nav a[aria-disabled="true"]')`, 'request finished') }
async function checkHeading(pattern) { await until(`${JSON.stringify(pattern)}.split('|').some(t => document.querySelector('h1')?.textContent.includes(t))`, `heading ${pattern}`) }
async function screenshot(name) {
  if (process.env.SMOKE_SCREENSHOTS !== '1') return
  const directory = join(projectRoot, '.local', 'browser-smoke')
  await mkdir(directory, { recursive: true })
  const { data } = await cdp('Page.captureScreenshot', { format: 'png', captureBeyondViewport: false })
  await writeFile(join(directory, `${name}.png`), Buffer.from(data, 'base64'))
}

try {
  await new Promise((resolve, reject) => { api.once('error', reject); api.listen(apiPort, '127.0.0.1', resolve) })
  vite = await createViteServer({ root: projectRoot, server: { host: '127.0.0.1', port: frontendPort, strictPort: true, proxy: { '/api': { target: `http://127.0.0.1:${apiPort}`, changeOrigin: false } } }, logLevel: 'error' })
  await vite.listen()
  // Refuse an occupied debug port instead of attaching to another browser's session.
  const debugCheck = createHttpServer()
  await new Promise((resolve, reject) => { debugCheck.once('error', reject); debugCheck.listen(debugPort, '127.0.0.1', resolve) })
  await new Promise((resolve) => debugCheck.close(resolve))
  profile = await mkdtemp(join(tmpdir(), 'radar-browser-smoke-'))
  browser = spawn(browserPath, ['--headless=new', '--disable-gpu', '--no-first-run', '--no-default-browser-check', '--disable-background-networking', '--disable-features=msEdgeSidebarV2', '--remote-debugging-address=127.0.0.1', `--remote-debugging-port=${debugPort}`, `--user-data-dir=${profile}`, 'about:blank'], { windowsHide: true, stdio: 'ignore' })
  let target
  for (let attempt = 0; attempt < 80; attempt++) {
    try { target = (await (await fetch(`http://127.0.0.1:${debugPort}/json/list`)).json()).find((entry) => entry.type === 'page'); if (target) break } catch {}
    await delay(100)
  }
  assert(target, 'Browser did not expose its local debugging endpoint.')
  socket = new WebSocket(target.webSocketDebuggerUrl)
  await new Promise((resolve, reject) => { socket.addEventListener('open', resolve, { once: true }); socket.addEventListener('error', reject, { once: true }) })
  socket.addEventListener('message', ({ data }) => {
    const message = JSON.parse(data)
    if (message.id) {
      const promise = pending.get(message.id)
      pending.delete(message.id)
      if (message.error) promise?.reject(new Error(message.error.message)); else promise?.resolve(message.result)
    } else if (message.method === 'Runtime.exceptionThrown') errors.push(message.params.exceptionDetails)
    else if (message.method === 'Runtime.consoleAPICalled' && message.params.type === 'error') errors.push(message.params.args.map((arg) => arg.value || arg.description).join(' '))
  })
  await cdp('Runtime.enable')
  await cdp('Page.enable')
  await cdp('Page.addScriptToEvaluateOnNewDocument', { source: 'window.confirm = () => true' })
  await cdp('Emulation.setDeviceMetricsOverride', { width: 1440, height: 1100, deviceScaleFactor: 1, mobile: false })
  await cdp('Page.navigate', { url: `http://127.0.0.1:${frontendPort}/#account` })
  await until(`document.querySelector('.session')?.textContent.includes('system_administrator')`, 'isolated admin session')

  // References: list all pages, inspect, create, edit with nullable clearing, and delete.
  await navigate('references')
  await checkHeading('Autores, gêneros e editoras')
  await until(`document.querySelectorAll('tbody tr').length === 20`, 'first page of authors')
  await clickText('Próxima →')
  await until(`document.querySelector('tbody')?.textContent.includes('Autor 23')`, 'second author page')
  await rowButton('Autor 23', 'Detalhes')
  await until(`Array.from(document.querySelectorAll('h2')).some(e => e.textContent.includes('Detalhes: Autor 23'))`, 'author detail')
  await clickText('Novo autor')
  await fillLabel('Nome', 'Autor do smoke')
  await fillLabel('Nascimento (opcional)', '1950-01-01')
  await fillLabel('Sexo (opcional)', 'Texto de teste')
  await fillLabel('Biografia (opcional)', 'Biografia criada pelo browser isolado.')
  await clickText('Cadastrar')
  await stateMatches(() => state.authors.some((a) => a.name === 'Autor do smoke'), 'create author')
  await clickText('Editar cadastro')
  await fillLabel('Nome', 'Autor do smoke editado')
  await fillLabel('Nascimento (opcional)', '')
  await fillLabel('Sexo (opcional)', '')
  await clickText('Salvar alterações')
  await stateMatches(() => state.authors.some((a) => a.name === 'Autor do smoke editado' && a.birth_date === null && a.sex === null), 'edit author and clear nullable fields')
  await rowButton('Autor do smoke editado', 'Excluir')
  await stateMatches(() => !state.authors.some((a) => a.name === 'Autor do smoke editado'), 'delete author')

  await clickText('Gêneros')
  await until(`document.querySelector('tbody')?.textContent.includes('Ficção')`, 'genre listing')
  await fillLabel('Nome', 'Ficção')
  await clickText('Cadastrar')
  await until(`document.querySelector('.notice.error')?.textContent.includes('conflita')`, 'duplicate genre conflict')
  assert(await evaluate(`Array.from(document.querySelectorAll('label')).find(e => e.firstChild?.textContent.trim() === 'Nome').querySelector('input').value === 'Ficção'`), 'The form must retain its values after HTTP 409.')
  await screenshot('references-conflict')
  await fillLabel('Nome', 'Gênero do smoke')
  await clickText('Cadastrar')
  await stateMatches(() => state.genres.some((g) => g.name === 'Gênero do smoke'), 'create genre')
  await clickText('Editar cadastro')
  await fillLabel('Nome', 'Gênero do smoke editado')
  await clickText('Salvar alterações')
  await stateMatches(() => state.genres.some((g) => g.name === 'Gênero do smoke editado'), 'edit genre')
  await rowButton('Gênero do smoke editado', 'Desativar')
  await stateMatches(() => state.genres.some((g) => g.name === 'Gênero do smoke editado' && !g.active), 'deactivate genre')
  await rowButton('Gênero do smoke editado', 'Reativar')
  await stateMatches(() => state.genres.some((g) => g.name === 'Gênero do smoke editado' && g.active), 'reactivate genre')

  await clickText('Editoras')
  await until(`document.querySelector('tbody')?.textContent.includes('Editora de teste')`, 'publisher listing')
  await fillLabel('Nome', 'Editora do smoke')
  await fillLabel('Site (opcional)', 'https://example.test')
  await fillLabel('País de origem (opcional)', 'Brasil')
  await clickText('Cadastrar')
  await stateMatches(() => state.publishers.some((p) => p.name === 'Editora do smoke'), 'create publisher')
  await clickText('Editar cadastro')
  await fillLabel('Nome', 'Editora do smoke editada')
  await clickText('Salvar alterações')
  await stateMatches(() => state.publishers.some((p) => p.name === 'Editora do smoke editada'), 'edit publisher')
  await rowButton('Editora do smoke editada', 'Excluir')
  await stateMatches(() => !state.publishers.some((p) => p.name === 'Editora do smoke editada'), 'delete publisher')
  await ready()
  await screenshot('references-desktop')

  // Works: exercise creation, relations beyond the first 20 options, detail and edit.
  await navigate('create-work')
  await checkHeading('Adicionar obra')
  await until(`document.querySelector('input[name="author_ids"][value="23"]') && !document.querySelector('fieldset:disabled')`, 'all catalog relations loaded')
  await fill('title', 'Obra do smoke')
  await fill('synopsis', 'Sinopse criada pelo browser isolado.')
  await evaluate(`document.querySelector('input[name="author_ids"][value="23"]').click(); document.querySelector('input[name="genre_ids"][value="1"]').click()`)
  await clickText('Publicar obra')
  await stateMatches(() => state.works.some((w) => w.title === 'Obra do smoke'), 'create work')
  const smokeWork = state.works.find((w) => w.title === 'Obra do smoke')
  assert.deepEqual(smokeWork.author_ids, [23])
  await navigate('works')
  await checkHeading('Gerenciar obras')
  await rowButton('Obra do smoke', 'Detalhes')
  await until(`Array.from(document.querySelectorAll('h2')).some(e => e.textContent.includes('Obra do smoke'))`, 'work detail')
  await rowButton('Obra do smoke', 'Editar')
  await until(`Array.from(document.querySelectorAll('h2')).some(e => e.textContent.includes('Editar obra'))`, 'work editor')
  await fill('title', 'Obra do smoke editada')
  await clickText('Salvar alterações')
  await stateMatches(() => smokeWork.title === 'Obra do smoke editada', 'edit work')
  await ready()
  await screenshot('works-desktop')

  // Editions: query-string navigation preserves the selected work; POST and PATCH are distinct.
  await navigate(`editions?work_id=${smokeWork.id}`)
  await checkHeading('Gerenciar edições')
  await clickText('+ Nova edição')
  await fill('title', 'Edição do smoke')
  await fill('publisher_id', 1)
  await fill('page_count', 345)
  await fill('isbn_absence_reason', 'Edição independente do teste isolado')
  await fill('status', 'published')
  await clickText('Cadastrar edição')
  await stateMatches(() => state.editions.some((e) => e.title === 'Edição do smoke'), 'create edition')
  const smokeEdition = state.editions.find((e) => e.title === 'Edição do smoke')
  assert.equal(smokeEdition.work_id, smokeWork.id)
  await rowButton('Edição do smoke', 'Editar')
  await until(`Array.from(document.querySelectorAll('h2')).some(e => e.textContent.includes('Editar edição'))`, 'edition editor')
  await fill('title', 'Edição do smoke editada')
  await fill('page_count', 346)
  await clickText('Salvar alterações')
  await stateMatches(() => smokeEdition.title === 'Edição do smoke editada' && smokeEdition.page_count === 346, 'edit edition')
  const editionPatch = requests.findLast((r) => r.method === 'PATCH' && r.path === `/api/v1/admin/editions/${smokeEdition.id}`)
  assert(!Object.hasOwn(editionPatch.body, 'work_id'), 'Edition PATCH must preserve the immutable work relation.')
  await rowButton('Edição do smoke editada', 'Arquivar')
  await stateMatches(() => smokeEdition.status === 'archived', 'archive edition')
  await navigate('works')
  await checkHeading('Gerenciar obras')
  await rowButton('Obra do smoke editada', 'Arquivar')
  await stateMatches(() => smokeWork.status === 'archived', 'archive work')

  // Verify every new route and a narrow viewport after the complete CRUD cycle.
  for (const hash of ['works', 'editions', 'references', 'console']) {
    await navigate(hash)
    await ready()
    assert.equal(await evaluate(`document.querySelector('nav a[aria-current="page"]')?.getAttribute('href')`), `#${hash}`)
  }
  await checkHeading('Console da API')
  await fillLabel('Caminho após /api/v1', '/admin/works/{id}')
  await fillLabel('Parâmetro id', smokeWork.id)
  await clickText('Enviar requisição')
  await stateMatches(() => requests.some((r) => r.method === 'GET' && r.path === `/api/v1/admin/works/${smokeWork.id}`), 'console path variable request')
  await ready()
  await fillLabel('Método', 'PATCH')
  await fillLabel('Corpo JSON (deixe vazio para enviar sem corpo)', '{invalid json')
  const beforeInvalidJson = requests.length
  await clickText('Enviar requisição')
  await until(`document.querySelector('.inline-error')?.textContent.includes('JSON válido')`, 'invalid JSON validation')
  assert.equal(requests.length, beforeInvalidJson, 'Invalid console JSON must not be sent.')
  await fillLabel('Corpo JSON (deixe vazio para enviar sem corpo)', '{"title":"Obra editada pelo console"}')
  await clickText('Formatar JSON')
  await clickText('Enviar requisição')
  await stateMatches(() => smokeWork.title === 'Obra editada pelo console', 'console PATCH request')
  await ready()
  await evaluate(`document.querySelector('details.response').open = true`)
  await until(`document.querySelector('details.response')?.textContent.includes('Histórico da sessão')`, 'request history')
  await screenshot('console-desktop')
  await cdp('Emulation.setDeviceMetricsOverride', { width: 390, height: 844, deviceScaleFactor: 1, mobile: true })
  await screenshot('console-mobile')
  assert(await evaluate('document.documentElement.scrollWidth <= innerWidth'), 'Page overflows the mobile viewport.')

  // The public catalog uses its own detail endpoint and paginates editions independently.
  await cdp('Emulation.setDeviceMetricsOverride', { width: 1440, height: 1100, deviceScaleFactor: 1, mobile: false })
  await navigate('catalog')
  await checkHeading('Catálogo público')
  await until(`document.querySelector('.book h2')?.textContent === 'Obra publicada'`, 'published public work')
  await ready()
  await fill('title', 'Obra publicada')
  await fill('author', 'Autor 01')
  await fill('genre_id', 1)
  await fill('language', 'pt-br')
  await fill('order', 'newest')
  await fillLabel('Obras por página', 10)
  await clickText('Pesquisar')
  await ready()
  const catalogQuery = requests.findLast((entry) => entry.path === '/api/v1/works').query
  assert.deepEqual(catalogQuery, { page: '1', page_size: '10', title: 'Obra publicada', author: 'Autor 01', genre_id: '1', language: 'pt-br', order: 'newest' })
  await clickText('Ver detalhes e edições →')
  await until(`document.querySelectorAll('.edition').length === 10`, 'first page of public editions')
  assert(requests.some((entry) => entry.path === '/api/v1/works/1'), 'Public details must be loaded from the detail endpoint.')
  await clickText('Próxima →')
  await until(`document.querySelectorAll('.edition').length === 2 && document.querySelector('.edition strong')?.textContent === 'Edição de teste 11'`, 'second page of public editions')
  assert(requests.some((entry) => entry.path === '/api/v1/works/1/editions' && entry.query.page === '2' && entry.query.page_size === '10'))
  await screenshot('catalog-desktop')
  await clickText('Usar na estante')
  await checkHeading('Minha estante')
  await until(`document.querySelector('select[name="work_id"]')?.value === '1' && document.querySelector('select[name="edition_id"]')?.value === '11' && !document.querySelector('select[name="edition_id"]')?.disabled`, 'catalog edition selected on shelf')
  assert.equal(await evaluate(`document.querySelector('.inline-error')?.textContent || ''`), '')

  // Losing the admin session must stop administrative data loading and preserve public browsing.
  await navigate('account')
  await checkHeading('Conta e sessão')
  await clickText('Sair')
  await until(`document.querySelector('.session')?.textContent === 'Visitante'`, 'logged out session')
  await ready()
  const afterLogout = requests.length
  for (const [hash, heading] of [['works', 'Gerenciar obras'], ['editions', 'Gerenciar edições'], ['create-work', 'Adicionar obra']]) {
    await navigate(hash)
    await checkHeading(heading)
    await until(`Array.from(document.querySelectorAll('h2')).some((element) => /Permissão de catálogo necessária|Entre para cadastrar obras/.test(element.textContent))`, `guard ${hash}`)
  }
  await navigate('references')
  await checkHeading('Autores, gêneros e editoras')
  await until(`document.querySelector('tbody')?.textContent.includes('Autor 01') && document.body.textContent.includes('Consulta pública')`, 'public author list')
  assert.equal(await evaluate(`Array.from(document.querySelectorAll('button')).some((button) => ['Editar', 'Excluir', 'Cadastrar'].includes(button.textContent.trim()))`), false, 'Visitor references must hide mutation controls.')
  await clickText('Gêneros')
  await until(`document.querySelector('tbody')?.textContent.includes('Ficção') && !document.querySelector('tbody')?.textContent.includes('Inativo')`, 'public active genres only')
  await rowButton('Ficção', 'Detalhes')
  await until(`Array.from(document.querySelectorAll('h2')).some((element) => element.textContent === 'Detalhes: Ficção')`, 'public genre detail from list')
  assert(!requests.slice(afterLogout).some((entry) => entry.path.startsWith('/api/v1/admin/')), 'Administrative requests must stop after logout.')
  assert(requests.slice(afterLogout).some((entry) => entry.path === '/api/v1/authors'))
  assert(requests.slice(afterLogout).some((entry) => entry.path === '/api/v1/genres'))
  assert(!requests.slice(afterLogout).some((entry) => /^\/api\/v1\/genres\/\d+$/.test(entry.path)), 'The API has no public genre detail endpoint.')
  assert.deepEqual(errors, [], `Browser errors: ${JSON.stringify(errors)}`)
  console.log(`Browser smoke passed: ${requests.length} isolated API requests; no browser console errors.`)
} finally {
  try { if (socket?.readyState === WebSocket.OPEN) await cdp('Browser.close') } catch {}
  socket?.close()
  if (browser && browser.exitCode === null) { browser.kill(); await new Promise((resolve) => { browser.once('exit', resolve); setTimeout(resolve, 2000).unref() }) }
  await vite?.close()
  await new Promise((resolve) => api.close(resolve))
  if (profile) {
    const safeRoot = resolve(tmpdir())
    const safeProfile = resolve(profile)
    assert(safeProfile.startsWith(`${safeRoot}\\`) || safeProfile.startsWith(`${safeRoot}/`), 'Temporary browser profile must stay inside the OS temp directory.')
    try { await rm(safeProfile, { recursive: true, force: true, maxRetries: 5, retryDelay: 200 }) } catch (error) { console.warn(`Temporary browser profile cleanup: ${error.message}`) }
  }
}
