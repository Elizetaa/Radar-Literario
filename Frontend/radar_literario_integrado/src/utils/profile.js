// Conversão entre o usuário devolvido pela API (GET /me) e o formato usado nas telas.

// Perfil "vazio": usado antes do login e depois do logout, para as telas nunca receberem undefined.
export const emptyProfile = { name: '', handle: '', bio: '', location: '', joined: '', favoriteBook: '', initials: '' }

// API -> tela. A API manda `nickname`; a tela mostra `@nickname` (handle).
export const toProfile = (me) => ({
  name: me.name,
  handle: `@${me.nickname}`,
  bio: me.biography || '',
  location: me.country_of_origin || '',
  joined: new Date(me.created_at).toLocaleDateString('pt-BR', { month: 'short', year: 'numeric' }),
  favoriteBook: '', // a API só devolve o id da obra favorita; o título ainda não é resolvido
  initials: me.name.slice(0, 2).toUpperCase(),
})
