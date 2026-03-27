package main

import (
	"flag"
	"log"
	"log/slog"
	"os"

	"github.com/nuvivo314/mcp-executor/internal/application"
	"github.com/nuvivo314/mcp-executor/internal/config"
	"github.com/nuvivo314/mcp-executor/internal/infrastructure/httpclient"
	mcpinfra "github.com/nuvivo314/mcp-executor/internal/infrastructure/mcp"
	"github.com/nuvivo314/mcp-executor/internal/infrastructure/openapi"
	"github.com/nuvivo314/mcp-executor/internal/infrastructure/sandbox"
	"github.com/nuvivo314/mcp-executor/internal/infrastructure/search"
)

func main() {
	cfgPath := flag.String("config", "config.yaml", "path to config file")
	verbose := flag.Bool("verbose", false, "enable debug logging")
	flag.Parse()

	level := slog.LevelInfo
	if *verbose {
		level = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})))

	slog.Debug("starting mcp-executor", "config", *cfgPath, "verbose", *verbose)

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatalf("loading config: %v", err)
	}
	slog.Debug("config loaded", "apis", len(cfg.APIs), "transport", cfg.Server.Transport, "address", cfg.Server.Address)

	// Infrastructure: adapters
	registry, err := openapi.NewRegistry(cfg.APIs)
	if err != nil {
		log.Fatalf("initialising API registry: %v", err)
	}
	slog.Debug("API registry initialised", "count", len(cfg.APIs))

	searchEngine := search.NewFuzzySearch(registry)

	allowedHosts := cfg.Sandbox.AllowedHosts
	if len(allowedHosts) == 0 {
		for _, api := range cfg.APIs {
			allowedHosts = append(allowedHosts, hostOf(api.BaseURL))
		}
	}
	slog.Debug("sandbox allowed hosts", "hosts", allowedHosts)
	httpClient := httpclient.NewHTTPClient(allowedHosts, cfg.Sandbox.MaxResponseBodyBytes)

	sb := sandbox.NewSandbox(registry, searchEngine, httpClient)

	// Application: use cases
	listUC := application.NewListApiUseCase(registry)
	searchUC := application.NewSearchUseCase(sb)
	executeUC := application.NewExecuteUseCase(sb, cfg.Sandbox.ExecutionTimeout)

	// Infrastructure: MCP wiring
	handlers := mcpinfra.NewHandlers(listUC, searchUC, executeUC)
	mcpSrv := mcpinfra.NewMCPServer(handlers)

	transport, err := mcpinfra.NewStreamableServer(mcpSrv, cfg.Server)
	if err != nil {
		log.Fatalf("creating transport: %v", err)
	}

	slog.Info("mcp-executor listening", "address", cfg.Server.Address, "transport", cfg.Server.Transport)
	if err := transport.Start(cfg.Server.Address); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

// hostOf extracts the hostname (without scheme) from a URL string.
func hostOf(rawURL string) string {
	// Simple extraction: strip scheme prefix.
	for _, prefix := range []string{"https://", "http://"} {
		if len(rawURL) > len(prefix) && rawURL[:len(prefix)] == prefix {
			rest := rawURL[len(prefix):]
			for i, c := range rest {
				if c == '/' {
					return rest[:i]
				}
			}
			return rest
		}
	}
	return rawURL
}
