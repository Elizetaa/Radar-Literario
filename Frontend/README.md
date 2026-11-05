# Frontend do Radar Literário

## Executar localmente

No terminal, inicie a API em um terminal:

```powershell
cd C:\Projetos\Radar-Literario\Backend
go run ./cmd/api
```

Em outro terminal, inicie o frontend:

```powershell
cd C:\Projetos\Radar-Literario\Frontend\radar_literario
npm ci
npm run dev
```

Abra `http://localhost:5173`. O botão **Testar backend** abre uma área de integração que consulta o catálogo e permite criar uma conta, entrar, adicionar obras à estante e publicar diários. As ações pessoais precisam de uma conta autenticada; para conferir cadastro, use um e-mail ainda não cadastrado.

A URL padrão da API é `http://127.0.0.1:8080/api/v1`. Para usar outra, copie `.env.example` para `.env.local`, ajuste `VITE_API_BASE_URL` e reinicie o Vite. A origem do frontend também precisa constar em `CORS_ALLOWED_ORIGINS` e `CSRF_ALLOWED_ORIGINS` no ambiente da API.
