# Cobertura do backend

Conferência em 27/10/2026: `Prompt_Implementacao_Backend.md`, PDF `RadarLiterárioV2.pdf` (incluindo UC02/UC03), `Class Diagram0.png` inspecionado visualmente, arquitetura, plano, README/serviço do protótipo, flake e estado Git. O início continha somente três rotas e handlers fictícios, sem persistência. O prompt de implementação era uma alteração local não versionada e foi preservado.

Implementação: Go 1.27.3, Gin 1.12.0, GORM 1.31.2/driver MySQL 1.6.0, driver SQL 1.10.1, Goose 3.29.0, MariaDB Windows 13.0.2. São **137 operações HTTP em 96 caminhos**, descritas em [api.yaml](api.yaml), com 76 schemas. SQL versionado é a referência; não há AutoMigrate.

Os detalhes e decisões por módulo estão em [contas](cobertura_contas.md), [catálogo](cobertura_catalogo.md), [leitura/comunidade](cobertura_leitura.md), [administração](cobertura_admin.md) e [backup](backup.md). Estes documentos distinguem testes de interfaces controladas, HTTP e persistência real. Nenhuma validação de frontend está incluída.

## Matriz de rastreabilidade

Os caminhos abaixo são relativos a `/api/v1`, exceto saúde.

