# Prompt para implementar todo o backend do Radar Literário

Este documento é uma instrução de execução para uma IA com acesso ao repositório. Foi preparado a partir do código e dos documentos disponíveis em 27/10/2026. Ele descreve trabalho futuro: a existência deste arquivo não significa que as funcionalidades estejam implementadas.

**Como usar:** peça à IA: “Leia `Docs/Prompt_Implementacao_Backend.md`, confira as fontes indicadas e execute a implementação completa descrita nele”. As decisões propostas abaixo são premissas explícitas deste prompt, não decisões históricas já aprovadas pela equipe. Se o responsável informar outra regra, ela prevalece e deve ser registrada.

---

## 1. Sua tarefa e o resultado esperado

Implemente o backend completo do Radar Literário neste repositório, com Go, Gin, GORM e MariaDB, preservando MVC e a organização atual em `Backend/cmd` e `Backend/internal`.

Entregue código executável e persistência real para contas, autenticação, perfil, catálogo, estante, diários, avaliações, comunidade, administração, denúncias, punições, sugestões de catálogo, estatísticas, recomendações e suporte. Inclua contratos HTTP, migrações SQL, testes relevantes e instruções reproduzíveis de execução.

Execute a implementação por etapas até concluir todos os módulos. Não encerre a tarefa após produzir apenas um plano, entidades ou endpoints demonstrativos. Não considere implementada uma função que devolve dados fixos, usa memória como armazenamento definitivo, ignora permissões ou contém `TODO`, `panic("not implemented")` ou resposta de sucesso sem efetuar a operação.

O escopo é o backend e sua infraestrutura/documentação. Consulte o frontend como referência de fluxos e consumidores da API; a reescrita ou integração completa da interface não faz parte desta tarefa. Não publique serviços nem execute operações sobre dados de produção. Preserve alterações locais de outras pessoas e não faça limpeza destrutiva do repositório.

## 2. Fontes obrigatórias e tratamento das divergências

Leia antes de implementar:

1. [`Backend/`](../Backend/), especialmente `go.mod`, `cmd/api/main.go`, `internal/router/router.go` e `internal/controller/user.go`.
2. [`Arquitetura_API.md`](../Arquitetura_API.md).
3. [`Class Diagram0.png`](../Class%20Diagram0.png), inspecionando a imagem; não deduza seu conteúdo apenas pelo nome.
4. [`RadarLiterárioV2.pdf`](RadarLiterárioV2.pdf), para requisitos funcionais, regras, requisitos não funcionais e casos de uso.
5. [`Plano_Desenvolvimento.md`](Plano_Desenvolvimento.md).
6. [`Frontend/radar_literario/README.md`](../Frontend/radar_literario/README.md) e, quando necessário, `src/services/libraryService.js` e os fluxos do protótipo. [`imagens.pdf`](imagens.pdf) é referência visual, sem definir o esquema relacional.
7. `flake.nix`, `.gitignore` e instruções locais aplicáveis, se existirem.

Separe três categorias em `Docs/decisoes.md`: **requisito de fonte**, **compatibilização entre fontes** e **premissa proposta para lacuna**. Para cada decisão, registre contexto, fonte, escolha, impacto e testes de aceite. Não transforme comportamentos dos mocks em requisitos definitivos.

Instruções explícitas do responsável prevalecem. Este prompt resolve provisoriamente as divergências que enumera; em outros conflitos, confronte especificação, diagrama e arquitetura, registre a escolha antes do código dependente e identifique qualquer mudança de requisito. Pergunte apenas quando não houver premissa razoável ou quando a escolha exigir uma decisão externa irreversível. Continue as partes independentes.

Crie uma matriz de rastreabilidade em `Docs/cobertura_backend.md`: identificador do requisito/caso de uso, regra resumida, serviço, rotas, migração/entidades e teste ou evidência. Não marque o frontend, entrega de e-mail ou serviços externos como validados sem tê-los executado.

### Checklist de cobertura da especificação

Esta síntese facilita a conferência; consulte os fluxos e condições no PDF, especialmente UC02/UC03.

| Requisitos | Cobertura exigida |
| --- | --- |
| RF01, RF02, RF03 | CRUD de leitor, login/logout e recuperação de senha. |
| RF04 | Perfil público com informações básicas, diários e avaliações, respeitando privacidade. |
| RF05, RF29 | Cadastro de obra via ISBN para leitor/administrador; ISBN associado à edição e triagem para leitor. |
| RF06, RF07 | Busca por título/autor/gênero e detalhe de obra com média e avaliações. |
| RF08, RF09, RF10, RF12 | Avaliação textual com nota 1–5, média, consulta e exclusão própria. |
| RF11, RF18, RF19, RF20 | Comentários em diários/avaliações, respostas e edição/exclusão própria. |
| RF13 | CRUD de diário vinculado a trecho/progresso da obra. |
| RF14, RF15, RF16 | Estante única por obra/status e listas de completos, incompletos, abandonados, recomendados e desejados. |
| RF17 | Quantidade de obras e páginas por mês e por ano, a partir de dados próprios. |
| RF21, RF22 | Curtidas em comentários, diários e avaliações. |
| RF23, RF24 | Conteúdo sensível e ocultação de spoiler/sensível até confirmação. |
| RF25, RF26, RF27 | Recomendações por leituras/gêneros, popularidade e melhores avaliações. |
| RF28 | Cadastro e gestão de administradores. |
| RF30, RF31 | Punição temporária/permanente de leitores e remoção administrativa de conteúdo. |
| RF32 | Gestão de FAQ e suporte. |
| RF33 | Triagem de inclusão, alteração e exclusão solicitadas por leitor. |
| RF34, RF35 | Denúncias de publicação/comentário/leitor e consulta administrativa. |
| RN01–RN05 | Escala de nota, spoiler explícito, autoria, status único e estatísticas próprias. |
| RN06, RN07, NR08 | Administrador pode banir leitor, não pode banir administrador e deve haver restrição etária de conteúdo. |
| RNF02, RNF03 | Controle de acesso por perfil e sessão expirada por inatividade configurável. |
| RNF04, RNF05, RNF06 | Backup periódico, auditoria administrativa e proteção técnica de dados pessoais; política de privacidade/retenção explicitada. |
| RNF01 | Responsividade é responsabilidade da View; o backend fornece contratos adequados, sem declarar a interface validada. |
| UC01, UC02, UC03 | Perfil, diário e avaliação, incluindo edição/exclusão e notificações de comentário/curtida. |
| UC04 | Relatórios/listas, incluindo preferidos, idioma mais lido e páginas. |
| UC05, UC08, UC12 | Obra, triagem e edição de obra; leitores também podem sugerir CRUD de edições. |
| UC6, UC07, UC09, UC10, UC11 | Punições, suporte/FAQ, administradores, estante e denúncia. |

Preserve a grafia `NR08` e `UC6` ao rastrear os identificadores, pois é assim que aparecem na especificação.

## 3. Estado real do projeto e decisões de compatibilização

O estado observado ao preparar este prompt é:

| Item | Situação real |
| --- | --- |
| Módulo | `Backend/go.mod`, módulo `radar-literario/backend`; declara `go 1.27.3` e Gin `v1.12.0`; `go.sum` já existe. |
| Entrada | `Backend/cmd/api/main.go`, chamando `router.Novo()` e escutando em `127.0.0.1:8080`. |
| Rotas | `GET /health`, `GET /api/usuarios/:id` e `POST /api/usuarios`. |
| Controller | `internal/controller/user.go` devolve usuário fictício ou ecoa nome/e-mail; não persiste. |
| Infraestrutura | `internal/config` e `internal/database` contêm apenas `.gitkeep`. |
| Ausências | Sem GORM, MariaDB configurado, migrações, entidades, serviços, repositórios, sessão ou testes de backend. |
| Ambiente | `flake.nix` ainda descreve Python/FastAPI/SQLAlchemy/Pydantic, Node e DBeaver; `.gitignore` está vazio. |
| Documentação antiga | `Arquitetura_API.md` ainda descreve backend Python e uma futura árvore `Backend/src`; esses trechos estão desatualizados. |

Reconfira esse inventário: outra pessoa pode ter avançado o projeto depois desta data.

