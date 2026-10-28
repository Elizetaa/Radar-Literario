export const catalogRoles = ['catalog_manager', 'administrator', 'system_administrator']

export function buildWorkPayload(fields) {
  const text = (name) => String(fields.get(name) || '').trim()
  const positive = (name) => {
    const value = Number(text(name))
    if (!Number.isSafeInteger(value) || value < 1) throw new Error('Informe uma editora e uma quantidade válida de páginas.')
    return value
  }
  const body = {
    title: text('title'), original_language: text('original_language').toLowerCase(),
    synopsis: text('synopsis'), status: text('status'), minimum_age: Number(text('minimum_age')),
    author_ids: fields.getAll('author_ids').map(Number), genre_ids: fields.getAll('genre_ids').map(Number),
  }
  if (!body.title) throw new Error('Informe o título da obra.')
  if (!/^[a-z]{2,3}(-[a-z0-9]{2,8})*$/.test(body.original_language)) throw new Error('Informe um idioma válido, como pt-br ou en.')
  if (body.status === 'published' && (!body.synopsis || !body.author_ids.length || !body.genre_ids.length)) {
    throw new Error('Para publicar, informe a sinopse e selecione pelo menos um autor e um gênero ativo.')
  }
  if (text('first_publication_date')) body.first_publication_date = text('first_publication_date')
  if (fields.has('with_edition')) {
    const isbn = text('isbn')
    const reason = text('isbn_absence_reason')
    if (!isbn && !reason) throw new Error('Informe o ISBN da edição ou o motivo de sua ausência.')
    body.initial_edition = {
      title: text('edition_title') || body.title,
      language: text('edition_language').toLowerCase() || body.original_language,
      publisher_id: positive('publisher_id'), page_count: positive('page_count'),
      status: body.status, minimum_age: body.minimum_age,
      ...(isbn ? { isbn } : { isbn: null, isbn_absence_reason: reason }),
    }
    if (text('edition_publication_date')) body.initial_edition.publication_date = text('edition_publication_date')
    if (text('image_url')) body.initial_edition.image_url = text('image_url')
  }
  return body
}
