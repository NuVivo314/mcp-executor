# MCP Executor - Plan

## Context

Serveur MCP HTTP en Go exposant 3 tools (`list_api`, `search`, `execute`) permettant a un LLM d'explorer et d'appeler des APIs OpenAPI via du code JavaScript sandboxe. QuickJS tourne dans WASM (via Wazero) pour isoler l'execution du host. Architecture hexagonale, nommage en-UK, pas de stdio.

## Libraries

| Role | Library | Import |
|------|---------|--------|
| MCP server | mark3labs/mcp-go | `github.com/mark3labs/mcp-go` |
| QuickJS | modernc.org/quickjs | `modernc.org/quickjs` |
| OpenAPI parsing | getkin/kin-openapi | `github.com/getkin/kin-openapi/openapi3` |
| Fuzzy search | sahilm/fuzzy | `github.com/sahilm/fuzzy` |
| YAML | gopkg.in/yaml.v3 | `gopkg.in/yaml.v3` |

## Directory Structure

```
mcp-executor/
  main.go
  config.yaml

  internal/
    config/
      config.go                          # YAML config structs + loader

    domain/
      model/
        api.go                           # APISpec, Endpoint, Parameter
        search_result.go                 # SearchResult
        execution_result.go              # ExecutionResult
      port/
        api_registry.go                  # ApiRegistry interface
        search_engine.go                 # SearchEngine interface
        sandbox.go                       # Sandbox interface
        http_client.go                   # HttpClient interface

    application/
      list_api.go                        # ListApiUseCase
      search.go                          # SearchUseCase
      execute.go                         # ExecuteUseCase (60s timeout)

    infrastructure/
      openapi/
        loader.go                        # kin-openapi loader -> ApiRegistry
        converter.go                     # openapi3.T -> domain model
      search/
        fuzzy.go                         # sahilm/fuzzy -> SearchEngine
      sandbox/
        runtime.go                       # QuickJS WASM runtime lifecycle
        search_sandbox.go                # JS context avec search(), getEndpoints(), getSpec()
        execute_sandbox.go               # JS context avec http*, search(), getEndpoints()
        js_helpers.go                    # Go<->JS conversion helpers
      httpclient/
        client.go                        # net/http wrapper -> HttpClient
      mcp/
        server.go                        # MCP server wiring + transport
        tools.go                         # Tool definitions + handlers
        prompts.go                       # MCP prompt (search example)
```

## Configuration (config.yaml)

```yaml
server:
  address: ":8080"
  transport: "streamable-http"  # or "sse"

apis:
  - name: "petstore"
    description: "Petstore sample API"
    spec_path: "./specs/petstore.yaml"
    base_url: "https://petstore.example.com"
    auth:
      type: "bearer"           # bearer | api-key | basic | none
      token_env: "PETSTORE_TOKEN"

sandbox:
  execution_timeout: 60s
  max_response_body_bytes: 1048576  # 1 MiB
  allowed_hosts: []                 # vide = uniquement base_url de chaque API
```

## Ports (interfaces)

- **ApiRegistry**: `ListAPIs() []APISpec`, `GetAPI(name) (*APISpec, error)`, `GetEndpoints(apiName) ([]Endpoint, error)`
- **SearchEngine**: `Search(apiName, query string) ([]SearchResult, error)`
- **Sandbox**: `EvalSearch(ctx, code, apiName) (string, error)`, `EvalExecute(ctx, code, apiName) (string, error)`
- **HttpClient**: `Do(ctx, method, url string, headers map[string]string, body string) (*ExecutionResult, error)`

## JS Sandbox Design

**Isolation**: Runtime QuickJS frais par requete via `modernc.org/quickjs`. Zero acces au filesystem/network du host. Credentials injectees cote Go, jamais exposees au JS.

### Fonctions injectees dans `search`:
- `search(query)` -> fuzzy match sur les endpoints, retourne `[{path, method, operationId, summary, score}]`
- `getEndpoints()` -> liste complete des endpoints
- `getSpec()` -> spec complete en JSON

### Fonctions injectees dans `execute`:
- `httpGet(path, headers?)`, `httpPost(path, body, headers?)`, `httpPut`, `httpPatch`, `httpDelete`
- `httpRequest(method, path, body?, headers?)` - version generique
- `search(query)` et `getEndpoints()` aussi disponibles

### Securite:
- Host allowlisting (SSRF prevention)
- Auth injectee cote Go (jamais dans le JS)
- Timeout 60s via context.Context (couvre toute l'invocation JS, pas par appel HTTP)
- Response body cap (1 MiB via io.LimitReader)
- Runtime detruit apres chaque requete

## MCP Tools

1. **list_api** - Aucun parametre. Retourne la liste des APIs configurees.
2. **search** - Params: `api_name` (required), `exec_code` (required). Execute le JS avec les helpers de recherche.
3. **execute** - Params: `api_name` (required), `exec_code` (required). Execute le JS avec les helpers HTTP. Timeout 60s.

## MCP Prompt

Un prompt `search_api` avec args `api_name` et `query` qui genere un exemple de code JS pour la recherche.

## Taches et avancement

| # | Tache | Statut |
|---|-------|--------|
| 1 | Scaffold: go mod init + directory structure | ✅ done |
| 2 | Config: structs YAML + loader | ⬜ pending |
| 3 | Domain: value objects (model) | ⬜ pending |
| 4 | Domain: port interfaces | ⬜ pending (blocked by #3) |
| 5 | Infra: OpenAPI loader + converter | ⬜ pending (blocked by #4) |
| 6 | Infra: fuzzy search adapter | ⬜ pending (blocked by #4) |
| 7 | Infra: HTTP client adapter | ⬜ pending (blocked by #4) |
| 8 | Infra: QuickJS sandbox - runtime + helpers | ⬜ pending (blocked by #4) |
| 9 | Infra: QuickJS sandbox - search context | ⬜ pending (blocked by #5,6,8) |
| 10 | Infra: QuickJS sandbox - execute context | ⬜ pending (blocked by #5,6,7,8) |
| 11 | Application: 3 use cases | ⬜ pending (blocked by #9,10) |
| 12 | Infra: MCP tools + prompts | ⬜ pending (blocked by #11) |
| 13 | Infra: MCP server wiring | ⬜ pending (blocked by #12) |
| 14 | Entrypoint: main.go + install deps | ⬜ pending (blocked by #2,13) |
| 15 | Config exemple + spec de test | ⬜ pending (blocked by #14) |

## Verification finale

- `go run main.go` demarre le serveur
- Tester les 3 tools avec un client MCP HTTP
- Verifier le timeout (JS qui depasse 60s)
- Verifier le host allowlisting (appel vers host non autorise)
