package application

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestExecuteUseCase_Success(t *testing.T) {
	sb := &stubSandbox{executeResult: `{"id":1}`}
	uc := NewExecuteUseCase(sb, 60*time.Second)
	result, err := uc.Execute(context.Background(), "petstore", `httpGet("/pets/1")`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != `{"id":1}` {
		t.Errorf("result = %q", result)
	}
}

func TestExecuteUseCase_SandboxError(t *testing.T) {
	sb := &stubSandbox{executeErr: errors.New("timeout")}
	uc := NewExecuteUseCase(sb, 60*time.Second)
	_, err := uc.Execute(context.Background(), "petstore", `httpGet("/pets")`)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestExecuteUseCase_AppliesTimeout(t *testing.T) {
	sb := &stubSandbox{executeResult: "ok"}
	uc := NewExecuteUseCase(sb, 60*time.Second)
	_, _ = uc.Execute(context.Background(), "petstore", `"ok"`)
	// Verify the context passed to EvalExecute had a deadline
	if sb.lastCtx == nil {
		t.Fatal("context not passed")
	}
	deadline, ok := sb.lastCtx.Deadline()
	if !ok {
		t.Fatal("expected context to have a deadline")
	}
	if time.Until(deadline) > 61*time.Second {
		t.Errorf("deadline too far: %v", time.Until(deadline))
	}
}

func TestExecuteUseCase_ParentContextCancelled(t *testing.T) {
	sb := &stubSandbox{executeResult: "ok"}
	uc := NewExecuteUseCase(sb, 60*time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	// Even with cancelled parent, Execute should propagate the error via context
	// The sandbox stub just returns, but the derived context should be cancelled
	_, _ = uc.Execute(ctx, "petstore", `"ok"`)
	if sb.lastCtx != nil && sb.lastCtx.Err() == nil {
		t.Error("derived context should reflect parent cancellation")
	}
}

func TestExecuteUseCase_PassesAPIName(t *testing.T) {
	sb := &stubSandbox{executeResult: "ok"}
	uc := NewExecuteUseCase(sb, 60*time.Second)
	_, _ = uc.Execute(context.Background(), "myapi", `httpGet("/")`)
	if sb.lastAPIName != "myapi" {
		t.Errorf("api name passed = %q, want %q", sb.lastAPIName, "myapi")
	}
}
