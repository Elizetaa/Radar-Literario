import test from 'node:test'
import assert from 'node:assert/strict'
import { buildConsolePath, consolePresets, consoleVariables, parseConsoleBody, splitConsolePath } from '../src/consoleRequests.js'

test('monta consultas com acentos, caracteres reservados, repetição e parâmetros desmarcados', () => {
  assert.equal(buildConsolePath('/works', [
    { name: 'title', value: 'Memórias & cartas', enabled: true },
    { name: 'page', value: '1', enabled: true },
    { name: 'page', value: '2', enabled: true },
    { name: 'status', value: 'draft', enabled: false },
    { name: ' ', value: 'ignorar', enabled: true },
  ]), '/works?title=Mem%C3%B3rias+%26+cartas&page=1&page=2')
})

test('preserva query colada no caminho e resolve variáveis sem permitir injeção no caminho', () => {
  assert.deepEqual(splitConsolePath('/works?title=Ol%C3%A1&page=2').query, [
    { name: 'title', value: 'Olá', enabled: true }, { name: 'page', value: '2', enabled: true },
  ])
  assert.equal(buildConsolePath('/catalog/isbn/{isbn}', [], { isbn: 'a/b?c' }), '/catalog/isbn/a%2Fb%3Fc')
  assert.equal(buildConsolePath('/admin/editions?work_id={work_id}', [{ name: 'page', value: '2' }], { work_id: 7 }), '/admin/editions?work_id=7&page=2')
  assert.deepEqual(consoleVariables('/works/{id}/{id}/editions/{edition_id}'), ['id', 'edition_id'])
  assert.throws(() => buildConsolePath('/admin/works/{id}', [], {}), /parâmetro id/)
})

test('rejeita caminhos externos, caminhos vazios, barras invertidas e fragmentos', () => {
  for (const path of ['https://example.com', '//example.com', '', '/works#fragment', '/\\example.com']) {
    assert.throws(() => buildConsolePath(path), /caminho relativo/)
  }
})

test('não envia corpo em GET e mantém DELETE sem corpo; JSON inválido não é enviado', () => {
  assert.equal(parseConsoleBody('GET', '{inválido'), undefined)
  assert.equal(parseConsoleBody('DELETE', ' '), undefined)
  assert.deepEqual(parseConsoleBody('PATCH', '{"active":false,"isbn":null}'), { active: false, isbn: null })
  assert.throws(() => parseConsoleBody('POST', '{inválido'), /JSON válido/)
})

test('todos os exemplos de CRUD geram caminhos válidos e corpos JSON serializáveis', () => {
  for (const resource of ['works', 'editions', 'authors', 'genres', 'publishers']) {
    const presets = consolePresets.filter((preset) => preset.resource === resource)
    for (const method of ['GET', 'POST', 'PATCH', 'DELETE']) assert.ok(presets.some((preset) => preset.method === method))
    for (const preset of presets) {
      const variables = Object.fromEntries(consoleVariables(preset.path).map((name) => [name, '1']))
      assert.ok(buildConsolePath(preset.path, [], variables).startsWith('/'))
      if (preset.body !== undefined) assert.deepEqual(parseConsoleBody(preset.method, JSON.stringify(preset.body)), preset.body)
    }
  }
})
