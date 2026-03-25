package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeTestConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writing test config: %v", err)
	}
	return path
}

func TestLoad_FullConfig(t *testing.T) {
	yaml := `
server:
  address: ":9090"
  transport: "sse"
apis:
  - name: "petstore"
    description: "Pet API"
    spec_path: "./specs/petstore.yaml"
    base_url: "https://petstore.example.com"
    auth:
      type: "bearer"
      token_env: "PET_TOKEN"
sandbox:
  execution_timeout: 30s
  max_response_body_bytes: 2097152
  allowed_hosts:
    - "petstore.example.com"
`
	cfg, err := Load(writeTestConfig(t, yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Server.Address != ":9090" {
		t.Errorf("address = %q, want %q", cfg.Server.Address, ":9090")
	}
	if cfg.Server.Transport != "sse" {
		t.Errorf("transport = %q, want %q", cfg.Server.Transport, "sse")
	}
	if len(cfg.APIs) != 1 {
		t.Fatalf("len(apis) = %d, want 1", len(cfg.APIs))
	}
	api := cfg.APIs[0]
	if api.Name != "petstore" {
		t.Errorf("api name = %q, want %q", api.Name, "petstore")
	}
	if api.Auth.Type != "bearer" {
		t.Errorf("auth type = %q, want %q", api.Auth.Type, "bearer")
	}
	if api.Auth.TokenEnv != "PET_TOKEN" {
		t.Errorf("token_env = %q, want %q", api.Auth.TokenEnv, "PET_TOKEN")
	}
	if cfg.Sandbox.ExecutionTimeout != 30*time.Second {
		t.Errorf("timeout = %v, want 30s", cfg.Sandbox.ExecutionTimeout)
	}
	if cfg.Sandbox.MaxResponseBodyBytes != 2097152 {
		t.Errorf("max_response_body_bytes = %d, want 2097152", cfg.Sandbox.MaxResponseBodyBytes)
	}
	if len(cfg.Sandbox.AllowedHosts) != 1 || cfg.Sandbox.AllowedHosts[0] != "petstore.example.com" {
		t.Errorf("allowed_hosts = %v, want [petstore.example.com]", cfg.Sandbox.AllowedHosts)
	}
}

func TestLoad_Defaults(t *testing.T) {
	yaml := `
apis:
  - name: "myapi"
    spec_path: "./spec.yaml"
    base_url: "https://api.example.com"
    auth:
      type: "api-key"
      key_env: "MY_KEY"
`
	cfg, err := Load(writeTestConfig(t, yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Server.Address != ":8080" {
		t.Errorf("default address = %q, want %q", cfg.Server.Address, ":8080")
	}
	if cfg.Server.Transport != "streamable-http" {
		t.Errorf("default transport = %q, want %q", cfg.Server.Transport, "streamable-http")
	}
	if cfg.Sandbox.ExecutionTimeout != 60*time.Second {
		t.Errorf("default timeout = %v, want 60s", cfg.Sandbox.ExecutionTimeout)
	}
	if cfg.Sandbox.MaxResponseBodyBytes != 1<<20 {
		t.Errorf("default max_response_body_bytes = %d, want %d", cfg.Sandbox.MaxResponseBodyBytes, 1<<20)
	}
	if cfg.APIs[0].Auth.Header != "X-API-Key" {
		t.Errorf("default auth header = %q, want %q", cfg.APIs[0].Auth.Header, "X-API-Key")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	yaml := `{{{not valid yaml`
	_, err := Load(writeTestConfig(t, yaml))
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
}

func TestLoad_EmptyFile(t *testing.T) {
	cfg, err := Load(writeTestConfig(t, ""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Defaults should still apply
	if cfg.Server.Address != ":8080" {
		t.Errorf("default address = %q, want %q", cfg.Server.Address, ":8080")
	}
	if cfg.Sandbox.ExecutionTimeout != 60*time.Second {
		t.Errorf("default timeout = %v, want 60s", cfg.Sandbox.ExecutionTimeout)
	}
}

func TestLoad_MultipleAPIs(t *testing.T) {
	yaml := `
apis:
  - name: "api1"
    spec_path: "./spec1.yaml"
    base_url: "https://api1.example.com"
    auth:
      type: "none"
  - name: "api2"
    spec_path: "./spec2.yaml"
    base_url: "https://api2.example.com"
    auth:
      type: "basic"
      username: "user"
      password_env: "API2_PASS"
`
	cfg, err := Load(writeTestConfig(t, yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.APIs) != 2 {
		t.Fatalf("len(apis) = %d, want 2", len(cfg.APIs))
	}
	if cfg.APIs[0].Name != "api1" {
		t.Errorf("api[0].name = %q, want %q", cfg.APIs[0].Name, "api1")
	}
	if cfg.APIs[1].Auth.Type != "basic" {
		t.Errorf("api[1].auth.type = %q, want %q", cfg.APIs[1].Auth.Type, "basic")
	}
	if cfg.APIs[1].Auth.Password != "API2_PASS" {
		t.Errorf("api[1].auth.password_env = %q, want %q", cfg.APIs[1].Auth.Password, "API2_PASS")
	}
	// Both should get default header
	for i, api := range cfg.APIs {
		if api.Auth.Header != "X-API-Key" {
			t.Errorf("api[%d].auth.header = %q, want %q", i, api.Auth.Header, "X-API-Key")
		}
	}
}
