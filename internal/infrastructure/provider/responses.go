package provider

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	domainprovider "codex-provider-hub/internal/domain/provider"
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
	if err := p.Validate(); err != nil {
		return err
	}
	token, err := resolveToken(p.AuthRef)
	if err != nil {
		return err
	}
	endpoint, err := modelsEndpoint(p.BaseURL)
	if err != nil {
		return err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("create provider request: %w", err)
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
		return fmt.Errorf("provider request failed: %w", err)
	}
	defer response.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 64<<10))
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("provider returned HTTP %d", response.StatusCode)
	}
	return nil
}

func modelsEndpoint(baseURL string) (string, error) {
	parsed, err := url.Parse(baseURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", domainprovider.ErrInvalidBaseURL
	}
	parsed.Path = strings.TrimRight(parsed.Path, "/") + "/models"
	return parsed.String(), nil
}

func resolveToken(reference string) (string, error) {
	reference = strings.TrimSpace(reference)
	if reference == "" {
		return "", nil
	}
	if strings.HasPrefix(reference, "credential:") {
		return "", fmt.Errorf("credential reference %q requires Windows Credential Manager support", reference)
	}
	reference = strings.TrimPrefix(reference, "env:")
	token := strings.TrimSpace(os.Getenv(reference))
	if token == "" {
		return "", fmt.Errorf("credential environment variable %q is not set", reference)
	}
	return token, nil
}
