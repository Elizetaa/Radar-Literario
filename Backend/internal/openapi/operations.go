package openapi

import (
	"radar-literario/backend/internal/model/schemas"
)

func operations() []operation {
	var all []operation
	add := func(method, path, summary, tag string, request, response any, list, private bool, status int, description string, params ...any) {
		if path != "/health" && path != "/ready" {
			path = "/api/v1" + path
		}
		all = append(all, operation{method: method, path: path, summary: summary, tag: tag, request: request, response: response, list: list, private: private, status: status, description: description, parameters: params})
	}
	add("GET", "/health", "Verificar vivacidade", "Saúde", nil, health{}, false, false, 200, "Não consulta o banco.")
	add("GET", "/ready", "Verificar prontidão", "Saúde", nil, health{}, false, false, 200, "Verifica conexão e versão necessária das migrações; falha com 503 sem revelar detalhes internos.")
	add("POST", "/readers", "Cadastrar leitor", "Contas", schemas.RegisterReader{}, schemas.AccountView{}, false, false, 201, "E-mail e apelido normalizados têm índices únicos. Cadastro cria somente leitor; senha protegida por bcrypt custo 12.")
	add("GET", "/auth/csrf", "Obter token CSRF", "Contas", nil, csrf{}, false, false, 200, "Antes do login, cria cookie HttpOnly de pré-autenticação por 30 minutos. Com sessão, retorna token vinculado à sessão atual.")
	add("POST", "/auth/login", "Entrar e rotacionar sessão", "Contas", schemas.Login{}, loginResult{}, false, false, 200, "Preserve Set-Cookie e utilize o novo csrf_token. Falha de credenciais tem mensagem genérica; limite configurável de tentativas por IP.")
	add("POST", "/auth/logout", "Encerrar sessão atual", "Contas", nil, nil, false, false, 204, "Revoga segredo atual e limpa cookie. Repetição não restaura sessão; obtenha CSRF de pré-autenticação após a revogação.")
	add("POST", "/auth/forgot-password", "Solicitar recuperação de senha", "Contas", schemas.ForgotPassword{}, message{}, false, false, 200, "Resposta genérica independentemente de conta existente. Token de uso único é enviado por outbox SMTP, nunca no JSON.")
	add("POST", "/auth/reset-password", "Redefinir senha", "Contas", schemas.ResetPassword{}, nil, false, false, 204, "Consome token atomicamente e revoga todas as sessões e tokens de recuperação da conta.")
	add("GET", "/me", "Consultar própria conta", "Contas", nil, schemas.AccountView{}, false, true, 200, "Dados pessoais são devolvidos somente ao titular.")
	add("PATCH", "/me", "Editar própria conta", "Contas", schemas.PatchAccount{}, schemas.AccountView{}, false, true, 200, "Campos ausentes são preservados; sex/country_of_origin/favorite_work_id aceitam null para limpar. Livro favorito deve estar publicado e ser elegível pela idade.")
	add("DELETE", "/me", "Excluir própria conta", "Contas", schemas.DeleteAccount{}, nil, false, true, 204, "Exige senha atual. Anonimiza identidade, remove dados privados, oculta conteúdo, revoga sessões e preserva referências mínimas de moderação pelo prazo configurado.")
	add("PUT", "/me/password", "Trocar própria senha", "Contas", schemas.ChangePassword{}, nil, false, true, 204, "Exige senha atual e revoga todas as sessões. Faça login novamente.")
	add("GET", "/readers/{id}", "Consultar perfil público", "Contas", nil, schemas.ReaderView{}, false, false, 200, "Respeita public/registered_users_only/private; não expõe e-mail ou nascimento. Favorito invisível retorna null.")

	workParams := []any{parameter("title", stringSchema(), "Trecho do título, até 300 caracteres."), parameter("author", stringSchema(), "Trecho do nome de autor, até 160 caracteres."), parameter("genre_id", idSchema(), "Identificador de gênero ativo."), parameter("isbn", stringSchema(), "ISBN-10/13 válido da edição."), parameter("language", stringSchema(), "Idioma da edição publicada, como pt-br."), parameter("order", enum("title", "newest", "popular", "top-rated"), "Ordenação estável; ID é o desempate final.")}
	add("GET", "/works", "Buscar obras publicadas", "Catálogo", nil, schemas.Work{}, true, false, 200, "Catálogo elegível pelo status e idade. Média inclui avaliações publicadas de contas ativas e classificação elegível.", workParams...)
	add("GET", "/works/{id}", "Consultar obra", "Catálogo", nil, schemas.Work{}, false, false, 200, "Retorna autores N:M, gêneros, média e quantidade de avaliações elegíveis; média null quando não há avaliações.")
	add("GET", "/works/{id}/editions", "Listar edições da obra", "Catálogo", nil, schemas.Edition{}, true, false, 200, "Somente edições e obra publicadas, respeitando ambas as classificações.")
	add("GET", "/editions/{id}", "Consultar edição", "Catálogo", nil, schemas.Edition{}, false, false, 200, "")
	add("GET", "/catalog/isbn/{isbn}", "Consultar edição local por ISBN", "Catálogo", nil, schemas.Edition{}, false, false, 200, "ISBN é validado e convertido para ISBN-13; ausência local retorna 404. O cadastro manual com dados completos continua disponível. Não há provedor bibliográfico externo obrigatório.")
	for _, r := range []struct {
		name, title string
		result      any
	}{{"authors", "autores", schemas.Author{}}, {"publishers", "editoras", schemas.Publisher{}}} {
		add("GET", "/"+r.name, "Listar "+r.title, "Catálogo", nil, r.result, true, false, 200, "")
		add("GET", "/"+r.name+"/{id}", "Consultar "+r.title, "Catálogo", nil, r.result, false, false, 200, "")
	}
	add("GET", "/genres", "Listar gêneros ativos", "Catálogo", nil, schemas.Genre{}, true, false, 200, "")
	add("GET", "/rankings/popular", "Consultar obras populares", "Catálogo", nil, schemas.Work{}, true, false, 200, "Popularidade = leitores ativos distintos com obra na estante. Desempate por média e ID; respeita classificação atual.")
	add("GET", "/rankings/top-rated", "Consultar obras melhor avaliadas", "Catálogo", nil, schemas.Work{}, true, false, 200, "Ordena por média, quantidade de avaliações e ID. Exige RANKING_MINIMUM_REVIEWS avaliações elegíveis (default 1).")
	add("GET", "/me/recommendations", "Consultar recomendações pessoais", "Catálogo", nil, schemas.Work{}, true, true, 200, "Afinidade = gêneros distintos compartilhados com interesses e gêneros de obras avaliadas com nota 4/5. Exclui toda estante; desempata por média, amostra e ID. Sem afinidade utiliza ranking de notas elegíveis.")
	add("GET", "/me/interests", "Consultar gêneros de interesse", "Catálogo", nil, []schemas.Genre{}, false, true, 200, "Lista própria, limitada a 100 relacionamentos; não utiliza paginação.")
	add("PUT", "/me/interests", "Substituir gêneros de interesse", "Catálogo", schemas.InterestsInput{}, []schemas.Genre{}, false, true, 200, "Valida gêneros ativos e pares únicos; substituição transacional. Lista vazia remove interesses.")
	add("GET", "/me/preferred-works", "Listar obras preferidas", "Catálogo", nil, schemas.Work{}, true, true, 200, "Preferência é relação explícita independente da estante e da recomendação.")
	add("PUT", "/me/preferred-works/{work_id}", "Marcar obra preferida", "Catálogo", nil, nil, false, true, 204, "Idempotente; valida obra elegível.")
	add("DELETE", "/me/preferred-works/{work_id}", "Desmarcar obra preferida", "Catálogo", nil, nil, false, true, 204, "Idempotente; não altera estante.")
	add("POST", "/catalog-suggestions", "Sugerir alteração do catálogo", "Catálogo", schemas.SuggestionInput{}, schemas.CatalogSuggestion{}, false, true, 201, "Leitor sugere create/update/delete de work/edition com DTO validado; nenhuma mudança de catálogo ocorre até aprovação. initial_edition permite sugerir obra e edição ISBN juntas. Relações referenciam autores/editoras/gêneros existentes.")
	add("GET", "/me/catalog-suggestions", "Acompanhar próprias sugestões", "Catálogo", nil, schemas.CatalogSuggestion{}, true, true, 200, "Não acessa propostas de outros leitores.", parameter("status", enum("pending", "approved", "rejected"), "Estado da sugestão."))
	add("GET", "/me/catalog-suggestions/{id}", "Consultar própria sugestão", "Catálogo", nil, schemas.CatalogSuggestion{}, false, true, 200, "")

	shelfParams := []any{parameter("status", enum("planned", "in_progress", "paused", "completed", "abandoned"), "Estado atual de leitura.")}
	add("GET", "/me/shelf", "Consultar própria estante", "Leitura", nil, schemas.ShelfItem{}, true, true, 200, "Uma relação por leitor/obra; edição escolhida e progresso pertencem ao leitor.", shelfParams...)
	add("GET", "/readers/{id}/shelf", "Consultar estante pública", "Leitura", nil, schemas.ShelfItem{}, true, false, 200, "Respeita a visibilidade do perfil e a classificação atual do catálogo.", shelfParams...)
	add("PUT", "/me/shelf/{work_id}", "Criar ou atualizar item da estante", "Leitura", schemas.ShelfInput{}, schemas.ShelfItem{}, false, true, 200, "Idempotente por leitor/obra. Página deve caber na edição e edição deve pertencer à obra. planned exige zero; completed completa páginas e define conclusão. Redução exige correction; troca de edição exige edition_change e não credita páginas. Retornar de completed limpa finished_at. Avanços abaixo de pico já creditado não duplicam páginas.")
	add("DELETE", "/me/shelf/{work_id}", "Remover item da estante", "Leitura", nil, nil, false, true, 204, "Remove registro e eventos que contribuem às estatísticas; preserva diários, edições e obras. Reinserção não infere releitura histórica.")
	for _, r := range []struct{ path, name, input, patch string }{{"reading-entries", "diário", "DiaryInput", "DiaryPatch"}, {"reviews", "avaliação", "ReviewInput", "ReviewPatch"}} {
		add("POST", "/"+r.path, "Criar "+r.name, "Leitura", schemas.PostInput{}, schemas.Post{}, false, true, 201, "Autoria vem da sessão. contains_spoiler é obrigatório. Rascunho não altera estante; primeira publicação de diário aplica progresso uma vez na mesma transação. Avaliação exige leitura em progresso/concluída e tem slot único por leitor/obra.")
		all[len(all)-1].requestSchema = r.input
		add("GET", "/"+r.path+"/{id}", "Consultar "+r.name, "Leitura", nil, schemas.Post{}, false, false, 200, "ID é compartilhado com post_id. Rascunho só é visto pelo autor. Status, idade, spoiler e sensível também se aplicam ao acesso direto.", revealParams()...)
		add("PATCH", "/"+r.path+"/{id}", "Editar própria "+r.name, "Leitura", schemas.PostPatch{}, schemas.Post{}, false, true, 200, "Não permite alterar autoria/tipo/obra. Autor não pode republicar conteúdo oculto/removido. Edição de diário já publicado não reescreve progresso histórico; média é consultada das avaliações vigentes.")
		all[len(all)-1].requestSchema = r.patch
		add("DELETE", "/"+r.path+"/{id}", "Excluir própria "+r.name, "Leitura", nil, nil, false, true, 204, "Exclusão lógica. Avaliação excluída pelo autor libera novo post; moderação preserva bloqueio do slot e não pode ser contornada por recriação.")
		add("GET", "/me/"+r.path, "Listar próprios "+r.path, "Leitura", nil, schemas.Post{}, true, true, 200, "Inclui rascunhos próprios autorizados e aplica política atual de conteúdo.", revealParams()...)
	}
	postParams := append(revealParams(), parameter("kind", enum("diary", "review"), "Tipo de publicação."), parameter("work_id", idSchema(), "Filtrar por obra."))
	add("GET", "/posts", "Consultar feed", "Comunidade", nil, schemas.Post{}, true, false, 200, "Ordem published_at DESC, id DESC. Inclui autoria, tipo, obra, contadores e curtida do solicitante. Não retorna corpo protegido por padrão.", postParams...)
	add("GET", "/posts/{id}", "Consultar publicação", "Comunidade", nil, schemas.Post{}, false, false, 200, "Classificação efetiva é o máximo entre obra atual e publicação; confirmação não contorna idade.", revealParams()...)
	add("GET", "/readers/{id}/posts", "Consultar publicações do leitor", "Comunidade", nil, schemas.Post{}, true, false, 200, "Respeita perfil e conteúdo; não expõe rascunhos alheios.", postParams...)
	add("GET", "/works/{id}/reviews", "Listar avaliações da obra", "Leitura", nil, schemas.Post{}, true, false, 200, "Somente avaliações publicadas elegíveis.", revealParams()...)
	commentParams := append(revealParams(), parameter("parent_id", idSchema(), "Omitir lista raízes; informar lista somente respostas diretas desse pai. Não retorna árvore ilimitada."))
	add("GET", "/posts/{id}/comments", "Listar comentários ou respostas diretas", "Comunidade", nil, schemas.Comment{}, true, false, 200, "Invisibilidade do post impede acesso derivado. Tombstones preservam árvore sem texto.", commentParams...)
	add("POST", "/posts/{id}/comments", "Comentar publicação", "Comunidade", schemas.CommentInput{}, schemas.Comment{}, false, true, 201, "Pai deve ser comentário existente do mesmo post; pai é imutável, profundidade limitada por COMMENT_MAX_DEPTH (default 5). Notifica autor/pai sem corpo protegido.")
	add("PATCH", "/comments/{id}", "Editar próprio comentário", "Comunidade", schemas.CommentPatch{}, schemas.Comment{}, false, true, 200, "Pai e autoria não podem ser alterados.")
	add("DELETE", "/comments/{id}", "Excluir próprio comentário", "Comunidade", nil, nil, false, true, 204, "Mantém tombstone e referências; texto deixa de ser retornado.")
	for _, resource := range []string{"posts", "comments"} {
		add("PUT", "/"+resource+"/{id}/like", "Curtir "+resource, "Comunidade", nil, nil, false, true, 204, "Idempotente e transacional; par único e identidade de evento impedem curtidas/notificações duplicadas. Não notifica o próprio ator.")
		add("DELETE", "/"+resource+"/{id}/like", "Descurtir "+resource, "Comunidade", nil, nil, false, true, 204, "Idempotente; recurso deve permanecer acessível.")
	}
	add("GET", "/me/notifications", "Consultar próprias notificações", "Comunidade", nil, schemas.Notification{}, true, true, 200, "Não inclui texto de conteúdo; target_available é recalculado com idade e status atuais.", parameter("unread", object{"type": "boolean"}, "true retorna não lidas; false retorna lidas."))
	add("PUT", "/me/notifications/{id}/read", "Marcar notificação como lida", "Comunidade", nil, nil, false, true, 204, "Somente destinatário; idempotente.")
	add("GET", "/me/statistics", "Consultar estatísticas pessoais", "Leitura", nil, schemas.Statistics{}, false, true, 200, "Somente titular. Intervalos [início,fim) em UTC. Retorna 12 meses com zeros, total anual, status, páginas correntes, média, diários, gêneros e idioma mais lido. Conclusões usam registros atualmente completed e finished_at; páginas usam eventos creditados sem soma de snapshots.", parameter("year", object{"type": "integer", "minimum": 1900, "maximum": 9998}, "Ano UTC; default é o ano atual do servidor."))
	add("GET", "/me/reading-report", "Listar relatório de leitura", "Leitura", nil, schemas.ShelfItem{}, true, true, 200, "completed, incomplete (in_progress/paused), abandoned, desired retornam ShelfItem; preferred/recommended retornam Work. Preferida é relação explícita, recomendada usa algoritmo.", parameter("group", object{"type": "string", "enum": []string{"completed", "incomplete", "abandoned", "desired", "preferred", "recommended"}, "default": "completed"}, "Grupo de relatório."))
	all[len(all)-1].responseSchema = "ReadingReportItem"

	add("POST", "/reports", "Registrar denúncia", "Comunidade", schemas.CreateReport{}, schemas.Report{}, false, true, 201, "Exatamente um de reported_user_id/reported_post_id/reported_comment_id. Valida alvo elegível e guarda evidências com acesso restrito.")
	add("GET", "/me/reports", "Acompanhar próprias denúncias", "Comunidade", nil, schemas.Report{}, true, true, 200, "Campos internos e identidade do denunciante não são divulgados a outros leitores.", parameter("status", enum("open", "in_review", "awaiting_information", "resolved", "dismissed", "canceled"), "Estado de análise."))
	add("GET", "/me/reports/{id}", "Consultar própria denúncia", "Comunidade", nil, schemas.Report{}, false, true, 200, "")
	add("POST", "/me/reports/{id}/cancel", "Cancelar própria denúncia ainda aberta", "Comunidade", nil, nil, false, true, 204, "Permitido somente em open, antes da atribuição/análise.")
	add("GET", "/me/reports/{id}/messages", "Consultar mensagens visíveis da denúncia", "Comunidade", nil, schemas.ReportMessage{}, true, true, 200, "Não inclui notas internas administrativas.")
	add("POST", "/me/reports/{id}/messages", "Complementar denúncia", "Comunidade", schemas.CreateReportMessage{}, schemas.ReportMessage{}, false, true, 201, "Permitido em awaiting_information; devolve à análise. Leitor só publica mensagem com visibilidade reporter.")
	add("GET", "/faq", "Consultar FAQ publicada", "Suporte", nil, schemas.FAQ{}, true, false, 200, "")
	add("GET", "/faq/{id}", "Consultar resposta publicada", "Suporte", nil, schemas.FAQ{}, false, false, 200, "")
	add("POST", "/support-tickets", "Abrir chamado de suporte", "Suporte", schemas.CreateSupportTicket{}, schemas.SupportTicket{}, false, true, 201, "Chamado e mensagem inicial são persistidos na mesma transação.")
	add("GET", "/me/support-tickets", "Consultar próprios chamados", "Suporte", nil, schemas.SupportTicket{}, true, true, 200, "Isolamento por solicitante.", parameter("status", enum("open", "in_progress", "resolved", "closed"), "Estado de atendimento."))
	add("GET", "/support-tickets/{id}", "Consultar chamado permitido", "Suporte", nil, schemas.SupportTicket{}, false, true, 200, "Somente solicitante ou atendente com permissão.")
	add("GET", "/support-tickets/{id}/messages", "Consultar conversa do chamado", "Suporte", nil, schemas.SupportMessage{}, true, true, 200, "Somente solicitante ou atendente permitido.")
	add("POST", "/support-tickets/{id}/messages", "Enviar mensagem no chamado", "Suporte", schemas.CreateSupportMessage{}, schemas.SupportMessage{}, false, true, 201, "Somente solicitante em chamado disponível ou atendente autorizado.")
	add("POST", "/support-tickets/{id}/close", "Encerrar próprio chamado", "Suporte", nil, nil, false, true, 204, "Encerramento pelo solicitante; reabertura depende da administração.")

	add("GET", "/admin/administrators", "Listar administradores", "Administração", nil, schemas.Administrator{}, true, true, 200, "")
	add("POST", "/admin/administrators", "Conceder acesso administrativo", "Administração", schemas.CreateAdministrator{}, schemas.Administrator{}, false, true, 201, "Informe user_id de conta existente OU dados completos de nova conta, nunca ambos. Sem senha fixa; revoga sessões em mudança de privilégio.")
	add("GET", "/admin/administrators/{id}", "Consultar administrador", "Administração", nil, schemas.Administrator{}, false, true, 200, "")
	add("PATCH", "/admin/administrators/{id}", "Alterar papel/estado administrativo", "Administração", schemas.PatchAdministrator{}, schemas.Administrator{}, false, true, 200, "Não remove/rebaixa o último system_administrator ativo; exige motivo e revoga sessões.")
	add("DELETE", "/admin/administrators/{id}", "Revogar perfil administrativo", "Administração", schemas.Reason{}, nil, false, true, 204, "Exige motivo. Protege último administrador de sistema e preserva trilha histórica.")
	add("GET", "/admin/users", "Consultar usuários com dados mínimos", "Administração", nil, schemas.AdministrativeUser{}, true, true, 200, "Sem e-mail/nascimento/hashes.", parameter("status", enum("pending_verification", "active", "inactive", "suspended", "deactivated", "banned"), "Estado da conta."))
	add("GET", "/admin/users/{id}", "Consultar usuário administrativo", "Administração", nil, schemas.AdministrativeUser{}, false, true, 200, "Dados minimizados.")
	for _, r := range []struct {
		name, title   string
		input, result any
	}{{"works", "obra", schemas.WorkInput{}, schemas.Work{}}, {"editions", "edição", schemas.EditionInput{}, schemas.Edition{}}, {"authors", "autor", schemas.AuthorInput{}, schemas.Author{}}, {"publishers", "editora", schemas.PublisherInput{}, schemas.Publisher{}}, {"genres", "gênero", schemas.GenreInput{}, schemas.Genre{}}} {
		params := []any{}
		if r.name == "works" {
			params = append(params, workParams...)
		}
		if r.name == "works" || r.name == "editions" {
			params = append(params, parameter("status", enum("draft", "pending_review", "published", "rejected", "hidden", "archived"), "Estado administrativo do catálogo."))
		}
		if r.name == "editions" {
			params = append(params, parameter("work_id", idSchema(), "Obra da edição."))
		}
		add("GET", "/admin/"+r.name, "Listar "+r.name+" administrativos", "Administração", nil, r.result, true, true, 200, "Consulta inclui registros não publicados.", params...)
		add("GET", "/admin/"+r.name+"/{id}", "Consultar "+r.title+" administrativo", "Administração", nil, r.result, false, true, 200, "Consulta inclui registros não publicados.")
		add("POST", "/admin/"+r.name, "Cadastrar "+r.title, "Administração", r.input, r.result, false, true, 201, "Obras exigem título/idioma e, quando publicadas, sinopse/autores/gêneros/classificação explícita. Edições exigem obra/editora/título/páginas/idioma/ISBN ou motivo; classificação explícita ao publicar. initial_edition permite criar obra e edição atomicamente.")
		all[len(all)-1].requestSchema = map[string]string{"works": "WorkCreate", "editions": "EditionCreate", "authors": "AuthorCreate", "publishers": "PublisherCreate", "genres": "GenreCreate"}[r.name]
		add("PATCH", "/admin/"+r.name+"/{id}", "Editar "+r.title, "Administração", r.input, r.result, false, true, 200, "Campos ausentes são preservados. Datas e ISBN anuláveis podem ser removidos com null; ISBN ausente exige justificativa. Edição não pode mudar de obra; redução de páginas não pode invalidar leituras/diários históricos.")
		add("DELETE", "/admin/"+r.name+"/{id}", "Remover "+r.title, "Administração", nil, nil, false, true, 204, "Obras/edições são arquivadas e gêneros desativados. Autor/editora com referências retorna 409.")
	}
	add("GET", "/admin/catalog-suggestions", "Consultar fila de sugestões", "Administração", nil, schemas.CatalogSuggestion{}, true, true, 200, "Consulta proposta e versão de alvo capturada.", parameter("status", enum("pending", "approved", "rejected"), "Estado de decisão."))
	add("GET", "/admin/catalog-suggestions/{id}", "Consultar sugestão administrativa", "Administração", nil, schemas.CatalogSuggestion{}, false, true, 200, "")
	add("POST", "/admin/catalog-suggestions/{id}/approve", "Aprovar sugestão de catálogo", "Administração", schemas.CatalogDecision{}, schemas.CatalogSuggestion{}, false, true, 200, "Transação com bloqueio da sugestão e do alvo, teste de versão, manutenção, decisão e auditoria. Repetição retorna 409 e não reaplica.")
	add("POST", "/admin/catalog-suggestions/{id}/reject", "Rejeitar sugestão de catálogo", "Administração", schemas.CatalogDecision{}, schemas.CatalogSuggestion{}, false, true, 200, "Decisão fundamentada e única; não altera catálogo.")
	add("GET", "/admin/reports", "Consultar fila de denúncias", "Administração", nil, schemas.Report{}, true, true, 200, "Evidências e identidade do denunciante são restritas à moderação.", parameter("status", enum("open", "in_review", "awaiting_information", "resolved", "dismissed", "canceled"), "Estado da denúncia."))
	add("GET", "/admin/reports/{id}", "Consultar denúncia administrativa", "Administração", nil, schemas.Report{}, false, true, 200, "Inspeção de evidências é auditada.")
	add("PATCH", "/admin/reports/{id}", "Atribuir e decidir denúncia", "Administração", schemas.PatchReport{}, schemas.Report{}, false, true, 200, "Valida transições, atendente com permissão e nota de resolução. Estados terminais não são reabertos.")
	add("GET", "/admin/reports/{id}/messages", "Consultar histórico de denúncia", "Administração", nil, schemas.ReportMessage{}, true, true, 200, "Inclui notas internas restritas.")
	add("POST", "/admin/reports/{id}/messages", "Solicitar informações ou registrar nota", "Administração", schemas.CreateReportMessage{}, schemas.ReportMessage{}, false, true, 201, "visibility=reporter torna a mensagem visível ao denunciante; internal permanece privada. request_information exige visibilidade reporter.")
	for _, kind := range []string{"posts", "comments"} {
		add("GET", "/admin/"+kind+"/{id}", "Inspecionar evidência de "+kind, "Administração", nil, schemas.ContentEvidence{}, false, true, 200, "Consulta privilegiada auditada; permite inspeção de conteúdo oculto/denunciado sem desativar filtros das rotas públicas.")
		add("POST", "/admin/"+kind+"/{id}/moderation", "Moderar "+kind, "Administração", schemas.Moderation{}, nil, false, true, 204, "Oculta/remove/restaura publicação permitida com motivo, auditoria e notificação sem trecho protegido. Conteúdo excluído pelo autor não é restaurado.")
	}
	add("GET", "/admin/penalties", "Consultar punições", "Administração", nil, schemas.Penalty{}, true, true, 200, "Vigência é reavaliada a cada operação protegida.", parameter("user_id", idSchema(), "Filtrar por leitor punido."))
	add("POST", "/admin/penalties", "Aplicar punição a leitor", "Administração", schemas.CreatePenalty{}, schemas.Penalty{}, false, true, 201, "Nenhum administrador pode punir outro administrador. warning registra aviso; restriction bloqueia escritas de comunidade/leitura; suspension bloqueia autenticação durante vigência; ban é permanente. Suspensão/ban revogam sessões. Vínculo a denúncias é opcional.")
	add("POST", "/admin/penalties/{id}/revoke", "Revogar punição", "Administração", schemas.Reason{}, nil, false, true, 204, "Motivo obrigatório. Revogar uma punição não elimina as demais nem reativa conta desativada.")
	add("GET", "/admin/audit-logs", "Consultar trilha administrativa", "Administração", nil, schemas.AuditLog{}, true, true, 200, "Acesso filtrado conforme papel; não inclui senhas/cookies/tokens.", parameter("target_type", stringSchema(), "Tipo do alvo de auditoria."))
	add("GET", "/admin/faq", "Consultar FAQ administrativa", "Administração", nil, schemas.FAQ{}, true, true, 200, "Inclui rascunhos.")
	add("POST", "/admin/faq", "Cadastrar entrada de FAQ", "Administração", schemas.SaveFAQ{}, schemas.FAQ{}, false, true, 201, "question e answer são obrigatórios na criação; estado default draft.")
	all[len(all)-1].requestSchema = "FAQCreate"
	add("GET", "/admin/faq/{id}", "Consultar entrada de FAQ administrativa", "Administração", nil, schemas.FAQ{}, false, true, 200, "")
	add("PATCH", "/admin/faq/{id}", "Editar ou publicar FAQ", "Administração", schemas.SaveFAQ{}, schemas.FAQ{}, false, true, 200, "")
	add("DELETE", "/admin/faq/{id}", "Remover entrada de FAQ", "Administração", nil, nil, false, true, 204, "")
	add("GET", "/admin/support-tickets", "Consultar fila de atendimento", "Administração", nil, schemas.SupportTicket{}, true, true, 200, "Não há chat em tempo real; histórico é persistido.", parameter("status", enum("open", "in_progress", "resolved", "closed"), "Estado de atendimento."))
	add("GET", "/admin/support-tickets/{id}", "Consultar chamado administrativo", "Administração", nil, schemas.SupportTicket{}, false, true, 200, "")
	add("PATCH", "/admin/support-tickets/{id}", "Atribuir/alterar estado de chamado", "Administração", schemas.PatchSupportTicket{}, schemas.SupportTicket{}, false, true, 200, "Valida estado e atendente; reabertura somente pela administração conforme transições permitidas.")
	add("GET", "/admin/support-tickets/{id}/messages", "Consultar conversa administrativa", "Administração", nil, schemas.SupportMessage{}, true, true, 200, "")
	add("POST", "/admin/support-tickets/{id}/messages", "Responder chamado como atendente", "Administração", schemas.CreateSupportMessage{}, schemas.SupportMessage{}, false, true, 201, "Autoria administrativa vem da sessão e é auditada.")
	return all
}
