import { discoveryBooks } from './discoveryBooks.js'

// Namespace separado: nenhum dado do back-end ou de outras aplicações é utilizado.
const storageKey = 'radar-literario:prototype:interested-books:v1'
const catalog = new Map(discoveryBooks.map((book) => [book.id, book]))

export function loadInterestedBooks() {
  try {
    const records = JSON.parse(globalThis.localStorage?.getItem(storageKey) ?? '[]')
    if (!Array.isArray(records)) return []
    const restored = new Map()
    for (const record of records) {
      const source = record && catalog.get(record.id)
      if (!source || restored.has(source.id)) continue
      const pages = Number.isInteger(record.pages) && record.pages >= 0 && record.pages <= source.totalPages ? record.pages : 0
      restored.set(source.id, {
        ...source,
        pages,
        status: pages === 0 ? 'interested' : pages >= source.totalPages ? 'completed' : 'reading',
        rating: Number.isFinite(record.rating) && record.rating >= 0 && record.rating <= 5 ? record.rating : source.rating,
        lastOpinion: typeof record.lastOpinion === 'string' ? record.lastOpinion : '',
      })
    }
    return [...restored.values()]
  } catch {
    // O protótipo continua em memória se o armazenamento não estiver disponível.
    return []
  }
}

export function persistInterestedBooks(books) {
  try {
    const records = books.filter((book) => catalog.has(book.id)).map(({ id, pages, rating, lastOpinion }) => ({ id, pages, rating, lastOpinion }))
    globalThis.localStorage?.setItem(storageKey, JSON.stringify(records))
  } catch {
    // Navegadores com armazenamento bloqueado ainda permitem usar a sessão atual.
  }
}
