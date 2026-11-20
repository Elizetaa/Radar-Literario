export const catalogStatuses = {
  draft: 'Rascunho', pending_review: 'Em análise', published: 'Publicado',
  rejected: 'Rejeitado', hidden: 'Oculto', archived: 'Arquivado',
}
export const minimumAges = [0, 10, 12, 14, 16, 18]

export function listPath(resource, filters = {}, page = 1, pageSize = 20) {
  const query = new URLSearchParams({ page, page_size: pageSize })
  for (const [key, value] of Object.entries(filters)) {
    if (value !== '' && value !== undefined && value !== null) query.set(key, String(value).trim())
  }
  return `${resource}?${query}`
}

export async function catalogOptions(perform, resources, isActive = () => true) {
  const options = {}
  for (const resource of resources) {
    const items = []
    for (let page = 1; isActive(); page++) {
      const result = await perform(listPath(`/admin/${resource}`, {}, page, 100))
      if (!isActive()) return null
      if (!result) throw new Error('Não foi possível carregar as opções. Tente novamente.')
      items.push(...result.data)
      if (!result.data.length || items.length >= result.pagination.total) break
    }
    if (!isActive()) return null
    options[resource] = items
  }
  return options
}

export function buildEditionPayload(fields, editing = false) {
  const text = (name) => String(fields.get(name) || '').trim()
  const positive = (name, label, max = Number.MAX_SAFE_INTEGER) => {
    const value = Number(text(name))
    if (!Number.isSafeInteger(value) || value < 1 || value > max) throw new Error(`Informe ${label} válido.`)
    return value
  }
  const title = text('title')
  const language = text('language').toLowerCase()
  if (!title) throw new Error('Informe o título da edição.')
  if (!/^[a-z]{2,3}(-[a-z0-9]{2,8})*$/.test(language)) throw new Error('Informe um idioma válido, como pt-br ou en.')
  const isbn = text('isbn')
  const isbn_absence_reason = text('isbn_absence_reason')
  if (!isbn && !isbn_absence_reason) throw new Error('Informe um ISBN válido ou o motivo de sua ausência.')
  return {
    ...(!editing ? { work_id: positive('work_id', 'uma obra') } : {}),
    publisher_id: positive('publisher_id', 'uma editora'), title, language,
    page_count: positive('page_count', 'um número de páginas', 1000000),
    edition_number: positive('edition_number', 'um número de edição', 100000),
    isbn: isbn || null, isbn_absence_reason: isbn ? '' : isbn_absence_reason,
    publication_date: text('publication_date') || null,
    image_url: text('image_url'), status: text('status'), minimum_age: Number(text('minimum_age')),
  }
}