Use as seguintes compatibilizações, registrando-as:

- **Estrutura:** mantenha `cmd/api` e `internal`; transporte para eles as responsabilidades MVC descritas na arquitetura. Não crie outra aplicação paralela em `Backend/src` ou um segundo `main.go` na raiz do backend.
- **Módulo:** preserve `radar-literario/backend` e ajuste os imports a esse caminho. Confira a toolchain instalada e a compatibilidade real das dependências; a versão escrita no arquivo não prova disponibilidade no ambiente. Justifique qualquer alteração de versão e gere `go.mod`/`go.sum` com ferramentas Go.
- **Execução:** o comando de entrada é `go run ./cmd/api` a partir de `Backend`; `go run .` não corresponde à estrutura atual.
- **Demonstrações:** substitua `/api/usuarios` e os dados fictícios pelo contrato `/api/v1` descrito adiante. Não mantenha esses handlers como uma segunda implementação de contas. Documente a mudança.
- **Autores:** use relação N:M entre `Work` e `Author`, conforme o diagrama, embora a arquitetura textual mencione um autor.
- **Leitura:** preserve também `paused`, que aparece no diagrama além dos quatro estados da arquitetura; não perca os estados exigidos por RN04.
- **Estante:** uma única relação por leitor/obra; a edição é a escolhida para aquele registro. Não permita uma duplicata por edição.
- **UML:** traduza herança para composição e relações SQL explícitas; não confunda embedding de structs Go com herança de tabelas.

## 4. Arquitetura e organização obrigatórias

```text
Backend/
├── cmd/
│   ├── api/main.go                  # Composição e servidor HTTP
│   └── admin/main.go                # Criação controlada do primeiro administrador
├── internal/
│   ├── config/                     # Variáveis, validação e defaults
│   ├── database/                   # GORM, pool e suporte transacional
│   ├── router/                     # Registro das rotas
│   ├── middleware/                 # Sessão, CSRF, CORS, logs e limites
│   ├── controller/                 # Entrada/saída HTTP por recurso
│   └── model/
│       ├── core/                   # Entidades, enums e erros de domínio
│       ├── interfaces/             # Contratos de repositórios/adaptadores
│       ├── tools/                  # GORM, SQL e integrações concretas
│       ├── schemas/                # DTOs de requests e responses
│       └── services/               # Casos de uso, permissões e transações
├── migrations/                     # SQL versionado
├── scripts/                        # Rotinas de apoio quando necessárias
├── .env.example
├── compose.yaml                    # MariaDB de desenvolvimento/teste
├── go.mod
├── go.sum
└── README.md
Docs/
├── api.yaml                        # OpenAPI completo
├── decisoes.md
└── cobertura_backend.md
```

Pastas auxiliares pequenas são permitidas quando tiverem uma responsabilidade concreta; evite criar camadas sem uso. Testes Go ficam junto aos arquivos testados, com sufixo `_test.go`.

Fluxo: **rota/middleware → controller → serviço → interface de repositório → implementação GORM/SQL → MariaDB**. React representa a View do MVC e não acessa o banco.

- Controllers fazem binding, validação de formato e conversão de erros para HTTP. Não consultam o GORM diretamente.
- Serviços aplicam regras, propriedade dos recursos e permissões; recebem dependências por interfaces e não importam Gin.
- Propague `context.Context` do request até as consultas e integrações externas.
- Repositórios concentram persistência; SQL manual usa parâmetros. Não concatene filtros ou ordenações recebidos do cliente.
- O serviço delimita transações; implemente uma unidade de trabalho ou executor transacional que forneça repositórios ligados à mesma transação. Não use o pool global por engano dentro de uma operação transacional.
- Entidades GORM e DTOs são separados. Nunca serialize entidades completas como resposta pública.
- `cmd/api/main.go` monta configuração, banco, repositórios, adaptadores, serviços, controllers e router. Inclua encerramento gracioso e fechamento do pool.
- Arquivos em `snake_case`, pacotes em minúsculas, identificadores e JSON em inglês, documentação e mensagens para o usuário em português.

## 5. Modelo de dados derivado do diagrama

### 5.1. Contas e perfis

Use `users` como base de identidade e credenciais, com `id`, `name`, `nickname`, `email`, `password_hash`, `status`, `birth_date`, `profile_image_url`, `email_verified_at` opcional, `created_at` e `updated_at`.

Use `readers` com `user_id` como PK/FK, `biography` e `profile_visibility`; use `administrators` com `user_id` como PK/FK e `role`. Não duplique senha/e-mail em cada perfil. Premissa: administrador pode ter também perfil de leitor na mesma conta; cadastro público cria apenas leitor e nunca aceita papel/status administrativo.

O minimundo antigo também menciona sexo, país de origem e livro favorito, ausentes ou diferentes no PNG atual. **Premissa de compatibilização:** sexo e país são dados opcionais de perfil, sem inferência nem exposição pública automática; registre a representação escolhida. Livro favorito e obras preferidas têm persistência e operações próprias: `readers.favorite_work_id` opcional e relação `reader_preferred_works` única por leitor/obra. Se mantiver o atributo sexo de autor do PDF, ele também é opcional e não condiciona o cadastro de obras.

Normalize e-mail e apelido para checagem de unicidade conforme uma política documentada. Defina comprimentos, campos obrigatórios e índices únicos no SQL; não dependa apenas de um `SELECT` antes do `INSERT`. Trate conflito concorrente como `409`. Data de nascimento não pode estar no futuro; não exponha nascimento, e-mail ou dados de segurança no perfil público.

Crie entidades adicionais para sessões, recuperação de senha e verificação de e-mail quando usada. Tokens devem ser aleatórios, expiráveis, de uso único quando aplicável e armazenados em forma de hash.

### 5.2. Catálogo

| Entidade | Campos e relações mínimos |
| --- | --- |
| `Publisher` | `id`, `name`, `website_url`, `country_of_origin`, timestamps. Uma editora tem várias edições. |
| `Author` | `id`, `name`, `biography`, `birth_date`, `death_date` opcional, `country_of_origin`. Autor e obra são N:M. |
| `Genre` | `id`, `name` único, `active`. Gênero e obra são N:M. |
| `Work` | `id`, `title`, `synopsis`, `original_language`, `first_publication_date`, `status`, timestamps, autores e gêneros. |
| `Edition` | `id`, `work_id`, `publisher_id`, `title`, `isbn`, `edition_number`, `page_count`, `publication_date`, `language`, `image_url`, `status`, timestamps. |

Crie `work_authors` e `work_genres` com unicidade dos pares e FKs. Uma edição pertence a uma única obra e uma editora; várias edições podem pertencer à mesma obra. Campos de página, idioma de edição, editora e ISBN pertencem à edição. Progresso e nota individual não são atributos globais da obra.

Valide páginas positivas, número de edição e coerência de datas. Não invente datas de nascimento/publicação para completar resultados incompletos de ISBN: use nulabilidade justificada ou mantenha a sugestão pendente. Normalize ISBN-10/ISBN-13, valide dígito verificador e evite cadastrar a mesma edição pelas duas representações equivalentes. Suporte obras/edições sem ISBN quando o cadastro manual justificar essa ausência.

Acrescente classificação etária necessária à política da seção 7. Registros de catálogo publicados devem ter os dados mínimos e classificação definidos. Consultas públicas não exibem rascunhos, pendências, rejeitados, ocultos ou arquivados.

### 5.3. Estante e progresso

Represente a estante e `ReadingRecord` por `reading_records`, contendo `id`, `reader_id`, `work_id`, `edition_id`, `current_page`, `started_at` opcional, `finished_at` opcional, `status` e timestamps.

- Garanta `UNIQUE(reader_id, work_id)`. `reader_id` referencia o perfil leitor.
- A edição escolhida deve pertencer à obra. Expresse essa integridade no banco por chave composta quando viável e valide no serviço.
- Premissa: cada item da estante tem uma edição válida escolhida; mudar de edição é explícito e exige revalidar o progresso, sem transferir páginas silenciosamente.
- `0 <= current_page <= edition.page_count`; status e progresso pertencem à relação pessoal.
- `planned` começa com página zero; iniciar leitura define início; `completed` define conclusão e progresso no total da edição. Documente os efeitos de voltar a outro estado e de corrigir datas.
- `paused` preserva o progresso; `abandoned` registra desistência. Nunca use strings livres como status.
- Remover um item da estante não exclui a obra do catálogo nem apaga automaticamente publicações. Defina preservação/remoção do registro e efeito nas estatísticas conforme a seção 7.

