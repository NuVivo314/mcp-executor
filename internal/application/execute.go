package application

import (
	"context"
	"time"

	"github.com/nuvivo314/mcp-executor/internal/domain/port"
)

// ExecuteUseCase executes JS code in the execute sandbox context with a timeout.
type ExecuteUseCase struct {
	sandbox port.Sandbox
	timeout time.Duration
}

func NewExecuteUseCase(sandbox port.Sandbox, timeout time.Duration) *ExecuteUseCase {
	return &ExecuteUseCase{sandbox: sandbox, timeout: timeout}
}

func (uc *ExecuteUseCase) Execute(ctx context.Context, apiName, code string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()
	return uc.sandbox.EvalExecute(ctx, code, apiName)
}
