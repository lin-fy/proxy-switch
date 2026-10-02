package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"codex-provider-hub/internal/application/ports"
	domainprovider "codex-provider-hub/internal/domain/provider"
	"codex-provider-hub/internal/infrastructure/credential"
)

type ResponsesTester struct {
	client *http.Client
}

func NewResponsesTester(client *http.Client) *ResponsesTester {
	if client == nil {
		client = &http.Client{Timeout: 8 * time.Second}
	}
	return &ResponsesTester{client: client}
}

func (t *ResponsesTester) Test(ctx context.Context, p domainprovider.Provider) error {
	_, err := t.ListModels(ctx, p)
	return err
}

func (t *ResponsesTester) ListModels(ctx context.Context, p domainprovider.Provider) ([]ports.RemoteModel, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	token, err := credential.Resolve(p.AuthRef)
	if err != nil {
		return nil, err
	}
	endpoint, err := modelsEndpoint(p.BaseURL)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create provider request: %w", err)
	}
	for key, value := range p.Headers {
		request.Header.Set(key, value)
	}
	if token != "" && request.Header.Get("Authorization") == "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	query := request.URL.Query()
	for key, value := range p.QueryParams {
		query.Set(key, value)
	}
	request.URL.RawQuery = query.Encode()
	response, err := t.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("provider request failed: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 64<<10))
		return nil, fmt.Errorf("provider returned HTTP %d", response.StatusCode)
	}
	var payload struct {
		Data []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"data"`
	}
	decoder := json.NewDecoder(io.LimitReader(response.Body, 2<<20))
	if err := decoder.Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode provider models: %w", err)
	}
	models := make([]ports.RemoteModel, 0, len(payload.Data))
	for _, item := range payload.Data {
		id := strings.TrimSpace(item.ID)
		if id == "" {
			continue
		}
		name := strings.TrimSpace(item.Name)
		if name == "" {
			name = id
		}
		models = append(models, ports.RemoteModel{ID: id, Name: name})
	}
	return models, nil
}

func (t *ResponsesTester) TestModel(ctx context.Context, p domainprovider.Provider, modelID string) error {
	if err := p.Validate(); err != nil {
		return err
	}
	modelID = strings.TrimSpace(modelID)
	if modelID == "" {
		return fmt.Errorf("model id is required")
	}
	token, err := credential.Resolve(p.AuthRef)
	if err != nil {
		return err
	}
	endpoint, err := responsesEndpoint(p.BaseURL)
	if err != nil {
		return err
	}
	body, err := json.Marshal(map[string]any{
		"model":             modelID,
		"input":             "ping",
		"max_output_tokens": 1,
	})
	if err != nil {
		return fmt.Errorf("encode model test request: %w", err)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create model test request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	for key, value := range p.Headers {
		request.Header.Set(key, value)
	}
	if token != "" && request.Header.Get("Authorization") == "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	query := request.URL.Query()
	for key, value := range p.QueryParams {
		query.Set(key, value)
	}
	request.URL.RawQuery = query.Encode()
	response, err := t.client.Do(request)
	if err != nil {
		return fmt.Errorf("model test request failed: %w", err)
	}
	defer response.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 64<<10))
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("model test returned HTTP %d", response.StatusCode)
	}
	return nil
}

func modelsEndpoint(baseURL string) (string, error) {
	return endpoint(baseURL, "models")
}

func responsesEndpoint(baseURL string) (string, error) {
	return endpoint(baseURL, "responses")
}

func endpoint(baseURL, suffix string) (string, error) {
	parsed, err := url.Parse(baseURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", domainprovider.ErrInvalidBaseURL
	}
	parsed.Path = strings.TrimRight(parsed.Path, "/") + "/" + suffix
	return parsed.String(), nil
}
