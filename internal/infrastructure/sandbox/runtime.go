package sandbox

import (
	"context"
	"fmt"

	"github.com/nuvivo314/mcp-executor/internal/domain/port"
	"modernc.org/quickjs"
)

// Sandbox implements port.Sandbox using a fresh QuickJS VM per request.
type Sandbox struct {
	registry port.ApiRegistry
	search   port.SearchEngine
	http     port.HttpClient
}

// NewSandbox creates a Sandbox with the given dependencies.
func NewSandbox(registry port.ApiRegistry, search port.SearchEngine, http port.HttpClient) *Sandbox {
	return &Sandbox{registry: registry, search: search, http: http}
}

// EvalSearch runs JS code with search helpers: search(), getEndpoints(), getSpec().
func (s *Sandbox) EvalSearch(ctx context.Context, code, apiName string) (string, error) {
	if _, err := s.registry.GetAPI(apiName); err != nil {
		return "", fmt.Errorf("unknown api %q: %w", apiName, err)
	}

	vm, err := quickjs.NewVM()
	if err != nil {
		return "", fmt.Errorf("creating js vm: %w", err)
	}
	cancelTimeout := applyTimeout(ctx, vm)
	defer func() {
		cancelTimeout()
		vm.Close()
	}()

	injectSearchHelpers(vm, s, apiName)

	return evalJS(ctx, vm, code)
}

// EvalExecute runs JS code with HTTP helpers and search helpers.
func (s *Sandbox) EvalExecute(ctx context.Context, code, apiName string) (string, error) {
	spec, err := s.registry.GetAPI(apiName)
	if err != nil {
		return "", fmt.Errorf("unknown api %q: %w", apiName, err)
	}

	vm, err := quickjs.NewVM()
	if err != nil {
		return "", fmt.Errorf("creating js vm: %w", err)
	}
	cancelTimeout := applyTimeout(ctx, vm)
	defer func() {
		cancelTimeout()
		vm.Close()
	}()

	injectSearchHelpers(vm, s, apiName)
	injectHTTPHelpers(ctx, vm, s, spec.BaseURL)

	return evalJS(ctx, vm, code)
}

// applyTimeout sets an eval timeout on the VM if the context has a deadline.
// Returns a cancel func that must be called when the VM is no longer needed.
func applyTimeout(ctx context.Context, vm *quickjs.VM) (cancel func()) {
	if deadline, ok := ctx.Deadline(); ok {
		_ = vm.SetEvalTimeout(deadline.Sub(deadlineNow()))
	}

	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			vm.Interrupt()
		case <-done:
		}
	}()
	return func() { close(done) }
}

// evalJS evaluates code and converts the result to a string.
func evalJS(ctx context.Context, vm *quickjs.VM, code string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("context error before eval: %w", err)
	}

	r, err := vm.Eval(code, quickjs.EvalGlobal)
	if err != nil {
		return "", fmt.Errorf("js error: %w", err)
	}

	return anyToString(r), nil
}

// anyToString converts a Go value returned by Eval to a string.
func anyToString(v any) string {
	if v == nil {
		return "null"
	}
	switch t := v.(type) {
	case string:
		return t
	case quickjs.Undefined:
		return "undefined"
	case quickjs.Unsupported:
		return "[unsupported]"
	case fmt.Stringer:
		return t.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}
