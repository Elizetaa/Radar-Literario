// Package openapi builds the reviewable HTTP contract from the actual router
// and explicit resource metadata. It never connects to a database.
package openapi

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"radar-literario/backend/internal/model/schemas"
)

type object = map[string]any
type operation struct {
	method, path, summary, tag    string
	request, response             any
	list, private                 bool
	status                        int
	description                   string
	parameters                    []any
	requestSchema, responseSchema string
}
type generator struct{ schemas object }
type health struct {
	Status string `json:"status"`
}
type csrf struct {
	CSRFToken string `json:"csrf_token"`
}
type loginResult struct {
	User      schemas.AccountView `json:"user"`
	CSRFToken string              `json:"csrf_token"`
}
type message struct {
	Message string `json:"message"`
}

func NormalizePath(path string) string {
	parts := strings.Split(path, "/")
	for i, v := range parts {
		if strings.HasPrefix(v, ":") {
			parts[i] = "{" + v[1:] + "}"
		}
	}
	return strings.Join(parts, "/")
}
func ref(name string) object       { return object{"$ref": "#/components/schemas/" + name} }
func stringSchema() object         { return object{"type": "string"} }
func enum(values ...string) object { return object{"type": "string", "enum": values} }
func idSchema() object             { return object{"type": "integer", "format": "int64", "minimum": 1} }
func parameter(name string, schema object, description string) object {
	return object{"name": name, "in": "query", "required": false, "description": description, "schema": schema}
}
func pageParams() []any {
	return []any{parameter("page", object{"type": "integer", "minimum": 1, "maximum": 1000000, "default": 1}, "Página, iniciando em 1."), parameter("page_size", object{"type": "integer", "minimum": 1, "maximum": 100, "default": 20}, "Quantidade máxima de itens da página.")}
}
func revealParams() []any {
	return []any{parameter("reveal_spoilers", object{"type": "boolean", "default": false}, "Confirmação explícita para revelar spoiler; não contorna a idade."), parameter("reveal_sensitive", object{"type": "boolean", "default": false}, "Confirmação explícita para revelar conteúdo sensível; não contorna a idade.")}
}

// Build fails when a registered route lacks explicit contract metadata or when
// metadata describes a route that the router does not expose.
func Build(routes gin.RoutesInfo, cookieName string) ([]byte, error) {
	if cookieName == "" {
		cookieName = "radar_session"
	}
	g := generator{schemas: object{}}
	ops := operations()
	registry := map[string]operation{}
	for _, op := range ops {
		key := op.method + " " + op.path
		if _, exists := registry[key]; exists {
			return nil, fmt.Errorf("duplicate contract operation %s", key)
		}
		registry[key] = op
	}
	paths := object{}
	seen := map[string]bool{}
	sort.Slice(routes, func(i, j int) bool {
		if routes[i].Path == routes[j].Path {
			return routes[i].Method < routes[j].Method
		}
		return routes[i].Path < routes[j].Path
	})
	for _, route := range routes {
		path := NormalizePath(route.Path)
		key := route.Method + " " + path
		op, ok := registry[key]
		if !ok {
			return nil, fmt.Errorf("route without OpenAPI metadata: %s", key)
		}
		seen[key] = true
		item, ok := paths[path].(object)
		if !ok {
			item = object{}
			paths[path] = item
		}
		item[strings.ToLower(route.Method)] = g.operation(op, cookieName)
	}
	for key := range registry {
		if !seen[key] {
			return nil, fmt.Errorf("contract operation without route: %s", key)
		}
	}
	g.customSchemas()
	doc := object{"openapi": "3.0.3", "info": object{"title": "Radar Literário — API", "version": "1.0.0", "description": "API Go/Gin/GORM com persistência MariaDB. IDs são inteiros positivos; datas civis usam YYYY-MM-DD e instantes usam RFC3339 em UTC. JSON de escrita rejeita campos desconhecidos. Listas têm paginação estável e limitada. Perfil, status e classificação etária são reavaliados no servidor. Publicações protegidas retornam metadados com title/text/data_url nulos até confirmação explícita; idade nunca é contornada. Não há verificação documental de nascimento nem verificação de e-mail nesta versão.\n\nSessão opaca em cookie HttpOnly. Obtenha GET /api/v1/auth/csrf antes de escrever: visitante recebe cookie de pré-autenticação e token; usuário autenticado recebe token ligado à sessão. Envie X-CSRF-Token em todas as escritas e preserve cookies. Login substitui a sessão e retorna novo token; logout/troca/recuperação revogam sessões. Origin deve estar nas listas configuradas; CORS permite somente origens explícitas. Cookie Secure é obrigatório em produção.\n\nPapéis: catalog_manager mantém catálogo; user_manager modera, atende suporte e mantém FAQ; administrator reúne os dois; system_administrator também gerencia administradores. Nenhum administrador pode banir outro administrador. A descrição das operações identifica a permissão necessária. Nome do cookie é configurável; este contrato registra o default local.\n\nGeração: em Backend, go run ./cmd/openapi -out ../Docs/api.yaml. O documento é JSON válido, também aceito como YAML 1.2."}, "servers": []any{object{"url": "http://127.0.0.1:8080", "description": "Desenvolvimento local; host/porta configuráveis."}}, "paths": paths, "components": object{"schemas": g.schemas, "securitySchemes": object{"SessionAuth": object{"type": "apiKey", "in": "cookie", "name": cookieName, "description": "Segredo opaco; banco armazena apenas hash. Expira por inatividade e duração absoluta configuráveis."}, "PreAuth": object{"type": "apiKey", "in": "cookie", "name": cookieName + "_preauth", "description": "Cookie de pré-autenticação criado por GET /auth/csrf."}, "CSRF": object{"type": "apiKey", "in": "header", "name": "X-CSRF-Token", "description": "Token obtido em GET /auth/csrf ou na resposta do login."}}}, "tags": []any{object{"name": "Saúde"}, object{"name": "Contas"}, object{"name": "Catálogo"}, object{"name": "Leitura"}, object{"name": "Comunidade"}, object{"name": "Administração"}, object{"name": "Suporte"}}}
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return nil, err
	}
	if err = Validate(data); err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