### 5.4. Publicações e interações

Use uma identidade comum `Post` para o feed e suas relações: `id`, `reader_id`, `kind`, `title`, `text`, `data_url` opcional, `status`, `contains_spoiler`, `contains_sensitive_content`, classificação etária e timestamps.

Implemente `ReadingDiary` e `Review` como subtipos 1:1 de `posts`, ou estrutura relacional equivalente documentada. Cada publicação deve ter exatamente um tipo válido; a criação do post e de seu subtipo é atômica. Se usar tabelas separadas, registre como garante a exclusividade no serviço e como testa a integridade.

- `reading_diaries`: `post_id`, `edition_id`, progresso correspondente a `pagesRead` do diagrama. Na API use `current_page` como posição absoluta na edição, distinguindo-a de incremento de páginas.
- `reviews`: `post_id`, `work_id`, `rating` inteiro de 1 a 5; unicidade da avaliação vigente por leitor/obra conforme a premissa da seção 7. A restrição precisa impedir concorrência entre duas criações, mesmo com o autor armazenado no post. Se repetir `reader_id` no subtipo para impor a chave, garanta consistência com a autoria do post.
- `comments`: `id`, `post_id`, `reader_id`, `parent_id` opcional, `text`, `status`, timestamps.
- `post_likes`: `id`, `reader_id`, `post_id`, `created_at`, com par único.
- `comment_likes`: `id`, `reader_id`, `comment_id`, `created_at`, com par único.

O comentário pai deve pertencer ao mesmo post; proíba autorreferência e ciclos. Premissa: pai não é alterado após a criação; encadeie somente comentários existentes e limite a profundidade, documentando o limite. Exclusões preservam a árvore por tombstone quando há respostas, sem devolver o texto removido.

Autoria é obtida da sessão, nunca de `reader_id`, `user_id` ou `author_id` fornecidos pelo cliente. Cadastro/edição não aceitam campos internos de moderação, contadores ou papéis.

### 5.5. Denúncias e punições

`reports`: `id`, `reporter_id`, `reported_user_id` opcional, `reported_post_id` opcional, `reported_comment_id` opcional, `assigned_to` opcional, `reason`, `description`, `report_data_url` opcional, `status`, `resolution_note` opcional, `resolved_at` opcional e timestamps.

Cada denúncia tem **exatamente um** alvo: usuário, post ou comentário. Implemente três FKs anuláveis e `CHECK` equivalente à soma das presenças igual a 1, além da validação do serviço. Não use um ID polimórfico sem integridade referencial.

`penalties`: `id`, `user_id`, `applied_by`, `revoked_by` opcional, `type`, `reason`, `starts_at`, `ends_at` opcional, `revoked_at` opcional, `revocation_reason` opcional, `created_at`. Crie `report_penalties` para a relação N:M do diagrama.

Persistam decisão, responsável, motivo, efeitos e revogação. Não apague denúncias/punições em cascata quando o alvo excluir conteúdo. Use exclusão lógica/referência preservada e defina o tratamento dos dados pessoais na política de exclusão. Restrinja o acesso a evidências e notas internas.

### 5.6. Entidades complementares

O diagrama não cobre todo o escopo. Acrescente, no mínimo:

- `catalog_suggestions`: autor, operação `create/update/delete`, recurso/alvo (`work` ou `edition`), proposta validada, estado, revisão, responsável e motivo da decisão; capture versão do alvo para detectar aprovação sobre dados alterados. UC12 exige também sugestões de criação, alteração e exclusão de edições por leitores, sempre com vínculo à obra.
- `report_messages` ou histórico equivalente: solicitação administrativa de informações e resposta do denunciante, com autoria, instante e visibilidade.
- `audit_logs`: ator, ação, alvo, motivo, data e alterações pertinentes, com dados sensíveis removidos.
- `reader_genres` ou equivalente para interesses explícitos usados em recomendações.
- `reader_preferred_works` e vínculo opcional de livro favorito no perfil, conforme os relatórios e o minimundo.
- `reading_progress_events`: leitor/registro de leitura, obra/edição, página anterior/nova, páginas creditadas, tipo `reading/correction`, origem e instante. A proposta de histórico é necessária para calcular páginas por mês/ano sem somar snapshots de diários.
- `notifications`: destinatário, ator, evento, recurso relacionado, data e leitura; use identidade única do evento para evitar duplicatas. O evento é exigido pelos casos de uso; a persistência e os endpoints são a solução técnica proposta.
- `faq_entries`: pergunta, resposta, ordem e estado de publicação.
- `support_tickets` e `support_messages`: solicitante, assunto, estado, atendente opcional, mensagens, autoria e timestamps. Essa estrutura é uma proposta para concretizar RF32, que não define um modelo de atendimento.
- Tabelas de sessão e tokens de conta descritas anteriormente.

## 6. Enums: preservar nomes conceituais e definir comportamento

Use constantes Go e valores JSON/SQL em inglês, `snake_case`; valide os valores no contrato e no banco. Não traduza silenciosamente estados diferentes para a mesma string.

| Enum do diagrama | Valores da API/persistência |
| --- | --- |
| `AdministratorRole` | `user_manager`, `catalog_manager`, `administrator`, `system_administrator` |
| `PenaltyType` | `warning`, `suspension`, `restriction`, `ban` |
| `ContentStatus` | `draft`, `published`, `hidden`, `removed`, `deleted` |
| `ReadingStatus` | `planned`, `in_progress`, `paused`, `completed`, `abandoned` |
| `ReportReason` | `spam`, `harassment`, `hate_speech`, `threats`, `inappropriate_content`, `privacy_violation`, `copyright_violation`, `impersonation`, `misinformation`, `other` |
| `ReportStatus` | `open`, `in_review`, `awaiting_information`, `resolved`, `dismissed`, `canceled` |
| `UserStatus` | `pending_verification`, `active`, `inactive`, `suspended`, `deactivated`, `banned` |
| `UserVisibility` | `public`, `registered_users_only`, `private` |
| `CatalogStatus` | `draft`, `pending_review`, `published`, `rejected`, `hidden`, `archived` |

Defina e teste transições, não apenas constantes:

- Publicação própria: criação em `draft` ou `published`, publicação de rascunho e edição/exclusão própria; `hidden/removed` são decisões de moderação. O autor não pode republicar conteúdo removido editando o status.
- Catálogo: rascunho/pendência passa por publicação ou rejeição; ocultar/arquivar retira da consulta pública sem quebrar relações históricas.
- Denúncia: abertura, atribuição/análise, solicitação de informação, resolução ou arquivamento fundamentado; cancelamento pelo denunciante somente enquanto ainda não analisada. Reabertura, se permitida, exige ação auditada.
- Punições: advertência registra aviso; restrição impede as ações configuradas; suspensão bloqueia acesso autenticado durante vigência; banimento bloqueia acesso sem prazo. Fim/revogação de uma punição não anula outras punições vigentes nem reativa conta desativada.
- Conta: inatividade de sessão não equivale a `UserStatus.inactive`. Defina quais estados autenticam; contas banidas/desativadas não obtêm sessões; suspensão é reavaliada conforme período e demais penalidades.

## 7. Regras de negócio e premissas para lacunas

As regras explicitadas pelas fontes devem ser preservadas. As escolhas marcadas como **premissa** fecham lacunas para uma primeira implementação completa e precisam constar em `Docs/decisoes.md`.

### 7.1. Leitura e avaliações

