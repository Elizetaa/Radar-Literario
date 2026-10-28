# Decisões do backend

Registro inicial em 27/10/2026, após conferir código, arquitetura, plano, PDF de requisitos, diagrama PNG e referência do protótipo. A implementação é de backend; frontend e entrega externa de e-mail não estão implicitamente validados.

| Categoria | Contexto e fonte | Escolha | Impacto | Aceite |
| --- | --- | --- | --- | --- |
| requisito de fonte | RF01–RF35, RN01–RN07, NR08, RNF02–RNF06 e UC01–UC12 do PDF | Persistência MariaDB, controle de autoria e papéis, sessões opacas, catálogo e comunidade | Serviços aplicam regras; controllers não acessam GORM | Testes de serviço, HTTP e integração em banco isolado |
| compatibilização entre fontes | Arquitetura ainda menciona Python/src; repositório já contém Go/cmd/internal | Preservar módulo radar-literario/backend, Gin e estrutura cmd/internal | Substituir /api/usuarios fictício por /api/v1 | Build dos comandos e contrato de rotas |
| compatibilização entre fontes | Diagrama PNG versus minimundo textual | Autores N:M, usuários base com perfis leitor/admin; paused; obra separada de edição | FKs e tabelas relacionais explícitas; sexo/país opcionais privados | Integridade de relações no MariaDB |
| premissa proposta para lacuna | Seções 5–8 do prompt resolvem ambiguidades do PDF | Adotar integralmente estados, papéis e políticas de spoiler/idade, avaliação única, progresso/eventos, recomendações, anonimização e retenção ali descritos | Classificação 0/10/12/14/16/18; idade autodeclarada; datas civis e períodos UTC; IDs BIGINT UNSIGNED | Cenários específicos documentados na matriz de cobertura |
| premissa proposta para lacuna | Ambiente local autorizado pelo responsável nesta sessão | Usar serviço MariaDB existente; criar bancos separados, usuários app/migração/teste com senhas aleatórias locais | API sem root; segredos em .env ignorado | Migrações e testes reais com credenciais limitadas |
| premissa proposta para lacuna | Responsável informou senha curta para primeiro administrador local | Exceção explícita somente no bootstrap em desenvolvimento; nome inferido como Luiz Elizeta e sem nascimento informado | Produção e cadastro público recusam senha fraca; admin sem idade informada acessa classificação 0 nas rotas públicas | Bootstrap idempotente e produção sem senha demonstrativa |
| premissa proposta para lacuna | Especificação não exige verificação de e-mail nem metadados externos | Cadastro ativo sem verificação; ISBN consultado localmente e cadastro manual validado; recuperação via SMTP/outbox persistida | Não cadastrar conta presa em pending_verification; importação bibliográfica externa não implementada | ISBN local e capturador SMTP de teste |
| premissa proposta para lacuna | Proteção de credenciais | Bcrypt custo 12; tokens aleatórios de 32 bytes armazenados por SHA-256; CSRF ligado à sessão e preauth com cookie double-submit, checagem de Origin | Tokens não retornam no JSON de recuperação nem logs; logout/troca revogam sessões | Testes de expiração, CSRF e consumo concorrente |
| premissa proposta para lacuna | Política técnica de retenção | Anonimização imediata da identidade/remoção de dados privados; conteúdo indisponível e evidências restritas por prazo configurável, limpeza pelo comando maintenance | Política técnica revisável pelo responsável, sem declaração de conformidade jurídica automática | Exclusão e limpeza em banco isolado |

Os documentos de contrato e cobertura descrevem as escolhas finais de cada módulo e a evidência executada; testes não executados permanecem identificados como tal.

## Complementos e decisões finais

As decisões específicas, com as mesmas três categorias, estão nos documentos de [contas](cobertura_contas.md), [catálogo](cobertura_catalogo.md), [leitura](cobertura_leitura.md) e [administração](cobertura_admin.md). A matriz geral está em [cobertura_backend.md](cobertura_backend.md).

| Categoria | Contexto e fonte | Escolha | Impacto | Aceite |
| --- | --- | --- | --- | --- |
| compatibilização entre fontes | Toolchain escrita no go.mod precisava de comprovação | Preservar Go 1.27.3/Gin1.12; arquivo oficial Go baixado e SHA-256 conferido; dependências geradas por ferramentas Go | Sem downgrade presumido; GORM/mysql/Goose e validador OpenAPI fixados no go.mod/sum | go version, mod verify, vet, testes e builds |
| compatibilização entre fontes | MariaDB13 rejeitou CHECK referenciando id AUTO_INCREMENT em comentário | Antes de concluir a migração3 local, usar triggers para pai imutável/self-reference e FK composta para mesmo post | Árvores não podem ser reencadeadas/cíclicas; histórico já aplicado não foi reescrito | Migração vazia e violações reais de trigger/FK |
| premissa proposta para lacuna | Correção/troca de edição não pode criar crédito artificial (§7.1) | Preservar máximo de página já creditado por registro entre edições; nova contagem só acima desse máximo | Conservador entre edições com paginações diferentes; releitura independente não modelada | Teste ida/volta de edição e avanço sem crédito duplicado |
| requisito de fonte | RN03, RF31 e exclusão: moderação/autoria não podem sobrescrever decisão concorrente | Gravações do autor condicionadas a estado editável; moderação não altera deleted/draft; conflitos SQL1020/1213/1205 retornam409 | Snapshot antigo não restaura conteúdo moderado/excluído | Snapshots concorrentes reais e exclusão entre leitura/decisão administrativa |
| premissa proposta para lacuna | Login precisa rotacionar sessão inclusive troca de conta no navegador | Revogar segredo anterior pelo hash, independentemente do novo usuário | Replay do cookie da conta anterior falha após login na outra | Fluxo HTTP de troca de contas com replay |
| premissa proposta para lacuna | Retenção precisa eliminar texto sem violar CHECK de post não vazio | Substituir texto antigo removido por marcador genérico, limpar título/URL e evidências privadas conforme prazo | Preserva integridade e não mantém corpo original | Purga real e exclusão com suporte/outbox no MariaDB |
| premissa proposta para lacuna | Validação deve apontar campos e códigos estáveis | Errors400 fornecem fields com campo conhecido; body/pagination identificam grupos; `_request` indica erro transversal | Cliente pode exibir validação sem interpretar mensagens ou erros SQL | Testes HTTP e contrato de ErrorEnvelope |
| premissa proposta para lacuna | Testes de retenção alteram escopo amplo do banco isolado | Serializar pacotes de integração com -p=1; manter concorrência interna por conexões | Evita contenção artificial entre suítes sem reduzir testes concorrentes do domínio | Suíte completa contra banco separado |

Backup/restauração e sua política operacional estão em [backup.md](backup.md). Rotinas de backup/retenção e hospedagem não foram agendadas/publicadas automaticamente.