func (g *generator) operation(op operation, cookie string) object {
	params := append([]any{}, op.parameters...)
	for _, part := range strings.Split(op.path, "/") {
		if strings.HasPrefix(part, "{") {
			params = append(params, object{"name": strings.Trim(part, "{}"), "in": "path", "required": true, "schema": idSchema(), "description": "Identificador positivo do recurso."})
			if strings.Trim(part, "{}") == "isbn" {
				params[len(params)-1].(object)["schema"] = object{"type": "string", "example": "9780306406157"}
			}
		}
	}
	if op.list {
		params = append(params, pageParams()...)
	}
	responses := object{}
	for _, status := range []int{400, 401, 403, 404, 409, 429, 500, 503} {
		responses[strconv.Itoa(status)] = object{"description": map[int]string{400: "Entrada inválida ou restrição de integridade.", 401: "Sessão ausente, expirada, revogada ou conta indisponível.", 403: "Permissão, CSRF ou origem não autorizados.", 404: "Recurso inexistente ou invisível ao solicitante.", 409: "Duplicidade ou alteração concorrente; atualize antes de repetir.", 429: "Limite de tentativas excedido.", 500: "Falha interna sem detalhes sensíveis.", 503: "Dependência temporariamente indisponível."}[status], "content": object{"application/json": object{"schema": ref("ErrorEnvelope")}}}
	}
	response := object{"description": "Operação concluída."}
	if op.status != 204 {
		data := g.schema(reflect.TypeOf(op.response), false)
		if op.responseSchema != "" {
			data = ref(op.responseSchema)
		}
		if op.list {
			data = object{"type": "object", "required": []string{"data", "pagination"}, "properties": object{"data": object{"type": "array", "items": data}, "pagination": g.schema(reflect.TypeOf(schemas.Pagination{}), false)}}
		} else if op.path != "/health" && op.path != "/ready" {
			data = object{"type": "object", "required": []string{"data"}, "properties": object{"data": data}}
		}
		response["content"] = object{"application/json": object{"schema": data}}
	}
	if op.path == "/api/v1/auth/login" || op.path == "/api/v1/auth/csrf" || op.path == "/api/v1/auth/logout" {
		response["headers"] = object{"Set-Cookie": object{"description": "Cookie HttpOnly, Path=/, SameSite configurado, Secure em produção. Login rotaciona o segredo; logout o remove.", "schema": stringSchema()}}
	}
	responses[strconv.Itoa(op.status)] = response
	security := []any{}
	write := op.method != "GET"
	if op.private {
		requirement := object{"SessionAuth": []string{}}
		if write {
			requirement["CSRF"] = []string{}
		}
		security = append(security, requirement)
	} else if write {
		security = append(security, object{"PreAuth": []string{}, "CSRF": []string{}}, object{"SessionAuth": []string{}, "CSRF": []string{}})
	}
	result := object{"operationId": strings.ToLower(op.method) + "_" + strings.NewReplacer("/", "_", "{", "", "}", "", "-", "_").Replace(strings.Trim(op.path, "/")), "summary": op.summary, "tags": []string{op.tag}, "responses": responses, "security": security}
	if len(params) > 0 {
		result["parameters"] = params
	}
	description := op.description
	if strings.Contains(op.path, "/admin/") {
		permission := "moderação/atendimento"
		if strings.Contains(op.path, "/administrators") {
			permission = "administração de sistema"
		} else if strings.Contains(op.path, "/works") || strings.Contains(op.path, "/editions") || strings.Contains(op.path, "/authors") || strings.Contains(op.path, "/publishers") || strings.Contains(op.path, "/genres") || strings.Contains(op.path, "catalog-suggestions") {
			permission = "catálogo"
		} else if strings.Contains(op.path, "audit-logs") {
			permission = "auditoria correspondente ao papel"
		}
		description += " Permissão administrativa exigida: " + permission + ". Operações de escrita são auditadas na transação."
	}
	if description != "" {
		result["description"] = strings.TrimSpace(description)
	}
	if op.request != nil {
		schema := g.schema(reflect.TypeOf(op.request), true)
		if op.requestSchema != "" {
			schema = ref(op.requestSchema)
		}
		result["requestBody"] = object{"required": true, "content": object{"application/json": object{"schema": schema}}}
	}
	return result
}

