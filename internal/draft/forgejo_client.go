package draft

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultForgejoUserAgent = "skillforge"

type forgejoClient struct {
	serverURL  string
	authMethod ForgejoAuthMethod
	token      string
	httpClient *http.Client
}

// NewForgejoClient constructs the live Forgejo pull-request client used by submission.
func NewForgejoClient(cfg SubmissionConfig) (ForgejoClient, error) {
	serverURL := strings.TrimSpace(cfg.ServerURL)
	if serverURL == "" {
		return nil, fmt.Errorf("forgejo client: %w", errForgejoServerURLRequired)
	}
	parsedURL, err := url.Parse(serverURL)
	if err != nil {
		return nil, fmt.Errorf("forgejo client: parse server URL: %w", err)
	}
	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return nil, fmt.Errorf("forgejo client: parse server URL: missing scheme or host")
	}

	authMethod := cfg.authMethod()
	switch authMethod {
	case ForgejoAuthMethodToken:
		if strings.TrimSpace(cfg.Token) == "" {
			return nil, fmt.Errorf("forgejo client: %w", errForgejoTokenRequired)
		}
	case ForgejoAuthMethodNone:
		// Allowed for tests and intentionally unauthenticated setups.
	default:
		return nil, fmt.Errorf("forgejo client: unsupported auth method %q", cfg.AuthMethod)
	}

	return &forgejoClient{
		serverURL:  strings.TrimRight(serverURL, "/"),
		authMethod: authMethod,
		token:      strings.TrimSpace(cfg.Token),
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}, nil
}

func (c *forgejoClient) CreatePullRequest(ctx context.Context, req PullRequestRequest) (PullRequest, error) {
	if strings.TrimSpace(req.Owner) == "" {
		return PullRequest{}, fmt.Errorf("forgejo create pull request: owner is required")
	}
	if strings.TrimSpace(req.Repo) == "" {
		return PullRequest{}, fmt.Errorf("forgejo create pull request: repo is required")
	}
	if strings.TrimSpace(req.HeadBranch) == "" {
		return PullRequest{}, fmt.Errorf("forgejo create pull request: head branch is required")
	}
	if strings.TrimSpace(req.BaseBranch) == "" {
		return PullRequest{}, fmt.Errorf("forgejo create pull request: base branch is required")
	}
	if strings.TrimSpace(req.Title) == "" {
		return PullRequest{}, fmt.Errorf("forgejo create pull request: title is required")
	}

	payload := struct {
		Head  string `json:"head"`
		Base  string `json:"base"`
		Title string `json:"title"`
		Body  string `json:"body,omitempty"`
	}{
		Head:  strings.TrimSpace(req.HeadBranch),
		Base:  strings.TrimSpace(req.BaseBranch),
		Title: strings.TrimSpace(req.Title),
		Body:  strings.TrimSpace(req.Body),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return PullRequest{}, fmt.Errorf("forgejo create pull request: marshal request: %w", err)
	}

	endpoint := c.serverURL + "/api/v1/repos/" + url.PathEscape(strings.TrimSpace(req.Owner)) + "/" + url.PathEscape(strings.TrimSpace(req.Repo)) + "/pulls"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return PullRequest{}, fmt.Errorf("forgejo create pull request: build request: %w", err)
	}
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", defaultForgejoUserAgent)
	switch c.authMethod {
	case ForgejoAuthMethodToken:
		httpReq.Header.Set("Authorization", "token "+c.token)
	case ForgejoAuthMethodNone:
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return PullRequest{}, fmt.Errorf("forgejo create pull request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		message, readErr := readForgejoError(resp.Body)
		if readErr != nil {
			return PullRequest{}, fmt.Errorf("forgejo create pull request: response %s (and failed to read error body: %v)", resp.Status, readErr)
		}
		if message == "" {
			return PullRequest{}, fmt.Errorf("forgejo create pull request: response %s", resp.Status)
		}
		return PullRequest{}, fmt.Errorf("forgejo create pull request: response %s: %s", resp.Status, message)
	}

	var decoded struct {
		Number  int    `json:"number"`
		ID      int64  `json:"id"`
		HTMLURL string `json:"html_url"`
		URL     string `json:"url"`
		WebURL  string `json:"web_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return PullRequest{}, fmt.Errorf("forgejo create pull request: decode response: %w", err)
	}

	return PullRequest{
		Number: decoded.Number,
		ID:     decoded.ID,
		URL:    firstNonEmpty(decoded.HTMLURL, decoded.WebURL, decoded.URL),
	}, nil
}

func readForgejoError(body io.Reader) (string, error) {
	data, err := io.ReadAll(body)
	if err != nil {
		return "", err
	}
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" {
		return "", nil
	}

	var decoded struct {
		Message string `json:"message"`
		Error   string `json:"error"`
	}
	if err := json.Unmarshal(data, &decoded); err == nil {
		if message := firstNonEmpty(decoded.Message, decoded.Error); message != "" {
			return message, nil
		}
	}
	return trimmed, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

var (
	errForgejoServerURLRequired = errors.New("forgejo server URL is required")
	errForgejoTokenRequired     = errors.New("forgejo token is required for token auth")
)