- RN04: existe um único estado por leitor/obra. `planned`, `in_progress`, `completed` e `abandoned` correspondem a quero ler, lendo, lido e abandonado; `paused` acrescenta o estado do diagrama.
- UC02/E01: diário exige conteúdo e progresso. Ao publicá-lo, obra que não está lendo nem lida passa para lendo. Se faltar item na estante, crie-o com a edição informada. Valide obra/edição, status do catálogo e progresso.
- **Premissa de rascunhos:** salvar/editar diário em `draft` não modifica estante ou progresso e não gera notificações públicas. A primeira transição para `published` aplica os efeitos de leitura uma única vez, na mesma transação. Criar diretamente como publicado tem o mesmo comportamento. Repetir publicação ou editar post já publicado não duplica eventos/páginas.
- Criação do post, subtipo, atualização/criação da estante e eventos persistidos relacionados pertencem à mesma transação. Use bloqueio ou controle de concorrência para evitar perda de progresso.
- **Premissa sobre páginas:** o diário registra página absoluta no momento da leitura; múltiplos diários com páginas 20 e 35 não representam 55 páginas. O progresso corrente fica no registro da estante. Reduções/correções devem ser explícitas; diário retroativo não reduz automaticamente o progresso atual. Editar/apagar um diário não reescreve implicitamente o progresso posterior. Documente e teste a regra adotada.
- **Premissa para histórico mensal/anual:** cada avanço real cria evento transacional com `credited_pages = max(new_page - previous_page, 0)`; repetir a mesma página não gera crédito adicional. Correções explícitas, troca de edição e mudanças meramente administrativas geram crédito zero. Uma redução exige `change_kind=correction`; não simule releitura zerando a página. A conclusão automática credita somente as páginas restantes quando declarada como avanço real. Eventos usam instante do servidor por padrão; diários retrospectivos preservam seu snapshot sem inventar páginas no mês corrente. Releituras múltiplas independentes não estão modeladas e não devem ser anunciadas como recurso pronto.
- RN03: somente o autor edita/exclui seus diários e avaliações, ressalvadas ações de moderação auditadas.
- RN01: notas inteiras entre 1 e 5. **Premissa:** uma avaliação por leitor/obra; pode ser publicada enquanto o item está `in_progress` ou `completed`, conciliando UC03 com a descrição de avaliação final. Duplicata retorna `409` e indica a possibilidade de editar a existente.
- **Premissa para recriação:** exclusão voluntária da avaliação permite criar outra, com novo post, sem ressuscitar curtidas/comentários antigos. Preserve histórico e imponha unicidade entre avaliações vigentes com chave de slot ativo/chave gerada compatível com MariaDB ou estratégia transacional equivalente. Um `UNIQUE(reader_id, work_id, deleted_at)` com `NULL` não garante por si só a unicidade desejada. Remoção administrativa não libera um caminho para contornar a moderação excluindo/recriando; sua liberação exige revisão autorizada.
- Média e quantidade de avaliações consideram apenas avaliações publicadas e elegíveis pela política de conta/conteúdo. Preferencialmente calcule `AVG`/`COUNT` na consulta para evitar contador inconsistente; se materializar, atualize na mesma transação, inclusive em moderação e exclusões. Sem avaliações, média é `null`, quantidade é zero.
- Concluir leitura preenche datas coerentes. Corrigir edição/progresso não pode tornar diários históricos inválidos silenciosamente; preserve a edição registrada em cada diário.

### 7.2. Spoilers, conteúdo sensível, idade e privacidade

- RN02: criação de publicação exige informar explicitamente `contains_spoiler: true/false`. Use DTO que distinga ausência de `false`. Diários e avaliações aceitam também indicação explícita de conteúdo sensível; comentários recebem política coerente com a publicação e não podem expor trecho protegido em notificações.
- RF23/RF24: por padrão, devolva metadados e indicação de conteúdo oculto, sem o corpo/trechos protegidos. Uma leitura explícita com `reveal_spoilers=true` e/ou `reveal_sensitive=true` pode revelar o conteúdo após confirmação pelo consumidor. Não copie corpo protegido para busca, prévias, comentários embutidos ou notificações sem aplicar a mesma política.
- NR08, conforme nome usado na especificação: a API deve impedir acesso a diários/avaliações acima da idade do leitor em feed, detalhes por ID, busca, comentários, curtidas e demais caminhos derivados. Confirmação de spoiler não contorna idade.
- **Premissa etária:** classifique conteúdo com idade mínima em `0, 10, 12, 14, 16, 18`; obra publicada exige classificação administrativa. Classificação efetiva do post é `max(classificação atual da obra, classificação própria do post)`, inclusive quando a classificação da obra for aumentada depois da publicação. Visitante e leitor sem idade verificável acessam apenas classificação 0; calcule idade pela data civil de nascimento, com regra de aniversário testada. Não presuma verificação documental de idade: registre que o dado é autodeclarado.
- **Premissa de visibilidade:** perfil `public` pode ser consultado por visitantes; `registered_users_only` exige sessão; `private` é visível apenas ao titular e à administração autorizada. Aplique o mesmo limite a estante e estatísticas públicas do perfil. Somente o titular consulta dados pessoais, interesses privados e seu histórico completo. Publicações marcadas públicas continuam sujeitas às regras de idade/conteúdo; documente separadamente visibilidade do perfil e dos posts.
- Conteúdo removido/deletado não retorna corpo nem fica acessível por IDs de comentários ou curtidas. Use `404` quando necessário para não revelar a existência de recurso invisível; use `403` para operação explicitamente proibida sobre recurso que o solicitante pode conhecer.
- Inspeção de evidências/conteúdo pelas rotas administrativas exige permissão de moderação e auditoria. Ter um papel administrativo não desativa automaticamente filtros de privacidade/idade nas rotas públicas.

### 7.3. Exclusões e retenção

- **Premissa:** excluir publicação remove-a logicamente das consultas e agregações, retira seu corpo das respostas e preserva referências necessárias para moderação. Comentários com respostas mantêm marcador sem texto; curtidas de conteúdo indisponível não são expostas.
- **Premissa:** excluir conta requer confirmação de senha, revoga todas as sessões e tokens, remove dados privados/estante/interesses e anonimiza identificadores pessoais; publicações e comentários deixam de ser exibidos. Preserve somente a trilha mínima de denúncia/auditoria necessária, com acesso restrito e política de retenção documentada. Não prometa conformidade legal automática nem retenção indefinida como requisito da especificação.
- Defina no README o que é anonimizado imediatamente e como conteúdo/evidências preservados são eliminados no prazo configurado. Caso o ambiente exija uma política jurídica específica, sinalize esse ponto e mantenha a implementação técnica configurável.
- **Premissa sobre estante:** remoção elimina o registro corrente e seus efeitos nas estatísticas derivadas, inclusive seus eventos de progresso; os diários publicados continuam com autor e edição. Reinserir a obra cria nova relação corrente, sem inferir uma releitura histórica inexistente. Explique esse efeito no contrato para que o consumidor possa avisar antes da remoção.
- Não use `ON DELETE CASCADE` indiscriminadamente. Documente a ação de cada FK e valide exclusão de conta, catálogo, conteúdo denunciado e comentários com respostas.

### 7.4. Administração, sugestões e moderação

- RN07: nenhum administrador pode banir outro administrador, inclusive por ação originada em denúncia ou API indireta. Cadastro público nunca concede privilégios.
- Leitores encaminham inclusão/alteração/exclusão de catálogo por sugestões; só a aprovação administrativa altera o catálogo. Uma solicitação de excluir obra é diferente de remover da estante.
- Aprovação da sugestão, alteração do catálogo, decisão e auditoria são atômicas. Detecte conflito se a obra/edição mudou desde a proposta; não sobrescreva alterações recentes sem revisão.
- Moderação valida alvo, motivo, permissão e transição. Conteúdo escondido/removido sai do feed e dos acessos derivados. Notifique interessados sem divulgar denunciante ou evidência privada.
- Em `awaiting_information`, o denunciante pode responder à solicitação pela própria denúncia; mantenha histórico de mensagens e devolva à análise conforme transição documentada. Notas internas administrativas continuam invisíveis ao denunciante.
- **Premissa de filtragem automática:** implemente verificação determinística configurável de spam, termos bloqueados e links de domínios bloqueados nos fluxos previstos pelos casos de uso. Valide URLs e rejeite esquemas indevidos; não execute/busque URLs enviadas por leitores. Documente limites desse filtro e mantenha denúncia/análise humana; não o apresente como detecção infalível de linguagem ofensiva ou malware.
- Auditoria administrativa é persistida e não editável por leitores. Remova senhas, cookies, tokens e dados excessivos dos registros.

