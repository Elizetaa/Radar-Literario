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

## Telas e fluxos

Início/feed, busca e filtro de catálogo, estante com progresso, sugestões, perfil com abas, cadastro de livro, diário de leitura, avaliação, comentários, denúncia de publicação e confirmações para remover livro ou excluir conta. O menu de criação também permite iniciar cadastro, diário ou avaliação.

## Dados e decisões

O serviço em `src/services/libraryService.js` é o ponto de troca por chamadas HTTP. Os registros são mantidos apenas em memória e reiniciam ao recarregar a página. Não há login, autorização ou segurança real; a sessão e as ações de conta são demonstrações. As capas são ilustrações locais em CSS, então não dependem de chamadas externas.

As imagens do protótipo são capturas em formato mobile; a interface preserva os elementos e fluxos visíveis e reorganiza o conteúdo para desktop, mantendo navegação compacta em telas estreitas. O perfil aberto corresponde à sessão já iniciada nas capturas, por isso não foi criada uma tela de login. A exclusão de conta é tratada como confirmação de demonstração. A referência inclui telas de registro de leitura, avaliação, cadastro de obra, feed, perfil, comentários e recomendações; os rótulos inconsistentes são normalizados para “diário de leitura”, “avaliação” e “livro”.

## Limitações

Sem backend, persistência ou autenticação real. A especificação `../../Docs/RadarLiterárioV2.pdf` não pôde ser extraída neste ambiente por falta de ferramentas de leitura de PDF; as notas conceituais disponíveis no backend não detalham regras funcionais.
