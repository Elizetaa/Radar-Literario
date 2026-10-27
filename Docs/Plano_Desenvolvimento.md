# Plano de entrega — três semanas

Este plano substitui o cronograma anterior. Objetivo: entregar o Radar Literário funcional com quatro pessoas e apoio de IA, começando pelo backend, em três semanas.

## Capacidade e premissas

- Planejar 15 dias úteis; finais de semana não entram na capacidade prevista.
- A disponibilidade ainda é N horas por pessoa/dia. Capacidade bruta: `4 × N × 15 = 60N` horas. Com 3 horas/dia são 180 horas; com 4 horas/dia são 240 horas. Estudo, revisão, reuniões e correções consomem parte desse total.
- Usar provisoriamente 4 horas por pessoa/dia para distribuir tarefas, sem tratar isso como disponibilidade confirmada.
- Manter Go, Gin, GORM, MariaDB, React, TypeScript e Tailwind, conforme a arquitetura.
- Você começa pelo backend; os demais trabalham em paralelo. O frontend começa pelos contratos e integra as rotas assim que estiverem disponíveis.
- Todos os módulos entram no plano, com implementações simples, persistência real e permissões verificadas. O prazo não garante que todo o escopo caiba: confirmar capacidade e requisitos no primeiro dia e reestimar nos dias 5 e 10.

Base: `Arquitetura_API.md`, estrutura do repositório e README do protótipo. Os PDFs devem ser conferidos no dia 1; este planejamento não equivale a uma auditoria completa dos requisitos deles.

## Situação atual

O frontend é um protótipo com dados simulados. O backend está em migração para Go com `cmd/api`, `internal`, módulo Go e Gin já adicionados. Há código inicial de rotas e alterações locais em andamento; este plano não modifica esses arquivos.

A arquitetura ainda descreve o esboço Python e uma estrutura em `Backend/src/`. Alinhar o documento à estrutura escolhida. Conferir imports/pacotes com o caminho de módulo do `go.mod`. O `.gitignore` está vazio e o flake ainda declara Python em um ambiente Linux: definir um único caminho de execução reproduzível para a entrega.

## Divisão da equipe

P1–P4 são pessoas da equipe. Substituir pelos nomes no dia 1. Cada tarefa deve ter um responsável e outro integrante como revisor.

| Pessoa | Responsabilidade principal |
| --- | --- |
| P1 — você | Backend: inicialização, configuração, contas, sessão, perfil, estante, diários e avaliações |
| P2 | Backend: catálogo, comunidade, administração, estatísticas, recomendações e suporte |
| P3 | Frontend: base TypeScript, cliente HTTP e fluxos do leitor |
| P4 | Frontend administrativo e complementar, ambiente de demonstração, validação e documentação |

P2 concentra muitos módulos. P4 deve preparar contratos e cenários com P2 e ajudar na implementação quando tiver capacidade. Redistribuir tarefas no dia 5 conforme o avanço real. A divisão define responsáveis, não impede colaboração.

## Linha do tempo

| Período | Meta verificável |
| --- | --- |
| Semana 1 — dias 1–5 | Cadastro → login → catálogo → estante pela interface, com persistência |
| Semana 2 — dias 6–10 | Primeira implementação integrada de leitura, comunidade, administração e complementos |
| Semana 3 — dias 11–15 | Fechar lacunas, validar todos os fluxos, corrigir e entregar |

## Semana 1 — base e primeiro fluxo real

| Dia | P1 — backend principal | P2 — backend complementar | P3 — frontend leitor | P4 — administração e entrega |
| --- | --- | --- | --- | --- |
| 1 | Executar API; alinhar módulo/imports, configuração, estrutura e rota de saúde | Modelagem central e preparação de MariaDB/migrações com P1 | Base TypeScript, navegação e referência visual do protótipo | Conferir requisitos, criar checklist de aceite e preparar ambiente de demonstração |
| 2 | Cadastro com validação, hash de senha e unicidade | Obras/edições/autores/gêneros, migrações e carga fictícia | Cliente HTTP e telas de cadastro/login conforme contrato | Validar instalação; preparar perfil e estrutura das telas administrativas |
| 3 | Login/logout, sessão, middleware de autenticação e CSRF | Busca, paginação e consulta de catálogo/edições | Integrar cadastro/login e iniciar catálogo real | Perfil, estados de erro e acesso por papel; validar autenticação com P1 |
| 4 | Perfil e estante com identidade obtida pela sessão | Autorização administrativa e manutenção inicial de obras/edições | Integrar catálogo, detalhes e estante | Integrar perfil e manutenção administrativa inicial |
| 5 | Concluir isolamento entre leitores e corrigir fluxo principal | Concluir catálogo e apoiar integração | Demonstrar cadastro → login → catálogo → estante → logout | Testar com duas contas, registrar falhas e atualizar instruções |

