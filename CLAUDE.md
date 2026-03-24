# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Language

- **Code, comments, identifiers, commit messages**: English (en-UK) — e.g. `initialise`, `serialise`, `colour`
- **Conversations with the user**: Français inclusif avec point médian (ex : « les développeur·euses »)

## Commands

```bash
# Build
go build ./...

# Run
go run main.go

# Test all
go test ./...

# Test a single package
go test ./internal/application/...

# Test a single test
go test -run TestName ./internal/...

# Lint (requires golangci-lint)
golangci-lint run
```

## Architecture

Hexagonal architecture. Dependencies flow inward: `infrastructure` → `application` → `domain`.

```
domain/model/     — pure value objects (APISpec, Endpoint, SearchResult, ExecutionResult)
domain/port/      — interfaces (ApiRegistry, SearchEngine, Sandbox, HttpClient)
application/      — use cases (ListApiUseCase, SearchUseCase, ExecuteUseCase)
infrastructure/   — adapters implementing the ports
  openapi/        — kin-openapi loader → ApiRegistry
  search/         — sahilm/fuzzy → SearchEngine
  sandbox/        — QuickJS/WASM runtime → Sandbox
  httpclient/     — net/http → HttpClient
  mcp/            — mark3labs/mcp-go wiring (tools, prompts, server)
config/           — YAML config structs + loader
```

## MCP Tools exposed

| Tool | Parameters | Purpose |
|------|-----------|---------|
| `list_api` | — | List configured APIs |
| `search` | `api_name`, `exec_code` | Run JS with search helpers |
| `execute` | `api_name`, `exec_code` | Run JS with HTTP helpers (60 s timeout) |

## JS Sandbox design

A **fresh QuickJS runtime per request** runs via `modernc.org/quickjs`. The sandbox has zero access to the host filesystem or network — all HTTP calls are proxied through Go.

- **search context** injects: `search(query)`, `getEndpoints()`, `getSpec()`
- **execute context** injects: `httpGet`, `httpPost`, `httpPut`, `httpPatch`, `httpDelete`, `httpRequest`, plus the search helpers
- Credentials are injected by Go; they are **never** exposed to JS
- Host allowlisting prevents SSRF (`sandbox.allowed_hosts` in config, defaults to each API's `base_url`)
- Response bodies are capped at 1 MiB via `io.LimitReader`
- The 60 s timeout wraps the **entire invocation** via `context.Context`, not individual HTTP calls

## Configuration

`config.yaml` at project root — see `.claude/plans/mcp-executor.md` for the full schema. Transport is `streamable-http` by default (can be set to `sse`). No stdio transport.

## Key decisions

- Port interfaces live in `domain/port/` — always depend on these, never on concrete adapters
- `ExecuteUseCase` enforces the 60 s deadline via a derived `context.WithTimeout`
- Module path: `github.com/nuvivo314/mcp-executor`
