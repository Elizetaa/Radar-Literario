# Catálogo: decisões, contratos e evidências

Este complemento detalha o módulo de catálogo. A matriz geral está em `cobertura_backend.md`; a verificação da interface permanece fora desta entrega.

| Categoria | Contexto/fonte | Escolha e impacto | Aceite |
| --- | --- | --- | --- |
| Requisito de fonte | RF05/RF29, UC05/UC12 e diagrama | Obra e edição separadas; ISBN, páginas, editora e idioma pertencem à edição; autores e gêneros são N:M. | Cadastro administrativo e sugestão com edição vinculada; FKs SQL. |
| Compatibilização entre fontes | Diagrama versus descrição de autor único em `Arquitetura_API.md` | `work_authors` e `work_genres` com pares únicos. | Cadastro com vários autores, sem repetir IDs. |
| Premissa proposta para lacuna | ISBN-10/13 equivalentes | Normalização para ISBN-13; validação dos verificadores; índice único também protege concorrência. ISBN ausente exige justificativa. | `TestNormalizeISBNCanonicalAndChecksums`; edição manual sem ISBN no teste de validação. |
| Premissa proposta para lacuna | RF33 e UC08 | Leitor envia proposta JSON estrita para obra ou edição. Aprovação bloqueia a sugestão e verifica `version` do alvo; catálogo, decisão e auditoria têm a mesma transação. Alteração concorrente exige nova sugestão. | Testes de conflito, repetição, concorrência e rollback. |
| Premissa proposta para lacuna | Datas bibliográficas incompletas | Nascimento/morte/publicação são datas civis opcionais; não são inventadas. `null` remove datas em PATCH. Edição datada não antecede a primeira publicação da obra. | Teste de PATCH e validações de datas. |
| Premissa proposta para lacuna | Dados mínimos de publicação | Publicar obra exige título, sinopse, código de idioma, autores, gêneros e classificação explícita na criação. Publicar edição exige vínculo válido, páginas positivas, ISBN ou motivo, idioma e classificação explícita. | Validação de obra incompleta e edição inválida. |
| Premissa proposta para lacuna | Exclusão com referências históricas | Obra/edição são arquivadas; gênero é desativado. Autor/editora só podem ser removidos sem vínculos. Reduzir páginas não pode invalidar estante/diários. | FKs `RESTRICT` e validação transacional. |
| Premissa proposta para lacuna | RF25–RF27 | Afinidade conta gêneros distintos em comum com interesses e avaliações publicadas de nota ≥4. Exclui estante; desempata por média, amostra e ID. Sem afinidade, usa a ordenação de notas. Popularidade conta leitores ativos distintos. Mínimo configurável para ranking de notas. | Filtros e identificação do leitor no teste de recomendações; SQL agregado e ordenação fixa no repositório. |
| Compatibilização entre fontes | Sexo opcional de autor no PDF | Texto opcional com até 80 caracteres; ausência não impede cadastro. | DTO próprio sem credenciais. |

## Rotas e parâmetros

O prefixo é `/api/v1`. Catálogo público aplica status `published` e classificação etária; possuir papel administrativo não altera automaticamente a visibilidade nessas rotas.

`GET /works` aceita `title`, `author`, `isbn`, `genre_id`, `language` e `order=title|newest|popular|top-rated`; todas as listas aceitam `page`/`page_size`. `GET /works/{id}`, `/works/{id}/editions`, `/editions/{id}`, `/catalog/isbn/{isbn}`, `/authors[/{id}]`, `/publishers[/{id}]`, `/genres` e os rankings são públicos para registros elegíveis. Detalhe de obra inclui autores, gêneros, média/amostra; nenhuma propriedade de estante de outro leitor é exposta.

`GET /me/recommendations`, `GET|PUT /me/interests`, `GET /me/preferred-works`, `PUT|DELETE /me/preferred-works/{work_id}`, `POST /catalog-suggestions` e `GET /me/catalog-suggestions[/{id}]` exigem leitor. Preferida não adiciona obra à estante. Obras preferidas continuam sujeitas à classificação e ao status atual do catálogo.

`/admin/works`, `/admin/editions`, `/admin/authors`, `/admin/publishers` e `/admin/genres` possuem GET/POST na coleção e GET/PATCH/DELETE no item; consultas administrativas permitem rascunhos/arquivados. `/admin/catalog-suggestions[/{id}]` consulta a fila; POST `/admin/catalog-suggestions/{id}/approve|reject` exige `reason` de 3 a 1000 caracteres. Essas operações exigem permissão de catálogo.

