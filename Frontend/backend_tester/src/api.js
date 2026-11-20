export class ApiError extends Error {
  constructor(message, status, payload) {
    super(message)
    this.status = status
    this.payload = payload
  }
}

async function send(path, options = {}) {
  if (!path.startsWith('/') || path.startsWith('//') || path.includes('://')) {
    throw new Error('Informe um caminho relativo à API, como /works.')
  }
  const url = `/api/v1${path}`
  let response
  try {
    response = await fetch(url, { credentials: 'include', ...options })
  } catch {
    throw new Error('Não foi possível conectar à API. Confirme se o backend e o proxy estão disponíveis.')
  }
  const text = await response.text()
  let payload = null
  if (text) {
    try { payload = JSON.parse(text) } catch {
      throw new ApiError('A API retornou uma resposta que não é JSON. Confira o endereço do backend.', response.status, { response: text.slice(0, 1000) })
    }
  }
  if (!response.ok) {
    throw new ApiError(payload?.error?.message || `Erro HTTP ${response.status}`, response.status, payload)
  }
  return { status: response.status, payload }
}

export async function request(path, method = 'GET', body) {
  const headers = { Accept: 'application/json' }
  if (method !== 'GET') {
    const csrf = await send('/auth/csrf')
    headers['X-CSRF-Token'] = csrf.payload.data.csrf_token
  }
  if (body !== undefined) headers['Content-Type'] = 'application/json'
  return send(path, { method, headers, ...(body === undefined ? {} : { body: JSON.stringify(body) }) })
}