### 7.5. Notificações, estatísticas, recomendações e suporte

- Persista notificações para comentários, respostas e curtidas conforme os casos de uso; evite notificar o próprio ator. Repetir um `PUT like` não duplica evento. Não devolva textos protegidos por spoiler, faixa etária ou moderação nas notificações.
- RN05: estatísticas pessoais consultam exclusivamente dados do próprio leitor. Filtros de período, quando oferecidos, têm limites inclusivos/exclusivos e fuso documentados; não use contagens globais como pessoais.
- RF17 exige **obras e páginas lidas por mês e por ano**, não apenas totais gerais. **Premissa das fórmulas:** obras por período são registros atualmente concluídos com `finished_at` no intervalo; páginas por período são a soma de `credited_pages` dos eventos de avanço no intervalo. Use intervalos `[início, fim)` em UTC nesta primeira versão, com mês/ano explícitos no contrato, e retorne períodos sem leitura como zero. Corrigir datas/status pode alterar a contagem de conclusões; correção de página não gera novo crédito de leitura.
- Inclua também contagem por status; total de páginas correntes como métrica separada de páginas lidas por eventos; média de notas do leitor; gêneros mais lidos; contagem de diários; e **idioma mais lido**, definido como idioma da edição com maior quantidade de obras atualmente concluídas, desempate por código de idioma. Sem leitura concluída, idioma é `null`.
- UC04/RF15 exigem listas de completos (`completed`), incompletos (`in_progress` e `paused`), abandonados (`abandoned`), desejados (`planned`), preferidos (relação explícita) e recomendados (algoritmo). Não confunda preferência pessoal com recomendação do algoritmo. Ofereça filtros/listagem paginada para cada grupo, além dos indicadores numéricos.
- **Premissa para recomendações:** pontue cada obra publicada pelo número de gêneros compartilhados com interesses explícitos e gêneros de obras avaliadas com nota 4 ou 5; exclua toda obra já na estante e conteúdo inelegível. Ordene por afinidade decrescente, média decrescente, quantidade de avaliações decrescente e ID crescente. Sem histórico/interesses, use o ranking global elegível. Documente pesos, desempates e testes. Não use mocks, aleatoriedade não controlada ou serviço de IA externo como algoritmo obrigatório.
- **Premissa para rankings:** popularidade é quantidade de leitores distintos com a obra na estante; melhores avaliadas usam média das avaliações publicadas, quantidade como desempate e ID final. Exponha quantidade/amostra e documente um mínimo configurável para entrar no ranking de notas. Rankings e recomendações também respeitam status e classificação do catálogo.
- FAQ: leitura pública de entradas publicadas; manutenção administrativa. Suporte: abertura pelo leitor autenticado, consulta e mensagens apenas pelo solicitante e atendente autorizado; estados `open`, `in_progress`, `resolved`, `closed`, com decisão documentada para reabertura. Não acrescente chat em tempo real como substituto do fluxo persistido.

## 8. Autenticação, autorização e proteção HTTP

Use sessão opaca controlada pelo servidor, persistida no MariaDB; não troque a proposta da arquitetura por JWT sem registrar e justificar uma mudança explícita.

- Senhas são armazenadas com algoritmo próprio para senhas e parâmetros documentados; use biblioteca mantida e compare de maneira adequada. Nunca grave/devolva senha em texto, nem use hash rápido simples como senha.
- Cookie de sessão `HttpOnly`, `Secure` em produção, `SameSite` definido pelo ambiente, `Path=/` e domínio tão restrito quanto possível. Armazene apenas hash do segredo de sessão no banco; gere segredo com fonte criptográfica.
- Implemente expiração absoluta e por inatividade configuráveis, rotação de sessão no login/elevação de privilégio e revogação no logout, troca/recuperação de senha, suspensão e banimento. Sessão expirada/revogada deve falhar mesmo com cookie ainda presente.
- **Premissa de defaults locais:** inatividade de 30 minutos e duração absoluta de 7 dias, ajustáveis por configuração. Não renove uma sessão além do limite absoluto.
- Proteja escritas autenticadas por cookie com token CSRF ligado à sessão, validado em header, e checagem de origem conforme implantação. Defina como o frontend obtém o token e como ele muda na rotação de sessão. Documente proteção dos fluxos pré-autenticação, inclusive login.
- CORS usa lista explícita de origens; com credenciais, não use origem `*`. Trate preflight e headers CSRF. Configure proxies confiáveis de forma explícita.
- Login, cadastro, recuperação e tentativas inválidas têm limites configuráveis. Mensagens de login/recuperação não permitem descobrir se um e-mail está cadastrado. Use `429` quando o limite for atingido.
- Recuperação de senha envia token de uso único por adaptador de e-mail, com prazo configurável e consumo atômico. Ambiente de desenvolvimento pode usar capturador local de e-mail; produção exige configuração real e não devolve token no JSON nem nos logs comuns. Verificação de e-mail, se ativada para `pending_verification`, deve ter fluxo completo; não crie contas presas num estado sem saída.
- Extraia identidade e perfis da sessão validada. Reavalie estado/punições vigentes a cada operação protegida; não confie em papel enviado no corpo ou em cookie modificável.
- Limite tamanho de corpo e tempo de requisição; valide IDs, paginação, URLs, enums e campos desconhecidos nos DTOs de escrita conforme política documentada. URLs de imagem/anexo/evidência são opcionais, com esquemas permitidos e sem busca arbitrária pelo servidor; os atributos `dataURL`/`reportDataURL` do diagrama não exigem sistema de upload nem conteúdo base64 em JSON. Nunca permita mass assignment de papéis, hashes, autor ou moderação.
- Use logs estruturados com identificador da requisição, duração e status; recupere pânicos sem expor stack trace ao cliente. Não registre corpos sensíveis.

Matriz mínima de permissões, como premissa a documentar:

| Perfil | Permissões |
| --- | --- |
| Visitante | Saúde da API, cadastro/login/recuperação, FAQ e catálogo/perfis/publicações elegíveis às restrições de acesso. |
| Leitor ativo | Própria conta/estante, publicações próprias, interações permitidas, denúncias, sugestões, interesses, notificações, estatísticas e suporte próprios. |
| `catalog_manager` | Manutenção de catálogo, autores/editoras/gêneros, análise de sugestões e auditoria correspondente. |
| `user_manager` | Atendimento, gestão da FAQ, análise de denúncias, moderação e punições de leitores; sem gerenciar administradores. |
| `administrator` | União das permissões de catálogo e moderação/atendimento; sem conceder papéis administrativos. |
| `system_administrator` | Permissões anteriores e CRUD de administradores; mantém a proibição de banir administrador. |

Cadastre o primeiro `system_administrator` por comando administrativo autenticado pelas credenciais do ambiente, sem rota pública de promoção e sem senha fixa no seed. Não permita remover/rebaixar o último administrador de sistema ativo. Toda alteração de papel/status administrativo exige auditoria e invalidação de sessões quando necessário.

## 9. Contrato HTTP completo

Crie `Docs/api.yaml` em OpenAPI com todas as rotas efetivamente implementadas, schemas, campos obrigatórios/anuláveis, enums, autenticação por cookie, CSRF, exemplos, filtros, paginação e erros. O contrato deve corresponder ao comportamento real, inclusive restrições de visibilidade.

Convenções:

- Prefixo `/api/v1`, JSON `snake_case`, IDs consistentes, timestamps ISO 8601 em UTC e datas civis `YYYY-MM-DD`.
- Listagem: `{ "data": [], "pagination": { "page": 1, "page_size": 20, "total": 0 } }`; `page >= 1`, `1 <= page_size <= 100`, ordenação estável com ID como desempate.
- Recurso individual: envelope `{ "data": { ... } }`; `204` não tem corpo. Defina nulabilidade e representação dos subtipos `Post` sem formatos variáveis não documentados.
- `200` para consulta/atualização com corpo, `201` para criação, `204` para sucesso sem corpo, `400` para entrada inválida, `401` para ausência/sessão inválida, `403` para ação proibida, `404` para inexistente/invisível, `409` para conflito, `429` para limite e `500` para erro interno. Falhas de dependência temporária podem usar `503`, documentado.
- Validação aponta campos e códigos estáveis. Não devolva SQL, DSN, stack trace ou erro bruto do driver.

