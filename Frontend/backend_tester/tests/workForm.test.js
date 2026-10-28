import test from 'node:test'
import assert from 'node:assert/strict'
import { buildWorkPayload } from '../src/workForm.js'

function form(overrides = {}) {
  const values = {
    title: '  Obra de teste  ', original_language: 'PT-BR', synopsis: '  Sinopse  ',
    status: 'published', minimum_age: '0', author_ids: ['2', '3'], genre_ids: ['5'],
    first_publication_date: '', ...overrides,
  }
  const result = new FormData()
  for (const [key, value] of Object.entries(values)) {
    for (const item of Array.isArray(value) ? value : [value]) result.append(key, item)
  }
  return result
}

test('publicação envia relações numéricas, classificação explícita e omite datas vazias', () => {
  const body = buildWorkPayload(form())
  assert.deepEqual(body, {
    title: 'Obra de teste', original_language: 'pt-br', synopsis: 'Sinopse',
    status: 'published', minimum_age: 0, author_ids: [2, 3], genre_ids: [5],
  })
})

test('publicação exige sinopse, autor e gênero; rascunho permite dados incompletos', () => {
  for (const override of [{ synopsis: ' ' }, { author_ids: [] }, { genre_ids: [] }]) {
    assert.throws(() => buildWorkPayload(form(override)), /Para publicar/)
  }
  assert.equal(buildWorkPayload(form({ status: 'draft', synopsis: '', author_ids: [], genre_ids: [] })).status, 'draft')
})

test('edição inicial envia ISBN nulo e justificativa, sem work_id ou datas vazias', () => {
  const body = buildWorkPayload(form({
    with_edition: 'on', publisher_id: '7', page_count: '120', isbn: '', isbn_absence_reason: '  Edição artesanal  ',
  }))
  assert.deepEqual(body.initial_edition, {
    title: body.title, language: 'pt-br', publisher_id: 7, page_count: 120,
    status: 'published', minimum_age: 0, isbn: null, isbn_absence_reason: 'Edição artesanal',
  })
})

test('edição com ISBN usa idioma e título próprios e preserva datas preenchidas', () => {
  const body = buildWorkPayload(form({
    first_publication_date: '2000-01-01', with_edition: 'on', publisher_id: '7', page_count: '100',
    isbn: '9780306406157', isbn_absence_reason: 'Ignorar', edition_title: 'Tradução', edition_language: 'EN',
    edition_publication_date: '2001-01-01', image_url: 'https://example.test/capa.jpg',
  }))
  assert.equal(body.first_publication_date, '2000-01-01')
  assert.equal(body.initial_edition.language, 'en')
  assert.equal(body.initial_edition.title, 'Tradução')
  assert.equal(body.initial_edition.isbn, '9780306406157')
  assert.equal(body.initial_edition.publication_date, '2001-01-01')
  assert.ok(!('isbn_absence_reason' in body.initial_edition))
})

test('bloqueia edição sem ISBN ou justificativa e sem editora ou páginas', () => {
  const base = { with_edition: 'on', publisher_id: '7', page_count: '100', isbn: '9780306406157' }
  assert.throws(() => buildWorkPayload(form({ ...base, isbn: '' })), /ISBN/)
  assert.throws(() => buildWorkPayload(form({ ...base, publisher_id: '' })), /editora/)
  assert.throws(() => buildWorkPayload(form({ ...base, page_count: '0' })), /páginas/)
})

test('bloqueia título vazio e idioma inválido', () => {
  assert.throws(() => buildWorkPayload(form({ title: ' ' })), /título/)
  assert.throws(() => buildWorkPayload(form({ original_language: 'português' })), /idioma/)
})