| Requisito/caso de uso | Regra e serviço | Rotas | Migração/entidades | Teste ou evidência |
| --- | --- | --- | --- | --- |
| RF01, UC01 | `Accounts.Register/Me/Patch/Delete/Public`; normalização, senha protegida e anonimização | `/readers`, `/me`, `/readers/{id}` | 1: users/readers; relações 2–4 | Validações de conta, cadastro concorrente por e-mail/apelido, JSON privado/público e exclusão no fluxo HTTP real |
| RF02, RNF03 | `Accounts.Login/Authenticate/Logout`; sessão opaca, rotação, expiração absoluta/inatividade e revogação | `/auth/login`, `/auth/logout`, `/auth/csrf` | 1: sessions | `TestSessionsEnforceAbsoluteIdleRevocationAndPenalties`, HTTP com replay da sessão ao trocar de conta, cookie/CSRF/CORS |
| RF03 | Recuperação com token hash, consumo único e outbox AES-GCM; SMTP fora de transação | `/auth/forgot-password`, `/auth/reset-password` | 1: password_resets/mail_outbox | Testes MariaDB: consumo concorrente, expiração, revogação, SMTP real loopback, falha/retry e exclusão intercalada |
| RF04, UC01 | Perfil público minimizado e regras de visibilidade, diário/avaliação elegíveis | `/readers/{id}`, `/readers/{id}/posts`, `/readers/{id}/shelf` | 1–3: perfis/posts/estante | HTTP de duas contas e consultas MariaDB; não expõe nascimento/e-mail no perfil público |
| RF05, RF29, UC05, UC12 | `Catalog.SaveWork/SaveEdition/ISBN/CreateSuggestion`; ISBN canônico e obra/edição distintas | `/catalog/isbn/{isbn}`, `/catalog-suggestions`, `/admin/works`, `/admin/editions` | 2: works/editions/authors/publishers/genres, N:M | ISBN-10/13 equivalentes concorrentes, autores N:M, criação administrativa e consulta ISBN pelo HTTP real |
| RF06, RF07 | Pesquisa parametrizada, detalhe e agregação SQL de avaliações elegíveis | `/works`, `/works/{id}`, `/works/{id}/editions`, `/authors`, `/publishers`, `/genres` | 2–3: catálogo e reviews | Filtros/limites/idade, busca pelo segundo autor, média e amostra em MariaDB |
| RF08, RF09, RF10, RF12, RN01, UC03 | Review única vigente por leitor/obra, nota inteira 1–5; exclusão própria libera slot, moderação não | `/reviews`, `/reviews/{id}`, `/me/reviews`, `/works/{id}/reviews` | 3: posts/reviews, unique active_slot | Concorrência: uma avaliação; criação/edição/exclusão/moderação e média; recriação e estado de leitura |
| RF11, RF18, RF19, RF20, UC02, UC03 | Comentários/respostas, autoria, pai no mesmo post, profundidade e tombstone | `/posts/{id}/comments`, `/comments/{id}` | 3: comments, FK composta e triggers | Testes de árvore, FK/pai inválido, edição alheia, tombstone e máscara do post; snapshot antigo não desfaz moderação |
| RF13, RN02, RN03, UC02 | Diário exige conteúdo/progresso/spoiler explícito; primeira publicação atualiza leitura atomicamente | `/reading-entries`, `/reading-entries/{id}`, `/me/reading-entries` | 3: posts/reading_diaries/reading_records/events | Rascunho sem efeito, publicação única, rollback por falha de subtipo, snapshot histórico e autoria; HTTP completo |
| RF14, RF16, RN04, UC10 | Uma relação por leitor/obra, edição válida, cinco estados e progresso limitado | `/me/shelf`, `/me/shelf/{work_id}`, `/readers/{id}/shelf` | 3: reading_records, UNIQUE(reader_id,work_id), FK edição/obra | Concorrência real com 10 escritas, 1 registro/1 evento; FK/CHECK, limites e transições |
| RF15, UC04 | Listas completas/incompletas/abandonadas/desejadas/preferidas/recomendadas; favorito persistido | `/me/reading-report`, `/me/preferred-works`, `/me/recommendations`, `PATCH /me` | 1–3: favorito, preferred_works, estante | Listagem paginada, grupos, callbacks de catálogo e isolamento entre leitores |
| RF17, RN05, UC04 | Estatísticas próprias por mês/ano UTC, períodos zerados, idioma e páginas por eventos | `/me/statistics` | 3: reading_progress_events e registros atuais | Duas contas, virada UTC, snapshots/correções, ida/volta de edição sem crédito duplicado, métricas de idioma/status |
| RF21, RF22, UC02, UC03 | Curtidas idempotentes e notificações persistidas, sem notificar o próprio ator | `/posts/{id}/like`, `/comments/{id}/like`, `/me/notifications`, `/me/notifications/{id}/read` | 3: likes/notificações, chaves únicas | 12 curtidas concorrentes: 1 like/1 notificação; propriedade e repetição; HTTP comentário/curtida/notificação |
| RF23, RF24, RN02 | Metadados sem texto/URL/título protegido; confirmação explícita | `/posts`, `/posts/{id}`, diários/reviews/comentários | 3: flags e status | `reveal_spoilers`/`reveal_sensitive`, máscaras, moderação e notificações sem trechos protegidos |
| NR08 | Classificação efetiva usa máximo de obra atual/post; confirmação não contorna idade | Catálogo/feed/detalhe/comentários/likes/notificações/denúncias | 2–4: minimum_age, idade do perfil | Aniversário civil, visitante idade0 e aumento de classificação após publicar, IDs diretos e caminhos derivados |
| RF25, RF26, RF27, UC04 | Afinidade por gêneros e avaliações 4/5, exclusão da estante; rankings SQL determinísticos | `/me/recommendations`, `/rankings/popular`, `/rankings/top-rated`, `/me/interests` | 2–3: reader_genres, N:M, reviews/estante | Testes MariaDB de afinidade, nota/amostra, popularidade, desempates, rascunho e idade |
| RF28, UC09 | Papéis, concessão/revogação controladas e último sistema ativo protegido | `/admin/administrators`, `/admin/administrators/{id}`, `cmd/admin` | 1: administrators/sessions/audit_logs | Matriz de papéis, última conta de sistema, rollback de auditoria e revogação MariaDB; bootstrap real solicitado |
| RF30, RN06, RN07, UC6 | Punições temporárias/permanentes sobrepostas; administrador ativo não pode ser punido | `/admin/penalties`, `/admin/penalties/{id}/revoke` | 4: penalties/report_penalties | Períodos, CHECK/FK, expiração/revogação, sobreposição, sessões invalidadas e proibição de alvo administrativo |
| RF31, RN03, UC6 | Inspeção e moderação autorizadas/auditadas; conteúdo excluído não é restaurado | `/admin/posts/{id}`, `/admin/comments/{id}`, respectivos `/moderation` | 3–4: posts/comments/audit/notificações | Testes de evidência/autorização, exclusão intercalada com moderação e snapshots obsoletos reais |
| RF32, UC07 | FAQ pública publicada e manutenção administrativa; suporte persistido/propriedade/papéis | `/faq`, `/support-tickets`, `/me/support-tickets`, mensagens e equivalentes `/admin` | 4: faq_entries/support_tickets/support_messages | Serviços/HTTP/MariaDB de FAQ rascunho/publicação, duas contas, atribuição e transições de atendimento |
| RF33, UC08 | Sugestão não altera catálogo; aprovação/decisão/auditoria atômicas e versão do alvo | `/catalog-suggestions`, `/me/catalog-suggestions`, `/admin/catalog-suggestions/{id}/approve|reject` | 2: catalog_suggestions e version | 8 aprovações concorrentes: 1 sucesso/7 conflitos; versão antiga e rollback da auditoria |
| RF34, RF35, UC11 | Exatamente um alvo, FKs, privacidade, complementação e transições | `/reports`, `/me/reports`, `/admin/reports`, mensagens/cancelamento | 4: reports/report_messages | Zero/múltiplos alvos, FK inválida, mensagens internas e resposta solicitada, resolução e propriedade |
| RNF02 | Permissões nos serviços, identidade da sessão e separação administrativa | Todas as rotas protegidas | 1–4 | Matriz visitante/leitor/papéis e tentativas de mass assignment; duas contas no HTTP |
| RNF04 | Backup com frequência/retenção configuráveis e restauração verificada | Scripts PowerShell; agendamento documentado | Dump das migrações 1–4 | 34 tabelas, 55 FKs, 2 triggers, dados UTF-8 e restrições restaurados; **agendamento não executado** |
| RNF05 | Auditoria persistida, sanitizada e limitada por domínio do papel | `/admin/audit-logs` e operações administrativas | 1: audit_logs | Auditoria de decisões/inspeção, permissões por target_type e rollback em falha de auditoria |
| RNF06 | Proteção técnica, minimização, anonimização e retenção configurável | `/me`, `cmd/maintenance` e contratos públicos | 1–4 | Hashes/cookies/outbox, anonimização, remoção de suporte/outbox, purga real; política jurídica não certificada |
| RNF01 | Responsividade é responsabilidade da View | Contratos documentados para consumo futuro | Não se aplica | Frontend não reescrito, integrado ou validado nesta tarefa |

