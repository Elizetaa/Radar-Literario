export const referenceResources = {
  authors: { title: 'Autores', singular: 'autor', create: 'Novo autor' },
  genres: { title: 'Gêneros', singular: 'gênero', create: 'Novo gênero' },
  publishers: { title: 'Editoras', singular: 'editora', create: 'Nova editora' },
}

export function referenceValues(resource, record = {}) {
  const common = { name: record.name || '' }
  if (resource === 'authors') return {
    ...common, biography: record.biography || '', birth_date: record.birth_date || '',
    death_date: record.death_date || '', country_of_origin: record.country_of_origin || '', sex: record.sex || '',
  }
  if (resource === 'publishers') return { ...common, website_url: record.website_url || '', country_of_origin: record.country_of_origin || '' }
  return { ...common, active: record.active ?? true }
}

export function buildReferencePayload(resource, values, original) {
  if (!referenceResources[resource]) throw new Error('Tipo de cadastro desconhecido.')
  const body = Object.fromEntries(Object.entries(referenceValues(resource, values)).map(([key, value]) => [key, typeof value === 'string' ? value.trim() : value]))
  if (!body.name) throw new Error('Informe o nome do cadastro.')
  if (body.name.length > (resource === 'genres' ? 100 : 160)) throw new Error('O nome excede o tamanho permitido pela API.')
  if (resource === 'authors') {
    if (body.birth_date && body.death_date && body.death_date < body.birth_date) throw new Error('A data de falecimento não pode ser anterior ao nascimento.')
    body.birth_date ||= null
    body.death_date ||= null
    body.sex ||= null
  }
  if (!original) return body
  const previous = referenceValues(resource, original)
  return Object.fromEntries(Object.entries(body).filter(([key, value]) => value !== (previous[key] === '' && ['birth_date', 'death_date', 'sex'].includes(key) ? null : previous[key])))
}

export function filterReferences(items, query, active = '') {
  const needle = query.trim().toLocaleLowerCase('pt-BR')
  return items.filter((item) => (!needle || `${item.id} ${item.name}`.toLocaleLowerCase('pt-BR').includes(needle)) && (active === '' || item.active === (active === 'true')))
}