func (g *generator) schema(t reflect.Type, request bool) object {
	if t == nil {
		return object{"type": "object"}
	}
	if t == reflect.TypeOf(time.Time{}) {
		return object{"type": "string", "format": "date-time"}
	}
	if t == reflect.TypeOf(json.RawMessage{}) {
		return object{"type": "object", "description": "Objeto JSON validado segundo o recurso da proposta."}
	}
	if t.Kind() == reflect.Pointer {
		value := g.schema(t.Elem(), request)
		if reference, ok := value["$ref"]; ok {
			return object{"allOf": []any{object{"$ref": reference}}, "nullable": true}
		}
		value["nullable"] = true
		return value
	}
	if strings.HasPrefix(t.Name(), "Optional[") || t.Name() == "NullableText" || t.Name() == "AdminOptionalID" {
		field, _ := t.FieldByName("Value")
		value := g.schema(field.Type, true)
		value["description"] = "Omitir preserva o valor; null remove o valor."
		return value
	}
	switch t.Kind() {
	case reflect.Bool:
		return object{"type": "boolean"}
	case reflect.String:
		return stringSchema()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return object{"type": "integer", "format": "int64"}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return object{"type": "integer", "format": "int64", "minimum": 0}
	case reflect.Float32, reflect.Float64:
		return object{"type": "number", "format": "double"}
	case reflect.Slice, reflect.Array:
		return object{"type": "array", "items": g.schema(t.Elem(), request)}
	case reflect.Struct:
		name := t.Name()
		if name == "" {
			return object{"type": "object"}
		}
		if _, exists := g.schemas[name]; exists {
			return ref(name)
		}
		g.schemas[name] = object{}
		properties := object{}
		required := []string{}
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			tag := field.Tag.Get("json")
			if tag == "-" || tag == "" {
				continue
			}
			key := strings.Split(tag, ",")[0]
			property := g.schema(field.Type, request)
			g.constraints(name, key, property)
			properties[key] = property
			if !request && !strings.Contains(tag, "omitempty") {
				required = append(required, key)
			}
		}
		if request {
			required = requestRequired[name]
		}
		definition := object{"type": "object", "properties": properties, "additionalProperties": false}
		if len(required) > 0 {
			definition["required"] = required
		}
		g.schemas[name] = definition
		return ref(name)
	}
	return object{"type": "object"}
}