Checklist:

- [ ] API e frontend executam com comandos documentados.
- [ ] Migrações constroem o esquema em banco vazio; carga fictícia é reproduzível.
- [ ] Senhas têm hash e não aparecem em responses/logs.
- [ ] Sessões expiram e são revogadas no logout; escritas por cookie têm proteção CSRF.
- [ ] Catálogo separa obra de edição e tem paginação limitada.
- [ ] Estante tem um status por leitor/obra e não pode ser alterada por outra conta.
- [ ] Dados permanecem após reinício, e o fluxo principal funciona pela interface.

**Marco do dia 5:** se o fluxo ainda não funcionar, redistribuir pessoas para concluí-lo e simplificar a apresentação dos módulos seguintes. Registrar qualquer redução de requisito e obter concordância de quem avalia; uma tela simulada não conta como funcionalidade concluída.

## Semana 2 — completar os módulos

| Dia | P1 — backend principal | P2 — backend complementar | P3 — frontend leitor | P4 — administração e entrega |
| --- | --- | --- | --- | --- |
| 6 | Diários, progresso, autoria e transação com estante | Feed e comentários com controle de acesso ao conteúdo | Integrar diário e progresso | Integrar catálogo administrativo; preparar denúncia/moderação |
| 7 | Avaliações de 1 a 5, unicidade e média consistente | Curtidas em publicações/comentários conforme requisito e denúncias | Integrar avaliações e iniciar feed/comentários | Telas de denúncias e ações administrativas; testar permissões |
| 8 | Validar restrição etária com P2 e corrigir regras de leitura | Moderação, punições e sugestões de alteração do catálogo | Concluir comunidade, spoiler e conteúdo sensível | Integrar moderação, sugestões e registro das decisões |
| 9 | Exclusão de conta/conteúdo e efeitos nas relações existentes | Estatísticas pessoais e recomendações determinísticas | Integrar perfil completo, exclusão e sugestões de catálogo | Integrar estatísticas/recomendações e conferir cálculos |
| 10 | Apoiar FAQ/suporte com P2; revisar transações/permissões | Concluir FAQ/suporte e lacunas dos complementos | Fechar lacunas nas telas do leitor | Integrar FAQ/suporte e demonstrar todos os módulos |

Checklist:

- [ ] Diário exige conteúdo/progresso; diário e estante são atualizados na mesma transação.
- [ ] Diário e avaliação respeitam autoria na edição/exclusão, salvo ações administrativas previstas.
- [ ] Notas ficam entre 1 e 5 e a média confere após criar, alterar e excluir avaliações.
- [ ] Publicações declaram spoiler; interface pede confirmação para spoiler/conteúdo sensível.
- [ ] API aplica restrição etária em listagens e acesso direto por ID.
- [ ] Feed, comentários e curtidas persistem, com paginação e unicidade quando aplicável.
- [ ] Leitores denunciam e sugerem alterações de catálogo; administradores analisam e decidem.
- [ ] Administração mantém catálogo e registra moderação/punições; administrador não pode banir outro administrador.
- [ ] Estatísticas usam somente dados do próprio leitor e têm fórmulas verificáveis.
- [ ] Recomendações usam regra documentada e dados reais da aplicação.
- [ ] FAQ e solicitações de suporte seguem os fluxos previstos na especificação.
- [ ] Exclusão de conta/conteúdo segue política definida e preserva integridade.

**Marco do dia 10:** classificar cada requisito como integrado, incompleto ou não iniciado. Atribuir aceite e responsável para pendências dos dias 11–12. Se houver módulos não iniciados, registrar explicitamente o risco de não entregar todo o escopo.

## Semana 3 — fechamento e entrega

| Dia | Trabalho da equipe | Resultado |
| --- | --- | --- |
| 11 | Completar pendências e conferir fluxos de leitor e administrador | Todos os fluxos acordados acessíveis pela interface |
| 12 | Validar instalação em banco vazio, ambiente de destino, sessão, CORS/CSRF e HTTPS quando publicado | Execução fora da máquina principal |
| 13 | Congelar funcionalidades novas; testar duas contas, papéis, autoria, erros, transações e exclusões | Lista final de defeitos ordenada por impacto |
| 14 | Corrigir defeitos, executar verificações Go/frontend, testar backup/restauração e revisar documentação | Versão candidata à entrega |
| 15 | Repetir demonstração completa e preparar os materiais exigidos | Versão entregue identificada e roteiro de apresentação |

Checklist final:

