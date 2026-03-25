package application

import (
	"context"
	"errors"
	"testing"
)

func TestSearchUseCase_Success(t *testing.T) {
	sb := &stubSandbox{searchResult: `[{"operationId":"listPets"}]`}
	uc := NewSearchUseCase(sb)
	result, err := uc.Execute(context.Background(), "petstore", `search("pets")`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != `[{"operationId":"listPets"}]` {
		t.Errorf("result = %q", result)
	}
}

func TestSearchUseCase_SandboxError(t *testing.T) {
	sb := &stubSandbox{searchErr: errors.New("js syntax error")}
	uc := NewSearchUseCase(sb)
	_, err := uc.Execute(context.Background(), "petstore", `{{{`)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSearchUseCase_PassesAPIName(t *testing.T) {
	sb := &stubSandbox{searchResult: "ok"}
	uc := NewSearchUseCase(sb)
	_, _ = uc.Execute(context.Background(), "myapi", `getEndpoints()`)
	if sb.lastAPIName != "myapi" {
		t.Errorf("api name passed = %q, want %q", sb.lastAPIName, "myapi")
	}
}

func TestSearchUseCase_PassesCode(t *testing.T) {
	sb := &stubSandbox{searchResult: "ok"}
	uc := NewSearchUseCase(sb)
	code := `search("create")`
	_, _ = uc.Execute(context.Background(), "petstore", code)
	if sb.lastCode != code {
		t.Errorf("code passed = %q, want %q", sb.lastCode, code)
	}
}
