// O histórico fica apenas em memória; credenciais nunca entram no painel.
const secret = /^(?:.*password.*|.*token.*|authorization|cookie|set-cookie|secret)$/i
export function redact(value) {
  if (Array.isArray(value)) return value.map(redact)
  if (value && typeof value === 'object') return Object.fromEntries(Object.entries(value).map(([key, item]) => [key, secret.test(key) ? '[oculto]' : redact(item)]))
  return value
}

export function redactPath(path) {
  const index = path.indexOf('?')
  if (index < 0) return path
  const query = new URLSearchParams(path.slice(index + 1))
  for (const key of [...query.keys()]) if (secret.test(key)) query.set(key, '[oculto]')
  return `${path.slice(0, index)}?${query}`
}
