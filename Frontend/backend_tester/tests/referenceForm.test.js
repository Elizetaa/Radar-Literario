import test from 'node:test'
import assert from 'node:assert/strict'
import { buildReferencePayload, filterReferences, referenceValues } from '../src/referenceForm.js'

test('editar autor permite limpar datas e sexo explicitamente com null', () => {
  const original = { id: 2, name: 'Autora', biography: 'Biografia', birth_date: '1970-01-01', death_date: '2020-01-01', sex: 'Feminino', country_of_origin: 'Brasil' }
  const values = { ...referenceValues('authors', original), birth_date: '', death_date: '', sex: '' }
  assert.deepEqual(buildReferencePayload('authors', values, original), { birth_date: null, death_date: null, sex: null })
})

test('PATCH de editora preserva campos que não foram alterados e permite limpar site', () => {
  const original = { id: 3, name: 'Editora', country_of_origin: 'Brasil', website_url: 'https://example.com' }
  assert.deepEqual(buildReferencePayload('publishers', { ...referenceValues('publishers', original), website_url: '' }, original), { website_url: '' })
  assert.deepEqual(buildReferencePayload('publishers', referenceValues('publishers', original), original), {})
})

test('criação e edição de gênero mantêm false como boolean e não enviam ID', () => {
  assert.deepEqual(buildReferencePayload('genres', { id: 2, name: '  Ficção  ', active: false }), { name: 'Ficção', active: false })
  assert.deepEqual(buildReferencePayload('genres', { name: 'Ficção', active: true }, { name: 'Ficção', active: false }), { active: true })
})

test('valida nome obrigatório, tamanho específico de gênero e ordem das datas', () => {
  assert.throws(() => buildReferencePayload('genres', { name: '  ' }), /Informe o nome/)
  assert.throws(() => buildReferencePayload('genres', { name: 'a'.repeat(101) }), /tamanho permitido/)
  assert.throws(() => buildReferencePayload('authors', { name: 'Autor', birth_date: '2020-01-01', death_date: '2000-01-01' }), /anterior ao nascimento/)
})

test('filtro de página combina nome ou ID e estado do gênero', () => {
  const items = [{ id: 1, name: 'Ficção', active: true }, { id: 2, name: 'Ficção científica', active: false }]
  assert.deepEqual(filterReferences(items, 'FICÇÃO', 'false'), [items[1]])
  assert.deepEqual(filterReferences(items, '1'), [items[0]])
  assert.deepEqual(filterReferences(items, 'inexistente'), [])
})
