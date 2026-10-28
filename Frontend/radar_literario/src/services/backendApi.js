const API_BASE = (import.meta.env.VITE_API_BASE_URL || 'http://127.0.0.1:8080/api/v1').replace(/\/$/, '')

let csrfToken = ''

async function request(path, { method = 'GET', body, csrf = false } = {}) {
  const headers = { Accept: 'application/json' }
  if (body !== undefined) headers['Content-Type'] = 'application/json'
  if (csrf) {
    if (!csrfToken) await getCsrfToken()
    headers['X-CSRF-Token'] = csrfToken
  }
  const response = await fetch(`${API_BASE}${path}`, {
    method,
    headers,
    credentials: 'include',
    ...(body === undefined ? {} : { body: JSON.stringify(body) }),
  })
  const payload = response.status === 204 ? null : await response.json().catch(() => null)
  if (!response.ok) {
    const message = payload?.error?.message || payload?.message || `Falha HTTP ${response.status}`
    throw new Error(message)
  }
  return payload
}

export async function getCsrfToken() {
  const result = await request('/auth/csrf')
  csrfToken = result?.data?.csrf_token || ''
  return csrfToken
}

export const backendApi = {
  async register(reader) {
    await getCsrfToken()
    return request('/readers', { method: 'POST', body: reader, csrf: true })
  },
  async login(credentials) {
    await getCsrfToken()
    const result = await request('/auth/login', { method: 'POST', body: credentials, csrf: true })
    csrfToken = result?.data?.csrf_token || ''
    return result?.data
  },
  async logout() {
    await getCsrfToken()
    await request('/auth/logout', { method: 'POST', csrf: true })
    csrfToken = ''
  },
  me: () => request('/me'),
  works: (title = '') => request(`/works?page=1&page_size=20${title ? `&title=${encodeURIComponent(title)}` : ''}`),
  editions: (workId) => request(`/works/${workId}/editions?page=1&page_size=50`),
  shelf: () => request('/me/shelf?page=1&page_size=50'),
  updateShelf: (workId, input) => request(`/me/shelf/${workId}`, { method: 'PUT', body: input, csrf: true }),
  entries: () => request('/me/reading-entries?page=1&page_size=20'),
  createEntry: (input) => request('/reading-entries', { method: 'POST', body: input, csrf: true }),
}
