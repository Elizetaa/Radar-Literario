import { request } from './api.js'

// Os seletores precisam incluir também os registros além da primeira página.
export async function loadCatalogOptions(path, fetchPage = request) {
  const items = []
  for (let page = 1; ; page++) {
    const { payload } = await fetchPage(`${path}?page=${page}&page_size=100`)
    items.push(...payload.data)
    if (!payload.data.length || items.length >= payload.pagination.total) return items
  }
}

export function editionLabel(edition) {
  return [
    edition.title,
    edition.edition_number ? `${edition.edition_number}ª edição` : null,
    edition.language,
    edition.page_count ? `${edition.page_count} páginas` : null,
    edition.isbn ? `ISBN ${edition.isbn}` : null,
  ].filter(Boolean).join(' · ')
}
