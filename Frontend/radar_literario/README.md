# Radar Literário

Frontend em React 19, Vite e Tailwind CSS 4 para descobrir livros, acompanhar leituras e compartilhar avaliações. A interface combina utilitários Tailwind e estilos responsivos próprios inspirados nas cores roxas, cartões e hierarquia visual das telas de `../../Docs/imagens.pdf`.

## Executar

Requer Node.js e npm instalados.

```sh
npm install
npm run dev
```

Comandos disponíveis: `npm run build` cria a versão de produção; `npm run lint` executa o ESLint; `npm run preview` serve localmente a versão compilada.

## Estrutura

- `src/App.jsx`: navegação, telas e interações da aplicação.
- `src/App.css` e `src/index.css`: identidade visual, acessibilidade e adaptações responsivas.
- `src/data/mockData.js`: registros fictícios iniciais.
- `src/services/libraryService.js`: camada assíncrona em memória para livros, publicações e comentários.
- `src/features/book-discovery/`: tela de indicações isolada, seus estilos, catálogo demonstrativo e armazenamento local de interesses.

## Telas e fluxos

Início/feed, busca e filtro de catálogo, estante com progresso, sugestões, perfil com abas, cadastro de livro, diário de leitura, avaliação, comentários, denúncia de publicação e confirmações para remover livro ou excluir conta. O menu de criação também permite iniciar cadastro, diário ou avaliação.

### Indicações de livros

Em **Sugestões → Indicações para você**, os livros são apresentados em cartões no estilo Tinder, com título, autor, gênero, nota demonstrativa, páginas e sinopse.

- Arraste para a esquerda ou escolha **Não quero ler**: apenas passa para o próximo livro, sem salvar ou alterar a estante.
- Arraste para a direita ou escolha **Quero ler**: salva automaticamente o livro com o status **Tenho interesse**.
- Com o cartão em foco, as setas esquerda e direita do teclado executam as mesmas ações.
- Os livros aceitos aparecem em **Estante → Tenho interesse**, identificados também no cartão da estante. Ao começar a registrar progresso, passam a ser livros em leitura.
- Quando as indicações acabam, a tela permite abrir a estante. Os livros descartados podem reaparecer ao voltar à tela, pois descartá-los não registra uma preferência.

A opção **Destaques da comunidade** preserva a tela anterior de sugestões. A navegação principal continua a mesma em desktop e celular.

## Dados e decisões

O serviço em `src/services/libraryService.js` é o ponto de troca por chamadas HTTP. Os registros originais continuam em memória e reiniciam ao recarregar a página. Os livros aceitos nas indicações, incluindo seu progresso posterior e sua remoção, são mantidos no `localStorage` deste navegador, com a chave exclusiva `radar-literario:prototype:interested-books:v1`. Se o navegador bloquear esse armazenamento, a tela continua funcionando em memória durante a sessão.

Não há login, autorização ou segurança real; a sessão e as ações de conta são demonstrações. As capas são ilustrações locais em CSS, então não dependem de chamadas externas. O catálogo de indicações, suas notas e seus metadados são demonstrativos, sem integração com serviços externos. A funcionalidade permanece inteiramente dentro de `Frontend/radar_literario`, sem dependências novas ou alterações no back-end.

As imagens do protótipo são capturas em formato mobile; a interface preserva os elementos e fluxos visíveis e reorganiza o conteúdo para desktop, mantendo navegação compacta em telas estreitas. O perfil aberto corresponde à sessão já iniciada nas capturas, por isso não foi criada uma tela de login. A exclusão de conta é tratada como confirmação de demonstração. A referência inclui telas de registro de leitura, avaliação, cadastro de obra, feed, perfil, comentários e recomendações; os rótulos inconsistentes são normalizados para “diário de leitura”, “avaliação” e “livro”.

## Limitações

Sem backend, persistência em servidor ou autenticação real. O armazenamento dos interesses é local a este navegador e não sincroniza contas ou dispositivos. A especificação `../../Docs/RadarLiterárioV2.pdf` orienta a aplicação definitiva; os fluxos atuais são demonstrações do protótipo.
