import test from 'node:test'
import assert from 'node:assert/strict'
import { editionLabel, loadCatalogOptions } from '../src/catalogOptions.js'

test('carrega obras além da primeira página para não limitar o seletor a 100 opções', async () => {
  const paths = []
  const firstPage = Array.from({ length: 100 }, (_, index) => ({ id: index + 1 }))
  const secondPage = [{ id: 101 }, { id: 102 }]
  const result = await loadCatalogOptions('/works', async (path) => {
    paths.push(path)
    return { payload: { data: paths.length === 1 ? firstPage : secondPage, pagination: { total: 102 } } }
  })
  assert.deepEqual(paths, ['/works?page=1&page_size=100', '/works?page=2&page_size=100'])
  assert.equal(result.length, 102)
  assert.equal(result.at(-1).id, 102)
})

test('lista de edições termina na última página e aceita obra sem edições', async () => {
  let calls = 0
  const result = await loadCatalogOptions('/works/7/editions', async () => {
    calls++
    return { payload: { data: [{ id: 9 }], pagination: { total: 1 } } }
  })
  assert.equal(calls, 1)
  assert.deepEqual(result, [{ id: 9 }])
  assert.deepEqual(await loadCatalogOptions('/works/8/editions', async () => ({ payload: { data: [], pagination: { total: 0 } } })), [])
})

test('falha em uma página não apresenta opções incompletas como se o carregamento tivesse sucesso', async () => {
  let calls = 0
  await assert.rejects(loadCatalogOptions('/works', async () => {
    if (++calls === 2) throw new Error('API indisponível')
    return { payload: { data: [{ id: 1 }], pagination: { total: 2 } } }
  }), /API indisponível/)
})

test('edições com mesmo título são diferenciadas por idioma, número, páginas e ISBN', () => {
  const edition = { id: 999, title: 'Uma história', edition_number: 2, language: 'pt-br', page_count: 240, isbn: '9780306406157' }
  assert.equal(editionLabel(edition), 'Uma história · 2ª edição · pt-br · 240 páginas · ISBN 9780306406157')
  assert.notEqual(editionLabel(edition), editionLabel({ ...edition, language: 'en', page_count: 200 }))
  assert.equal(editionLabel({ title: 'Sem ISBN', language: 'pt', page_count: 100 }), 'Sem ISBN · pt · 100 páginas')
})
