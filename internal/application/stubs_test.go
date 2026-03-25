package application

import (
	"context"
	"errors"

	"github.com/nuvivo314/mcp-executor/internal/domain/model"
)

// stubRegistry implements port.ApiRegistry for tests.
type stubRegistry struct {
	apis []model.APISpec
}

func (s *stubRegistry) ListAPIs() []model.APISpec {
	return s.apis
}

func (s *stubRegistry) GetAPI(name string) (*model.APISpec, error) {
	for _, a := range s.apis {
		if a.Name == name {
			spec := a
			return &spec, nil
		}
	}
	return nil, errors.New("not found")
}

func (s *stubRegistry) GetEndpoints(apiName string) ([]model.Endpoint, error) {
	return nil, nil
}

// stubSandbox implements port.Sandbox for tests.
type stubSandbox struct {
	searchResult string
	searchErr    error
	executeResult string
	executeErr    error
	lastCode    string
	lastAPIName string
	lastCtx     context.Context
}

func (s *stubSandbox) EvalSearch(ctx context.Context, code, apiName string) (string, error) {
	s.lastCode = code
	s.lastAPIName = apiName
	s.lastCtx = ctx
	return s.searchResult, s.searchErr
}

func (s *stubSandbox) EvalExecute(ctx context.Context, code, apiName string) (string, error) {
	s.lastCode = code
	s.lastAPIName = apiName
	s.lastCtx = ctx
	return s.executeResult, s.executeErr
}
