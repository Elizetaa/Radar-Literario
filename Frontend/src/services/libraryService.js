import { initialBooks, initialComments, initialPosts } from '../data/mockData.js'
import { discoveryBooks } from '../features/book-discovery/discoveryBooks.js'
import { loadInterestedBooks, persistInterestedBooks } from '../features/book-discovery/interestStorage.js'

// Prototype adapter: original records stay in memory; discovery selections use local storage.
let books = [...loadInterestedBooks(), ...structuredClone(initialBooks)]
let posts = structuredClone(initialPosts)
let comments = structuredClone(initialComments)
let nextId = 20
const wait = () => new Promise((resolve) => setTimeout(resolve, 80))

export const libraryService = {
  async getBooks() { await wait(); return structuredClone(books) },
  async getRecommendations() { await wait(); return structuredClone(discoveryBooks.filter((candidate) => !books.some((book) => book.id === candidate.id))) },
  async saveInterest(id) {
    await wait()
    const existing = books.find((book) => book.id === id)
    if (existing) return structuredClone(existing)
    const candidate = discoveryBooks.find((book) => book.id === id)
    if (!candidate) throw new Error('Não foi possível encontrar esta indicação. Tente novamente.')
    const item = { ...candidate, pages: 0, status: 'interested' }
    books = [item, ...books]
    persistInterestedBooks(books)
    return structuredClone(item)
  },
  async getPosts() { await wait(); return structuredClone(posts) },
  async getComments() { await wait(); return structuredClone(comments) },
  async addBook(book) { await wait(); const item = { ...book, id: nextId++, pages: 0, totalPages: Number(book.totalPages) || 1, rating: 0, cover: book.cover || '' }; books = [item, ...books]; return item },
  async updateBook(id, changes) { await wait(); books = books.map((book) => { if (book.id !== id) return book; const updated = { ...book, ...changes }; if (book.status && changes.pages !== undefined) updated.status = updated.pages === 0 ? 'interested' : updated.pages >= updated.totalPages ? 'completed' : 'reading'; return updated }); persistInterestedBooks(books); return structuredClone(books.find((book) => book.id === id)) },
  async deleteBook(id) { await wait(); books = books.filter((book) => book.id !== id); persistInterestedBooks(books) },
  async addPost(post) { await wait(); const item = { ...post, id: nextId++, person: 'Pudim Nascimento', initials: 'PN', time: 'agora', likes: 0 }; posts = [item, ...posts]; return item },
  async toggleLike(id) { await wait(); posts = posts.map((post) => post.id === id ? { ...post, likes: post.likes + 1, liked: !post.liked } : post); return structuredClone(posts) },
  async reportPost(id) { await wait(); posts = posts.filter((post) => post.id !== id) },
  async addComment(text) { await wait(); const item = { id: nextId++, person: 'Pudim', handle: '@pudim', text, time: 'agora' }; comments = [...comments, item]; return item },
}