```json
{
  "error": {
    "code": "INVALID_READING_PROGRESS",
    "message": "O progresso não pode ultrapassar o total de páginas da edição.",
    "fields": {
      "current_page": "Informe uma página válida para esta edição."
    }
  }
}
```

### 9.1. Rotas de conta e perfil

Os caminhos desta e das próximas tabelas são relativos a `/api/v1`, exceto saúde. `:id` representa parâmetro de rota; no OpenAPI use `{id}`.

| Método | Caminho | Comportamento |
| --- | --- | --- |
| POST | `/readers` | Cadastro de leitor com nome, apelido, e-mail, senha e nascimento; nenhum papel administrativo. |
| POST | `/auth/login` | Valida credenciais/estado, cria/rotaciona sessão e cookie. |
| POST | `/auth/logout` | Revoga sessão atual e limpa cookie; repetição não restaura sessão. |
| GET | `/auth/csrf` | Fornece token segundo o protocolo definido, inclusive fluxo anterior ao login. |
| POST | `/auth/forgot-password` | Solicita recuperação com resposta genérica. |
| POST | `/auth/reset-password` | Consome token e troca senha, revogando sessões. |
| POST | `/auth/verify-email` | Confirma token, se a verificação for habilitada. |
| POST | `/auth/resend-verification` | Reenvio limitado e resposta sem enumeração de conta, se habilitado. |
| GET/PATCH/DELETE | `/me` | Consulta/edição/exclusão da própria conta segundo a política definida. |
| PUT | `/me/password` | Troca autenticada com senha atual e revogação/rotação documentadas. |
| GET | `/readers/:id` | Perfil público filtrado pela visibilidade; não é cópia de `/me`. |
| GET | `/readers/:id/shelf` | Estante permitida pela visibilidade do perfil. |
| GET | `/readers/:id/posts` | Diários e avaliações públicos/elegíveis do leitor, filtráveis por tipo. |
| GET/PUT | `/me/interests` | Consulta/substituição dos gêneros de interesse do titular. |
| GET | `/me/preferred-works` | Lista paginada de obras preferidas pelo titular. |
| PUT/DELETE | `/me/preferred-works/:work_id` | Marca/desmarca preferência de forma idempotente. |

`PATCH /me` permite definir/remover `favorite_work_id`, validando obra existente e elegível. Preferências e livro favorito não alteram o estado de leitura nem concedem acesso a conteúdo restrito.

### 9.2. Catálogo e ISBN

| Método | Caminho | Comportamento |
| --- | --- | --- |
| GET | `/works` | Busca por título/autor/ISBN/gênero/idioma, com filtros e ordenação permitidos. |
| GET | `/works/:id` | Obra, autores, gêneros, média e quantidade elegíveis, sem progresso de outro leitor. |
| GET | `/works/:id/editions` | Edições publicadas e paginadas da obra. |
| GET | `/editions/:id` | Detalhe da edição e relação com a obra. |
| GET | `/authors`, `/authors/:id` | Consulta de autores. |
| GET | `/publishers`, `/publishers/:id` | Consulta de editoras. |
| GET | `/genres` | Gêneros disponíveis. |
| GET | `/catalog/isbn/:isbn` | Consulta edição local pelo ISBN normalizado; importação externa é extensão opcional. |
| POST | `/catalog-suggestions` | Leitor sugere inclusão, alteração ou exclusão de obra ou edição com dados validados. |
| GET | `/me/catalog-suggestions`, `/me/catalog-suggestions/:id` | Acompanha somente suas sugestões e decisões permitidas. |

Cadastro por ISBN precisa funcionar tanto em sugestão de leitor quanto no CRUD administrativo: receba ISBN válido e os dados bibliográficos necessários, verifique duplicatas e vincule à edição. RF05/RF29 não exigem preencher metadados por serviço externo. Uma consulta local sem resultado não impede propor cadastro manual com esse ISBN.

**Extensão opcional:** se implementar importação bibliográfica, use adaptador de provedor com URL base controlada pela configuração, timeout, limites e tradução para DTO de prévia. Não aceite URL arbitrária como destino de consulta. Falta de resultado, dados incompletos e indisponibilidade têm respostas documentadas; cadastro manual continua funcional. Resultado externo não define classificação etária e passa pela triagem. Testes usam servidor simulado; indisponibilidade desse serviço não bloqueia a conclusão do escopo obrigatório.

### 9.3. Estante, diários e avaliações

| Método | Caminho | Comportamento |
| --- | --- | --- |
| GET | `/me/shelf` | Estante do titular, com filtro por estado e paginação. |
| PUT/DELETE | `/me/shelf/:work_id` | Criação/atualização idempotente ou remoção da relação, conforme regras de edição/progresso. |
| GET | `/me/reading-entries` | Diários do titular, incluindo rascunhos permitidos. |
| POST | `/reading-entries` | Cria diário/post; se publicado, atualiza estante/progresso na mesma transação. |
| GET/PATCH/DELETE | `/reading-entries/:id` | Consulta autorizada e manutenção pelo autor. |
| GET | `/me/reviews` | Avaliações do titular. |
| POST | `/reviews` | Cria avaliação única por leitor/obra e nota válida. |
| GET/PATCH/DELETE | `/reviews/:id` | Consulta autorizada e manutenção pelo autor; média consistente. |
| GET | `/works/:id/reviews` | Avaliações publicadas elegíveis e paginadas. |

Documente a relação entre ID de diário/avaliação e `post_id`; prefira PK compartilhada para permitir navegação inequívoca do feed ao subtipo. Um PATCH deve distinguir campo ausente, zero, `false` e `null` onde aplicável.

### 9.4. Comunidade, denúncias e notificações

| Método | Caminho | Comportamento |
| --- | --- | --- |
| GET | `/posts`, `/posts/:id` | Feed/detalhe de publicações elegíveis, com tipo, obra e filtros documentados. |
| GET/POST | `/posts/:id/comments` | Lista/cria comentários; criação pode indicar pai da mesma publicação. |
| PATCH/DELETE | `/comments/:id` | Edição/exclusão própria, com preservação da árvore. |
| PUT/DELETE | `/posts/:id/like` | Curtir/descurtir idempotentemente. |
| PUT/DELETE | `/comments/:id/like` | Curtir/descurtir comentário idempotentemente. |
| POST | `/reports` | Abre denúncia com exatamente um alvo válido. |
| GET | `/me/reports`, `/me/reports/:id` | Acompanha suas denúncias com campos internos omitidos. |
| POST | `/me/reports/:id/cancel` | Cancela denúncia própria quando a transição permitir. |
| GET/POST | `/me/reports/:id/messages` | Consulta mensagens visíveis e complementa a própria denúncia quando solicitado. |
| GET | `/me/notifications` | Notificações próprias, paginadas e filtráveis por leitura. |
| PUT | `/me/notifications/:id/read` | Marca leitura de forma idempotente; somente destinatário. |

Listagens de comentários não podem contornar invisibilidade do post. O feed retorna contadores e estado da curtida do solicitante sem uma consulta adicional por item. Defina como respostas de comentários são paginadas, sem devolver árvore ilimitada.

### 9.5. Estatísticas, recomendações e suporte

| Método | Caminho | Comportamento |
| --- | --- | --- |
| GET | `/me/statistics` | Indicadores calculados do titular, com fórmulas e período documentados. |
| GET | `/me/reading-report` | Lista paginada por grupo `completed/incomplete/abandoned/desired/preferred/recommended`, cobrindo UC04 e RF15. |
| GET | `/me/recommendations` | Obras recomendadas por regra determinística e com motivo resumido. |
| GET | `/rankings/popular`, `/rankings/top-rated` | Rankings de popularidade e notas com política explícita. |
| GET | `/faq`, `/faq/:id` | Perguntas/respostas publicadas. |
| POST | `/support-tickets` | Abre chamado autenticado. |
| GET | `/me/support-tickets` | Lista chamados do titular. |
| GET | `/support-tickets/:id` | Detalhes apenas ao solicitante ou atendente permitido. |
| GET/POST | `/support-tickets/:id/messages` | Conversa persistida, paginada e autorizada. |
| POST | `/support-tickets/:id/close` | Encerramento permitido pela política de atendimento. |

