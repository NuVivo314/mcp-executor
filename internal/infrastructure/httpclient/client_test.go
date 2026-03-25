package httpclient

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDo_GET(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want GET", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{"name":"Fido"}`)
	}))
	defer srv.Close()

	client := NewHTTPClient([]string{srv.Listener.Addr().String()}, 1<<20)
	resp, err := client.Do(context.Background(), "GET", srv.URL+"/pets/1", nil, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	if !strings.Contains(resp.Body, "Fido") {
		t.Errorf("body = %q, want to contain 'Fido'", resp.Body)
	}
}

func TestDo_POST(t *testing.T) {
	var receivedBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		receivedBody = string(b)
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{"id":42}`)
	}))
	defer srv.Close()

	client := NewHTTPClient([]string{srv.Listener.Addr().String()}, 1<<20)
	resp, err := client.Do(context.Background(), "POST", srv.URL+"/pets", nil, `{"name":"Rex"}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 201 {
		t.Errorf("status = %d, want 201", resp.StatusCode)
	}
	if receivedBody != `{"name":"Rex"}` {
		t.Errorf("received body = %q, want %q", receivedBody, `{"name":"Rex"}`)
	}
}

func TestDo_Headers(t *testing.T) {
	var receivedAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := NewHTTPClient([]string{srv.Listener.Addr().String()}, 1<<20)
	headers := map[string]string{"Authorization": "Bearer secret-token"}
	_, err := client.Do(context.Background(), "GET", srv.URL+"/pets", headers, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if receivedAuth != "Bearer secret-token" {
		t.Errorf("Authorization = %q, want %q", receivedAuth, "Bearer secret-token")
	}
}

func TestDo_ResponseHeaders(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{}`)
	}))
	defer srv.Close()

	client := NewHTTPClient([]string{srv.Listener.Addr().String()}, 1<<20)
	resp, err := client.Do(context.Background(), "GET", srv.URL+"/", nil, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Headers["Content-Type"] != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", resp.Headers["Content-Type"])
	}
}

func TestDo_AllowlistBlocked(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	// Allow a different host — request should be blocked.
	client := NewHTTPClient([]string{"allowed.example.com"}, 1<<20)
	_, err := client.Do(context.Background(), "GET", srv.URL+"/pets", nil, "")
	if err == nil {
		t.Fatal("expected error for blocked host")
	}
}

func TestDo_AllowlistEmpty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "ok")
	}))
	defer srv.Close()

	// Empty allowlist = allow all.
	client := NewHTTPClient([]string{}, 1<<20)
	resp, err := client.Do(context.Background(), "GET", srv.URL+"/", nil, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
}

func TestDo_BodyCap(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Write 10 bytes but cap is 5
		_, _ = io.WriteString(w, "0123456789")
	}))
	defer srv.Close()

	client := NewHTTPClient([]string{}, 5) // cap at 5 bytes
	resp, err := client.Do(context.Background(), "GET", srv.URL+"/", nil, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Body) > 5 {
		t.Errorf("body len = %d, want <= 5 (cap enforced)", len(resp.Body))
	}
}

func TestDo_ContextCancelled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	client := NewHTTPClient([]string{}, 1<<20)
	_, err := client.Do(ctx, "GET", srv.URL+"/", nil, "")
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}
