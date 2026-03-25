package openapi

import (
	"context"
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/nuvivo314/mcp-executor/internal/config"
	"github.com/nuvivo314/mcp-executor/internal/domain/model"
	"github.com/nuvivo314/mcp-executor/internal/domain/port"
)

// registry implements port.ApiRegistry backed by pre-loaded OpenAPI specs.
type registry struct {
	specs     []model.APISpec
	endpoints map[string][]model.Endpoint // keyed by API name
}

// NewRegistry loads all API specs from the given configs and returns an ApiRegistry.
func NewRegistry(cfgs []config.APIConfig) (port.ApiRegistry, error) {
	loader := openapi3.NewLoader()

	r := &registry{
		endpoints: make(map[string][]model.Endpoint, len(cfgs)),
	}

	for _, cfg := range cfgs {
		doc, err := loader.LoadFromFile(cfg.SpecPath)
		if err != nil {
			return nil, fmt.Errorf("loading spec %q for api %q: %w", cfg.SpecPath, cfg.Name, err)
		}
		if err := doc.Validate(context.Background()); err != nil {
			return nil, fmt.Errorf("validating spec %q for api %q: %w", cfg.SpecPath, cfg.Name, err)
		}

		r.specs = append(r.specs, model.APISpec{
			Name:        cfg.Name,
			Description: cfg.Description,
			BaseURL:     cfg.BaseURL,
			SpecPath:    cfg.SpecPath,
		})
		r.endpoints[cfg.Name] = convertEndpoints(doc)
	}

	return r, nil
}

func (r *registry) ListAPIs() []model.APISpec {
	out := make([]model.APISpec, len(r.specs))
	copy(out, r.specs)
	return out
}

func (r *registry) GetAPI(name string) (*model.APISpec, error) {
	for i := range r.specs {
		if r.specs[i].Name == name {
			spec := r.specs[i]
			return &spec, nil
		}
	}
	return nil, fmt.Errorf("api %q not found", name)
}

func (r *registry) GetEndpoints(apiName string) ([]model.Endpoint, error) {
	eps, ok := r.endpoints[apiName]
	if !ok {
		return nil, fmt.Errorf("api %q not found", apiName)
	}
	out := make([]model.Endpoint, len(eps))
	copy(out, eps)
	return out, nil
}
