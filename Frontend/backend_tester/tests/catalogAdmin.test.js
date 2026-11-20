import test from 'node:test'
import assert from 'node:assert/strict'
import { buildEditionPayload, catalogOptions, listPath } from '../src/catalogAdmin.js'
import { redact, redactPath } from '../src/requestLog.js'

function fields(overrides = {}) {
  const values = { work_id: '12', publisher_id: '7', title: '  Edição  ', language: 'PT-BR', page_count: '230', edition_number: '2', isbn: '', isbn_absence_reason: 'Artesanal', publication_date: '', image_url: '', status: 'published', minimum_age: '0', ...overrides }
  const data = new FormData()
  for (const [key, value] of Object.entries(values)) data.set(key, value)
  return data
}

test('edição preserva classificação zero e envia ISBN/data null para limpeza', () => {
  assert.deepEqual(buildEditionPayload(fields()), {
    work_id: 12, publisher_id: 7, title: 'Edição', language: 'pt-br', page_count: 230, edition_number: 2,
    isbn: null, isbn_absence_reason: 'Artesanal', publication_date: null, image_url: '', status: 'published', minimum_age: 0,
  })
})

test('editar edição omite obra imutável e permite substituir justificativa por ISBN', () => {
  const result = buildEditionPayload(fields({ work_id: '', isbn: '9780306406157', publication_date: '2020-01-01' }), true)
  assert.ok(!('work_id' in result))
  assert.equal(result.isbn_absence_reason, '')
  assert.equal(result.isbn, '9780306406157')
  assert.equal(result.publication_date, '2020-01-01')
})

test('edição valida vínculos e limites numéricos antes de enviar', () => {
  for (const changes of [{ work_id: '' }, { publisher_id: '' }, { page_count: '0' }, { page_count: '1000001' }, { edition_number: '1.5' }, { edition_number: '100001' }, { isbn_absence_reason: '' }, { language: 'português' }]) {
    assert.throws(() => buildEditionPayload(fields(changes)))
  }
})

test('filtros escapam acentos e caracteres de consulta sem perder paginação', () => {
  const path = listPath('/admin/works', { title: ' A & B ', status: '', minimum_age: 0, author: null }, 3, 50)
  const url = new URL(path, 'http://test.local')
  assert.equal(url.searchParams.get('title'), 'A & B')
  assert.equal(url.searchParams.get('minimum_age'), '0')
  assert.equal(url.searchParams.get('page'), '3')
  assert.equal(url.searchParams.get('page_size'), '50')
  assert.equal(url.searchParams.has('status'), false)
})

test('opções administrativas incluem todas as páginas e falham sem expor lista parcial', async () => {
  const paths = []
  const result = await catalogOptions(async (path) => {
    paths.push(path)
    return paths.length === 1 ? { data: [{ id: 1 }], pagination: { total: 2 } } : { data: [{ id: 2 }], pagination: { total: 2 } }
  }, ['authors'])
  assert.deepEqual(result.authors, [{ id: 1 }, { id: 2 }])
  assert.equal(paths.length, 2)
  await assert.rejects(catalogOptions(async () => undefined, ['genres']), /Não foi possível/)
})

test('histórico oculta senhas e tokens aninhados e em query sem alterar dados originais', () => {
  const input = { password: 'senha', data: { csrf_token: 'secret', user: { name: 'Leitor' } }, values: [{ reset_token: 'private' }] }
  const safe = redact(input)
  assert.equal(safe.password, '[oculto]')
  assert.equal(safe.data.csrf_token, '[oculto]')
  assert.equal(safe.values[0].reset_token, '[oculto]')
  assert.equal(input.password, 'senha')
  assert.equal(safe.data.user.name, 'Leitor')
  const path = redactPath('/auth/reset?token=private&page=2')
  assert.equal(new URL(path, 'http://test.local').searchParams.get('token'), '[oculto]')
  assert.ok(!path.includes('private'))
})
