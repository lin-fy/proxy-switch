package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	domainprovider "codex-provider-hub/internal/domain/provider"
)

func TestResponsesTesterChecksModelsEndpoint(t *testing.T) {
	t.Setenv("CPA_API_KEY", "secret-token")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/models" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if r.URL.Query().Get("tenant") != "demo" {
			t.Errorf("tenant query = %q", r.URL.Query().Get("tenant"))
		}
		if r.Header.Get("Authorization") != "Bearer secret-token" {
			t.Errorf("authorization = %q", r.Header.Get("Authorization"))
		}
		if r.Header.Get("X-Test") != "provider-hub" {
			t.Errorf("X-Test = %q", r.Header.Get("X-Test"))
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	defer server.Close()

	p, err := domainprovider.New("cpa", "CPA", server.URL+"/v1", "CPA_API_KEY")
	if err != nil {
		t.Fatal(err)
	}
	p.Headers = map[string]string{"X-Test": "provider-hub"}
	p.QueryParams = map[string]string{"tenant": "demo"}
	if err := NewResponsesTester(server.Client()).Test(context.Background(), p); err != nil {
		t.Fatal(err)
	}
}

func TestResponsesTesterReportsHTTPStatusWithoutBody(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	defer server.Close()
	p, err := domainprovider.New("custom", "Custom", server.URL, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := NewResponsesTester(server.Client()).Test(context.Background(), p); err == nil || err.Error() != "provider returned HTTP 404" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResponsesTesterRequiresConfiguredCredential(t *testing.T) {
	t.Setenv("MISSING_PROVIDER_KEY", "")
	p, err := domainprovider.New("custom", "Custom", "https://example.com/v1", "MISSING_PROVIDER_KEY")
	if err != nil {
		t.Fatal(err)
	}
	if err := NewResponsesTester(nil).Test(context.Background(), p); err == nil {
		t.Fatal("expected missing credential error")
	}
}