### 9.6. Administração

Todas as rotas abaixo exigem papel adequado; prefixo `/admin` não substitui autorização no serviço.

| Método | Caminho | Comportamento |
| --- | --- | --- |
| GET/POST | `/admin/administrators` | Listagem e criação controlada de administradores. |
| GET/PATCH/DELETE | `/admin/administrators/:id` | Consulta, papel/status e revogação administrativa; protege último administrador de sistema. |
| GET | `/admin/users`, `/admin/users/:id` | Consulta administrativa com minimização de dados. |
| POST/PATCH/DELETE | `/admin/works[/:id]` | Criação, alteração e arquivamento de obra, conforme método/caminho concreto no OpenAPI. |
| POST/PATCH/DELETE | `/admin/editions[/:id]` | Manutenção de edição; vínculo obra/editora e ISBN validados. |
| POST/PATCH/DELETE | `/admin/authors[/:id]`, `/admin/publishers[/:id]`, `/admin/genres[/:id]` | Manutenção de entidades do catálogo sem romper FKs. |
| GET | `/admin/catalog-suggestions`, `/admin/catalog-suggestions/:id` | Fila e detalhe das sugestões. |
| POST | `/admin/catalog-suggestions/:id/approve`, `/admin/catalog-suggestions/:id/reject` | Decisão fundamentada, sem aplicar duas vezes. |
| GET | `/admin/reports`, `/admin/reports/:id` | Fila, filtros por estado e evidências autorizadas. |
| PATCH | `/admin/reports/:id` | Atribuição e transições válidas, com resolução quando exigida. |
| GET/POST | `/admin/reports/:id/messages` | Histórico/solicitação de informações com visibilidade explícita. |
| POST | `/admin/posts/:id/moderation`, `/admin/comments/:id/moderation` | Ocultação/remoção/restauração permitida e auditada. |
| GET/POST | `/admin/penalties` | Consulta/aplicação de punição a alvo permitido, com vínculo opcional a denúncias. |
| POST | `/admin/penalties/:id/revoke` | Revoga com motivo e recalcula efeitos das punições restantes. |
| GET | `/admin/audit-logs` | Consulta de trilha conforme permissão. |
| GET/POST | `/admin/faq` | Administração da FAQ. |
| GET/PATCH/DELETE | `/admin/faq/:id` | Manutenção/publicação de entrada. |
| GET | `/admin/support-tickets`, `/admin/support-tickets/:id` | Atendimento e filtros. |
| PATCH | `/admin/support-tickets/:id` | Atribuição e mudança válida de estado. |
| POST | `/admin/support-tickets/:id/messages` | Resposta administrativa com autoria auditada. |

A notação `[/:id]` acima é apenas resumo: implemente `POST` na coleção e `PATCH/DELETE` no item e descreva cada operação separadamente no OpenAPI. Acrescente consultas administrativas de rascunhos/ocultos quando necessárias ao CRUD e à revisão; não exponha esses estados pelas rotas públicas.

### 9.7. Saúde

- Preserve `GET /health` como vivacidade, sem depender de consulta demorada ao banco.
- Adicione `GET /ready` como prontidão: verifica banco e estado necessário à operação; retorna `503` quando indisponível, sem expor credenciais/detalhes internos.

## 10. MariaDB, migrações e consistência

- Use `gorm.io/driver/mysql` com MariaDB real e SQL versionado em `Backend/migrations` como referência do esquema. Escolha uma ferramenta de migração mantida e documente comandos exatos de aplicar, consultar versão e reverter quando suportado.
- Não use `AutoMigrate` como mecanismo definitivo. Migração aplicada em ambiente compartilhado é corrigida por nova migração; não reescreva histórico.
- Use `utf8mb4`, política de collation explícita, datas/horas em UTC, índices para filtros/joins e tipos consistentes entre Go e SQL. Defina IDs, campos opcionais e precisão temporal.
- Inclua PKs, FKs, `NOT NULL`, `UNIQUE` e `CHECK` nas regras que o banco consegue expressar. Confira suporte real na versão MariaDB escolhida; teste restrições efetivamente aplicadas, não apenas a sintaxe gerada.
- Planeje índices para feed por estado/data/ID, comentários por post/pai, estante por leitor/estado, ISBN, relações autor/gênero, sessões por hash/expiração, notificações por destinatário/leitura e filas administrativas por estado/data.
- Use consultas agregadas e preloads controlados para evitar N+1. Nunca carregue tabela inteira para paginar ou filtrar em Go.
- Teste concorrência em estante, avaliação única, curtidas, consumo de tokens e aprovação de sugestão; conflitos não geram duplicata nem `500` indevido.
- Transações cobrem operações compostas: diário/estante/post; subtipo/post; aprovação/catálogo/auditoria; moderação/estado/eventos; recuperação de senha/revogação de sessões.
- Chamadas de e-mail/ISBN não seguram transação SQL durante rede. Para eventos que exigem entrega externa após commit, use outbox persistida ou estratégia equivalente documentada. Notificações internas podem ser gravadas na própria transação.
- Configure pool, tempos de conexão/consulta e retentativas limitadas quando apropriadas. Não repita operação com efeito sem tratar idempotência.
- Separe usuário de execução, credencial de migração e banco de testes. A API não usa root e não aplica alterações destrutivas automaticamente no startup.
- Inclua seed opcional e idempotente de dados fictícios para demonstração, sem contas reais ou senha administrativa fixa. Aplicar migrações não deve inserir conteúdo de demonstração em produção.

## 11. Configuração e execução reproduzíveis

### MariaDB já instalado no ambiente local

Configuração informada pelo responsável pelo projeto:

| Parâmetro | Valor |
| --- | --- |
| Usuário administrativo | `root` |
| Senha do `root` | `123` |
| Porta TCP | `3306` |
| Nome do serviço no Windows (service name) | `MariaDB` |

Considere essa instalação local ao documentar a conexão pelo DBeaver, a preparação do banco e a execução das migrações. `MariaDB` é o nome do serviço do Windows; o nome do banco da aplicação continua sendo `radar_literario`. Use o acesso administrativo para criar o banco e os usuários com as permissões previstas neste documento; a conexão da API utiliza o usuário próprio `radar_app`.

### Variáveis de ambiente da API

Crie `Backend/.env.example`, sem segredos reais, cobrindo pelo menos:

```dotenv
APP_ENV=development
HTTP_HOST=127.0.0.1
HTTP_PORT=8080
DB_HOST=127.0.0.1
DB_PORT=3306
DB_NAME=radar_literario
DB_USER=radar_app
DB_PASSWORD=defina_localmente
CORS_ALLOWED_ORIGINS=http://localhost:5173
SESSION_COOKIE_NAME=radar_session
SESSION_IDLE_TIMEOUT=30m
SESSION_ABSOLUTE_TIMEOUT=168h
COOKIE_SECURE=false
COOKIE_SAME_SITE=lax
CSRF_ALLOWED_ORIGINS=http://localhost:5173
SMTP_HOST=localhost
SMTP_PORT=1025
SMTP_FROM=nao-responda@example.test
PUBLIC_FRONTEND_URL=http://localhost:5173
```

Acrescente variáveis efetivamente usadas para TLS/SMTP, ISBN, pool, timeouts, limites, retenção e testes. Não mantenha variáveis sem efeito. Documente unidades, obrigatoriedade, defaults e diferenças de produção. Go não lê `.env` automaticamente: implemente carregamento explícito apenas onde desejado ou documente a exportação de variáveis.

Valide configuração ao iniciar e falhe com mensagem clara se faltar algo obrigatório. Produção não pode iniciar silenciosamente com cookie inseguro, senha de demonstração ou envio de recuperação sem configuração funcional. Construa DSN com API de configuração do driver para tratar caracteres especiais, sem imprimir senha.

Forneça um caminho principal reproduzível com `compose.yaml` para MariaDB e, se necessário, capturador de e-mail de desenvolvimento. Use volume persistente, healthcheck, versão explícita verificada e credenciais externas. Serviço de testes deve usar banco/volume isolados. DBeaver é cliente de administração opcional, não dependência de execução da API.

