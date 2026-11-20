const preset = (resource, label, method, path, body, description = '') => ({
  id: `${resource}-${method}-${path}`, resource, label, method, path, body, description,
})

export const consoleResources = [
  ['works', 'Obras'], ['editions', 'Edições'], ['authors', 'Autores'],
  ['genres', 'Gêneros'], ['publishers', 'Editoras'], ['reader', 'Leitor e leitura'],
]

const workExample = {
  title: 'Obra de teste', original_language: 'pt-br', synopsis: 'Sinopse da obra de teste.',
  status: 'draft', minimum_age: 0, author_ids: [], genre_ids: [],
}
const editionExample = {
  work_id: 1, publisher_id: 1, title: 'Edição de teste', language: 'pt-br',
  page_count: 200, edition_number: 1, isbn: null,
  isbn_absence_reason: 'Edição de teste sem ISBN', status: 'draft', minimum_age: 0,
}

export const consolePresets = [
  preset('works', 'Listar obras publicadas', 'GET', '/works?page=1&page_size=20'),
  preset('works', 'Listar obras administrativas', 'GET', '/admin/works?page=1&page_size=20', undefined, 'Inclui rascunhos e obras arquivadas. Adicione filtros title, author, genre_id, isbn, language ou status.'),
  preset('works', 'Consultar obra administrativa', 'GET', '/admin/works/{id}'),
  preset('works', 'Criar obra em rascunho', 'POST', '/admin/works', workExample, 'Associe os IDs reais em author_ids e genre_ids. Para publicar, preencha a sinopse e essas relações.'),
  preset('works', 'Alterar título da obra', 'PATCH', '/admin/works/{id}', { title: 'Título atualizado' }),
  preset('works', 'Publicar obra', 'PATCH', '/admin/works/{id}', { status: 'published' }, 'A obra precisa atender aos requisitos de publicação.'),
  preset('works', 'Arquivar obra', 'DELETE', '/admin/works/{id}', undefined, 'Arquiva a obra; suas edições deixam de aparecer no catálogo público. O backend preserva o histórico.'),
  preset('editions', 'Listar todas as edições', 'GET', '/admin/editions?page=1&page_size=20', undefined, 'Filtros disponíveis: work_id e status.'),
  preset('editions', 'Listar edições públicas da obra', 'GET', '/works/{work_id}/editions?page=1&page_size=20'),
  preset('editions', 'Consultar edição administrativa', 'GET', '/admin/editions/{id}'),
  preset('editions', 'Criar edição em rascunho', 'POST', '/admin/editions', editionExample, 'Troque work_id e publisher_id pelos IDs reais. ISBN deve ser válido ou nulo com justificativa.'),
  preset('editions', 'Alterar edição', 'PATCH', '/admin/editions/{id}', { title: 'Edição revisada', page_count: 220 }),
  preset('editions', 'Publicar edição', 'PATCH', '/admin/editions/{id}', { status: 'published' }),
  preset('editions', 'Arquivar edição', 'DELETE', '/admin/editions/{id}'),
  preset('editions', 'Consultar ISBN', 'GET', '/catalog/isbn/{isbn}'),
  ...[
    ['authors', 'autores', 'autor', { name: 'Autor de teste', biography: 'Biografia do autor.', country_of_origin: 'Brasil' }],
    ['genres', 'gêneros', 'gênero', { name: 'Gênero de teste', active: true }],
    ['publishers', 'editoras', 'editora', { name: 'Editora de teste', country_of_origin: 'Brasil', website_url: 'https://example.com' }],
  ].flatMap(([resource, plural, singular, body]) => [
    preset(resource, `Listar ${plural}`, 'GET', `/admin/${resource}?page=1&page_size=20`),
    preset(resource, `Consultar ${singular}`, 'GET', `/admin/${resource}/{id}`),
    preset(resource, `Criar ${singular}`, 'POST', `/admin/${resource}`, body),
    preset(resource, `Alterar ${singular}`, 'PATCH', `/admin/${resource}/{id}`, { name: 'Nome atualizado' }),
    preset(resource, resource === 'genres' ? 'Desativar gênero (DELETE)' : `Excluir ${singular}`, 'DELETE', `/admin/${resource}/{id}`, undefined, resource === 'genres' ? 'DELETE desativa o gênero e preserva as relações existentes.' : 'O backend pode impedir a exclusão quando existem relações com outros registros.'),
  ]),
  preset('genres', 'Desativar gênero', 'PATCH', '/admin/genres/{id}', { active: false }),
  preset('reader', 'Consultar perfil', 'GET', '/me'),
  preset('reader', 'Listar minha estante', 'GET', '/me/shelf?page=1&page_size=20'),
  preset('reader', 'Listar minhas avaliações', 'GET', '/me/reviews?page=1&page_size=20'),
  preset('reader', 'Listar meus diários', 'GET', '/me/reading-entries?page=1&page_size=20'),
  preset('reader', 'Consultar estatísticas', 'GET', '/me/statistics'),
  preset('reader', 'Listar notificações não lidas', 'GET', '/me/notifications?page=1&page_size=20&unread=true'),
  preset('reader', 'Listar recomendações', 'GET', '/me/recommendations?page=1&page_size=20'),
  preset('reader', 'Consultar interesses', 'GET', '/me/interests'),
]

export function splitConsolePath(value) {
  const [path, ...query] = value.trim().split('?')
  return {
    path,
    query: [...new URLSearchParams(query.join('?'))].map(([name, value]) => ({ name, value, enabled: true })),
  }
}

export function consoleVariables(value) {
  return [...new Set([...value.matchAll(/\{([a-z_]+)\}/g)].map((match) => match[1]))]
}

export function buildConsolePath(value, query = [], variables = {}) {
  const parsed = splitConsolePath(value)
  const replace = (text) => text.replace(/\{([a-z_]+)\}/g, (_, name) => {
    const replacement = String(variables[name] ?? '').trim()
    if (!replacement) throw new Error(`Preencha o parâmetro ${name} do caminho.`)
    return replacement
  })
  const path = parsed.path.replace(/\{([a-z_]+)\}/g, (_, name) => encodeURIComponent(replace(`{${name}}`)))
  if (!path.startsWith('/') || path.startsWith('//') || path.includes('://') || /[\\#]/.test(path)) {
    throw new Error('Informe um caminho relativo à API, como /admin/works. Não use endereço completo ou fragmento (#).')
  }
  const parameters = new URLSearchParams()
  for (const item of [...parsed.query, ...query]) {
    if (item.enabled !== false && item.name.trim()) parameters.append(item.name.trim(), replace(item.value))
  }
  const suffix = parameters.toString()
  return `${path}${suffix ? `?${suffix}` : ''}`
}

export function parseConsoleBody(method, value) {
  if (method === 'GET' || !value.trim()) return undefined
  try { return JSON.parse(value) } catch { throw new Error('O corpo precisa conter JSON válido.') }
}