var requestRequired = map[string][]string{
	"RegisterReader": {"name", "nickname", "email", "password", "birth_date"}, "Login": {"email", "password"}, "ForgotPassword": {"email"}, "ResetPassword": {"token", "password"}, "ChangePassword": {"current_password", "new_password"}, "DeleteAccount": {"password"}, "ShelfInput": {"edition_id", "current_page", "status"}, "PostInput": {"text", "contains_spoiler"}, "CommentInput": {"text"}, "CommentPatch": {"text"}, "SuggestionInput": {"operation", "resource", "proposal"}, "CatalogDecision": {"reason"}, "InterestsInput": {"genre_ids"}, "CreateAdministrator": {"role", "reason"}, "PatchAdministrator": {"reason"}, "CreateReport": {"reason", "description"}, "CreateReportMessage": {"text"}, "Moderation": {"status", "reason"}, "CreatePenalty": {"user_id", "type", "reason"}, "Reason": {"reason"}, "CreateSupportTicket": {"subject", "text"}, "CreateSupportMessage": {"text"},
}

func (g *generator) constraints(name, key string, p object) {
	if key == "id" || strings.HasSuffix(key, "_id") {
		p["minimum"] = 1
	}
	if key == "minimum_age" {
		p["enum"] = []int{0, 10, 12, 14, 16, 18}
	}
	if key == "birth_date" || key == "death_date" || key == "first_publication_date" || key == "publication_date" {
		p["type"] = "string"
		p["format"] = "date"
	}
	if key == "email" {
		p["format"] = "email"
		p["maxLength"] = 254
	}
	if key == "nickname" {
		p["pattern"] = `^[a-z0-9_]{3,30}$`
	}
	if key == "password" || key == "new_password" || key == "current_password" {
		p["writeOnly"] = true
		p["format"] = "password"
		p["description"] = "Senha: cadastros/novas senhas exigem 8 a 72 bytes UTF-8."
	}
	if key == "rating" {
		p["minimum"] = 1
		p["maximum"] = 5
	}
	if key == "current_page" {
		p["minimum"] = 0
	}
	if key == "page_count" {
		p["minimum"] = 1
		p["maximum"] = 1000000
	}
	if key == "edition_number" {
		p["minimum"] = 1
		p["maximum"] = 100000
	}
	if key == "isbn" {
		p["type"] = "string"
		p["nullable"] = true
		p["description"] = "ISBN-10/13 válido, normalizado para ISBN-13. null exige isbn_absence_reason."
		p["example"] = "9780306406157"
	}
	if key == "isbn_absence_reason" {
		p["maxLength"] = 300
	}
	if key == "original_language" || key == "language" {
		p["maxLength"] = 35
		p["pattern"] = `^[a-z]{2,3}(-[a-z0-9]{2,8})*$`
		p["example"] = "pt-br"
	}
	if strings.HasSuffix(key, "_url") || key == "website_url" {
		p["maxLength"] = 2048
		p["description"] = "URL HTTP/HTTPS opcional sem credenciais; a API não busca URLs enviadas pelo usuário."
	}
	if key == "contains_spoiler" {
		p["description"] = "Obrigatório explicitamente na criação, inclusive false."
	}
	if key == "profile_visibility" {
		p["enum"] = []string{"public", "registered_users_only", "private"}
	}
	if key == "role" {
		p["enum"] = []string{"user_manager", "catalog_manager", "administrator", "system_administrator"}
	}
	if key == "operation" {
		p["enum"] = []string{"create", "update", "delete"}
	}
	if key == "resource" {
		p["enum"] = []string{"work", "edition"}
	}
	if key == "status" {
		switch name {
		case "Work", "WorkInput", "Edition", "EditionInput":
			p["enum"] = []string{"draft", "pending_review", "published", "rejected", "hidden", "archived"}
		case "PostInput", "PostPatch":
			p["enum"] = []string{"draft", "published"}
		case "Post", "Comment":
			p["enum"] = []string{"draft", "published", "hidden", "removed", "deleted"}
		case "Moderation":
			p["enum"] = []string{"published", "hidden", "removed"}
		case "ShelfInput", "ShelfItem", "StatusCount":
			p["enum"] = []string{"planned", "in_progress", "paused", "completed", "abandoned"}
		case "CatalogSuggestion":
			p["enum"] = []string{"pending", "approved", "rejected"}
		case "Report", "PatchReport":
			p["enum"] = []string{"open", "in_review", "awaiting_information", "resolved", "dismissed", "canceled"}
		case "SupportTicket", "PatchSupportTicket":
			p["enum"] = []string{"open", "in_progress", "resolved", "closed"}
		case "FAQ", "SaveFAQ":
			p["enum"] = []string{"draft", "published"}
		case "PatchAdministrator", "Administrator":
			p["enum"] = []string{"active", "inactive"}
		case "AdministrativeUser":
			p["enum"] = []string{"pending_verification", "active", "inactive", "suspended", "deactivated", "banned"}
		}
	}
	if key == "change_kind" {
		p["enum"] = []string{"advance", "correction", "edition_change"}
		p["default"] = "advance"
	}
	if key == "type" && (name == "CreatePenalty" || name == "Penalty") {
		p["enum"] = []string{"warning", "suspension", "restriction", "ban"}
	}
	if key == "visibility" {
		p["enum"] = []string{"reporter", "internal"}
	}
	if key == "reason" && (name == "CreateReport" || name == "Report") {
		p["enum"] = []string{"spam", "harassment", "hate_speech", "threats", "inappropriate_content", "privacy_violation", "copyright_violation", "impersonation", "misinformation", "other"}
	}
	if key == "proposal" {
		p["anyOf"] = []any{ref("WorkInput"), ref("EditionInput"), object{"type": "object", "maxProperties": 0}}
		delete(p, "type")
		p["description"] = "DTO de obra ou edição segundo resource/operation. Exclusão exige {}. Não aceita autoria ou permissões."
	}
	if key == "text" {
		p["maxLength"] = 10000
	}
	if key == "title" {
		p["maxLength"] = 300
		if name == "PostInput" || name == "PostPatch" || name == "Post" {
			p["maxLength"] = 200
		}
	}
	if key == "author_ids" || key == "genre_ids" {
		p["maxItems"] = 100
		p["uniqueItems"] = true
	}
}

