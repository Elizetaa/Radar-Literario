# Validação mobile do Radar Literário

O frontend usa a API Go do ramo `origin/Integração-com-Backend` com MariaDB. Para preparar a API, use as instruções de configuração, migração e carga inicial disponíveis no README daquele ramo.

```bash
git show origin/Integração-com-Backend:README.md
cd Frontend/radar_literario_integrado
npm ci
npm run lint
npm run build
npm run dev
```

Com a API em `127.0.0.1:8080`, acesse `http://localhost:5173`. Use contas de teste diferentes para publicar e para curtir, comentar ou denunciar. Tenha pelo menos uma obra com progresso na estante para testar a avaliação.

## Checklist com a API real

Os testes foram executados no Brave com emulação de toque e com a API Go e o MariaDB locais. Em 768 px, use os controles do tablet. A folha inferior e o botão flutuante aparecem até 767 px.

| Fluxo | Passos de verificação | 360 px | 390 px | 430 px | 768 px |
| --- | --- | :---: | :---: | :---: | :---: |
| Autenticação | Criar conta, sair, tentar uma senha incorreta e entrar com a senha correta. | ✓ | ✓ | ✓ | ✓ |
| Início | Conferir a saudação, a seção de retomada e o feed. Curtir e denunciar uma publicação de outra conta. | ✓ | ✓ | ✓ | ✓ |
| Sugestões | Conferir a contagem de páginas, passar um cartão por arrasto e salvar outro na estante. | ✓ | ✓ | ✓ | ✓ |
| Pesquisar | Consultar pelo nome do autor e pelo título, selecionar um gênero, abrir o detalhe e adicionar a obra. | ✓ | ✓ | ✓ | ✓ |
| Estante | Alternar filtros de estado e gênero, registrar progresso e remover uma obra. | ✓ | ✓ | ✓ | ✓ |
| Diário e avaliação | Enviar um diário, tentar avaliar sem nota, selecionar as estrelas e publicar. | ✓ | ✓ | ✓ | ✓ |
| Comentários | Abrir a discussão de uma publicação, enviar um comentário e confirmar que ele aparece no painel. | ✓ | ✓ | ✓ | ✓ |
| Perfil | Abrir as quatro abas, conferir uma curtida, editar o nome, sair e conferir a mensagem de indisponibilidade da exclusão. | ✓* | ✓* | ✓* | ✓* |
| Adicionar livro | Abrir a lista inicial de obras, buscar por autor e adicionar um resultado. | ✓ | ✓ | ✓ | ✓ |
| Feedback | Conferir busca vazia, erro de autenticação e mensagens de confirmação. | ✓ | ✓ | ✓ | ✓ |

\* A rota de exclusão existe no Go. A chamada depende de um método no cliente da API, que ficou fora do escopo das mudanças aos serviços.

- Em 320 px, verificou-se a largura de autenticação, início e destinos da barra inferior. A mesma bateria de fluxos passou nessa largura.
- Em 390 px, verificaram-se arrasto nos dois sentidos, fechamento da folha com restauração do foco, espaço de envio com altura reduzida e recuperação após falha de rede no catálogo.
- Em paisagem, verificou-se o layout de Sugestões em 720 × 390 px.
- Em 1024 e 1440 px, compararam-se as posições e larguras do cabeçalho, menu lateral, início, feed e coluna de leitura com a versão integrada original.
- O axe-core não sinalizou violações nas verificações de autenticação, início, folha de leitura, sugestões, busca e perfil em 390 px. A validação manual em aparelhos físicos, incluindo o teclado do Safari no iOS, continua pendente.