- [ ] Frontend, API e MariaDB integrados; fluxos finais não dependem de mocks.
- [ ] Go formatado; `go vet ./...` e `go test ./...` aprovados, com configuração documentada dos testes de integração relevantes.
- [ ] Frontend aprovado em lint, verificação de tipos e build.
- [ ] Testes de integração usam banco separado e dados fictícios.
- [ ] `.env.example` atualizado e segredos reais fora do Git.
- [ ] README com ambiente, migrações, execução, testes e criação controlada do administrador.
- [ ] Contrato da API corresponde às rotas, payloads e erros reais.
- [ ] Backup e restauração verificados para o ambiente de entrega.
- [ ] Equipe consegue reproduzir a demonstração seguindo as instruções.

## Decisões antes de implementar

Registrar decisões em `Docs/decisoes.md` após conferir a especificação. Os prazos vêm antes das funcionalidades dependentes.

| Até quando | Decisão |
| --- | --- |
| Dia 1 | Checklist completo da avaliação, horas disponíveis por integrante e necessidade de hospedagem pública |
| Dia 1 | Estrutura `cmd/api` + `internal`, responsabilidades MVC e contratos de conta, catálogo e estante |
| Dia 1 | Frontend: arquitetura prevê reescrita em TypeScript. Usar o protótipo como referência visual reduz decisões de design; reaproveitar código JSX seria mudança da estratégia e precisa ser registrada pela equipe |
| Dia 2 | Sessão, expiração, cookies, CSRF e criação controlada do administrador |
| Dia 2 | Política de exclusão de conta/conteúdo, antecipando comentários, denúncias e histórico |
| Dia 3 | Classificação etária e controle de acesso a publicações |
| Dia 5 | Avaliação durante leitura ou somente após conclusão; unicidade por leitor/obra ou leitor/edição |
| Dia 5 | Relação entre progresso, edição e total de páginas; representação das publicações no feed |
| Dia 7 | Fórmulas de estatísticas, regra de recomendação e fluxo/permissões de suporte |

## Simplificações para preservar a cobertura

- Recomendações: propor seleção por interesses/gêneros e exclusão de obras já na estante, com ordenação e alternativa sem histórico definidas. Confirmar a regra; aprendizado de máquina não é necessário para essa proposta.
- Estatísticas: calcular os indicadores exigidos por SQL e mostrar números/tabelas. Gráficos somente se necessários à avaliação.
- Administração: tabelas, formulários e aprovação/rejeição com histórico exigido; apresentação visual simples.
- Suporte: FAQ e solicitações com atendimento conforme requisito. Não acrescentar chat em tempo real sem exigência.
- Catálogo: manutenção administrativa manual e carga fictícia. Integração externa somente se obrigatória.
- Interface: componentes compartilhados e navegação simples. Evitar redesenhar todas as telas antes de integrar os fluxos.
- Ambiente: documentar um caminho principal de instalação e execução.

Se a especificação exigir mais que essas soluções, o requisito permanece no escopo e precisa entrar na estimativa. Completo será medido pelo checklist da avaliação, não pela quantidade de telas.

## Aprendizado e uso de IA

1. Reunião diária de até 10 minutos: entrega anterior, entrega do dia e impedimentos.
2. Tarefas pequenas com responsável, dependências e aceite observável.
3. Estudo de 20–30 minutos sobre o conceito necessário, seguido de aplicação no projeto.
4. Pedidos à IA com contrato, estrutura real e regras específicas. Dividir por funcionalidade e responsabilidade.
5. Ler e revisar o código gerado; validar sucesso, falha e acesso indevido. Não integrar código que ninguém da equipe consegue explicar.
6. Integrar diariamente e conferir a comunicação entre API e frontend. Combinar alterações de contrato antes de mudar as duas pontas.

Exemplo de tarefa: criar `PUT /api/v1/me/shelf/:work_id`, obter leitor da sessão, validar edição e garantir unicidade leitor/obra. Aceite: repetir não duplica e outra conta não altera o registro.

## Primeiras ações

- [ ] Atribuir P1–P4 e registrar N horas diárias reais de cada pessoa.
- [ ] Conferir PDFs e transformar exigências em critérios verificáveis.
- [ ] Executar a API atual e corrigir inconsistências de módulo/imports se necessário.
- [ ] Definir modelo inicial e contratos de cadastro, sessão, catálogo e estante.
- [ ] Preparar MariaDB e migrações enquanto a base frontend é organizada.
- [ ] Demonstrar uma chamada real entre interface e API ainda na primeira semana.

O prazo tem risco elevado para uma equipe em aprendizado. IA pode apoiar implementação e revisão, mas a viabilidade precisa ser reavaliada com a disponibilidade real e os marcos dos dias 5 e 10. Nenhum módulo deve ser retirado silenciosamente.