## Verificações de aceitação

**Resultado final:** 73 testes principais e 23 subcasos aprovados, sem testes ignorados, na execução completa com MariaDB e `-tags=integration -p=1`. `go mod tidy`, `go mod verify`, `go fmt`, `go vet`, `go test ./...`, builds de `cmd/api`/`cmd/admin` e verificação do OpenAPI também concluíram com sucesso. O contrato passou pelo validador padrão kin-openapi, além da checagem de correspondência com rotas/DTOs.

O processo HTTP foi iniciado duas vezes em `127.0.0.1:8080`: `/health` e `/ready` responderam200; a sessão do administrador id1 continuou válida após reiniciar o processo; o logout a revogou. O encerramento por Ctrl+C produziu o registro de conclusão e fechou o listener. `TestGracefulShutdownWaitsForInflightRequest` também confirmou que uma requisição em andamento termina antes do fechamento.

| Cenário do prompt | Evidência executada |
| --- | --- |
| 1–4: contas, sessão, recuperação e isolamento | Serviços com fakes, HTTP, sete testes MariaDB de contas e dois capturadores SMTP; cookies/CORS/CSRF, expiração e credenciais protegidas |
| 5: catálogo | Testes de catálogo/ISBN, N:M, filtros, acesso/estados e operações reais no fluxo HTTP |
| 6–10: leitura/comunidade | Testes de serviço/HTTP e MariaDB de concorrência, transação, progressos, slot de review, máscaras, árvores, curtidas/notificações |
| 11–14: administração | Serviços/HTTP/MariaDB de denúncia/punição/sugestão/papéis e proteção do último administrador; concorrência real da última revogação administrativa não exercitada separadamente |
| 15: estatísticas/recomendações | Consultas reais, duas contas/períodos UTC, páginas sem duplicação por correção/edição, listas, idioma e rankings determinísticos |
| 16–17: suporte/exclusão | FAQ e chamados persistidos/isolados; exclusão e purga em banco de teste mantêm FKs e removem dados privados |
| 18: esquema e reinício | Instalação nova em instância MariaDB temporária na porta3310, setup real, migrações up/down/up, constraints reais; recomposição de serviços e reinício do processo HTTP |
| 19: integrações | ISBN local/cadastro manual; SMTP TCP loopback com sucesso/falha/retry, tokens e duplicatas; sem provedor externo de ISBN |
| 20: operação | Backup/restauração real com dados e triggers; health/ready; encerramento aguarda requisição em andamento em teste TCP |

Comandos principais, a partir de `Backend`, conforme o README:

```powershell
go mod tidy
go mod verify
go fmt ./...
go vet ./...
go test ./...
.\scripts\with_env.ps1 -EnvFile .env.test -GoArgs @('test','-tags=integration','-p=1','./...','-count=1')
go build ./cmd/api
go build ./cmd/admin
go run ./cmd/openapi -check
```

As integrações serializam pacotes (`-p=1`) porque a purga de retenção tem escopo global no banco exclusivo; os testes de concorrência continuam abrindo conexões independentes dentro do cenário. Uma execução inicial com pacotes paralelos encontrou conflito de locks entre fixtures de retenção e o fluxo HTTP; não era isolamento adequado da suíte. O comando documentado usa execução serial dos pacotes.

`go test -race ./internal/model/services` foi tentado: a toolchain retornou `-race requires cgo`; não há compilador C compatível disponível nesta execução Windows. A ausência do detector de corrida não é apresentada como aprovação; a concorrência SQL foi exercitada pelos testes reais descritos acima.

Não foram executados Docker Compose, Nix, frontend, envio para provedor externo de e-mail, verificação documental de idade, hospedagem/HTTPS ou agendamento de backup/retenção. Os comandos e imagens Compose foram preparados/verificados, mas sua execução fica para um ambiente com Docker. O filtro de spam/termos/domínios é determinístico e limitado; análise humana permanece necessária.