func (g *generator) customSchemas() {
	for _, v := range []any{schemas.WorkInput{}, schemas.EditionInput{}, schemas.PostInput{}, schemas.PostPatch{}} {
		g.schema(reflect.TypeOf(v), true)
	}
	g.schemas["ErrorEnvelope"] = object{"type": "object", "required": []string{"error"}, "properties": object{"error": object{"type": "object", "required": []string{"code", "message"}, "properties": object{"code": object{"type": "string", "example": "INVALID_READING_PROGRESS"}, "message": object{"type": "string", "example": "O progresso não pode ultrapassar o total de páginas da edição."}, "fields": object{"type": "object", "additionalProperties": stringSchema()}}}}}
	for _, kind := range []string{"DiaryInput", "ReviewInput"} {
		base, _ := json.Marshal(g.schemas["PostInput"])
		var schema object
		_ = json.Unmarshal(base, &schema)
		props := schema["properties"].(map[string]any)
		if kind == "DiaryInput" {
			delete(props, "rating")
			schema["required"] = []string{"text", "contains_spoiler", "edition_id", "current_page"}
		} else {
			delete(props, "edition_id")
			delete(props, "current_page")
			delete(props, "occurred_at")
			schema["required"] = []string{"text", "contains_spoiler", "work_id", "rating"}
		}
		g.schemas[kind] = schema
	}
	for _, kind := range []string{"DiaryPatch", "ReviewPatch"} {
		base, _ := json.Marshal(g.schemas["PostPatch"])
		var schema object
		_ = json.Unmarshal(base, &schema)
		props := schema["properties"].(map[string]any)
		if kind == "DiaryPatch" {
			delete(props, "rating")
		} else {
			delete(props, "current_page")
		}
		g.schemas[kind] = schema
	}
	g.schemas["ReadingReportItem"] = object{"oneOf": []any{ref("ShelfItem"), ref("Work")}, "description": "Grupos completed/incomplete/abandoned/desired retornam ShelfItem; preferred/recommended retornam Work."}
	for _, definition := range []struct {
		name, base string
		required   []string
	}{{"WorkCreate", "WorkInput", []string{"title", "original_language"}}, {"EditionCreate", "EditionInput", []string{"work_id", "publisher_id", "title", "page_count", "language"}}, {"InitialEditionCreate", "EditionInput", []string{"publisher_id", "title", "page_count", "language"}}, {"AuthorCreate", "AuthorInput", []string{"name"}}, {"PublisherCreate", "PublisherInput", []string{"name"}}, {"GenreCreate", "GenreInput", []string{"name"}}, {"FAQCreate", "SaveFAQ", []string{"question", "answer"}}} {
		data, _ := json.Marshal(g.schemas[definition.base])
		var schema object
		_ = json.Unmarshal(data, &schema)
		schema["required"] = definition.required
		props := schema["properties"].(map[string]any)
		for _, key := range definition.required {
			property := props[key].(map[string]any)
			delete(property, "nullable")
		}
		if definition.name == "InitialEditionCreate" {
			delete(props, "work_id")
		}
		if definition.name == "WorkCreate" {
			props["initial_edition"] = ref("InitialEditionCreate")
			schema["anyOf"] = []any{object{"properties": object{"status": enum("draft", "pending_review", "rejected", "hidden", "archived")}}, object{"required": []string{"status", "synopsis", "author_ids", "genre_ids", "minimum_age"}, "properties": object{"status": enum("published"), "synopsis": object{"type": "string", "minLength": 1}, "author_ids": object{"type": "array", "minItems": 1, "items": idSchema()}, "genre_ids": object{"type": "array", "minItems": 1, "items": idSchema()}}}}
		}
		if definition.name == "EditionCreate" || definition.name == "InitialEditionCreate" {
			schema["anyOf"] = []any{object{"required": []string{"isbn"}, "properties": object{"isbn": object{"type": "string", "minLength": 10}}}, object{"required": []string{"isbn_absence_reason"}, "properties": object{"isbn_absence_reason": object{"type": "string", "minLength": 1}}}}
		}
		g.schemas[definition.name] = schema
	}
	if report, ok := g.schemas["CreateReport"].(object); ok {
		report["oneOf"] = []any{object{"required": []string{"reported_user_id"}, "not": object{"anyOf": []any{object{"required": []string{"reported_post_id"}}, object{"required": []string{"reported_comment_id"}}}}}, object{"required": []string{"reported_post_id"}, "not": object{"anyOf": []any{object{"required": []string{"reported_user_id"}}, object{"required": []string{"reported_comment_id"}}}}}, object{"required": []string{"reported_comment_id"}, "not": object{"anyOf": []any{object{"required": []string{"reported_user_id"}}, object{"required": []string{"reported_post_id"}}}}}}
	}
}