Criação de obra aceita `initial_edition` para inclusão atômica com ISBN. Os relacionamentos `author_ids`, `genre_ids` e `publisher_id` referenciam cadastros existentes. O administrador mantém autores/editoras/gêneros; a sugestão de leitor não os cria implicitamente. Uma sugestão de obra com edição inicial aprovada publica ambos. Na manutenção administrativa, os estados são informados explicitamente conforme o cadastro desejado.

Proposta de sugestão: `operation=create|update|delete`, `resource=work|edition`, `target_id` apenas em alteração/exclusão e `proposal` com o DTO correspondente. Exclusão usa `proposal: {}`. Proposta não aceita autoria ou privilégio e não altera o catálogo antes da aprovação. `first_publication_date`, `publication_date`, datas de autor e ISBN têm suporte explícito a `null` quando anuláveis.

## Rastreabilidade

| Requisito | Serviço/rotas | Migração | Teste/evidência |
| --- | --- | --- | --- |
| RF05, RF29, UC05, UC12 | `Catalog.SaveWork`, `SaveEdition`, `ISBN`, `CreateSuggestion` | `00002_catalog.sql`: obras/edições/ISBN/FKs | Verificador ISBN, edição manual, campos obrigatórios, restrições MariaDB nos testes gerais. |
| RF06, RF07 | `Catalog.Works`, `Work`, `Editions`, autores/editoras/gêneros | Índices por título/status/obra/gênero/autor | HTTP: ID inválido, rascunho invisível, limites de paginação; serviço: idade e papéis. |
| RF25, RF26, RF27 | `Recommendations`, `/rankings/popular`, `/rankings/top-rated` | Relações de gêneros/interesses/estante/avaliações | `TestCatalogRecommendationsUseActorAndMinimumSample`; agregações no repositório. |
| RF33, UC08 | `CreateSuggestion`, `DecideSuggestion` | `catalog_suggestions`, versões em obra/edição | Propriedade, alvo alterado, decisão única, concorrência e rollback de auditoria. |
| UC04 | `Preferred`, `SetPreferred`, `Interests`, `SetInterests` | `reader_preferred_works`, `reader_genres` e FK de favorito | Paginação e elegibilidade; leitura utiliza callbacks do catálogo nos relatórios. |
| NR08, RNF02 | Catálogo público/edição/ISBN/sugestão/recomendação | `minimum_age` com CHECK | Detalhe protege rascunho e idade; leitor não acessa manutenção administrativa. |

## Evidência de integração real

Em 27/10/2026 foram executados e aprovados os três testes de `internal/model/tools/catalog_integration_test.go` contra MariaDB 13.0.2, banco isolado `radar_literario_test`, com migrações aplicadas. O comando, a partir de `Backend`, foi:

```powershell
Set-ExecutionPolicy -Scope Process Bypass -Force
.\scripts\with_env.ps1 -EnvFile .env.test -GoArgs @('test','-tags=integration','./internal/model/tools','-run','TestMariaDBCatalog','-count=1','-v')
```

`TestMariaDBCatalogSuggestionConcurrentDecisionVersionAndRollback` executa oito aprovações simultâneas sobre conexões reais: há uma decisão aplicada, sete conflitos, um incremento de versão e um evento de auditoria de aprovação. Verifica também que uma proposta desatualizada permanece pendente e não sobrescreve uma edição recente. Uma falha de auditoria injetada dentro do repositório transacional reverte tanto o catálogo quanto a decisão.

`TestMariaDBCatalogISBNEquivalenceAndConcurrentUnique` disputa oito cadastros alternando ISBN-10 e seu ISBN-13 equivalente: apenas uma edição persiste e ambas as consultas apontam para ela. Confere a relação com dois autores e busca pelo segundo autor.

`TestMariaDBCatalogRecommendationsRankingsAndEligibility` consulta agregações reais: afinidade por interesse e avaliação de nota alta, exclusão da estante, rascunhos e classificação etária, média sem avaliação acima da idade, amostra mínima no ranking de notas, popularidade por leitores distintos e ordem determinística repetida.

As fixtures usam IDs próprios, conexões do pool nas disputas concorrentes e limpeza limitada a esses registros; não há `TRUNCATE`. Os testes de serviço continuam utilizando fakes para falhas controladas e validação rápida. Frontend, Docker e serviços bibliográficos externos não foram validados por esses testes.
