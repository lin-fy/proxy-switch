package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	domainprovider "proxy-switch/internal/domain/provider"
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

func TestResponsesTesterListsModelsAndDefaultsNames(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":[{"id":" model-a ","name":" Model A "},{"id":"model-b"},{"name":"ignored"}]}`))
	}))
	defer server.Close()
	p, err := domainprovider.New("custom", "Custom", server.URL+"/v1", "")
	if err != nil {
		t.Fatal(err)
	}
	models, err := NewResponsesTester(server.Client()).ListModels(context.Background(), p)
	if err != nil {
		t.Fatal(err)
	}
	if len(models) != 2 || models[0].ID != "model-a" || models[0].Name != "Model A" || models[1].Name != "model-b" {
		t.Fatalf("models = %#v", models)
	}
}

func TestResponsesTesterRejectsBadModelsJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":`))
	}))
	defer server.Close()
	p, err := domainprovider.New("custom", "Custom", server.URL+"/v1", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewResponsesTester(server.Client()).ListModels(context.Background(), p); err == nil || !strings.Contains(err.Error(), "decode provider models") {
		t.Fatalf("bad JSON error = %v", err)
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

func TestResponsesTesterCanCallModel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/responses" {
			t.Errorf("request = %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("content type = %q", r.Header.Get("Content-Type"))
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"resp_test"}`))
	}))
	defer server.Close()
	p, err := domainprovider.New("custom", "Custom", server.URL+"/v1", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := NewResponsesTester(server.Client()).TestModel(context.Background(), p, "model-a"); err != nil {
		t.Fatal(err)
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
