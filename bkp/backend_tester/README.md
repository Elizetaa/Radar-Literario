# Testador do backend — Radar Literário

Aplicação independente em React e Vite, com CSS simples e sem dados simulados.

## Executar

Com o MariaDB ativo, inicie a API em outro terminal seguindo o [README do backend](../../Backend/README.md). Se o Go já estiver no PATH:

```powershell
cd C:\Projetos\Radar-Literario\Backend
go run ./cmd/api
```

Inicie este frontend:

```powershell
cd C:\Projetos\Radar-Literario\Frontend\backend_tester
npm install
npm run dev
```

Abra **http://localhost:5173**. Caso o protótipo `radar_literario` esteja usando essa porta, encerre seu servidor primeiro. A porta é fixa para corresponder às origens CORS e CSRF já configuradas no backend.

O Vite encaminha `/api` para `http://127.0.0.1:8080`. Para mudar o destino, copie `.env.example` para `.env.local`, ajuste `API_PROXY_TARGET` e reinicie o servidor. Nenhuma credencial do banco deve ser colocada neste projeto.

## Páginas e roteiro de teste

1. **Conta:** cadastre um leitor, entre e confira o perfil. O cadastro não inicia sessão automaticamente. Novas senhas precisam ter entre 8 e 72 bytes.
2. **Catálogo público:** busque obras publicadas por título, autor, gênero, ISBN ou idioma da edição; ordene e consulte detalhes e edições com paginação. Clique em “Usar na estante” para iniciar o registro de leitura. A classificação etária da conta continua valendo; visitantes veem apenas obras livres.
   Em **Adicionar obra**, cadastre o título, idioma, sinopse, autores, gêneros e classificação etária. Uma conta com permissão de catálogo pode publicar diretamente ou salvar como rascunho. O formulário oferece cadastros rápidos de autores, gêneros e editoras e permite incluir a primeira edição na mesma operação (com ISBN ou justificativa para sua ausência). Obras em rascunho não aparecem no catálogo público.
3. **Gerenciar obras:** liste todos os estados, inclusive rascunhos e arquivados, com filtros e total de resultados. Consulte o JSON completo, edite campos e relacionamentos, remova uma data deixando-a vazia e arquive obras. O atalho “Edições” abre a lista filtrada pela obra.
4. **Edições:** liste por obra/estado, cadastre novas versões, consulte, edite e arquive. Os seletores incluem todas as páginas de obras e editoras. É possível alterar editora, título, idioma, páginas, número da edição, ISBN/justificativa, data, capa e classificação. A obra de uma edição existente é imutável.
5. **Autores, gêneros e editoras:** consulte listas e detalhes; crie, edite e remova cadastros com formulários completos. Autores incluem biografia, nascimento, falecimento, sexo e país; editoras incluem país e site. Gêneros podem ser desativados e reativados. A consulta por ID funciona em todos os recursos administrativos e em autores/editoras públicos. Os filtros de nome/ID e estado são explicitamente locais à página: esses endpoints não oferecem busca no servidor.
6. **Estante:** escolha a obra pela lista de títulos e selecione uma de suas edições. Quando houver somente uma edição, ela será selecionada automaticamente. Salve como planejada com página zero. Depois altere para “Em leitura”, avance páginas e consulte as estatísticas. Também é possível corrigir o progresso, trocar a edição e remover o item. O botão “Editar” preenche as listas com a obra e a edição salvas; não é necessário informar IDs.
7. **Publicações:** crie um diário ou uma avaliação usando os IDs da obra/edição. Escolha rascunho ou publicado e marque spoiler/conteúdo sensível quando necessário. Uma avaliação exige leitura iniciada ou concluída. A revelação de conteúdo é explícita.
8. **Console da API:** escolha modelos de CRUD e consultas por recurso, preencha variáveis de caminho, edite/ative parâmetros de consulta e formate o JSON antes de enviar. Também aceita caminhos livres após `/api/v1`, conforme o [contrato OpenAPI](../../Docs/api.yaml). Os modelos apenas preenchem o formulário; os IDs do corpo JSON devem ser substituídos pelos reais.

As operações de manutenção exigem `catalog_manager`, `administrator` ou `system_administrator`. O cadastro público cria leitores; use uma conta administrativa existente ou siga o [bootstrap administrativo do backend](../../Backend/README.md#5-primeiro-administrador). A interface e a API verificam essas permissões.

O `DELETE` de obra/edição arquiva; o de gênero desativa. Autores e editoras vinculados não podem ser excluídos. Confirmações exibem o registro afetado; erros como HTTP 409 mantêm os campos preenchidos para correção. Requisições de escrita alteram os dados reais do banco local.

Cada página mostra erros retornados pelo backend. O painel “Última resposta” guarda as últimas 30 requisições em memória, com método, caminho, código HTTP, duração, corpo enviado e resposta JSON copiável. Senhas e tokens são ocultados no histórico. Cookies de sessão e CSRF são gerenciados automaticamente. Para haver livros no catálogo, cadastre e publique obras e edições; este frontend não insere dados automaticamente.

## Build

```powershell
npm run build
npm run preview
```

## Verificação

```powershell
npm test
npm run test:browser
```

Os testes unitários verificam os corpos dos formulários, limpeza de campos com `null`, relações, paginação de opções, parâmetros do console e ocultação de credenciais no histórico.

O teste de navegador inicia Vite e uma API simulada em memória em portas isoladas (5177 e 18087), usando Edge/Chromium headless e Node 24. Não usa banco nem credenciais reais e encerra os processos ao terminar. Verifica CRUD, conflito HTTP 409, paginação, formulários, console e largura móvel. Se necessário, defina `SMOKE_BROWSER_PATH` para o executável do navegador; `SMOKE_FRONTEND_PORT`, `SMOKE_API_PORT` e `SMOKE_DEBUG_PORT` substituem as portas. `SMOKE_SCREENSHOTS=1` grava capturas em `.local/browser-smoke`. No PowerShell, se os scripts `npm.ps1` forem bloqueados, use `npm.cmd`.

O preview também usa o proxy na porta 5173. O conteúdo de `dist` sozinho depende de um servidor que encaminhe `/api` para o backend.