// Validate checks JSON syntax, every local reference, operation identities and
// parameter declarations. Route correspondence is enforced by Build itself.
func Validate(data []byte) error {
	var doc object
	if err := json.Unmarshal(data, &doc); err != nil {
		return err
	}
	if doc["openapi"] != "3.0.3" {
		return fmt.Errorf("unsupported OpenAPI version")
	}
	paths, ok := doc["paths"].(map[string]any)
	if !ok || len(paths) == 0 {
		return fmt.Errorf("missing paths")
	}
	ids := map[string]bool{}
	for path, value := range paths {
		methods, ok := value.(map[string]any)
		if !ok {
			return fmt.Errorf("invalid path %s", path)
		}
		for method, raw := range methods {
			op, ok := raw.(map[string]any)
			if !ok {
				return fmt.Errorf("invalid operation %s %s", method, path)
			}
			id, _ := op["operationId"].(string)
			if id == "" || ids[id] {
				return fmt.Errorf("invalid/duplicate operationId %s", id)
			}
			ids[id] = true
			if _, ok := op["responses"].(map[string]any); !ok {
				return fmt.Errorf("missing responses %s", id)
			}
			params, _ := op["parameters"].([]any)
			for _, part := range strings.Split(path, "/") {
				if !strings.HasPrefix(part, "{") {
					continue
				}
				name := strings.Trim(part, "{}")
				found := false
				for _, p := range params {
					v, _ := p.(map[string]any)
					if v["name"] == name && v["in"] == "path" && v["required"] == true {
						found = true
					}
				}
				if !found {
					return fmt.Errorf("undeclared path parameter %s in %s", name, path)
				}
			}
		}
	}
	var walk func(any) error
	walk = func(node any) error {
		switch v := node.(type) {
		case map[string]any:
			if raw, ok := v["$ref"].(string); ok {
				if !strings.HasPrefix(raw, "#/") {
					return fmt.Errorf("external reference not supported %s", raw)
				}
				var target any = doc
				for _, part := range strings.Split(raw[2:], "/") {
					m, ok := target.(map[string]any)
					if !ok {
						return fmt.Errorf("invalid reference %s", raw)
					}
					target, ok = m[strings.ReplaceAll(strings.ReplaceAll(part, "~1", "/"), "~0", "~")]
					if !ok {
						return fmt.Errorf("unresolved reference %s", raw)
					}
				}
			}
			for _, child := range v {
				if err := walk(child); err != nil {
					return err
				}
			}
		case []any:
			for _, child := range v {
				if err := walk(child); err != nil {
					return err
				}
			}
		}
		return nil
	}
	return walk(doc)
}
