package application

import (
	"context"

	"github.com/nuvivo314/mcp-executor/internal/domain/port"
)

// SearchUseCase executes JS code in the search sandbox context.
type SearchUseCase struct {
	sandbox port.Sandbox
}

func NewSearchUseCase(sandbox port.Sandbox) *SearchUseCase {
	return &SearchUseCase{sandbox: sandbox}
}

func (uc *SearchUseCase) Execute(ctx context.Context, apiName, code string) (string, error) {
	return uc.sandbox.EvalSearch(ctx, code, apiName)
}
