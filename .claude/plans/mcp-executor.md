# MCP Executor - Plan

## Context

Serveur MCP HTTP en Go exposant 3 tools (`list_api`, `search`, `execute`) permettant a un LLM d'explorer et d'appeler des APIs OpenAPI via du code JavaScript sandboxe. QuickJS tourne dans WASM (via Wazero) pour isoler l'execution du host. Architecture hexagonale, nommage en-UK, pas de stdio.

## Approche de developpement

**Architecture hexagonale + TDD (Test-Driven Development)**

- Cycle **Red-Green-Refactor** : pour chaque composant, ecrire d'abord les tests (Red), puis l'implementation minimale pour les faire passer (Green), puis refactoriser si necessaire (Refactor).
- Les **ports** (interfaces dans `domain/port/`) facilitent le mocking : chaque use case et adaptateur est testable en isolation grace a des mocks/stubs des interfaces.
- Convention : les fichiers `*_test.go` sont crees **avant** le fichier d'implementation correspondant.
- `go test ./...` doit passer a chaque etape avant de committer.
- Les tests d'integration valident le cablage entre couches en fin de cycle.

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

Chaque tache d'implementation suit le cycle TDD : tests d'abord (#N), puis implementation (#Nb).

| # | Tache | Statut |
|---|-------|--------|
| 1 | Scaffold: go mod init + directory structure | ✅ done |
| 2 | 🧪 Tests Config: tests pour le loader YAML | ✅ done |
| 2b | Config: structs YAML + loader | ✅ done |
| 3 | 🧪 Tests Domain: tests des value objects | ✅ done |
| 3b | Domain: value objects (model) | ✅ done |
| 4 | Domain: port interfaces (pas de tests — interfaces pures) | ✅ done |
| 5 | 🧪 Tests OpenAPI: tests avec spec fixture | ✅ done |
| 5b | Infra: OpenAPI loader + converter | ✅ done |
| 6 | 🧪 Tests Fuzzy search: tests du search adapter | ✅ done |
| 6b | Infra: fuzzy search adapter | ✅ done |
| 7 | 🧪 Tests HTTP client: tests avec mock server | ✅ done |
| 7b | Infra: HTTP client adapter | ✅ done |
| 8 | 🧪 Tests QuickJS sandbox: tests runtime + helpers | ✅ done |
| 8b | Infra: QuickJS sandbox - runtime + helpers | ✅ done |
| 9 | 🧪 Tests Search context: tests sandbox search | ✅ done (couvert par #8) |
| 9b | Infra: QuickJS sandbox - search context | ✅ done (couvert par #8b) |
| 10 | 🧪 Tests Execute context: tests sandbox execute | ✅ done (couvert par #8) |
| 10b | Infra: QuickJS sandbox - execute context | ✅ done (couvert par #8b) |
| 11 | 🧪 Tests Use cases: tests avec mocks des ports | ✅ done |
| 11b | Application: 3 use cases | ✅ done |
| 12 | 🧪 Tests MCP: tests tools + prompts | ✅ done |
| 12b | Infra: MCP tools + prompts | ✅ done |
| 13 | Infra: MCP server wiring | ✅ done |
| 14 | Entrypoint: main.go + install deps | ✅ done |
| 15 | Config exemple + spec de test | ✅ done |
| 16 | 🧪 Tests Integration: tests end-to-end | ✅ done |

## Verification finale

- `go test ./...` — tous les tests passent (unitaires + integration)
- `go run main.go` demarre le serveur
- Tester les 3 tools avec un client MCP HTTP
- Verifier le timeout (JS qui depasse 60s)
- Verifier le host allowlisting (appel vers host non autorise)
