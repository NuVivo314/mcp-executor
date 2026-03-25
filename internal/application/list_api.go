package application

import (
	"github.com/nuvivo314/mcp-executor/internal/domain/model"
	"github.com/nuvivo314/mcp-executor/internal/domain/port"
)

// ListApiUseCase returns all registered API specs.
type ListApiUseCase struct {
	registry port.ApiRegistry
}

func NewListApiUseCase(registry port.ApiRegistry) *ListApiUseCase {
	return &ListApiUseCase{registry: registry}
}

func (uc *ListApiUseCase) Execute() []model.APISpec {
	return uc.registry.ListAPIs()
}