Atualize `flake.nix` para refletir Go/MariaDB e ferramentas realmente necessárias, retirando dependências Python obsoletas quando confirmado que não são usadas. Preserve o ambiente frontend. A execução em Windows não deve depender de Nix; documente PowerShell e, quando houver diferenças, Linux.

Atualize `.gitignore` para segredos `.env`, binários Go, cobertura, logs, volumes/dumps locais e dependências geradas; mantenha `.env.example`, migrações, `go.mod` e `go.sum` versionados. Não ignore arquivos de documentação ou fontes úteis.

O README deve apresentar, na ordem, pré-requisitos/versões, configuração, subida do banco, migrações, criação do administrador, seed opcional, execução da API e testes. Inclua o comando real da ferramenta de migração e exemplos de sessão/cookie/CSRF, não pseudocomandos.

Documente backup periódico com frequência/retenção configuráveis e procedimento de restauração. Forneça rotina ou instruções executáveis, registre como agendar no ambiente escolhido e teste restauração em banco descartável; não declare backup operacional apenas por existir um script não agendado.

## 12. Testes e evidências de aceite

Implemente testes que validem regras e falhas reais; não gere testes que apenas reproduzem getters ou a estrutura do código. Use fakes de interfaces nos testes de serviços, `httptest` nas rotas e MariaDB isolado nos testes de persistência/integração. SQLite não substitui a verificação de FKs, unicidade, SQL e transações MariaDB.

Cobertura mínima de cenários:

1. Cadastro válido/inválido, e-mail/apelido duplicados e concorrentes, senha protegida e ausência de dados sensíveis no JSON.
2. Login/logout, expiração por inatividade/absoluta, revogação, cookie, CSRF ausente/inválido e CORS permitido/negado.
3. Recuperação: resposta genérica, token expirado, consumo único concorrente e revogação de sessões.
4. Duas contas: uma não edita perfil, estante, post, comentário, chamado ou notificação da outra. Leitor não promove conta nem define autoria/status administrativo pelo corpo.
5. Catálogo: obra/edição separadas, autores N:M, ISBN válido/inválido/equivalente duplicado, filtros/paginação e exclusão pública de rascunhos/ocultos.
6. Estante: repetição idempotente, unicidade leitor/obra, edição de outra obra rejeitada, limites de página e transições.
7. Diário: exige conteúdo/progresso/spoiler explícito, rascunho não altera estante, publicação aplica efeitos uma vez, muda estado conforme UC02, preserva histórico e reverte toda a transação se um passo falhar.
8. Avaliação: escala/unicidade/estado permitido, recriação após exclusão própria, impedimento de contorno da moderação, concorrência e média correta após criação, edição, exclusão e moderação.
9. Feed: autoria/tipo corretos, ordem estável, paginação, proteção de spoiler/sensível e idade em todos os caminhos, inclusive após aumento de classificação da obra, ID direto, comentários e notificações.
10. Comentários: pai de outro post, profundidade, ciclos, edição própria e tombstone com respostas. Curtidas concorrentes não duplicam nem geram notificações repetidas.
11. Denúncia: zero/dois/três alvos rejeitados, FK inválida, privacidade do denunciante, complementação de informações e transições/atribuições autorizadas.
12. Punição: período, expiração/revogação, penalidades sobrepostas, bloqueio de sessão e proibição de administrador banir administrador.
13. Sugestão: leitor não altera catálogo diretamente; aprovação é atômica, repetição não reaplica e alteração concorrente do alvo gera conflito.
14. Papéis: cada perfil administrativo só executa sua matriz; CRUD de administradores protege o último administrador de sistema.
15. Estatísticas com dados de duas contas, meses/anos e viradas de período, meses zerados, snapshots sem soma duplicada, correções sem crédito artificial, idioma mais lido, listas UC04, obras preferidas/favorito, recomendações determinísticas e rankings com desempate/amostra testados.
16. FAQ e suporte persistem, respeitam publicação/propriedade/papel e não expõem mensagens de outro chamado.
17. Exclusão de conta/conteúdo denunciado preserva integridade e política de anonimização; referências não vazam textos removidos.
18. Migrações constroem banco vazio, correspondem aos modelos e rejeitam dados que violem restrições; reiniciar a API mantém os dados.
19. ISBN local/cadastro com dados completos e e-mail de recuperação: sucesso, falhas e duplicatas; quando houver provedor bibliográfico opcional, timeout/indisponibilidade/dado parcial/ausência de resultado com adaptador controlado.
20. Backup/restauração em ambiente descartável, prontidão sem banco e encerramento gracioso.

Execute a partir de `Backend`, com a toolchain compatível disponível:

```sh
go mod tidy
go mod verify
go fmt ./...
go vet ./...
go test ./...
go build ./cmd/api
go build ./cmd/admin
```

Se testes de integração dependerem de tag/configuração, forneça comando explícito e execute-os separadamente contra o MariaDB de testes. Acrescente teste de corrida onde suportado e relevante à concorrência implementada. Valide sintaxe/referências do OpenAPI e correspondência entre contrato e rotas.

Se faltar toolchain, banco ou acesso externo, registre o impedimento, os comandos tentados e o que continua sem validação; prossiga nas partes independentes. Não invente resultados de testes, migrações, e-mail, ISBN ou backup. Compilar não prova regras de negócio.

## 13. Sequência de implementação

1. **Reconciliação:** inspecione estado e alterações locais, leia fontes, produza decisões e matriz inicial de cobertura; alinhe a estrutura escolhida em `Arquitetura_API.md` sem apagar regras.
2. **Fundação:** configuração, conexão, pool, tratamento de erros, middlewares, composição, MariaDB local, migrações, saúde/prontidão e OpenAPI inicial.
3. **Contas:** esquema, cadastro, sessão/CSRF, recuperação, perfil, interesses, permissões e comando de administrador.
4. **Catálogo:** obras/edições/autores/editoras/gêneros, ISBN, pesquisa, CRUD administrativo e seeds fictícios.
5. **Estante/leitura:** relação única por obra, progresso/edição, diários e avaliações com transações.
6. **Comunidade:** feed comum, comentários/respostas, curtidas, filtros de acesso e notificações persistidas.
7. **Administração:** denúncias, moderação, punições, sugestões, auditoria e gestão de administradores.
8. **Complementos:** estatísticas, recomendações/rankings, FAQ e suporte completos.
9. **Consolidação:** exclusões/retenção, testes de concorrência/transação, instalação em banco vazio, backup/restauração e revisão final de contrato/documentação.

Em cada etapa, implemente o fluxo inteiro de migração/entidade → repositório → serviço → DTO/controller/rota → teste/contrato. As etapas são uma ordem de trabalho, não autorização para omitir as últimas funcionalidades.

## 14. Critérios de conclusão e resposta final

Só considere o backend concluído quando:

- Todos os módulos e os requisitos de backend das fontes estiverem cobertos e rastreados, com premissas e eventuais desvios explicitados.
- Não houver handler de demonstração apresentado como persistência, regra de negócio esquecida no controller ou autorização delegada ao frontend.
- Cadastro → login → catálogo → estante → diário → avaliação → comentário/curtida → denúncia → decisão administrativa funcionar com dados persistidos e isolamento entre duas contas.
- Recomendações, estatísticas, ISBN, recuperação de senha, FAQ/suporte, notificações e administração possuírem implementação concreta, inclusive falhas previsíveis.
- Banco vazio puder ser preparado, API executada e administrador criado seguindo o README, sem senha fixa nem edição manual de tabela.
- Migrações, testes relevantes e verificações disponíveis tiverem sido executados; limitações externas restantes estiverem descritas com precisão.
- `Docs/api.yaml`, `Docs/decisoes.md`, `Docs/cobertura_backend.md`, `Backend/README.md`, `.env.example`, ambiente e arquitetura refletirem a implementação final.

Ao terminar, responda em português com: o que foi implementado; principais decisões; como iniciar banco/aplicar migrações/executar a API/criar administrador; verificações executadas com resultado; e bloqueios reais, caso existam. Identifique pendências como pendências, nunca como conclusão. Não apresente implementação de frontend ou publicação como parte deste resultado.
