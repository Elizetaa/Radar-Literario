// Adapter entre a UI (formato do protótipo) e a API Go. Sem mockData / localStorage.
// Rotas e formatos conferidos no Backend (controller/*.go, schemas/*.go).
import { backendApi } from './backendApi.js'

const list = (res) => res?.data ?? []
const one = (res) => res?.data ?? res

// Estados reais do backend: planned | in_progress | paused | completed | abandoned
const STATUS_TO_UI = { planned: 'interested', in_progress: 'reading', paused: 'reading', completed: 'completed', abandoned: 'interested' }
const statusToApi = (pages, total) => (pages <= 0 ? 'planned' : pages >= total ? 'completed' : 'in_progress')

// Caches: shelf/posts só trazem work_id/edition_id; título, autor e páginas vêm de /works e /editions.
const workCache = new Map()
const editionCache = new Map()
const getWork = (id) => workCache.get(id) ?? workCache.set(id, backendApi.work(id).then(one)).get(id)
const getEdition = (id) => editionCache.get(id) ?? editionCache.set(id, backendApi.edition(id).then(one)).get(id)
const firstEdition = async (workId) => {
  const eds = list(await backendApi.editions(workId))
  if (!eds.length) throw new Error('Esta obra não tem edição publicada disponível.')
  return eds[0]
}

function workToBook(w, extra = {}) {
  return {
    id: w.id, title: w.title,
    author: (w.authors || []).map((a) => a.name).join(', '),
    genre: w.genres?.[0]?.name || 'Ficção',
    description: w.synopsis || '',
    rating: w.average_rating ?? 0,
    pages: 0, totalPages: 1, cover: '', status: undefined,
    ...extra,
  }
}

async function shelfToBook(item, myRatings) {
  const [w, ed] = await Promise.all([getWork(item.work_id), getEdition(item.edition_id)])
  return workToBook(w, {
    editionId: item.edition_id,
    pages: item.current_page,
    totalPages: ed.page_count || 1,
    cover: ed.image_url || '',
    status: STATUS_TO_UI[item.status] ?? 'interested',
    rating: myRatings.get(item.work_id) ?? 0, // nota do próprio usuário (review), não a média
  })
}

async function postToUi(p) {
  const w = await getWork(p.work_id)
  const ed = p.edition_id ? await getEdition(p.edition_id) : null
  const name = p.author_name || p.author_nickname || 'Leitor'
  return {
    id: p.id, bookId: p.work_id, title: w.title,
    author: (w.authors || []).map((a) => a.name).join(', '),
    text: p.content_hidden ? 'Conteúdo oculto (spoiler ou conteúdo sensível).' : (p.text || ''),
    pages: p.current_page ?? undefined, totalPages: ed?.page_count,
    rating: p.rating ?? 0, person: name, handle: `@${p.author_nickname}`, initials: name.slice(0, 2).toUpperCase(),
    time: new Date(p.published_at || p.created_at).toLocaleDateString('pt-BR'),
    likes: p.like_count, liked: p.liked, comments: p.comment_count, mine: false,
  }
}

export const libraryService = {
  // Estante do usuário (vazia para conta nova — a UI precisa tratar [] ).
  async getBooks() {
    const [shelf, reviews] = await Promise.all([backendApi.shelf(), backendApi.myReviews()])
    const ratings = new Map(list(reviews).map((r) => [r.work_id, r.rating]))
    return Promise.all(list(shelf).map((i) => shelfToBook(i, ratings)))
  },

  // "Indicações para você": /me/recommendations; se vier vazio (sem interesses), cai no ranking popular.
  async getRecommendations() {
    const [rec, shelf] = await Promise.all([backendApi.recommendations(), backendApi.shelf()])
    const onShelf = new Set(list(shelf).map((i) => i.work_id))
    let works = list(rec)
    if (!works.length) works = list(await backendApi.popular())
    return works.filter((w) => !onShelf.has(w.id)).map((w) => workToBook(w))
  },

  // "Tenho interesse" = obra na estante com status planned.
  async saveInterest(workId) {
    const ed = await firstEdition(workId)
    await backendApi.updateShelf(workId, { edition_id: ed.id, current_page: 0, status: 'planned' })
  },

  // Cadastro manual deixou de existir: UI deve escolher uma obra do catálogo (workId).
  async addBook({ workId }) {
    if (!workId) throw new Error('Escolha uma obra do catálogo.')
    return this.saveInterest(workId)
  },
  // Busca no catálogo (banco): por título OU autor, com filtro opcional de gênero.
  async searchCatalog(term = '', genreId = 0) {
    const [byTitle, byAuthor] = await Promise.all([
      backendApi.catalog({ title: term, genreId }),
      term ? backendApi.catalog({ author: term, genreId }) : { data: [] },
    ])
    const seen = new Map()
    for (const w of [...list(byTitle), ...list(byAuthor)]) if (!seen.has(w.id)) seen.set(w.id, w)
    return [...seen.values()].map((w) => workToBook(w))
  },
  async getGenres() { return list(await backendApi.genres()).filter((g) => g.active !== false) },

  // changes: { pages, rating, lastOpinion } — o progresso vai pra estante, a nota vira review.
  async updateBook(workId, changes) {
    if (changes.pages !== undefined) {
      const shelf = list(await backendApi.shelf()).find((i) => i.work_id === workId)
      const ed = shelf ? await getEdition(shelf.edition_id) : await firstEdition(workId)
      const total = ed.page_count || 1
      const pages = Math.min(changes.pages, total)
      await backendApi.updateShelf(workId, {
        edition_id: ed.id, current_page: pages, status: statusToApi(pages, total),
        change_kind: shelf && pages < shelf.current_page ? 'correction' : 'advance',
      })
    }
  },
  async deleteBook(workId) { await backendApi.deleteShelf(workId) },

  // Feed da comunidade (diários + avaliações publicados).
  async getPosts() { return Promise.all(list(await backendApi.feed()).map(postToUi)) },

  // Diário -> POST /reading-entries ; avaliação (rating>0) -> POST /reviews.
  // Obs.: o backend exige obra em in_progress/completed na estante para avaliar.
  async addPost(post) {
    if (post.rating > 0) {
      return backendApi.createReview({ work_id: post.bookId, rating: post.rating, title: '', text: post.text, contains_spoiler: false, status: 'published' })
    }
    const shelf = list(await backendApi.shelf()).find((i) => i.work_id === post.bookId)
    const ed = shelf ? { id: shelf.edition_id } : await firstEdition(post.bookId)
    return backendApi.createEntry({
      work_id: post.bookId, edition_id: ed.id, current_page: post.pages ?? 0,
      title: '', text: post.text, contains_spoiler: false, status: 'published',
    })
  },

  // Comentários agora são POR PUBLICAÇÃO: chame getComments(post.id) ao abrir o modal.
  async getComments(postId) {
    if (!postId) return []
    return list(await backendApi.comments(postId)).map((c) => ({
      id: c.id, person: c.author_name || c.author_nickname, handle: `@${c.author_nickname}`,
      text: c.content_hidden ? '(comentário oculto)' : c.text, time: new Date(c.created_at).toLocaleDateString('pt-BR'),
    }))
  },
  async addComment(postId, text) { return one(await backendApi.createComment(postId, text)) },

  async toggleLike(postId, currentlyLiked) { await backendApi.likePost(postId, !currentlyLiked) },
  async reportPost(postId) {
    await backendApi.report({ reported_post_id: postId, reason: 'inappropriate_content', description: 'Denúncia feita pelo feed.' })
  },
}
