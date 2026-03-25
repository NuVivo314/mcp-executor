package application

import (
	"testing"

	"github.com/nuvivo314/mcp-executor/internal/domain/model"
)

func TestListApiUseCase_Empty(t *testing.T) {
	reg := &stubRegistry{apis: nil}
	uc := NewListApiUseCase(reg)
	got := uc.Execute()
	if len(got) != 0 {
		t.Errorf("len = %d, want 0", len(got))
	}
}

func TestListApiUseCase_ReturnsAll(t *testing.T) {
	reg := &stubRegistry{apis: []model.APISpec{
		{Name: "petstore", BaseURL: "https://petstore.example.com"},
		{Name: "myapi", BaseURL: "https://api.example.com"},
	}}
	uc := NewListApiUseCase(reg)
	got := uc.Execute()
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].Name != "petstore" {
		t.Errorf("got[0].Name = %q, want %q", got[0].Name, "petstore")
	}
	if got[1].Name != "myapi" {
		t.Errorf("got[1].Name = %q, want %q", got[1].Name, "myapi")
	}
}
