export const initialBooks = [
  { id: 1, title: 'Quarta Asa', author: 'Rebecca Yarros', genre: 'Fantasia', pages: 108, totalPages: 896, rating: 4.8, description: 'Uma jovem enfrenta um mundo de dragões, alianças e escolhas impossíveis.' },
  { id: 2, title: 'Trono de Vidro', author: 'Sarah J. Maas', genre: 'Fantasia', pages: 310, totalPages: 392, rating: 4.7, description: 'Uma assassina disputa sua liberdade em um reino cheio de segredos.' },
  { id: 3, title: 'Reino de Cinzas', author: 'Sarah J. Maas', genre: 'Fantasia', pages: 0, totalPages: 980, rating: 4.9, description: 'O destino de Terrasen e de seus aliados está por um fio.' },
  { id: 4, title: 'Jogo de Amor para Dois', author: 'Ali Hazelwood', genre: 'Romance', pages: 0, totalPages: 368, rating: 4.3, description: 'Uma comédia romântica sobre trabalho, rivalidade e novas possibilidades.' },
  { id: 5, title: 'A Hipótese do Amor', author: 'Ali Hazelwood', genre: 'Romance', pages: 245, totalPages: 352, rating: 4.5, description: 'Uma experiência acadêmica improvável ganha contornos inesperados.' },
]

export const initialPosts = [
  { id: 1, person: 'Joana Moreira', handle: '@joanale', initials: 'JM', title: 'Alchemised', author: 'SenLinYu', rating: 4, text: 'Gostei do livro. Fiquei com pena do Kain durante boa parte, sempre que ele contava algo de mim. Porém, não sei como suportei o Luc.', likes: 24, bookId: 1, time: 'há 2 horas' },
  { id: 2, person: 'Matheus Costa', handle: '@matcosta', initials: 'MC', title: 'Revolução dos Bichos', author: 'George Orwell', rating: 0, pages: 42, totalPages: 120, text: 'Minha leitura da semana. Uma história curta que faz pensar por muito tempo.', likes: 12, bookId: 2, time: 'ontem' },
  { id: 3, person: 'Lia Nascimento', handle: '@lian', initials: 'LN', title: 'Quarta Asa', author: 'Rebecca Yarros', rating: 5, text: 'Não consegui parar de ler. Já quero começar o próximo!', likes: 38, bookId: 1, time: 'há 3 dias' },
]

export const initialComments = [
  { id: 1, person: 'Pudim', handle: '@pudim', text: 'CONCORDO!!!', time: 'há 10 min' },
  { id: 2, person: 'Cleiton', handle: '@cleiton', text: 'Luc está passando por problemas', time: 'há 8 min' },
]

export const currentUser = { name: 'Pudim Nascimento', handle: '@pudim', bio: 'Leitora de fantasia, café e boas histórias.', location: 'Brasil', joined: 'jun/2026', favoriteBook: 'Reino de Cinzas', initials: 'PN' }
