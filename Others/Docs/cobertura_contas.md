# Evidência de contas e recuperação

Testes implementados em `Backend/internal/model/tools/account_integration_test.go` (MariaDB real, tag `integration`) e `smtp_test.go` (transporte SMTP real em loopback).

| Requisito | Evidência |
| --- | --- |
| RF01 / UC01, unicidade concorrente | Seis cadastros simultâneos com e-mail igual e, separadamente, apelido igual: um sucesso e cinco conflitos `409` em cada cenário. Fixtures persistidas próprias; limpeza somente dos IDs criados pelo teste. |
| RF02 / RNF03 | Login persistido, sessão válida, revogação após recuperação e impossibilidade de autenticar conta excluída. Expiração absoluta/inatividade também testadas nas unidades de serviços. |
| RF03 | Outbox criptografada, token persistido como SHA-256, consumo concorrente por seis chamadas com exatamente um sucesso e cinco tokens inválidos, expiração e nova senha preservada após recriar o serviço. E-mail desconhecido mantém resposta genérica. Troca do e-mail invalida tokens e mensagens destinados ao endereço anterior; recuperação revalida a conta após uma exclusão intercalada com a primeira consulta. |
| Recuperação e falhas externas | Adaptador SMTP envia envelope/corpo reais para servidor TCP local; o teste integrado completa outbox → entrega SMTP → extração do link → recuperação de senha. Recusa injeção de cabeçalhos, rejeição de destinatário, ausência de STARTTLS obrigatório e contexto cancelado. Falha controlada de entrega mantém outbox com tentativa e reenvio agendado. |
| RNF06 / exclusão e retenção | Exclusão exige senha, anonimiza identidade, revoga sessões e elimina recuperação pendente e mensagens privadas de suporte. Limpeza por retenção remove texto original de publicação removida sem quebrar integridade. |

Não se afirma entrega em provedor de e-mail externo ou caixa de entrada real. O capturador SMTP usado nos testes é local e não exige Mailpit nem conexão com a Internet.

Executados com sucesso: sete testes MariaDB de contas/recuperação e dois testes do adaptador SMTP (mais os subcasos de unicidade concorrente), usando `go test -tags=integration ./internal/model/tools -run 'Test(MariaDBAccount|MariaDBRecovery|MariaDBEmail|SMTP)' -v`. Os testes não esvaziam tabelas: cenários simples usam rollback, e os de concorrência removem somente suas próprias fixtures.
