package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the full application configuration.
type Config struct {
	Server  ServerConfig `yaml:"server"`
	APIs    []APIConfig  `yaml:"apis"`
	Sandbox SandboxConfig `yaml:"sandbox"`
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Address   string `yaml:"address"`
	Transport string `yaml:"transport"` // "streamable-http" | "sse"
}

// APIConfig describes a single configured API.
type APIConfig struct {
	Name        string     `yaml:"name"`
	Description string     `yaml:"description"`
	SpecPath    string     `yaml:"spec_path"`
	BaseURL     string     `yaml:"base_url"`
	Auth        AuthConfig `yaml:"auth"`
}

// AuthConfig holds authentication settings for an API.
type AuthConfig struct {
	Type     string `yaml:"type"`      // "bearer" | "api-key" | "basic" | "none"
	TokenEnv string `yaml:"token_env"` // env var name for bearer token
	KeyEnv   string `yaml:"key_env"`   // env var name for API key
	Header   string `yaml:"header"`    // header name for api-key (default: X-API-Key)
	Username string `yaml:"username"`
	Password string `yaml:"password_env"` // env var name for basic auth password
}

// SandboxConfig holds sandbox execution settings.
type SandboxConfig struct {
	ExecutionTimeout    time.Duration `yaml:"execution_timeout"`
	MaxResponseBodyBytes int64        `yaml:"max_response_body_bytes"`
	AllowedHosts        []string      `yaml:"allowed_hosts"`
}

// Load reads and parses the YAML config file at the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file %q: %w", path, err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config file %q: %w", path, err)
	}

	cfg.applyDefaults()

	return cfg, nil
}

func (c *Config) applyDefaults() {
	if c.Server.Address == "" {
		c.Server.Address = ":8080"
	}
	if c.Server.Transport == "" {
		c.Server.Transport = "streamable-http"
	}
	if c.Sandbox.ExecutionTimeout == 0 {
		c.Sandbox.ExecutionTimeout = 60 * time.Second
	}
	if c.Sandbox.MaxResponseBodyBytes == 0 {
		c.Sandbox.MaxResponseBodyBytes = 1 << 20 // 1 MiB
	}
	for i := range c.APIs {
		if c.APIs[i].Auth.Header == "" {
			c.APIs[i].Auth.Header = "X-API-Key"
		}
	}
}
