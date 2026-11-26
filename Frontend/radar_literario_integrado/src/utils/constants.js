// Constantes compartilhadas entre os componentes (menu e filtros da estante).

// Itens do menu principal: [id da página, rótulo, ícone].
// O mesmo array alimenta a barra lateral (desktop) e a barra inferior (celular),
// então adicionar uma página aqui já a coloca nos dois menus.
export const navItems = [
  ['home', 'Início', '⌂'],
  ['search', 'Pesquisar', '⌕'],
  ['shelf', 'Estante', '▤'],
  ['suggestions', 'Sugestões', '♧'],
  ['profile', 'Perfil', '●'],
]

// Definindo o caminho pra cada página
export const pagePaths = {
  home: '/',
  search: '/pesquisar',
  shelf: '/estante',
  suggestions: '/sugestoes',
  profile: '/perfil',
}
// Filtros por situação na página "Minha estante".
// Qualquer outro valor de filtro é tratado como NOME DE GÊNERO (vem do menu "Suas leituras").
export const shelfStatusFilters = ['Todos', 'Tenho interesse', 'Em leitura', 'Quero ler', 'Concluídos']
