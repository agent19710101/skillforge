package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/agent19710101/skillforge/internal/catalog"
)

const defaultUserAgent = "skillforge-cli"

type Client struct {
	BaseURL    *url.URL
	HTTPClient *http.Client
}

type APIError struct {
	StatusCode int                    `json:"statusCode"`
	Code       string                 `json:"error"`
	Message    string                 `json:"message"`
	Validation *DraftValidation       `json:"validation,omitempty"`
	Submission *DraftSubmissionStatus `json:"submission,omitempty"`
}

func (e *APIError) Error() string {
	if e == nil {
		return "api error"
	}
	if strings.TrimSpace(e.Code) == "" {
		return fmt.Sprintf("api request failed (%d): %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("api request failed (%d, %s): %s", e.StatusCode, e.Code, e.Message)
}

type ListOptions struct {
	Validation string
	Offset     int
	Limit      int
}

type ListSkillsResponse struct {
	Skills []catalog.SkillRecord `json:"skills"`
	Total  int                   `json:"total"`
	Offset int                   `json:"offset"`
	Limit  int                   `json:"limit"`
}

type SearchResponse struct {
	Query  string                `json:"query"`
	Skills []catalog.SkillRecord `json:"skills"`
	Total  int                   `json:"total"`
}

type DraftCreateRequest struct {
	Operation string `json:"operation"`
	SkillName string `json:"skillName"`
	Content   string `json:"content,omitempty"`
}

type DraftValidation struct {
	Valid    bool              `json:"valid"`
	Findings []catalog.Finding `json:"findings,omitempty"`
}

type DraftSubmissionStatus struct {
	Enabled    bool   `json:"enabled"`
	BaseBranch string `json:"baseBranch,omitempty"`
	Reason     string `json:"reason,omitempty"`
}

type DraftResponse struct {
	ID         string                `json:"id"`
	Operation  string                `json:"operation"`
	SkillName  string                `json:"skillName"`
	BranchName string                `json:"branchName"`
	CreatedAt  string                `json:"createdAt"`
	Validation DraftValidation       `json:"validation"`
	Submission DraftSubmissionStatus `json:"submission"`
}

type PullRequestRef struct {
	Number int    `json:"number,omitempty"`
	ID     int64  `json:"id,omitempty"`
	URL    string `json:"url,omitempty"`
}

type DraftSubmissionResponse struct {
	ID          string          `json:"id"`
	Operation   string          `json:"operation"`
	SkillName   string          `json:"skillName"`
	BranchName  string          `json:"branchName"`
	BaseBranch  string          `json:"baseBranch"`
	CommitHash  string          `json:"commitHash,omitempty"`
	PullRequest *PullRequestRef `json:"pullRequest,omitempty"`
	Validation  DraftValidation `json:"validation"`
}

func New(baseURL string) (*Client, error) {
	trimmed := strings.TrimSpace(baseURL)
	if trimmed == "" {
		return nil, fmt.Errorf("base URL is required")
	}
	u, err := url.Parse(trimmed)
	if err != nil {
		return nil, fmt.Errorf("parse base URL: %w", err)
	}
	if u.Scheme == "" || u.Host == "" {
		return nil, fmt.Errorf("base URL must include scheme and host")
	}
	if u.Path != "" && !strings.HasSuffix(u.Path, "/") {
		u.Path += "/"
	}
	return &Client{BaseURL: u, HTTPClient: http.DefaultClient}, nil
}

func (c *Client) ListSkills(ctx context.Context, opts ListOptions) (ListSkillsResponse, error) {
	values := url.Values{}
	if strings.TrimSpace(opts.Validation) != "" {
		values.Set("validation", strings.TrimSpace(opts.Validation))
	}
	if opts.Offset > 0 {
		values.Set("offset", strconv.Itoa(opts.Offset))
	}
	if opts.Limit > 0 {
		values.Set("limit", strconv.Itoa(opts.Limit))
	}

	var resp ListSkillsResponse
	err := c.getJSON(ctx, "/api/v1/skills", values, &resp)
	return resp, err
}

func (c *Client) SearchSkills(ctx context.Context, query string) (SearchResponse, error) {
	values := url.Values{}
	values.Set("q", strings.TrimSpace(query))

	var resp SearchResponse
	err := c.getJSON(ctx, "/api/v1/search", values, &resp)
	return resp, err
}

func (c *Client) GetSkill(ctx context.Context, name string) (catalog.SkillRecord, error) {
	var skill catalog.SkillRecord
	err := c.getJSON(ctx, "/api/v1/skills/"+url.PathEscape(strings.TrimSpace(name)), nil, &skill)
	return skill, err
}

func (c *Client) CreateDraft(ctx context.Context, req DraftCreateRequest) (DraftResponse, error) {
	var resp DraftResponse
	err := c.sendJSON(ctx, http.MethodPost, "/api/v1/drafts", nil, req, &resp)
	return resp, err
}

func (c *Client) GetDraft(ctx context.Context, id string) (DraftResponse, error) {
	var resp DraftResponse
	err := c.sendJSON(ctx, http.MethodGet, "/api/v1/drafts/"+url.PathEscape(strings.TrimSpace(id)), nil, nil, &resp)
	return resp, err
}

func (c *Client) SubmitDraft(ctx context.Context, id string) (DraftSubmissionResponse, error) {
	var resp DraftSubmissionResponse
	err := c.sendJSON(ctx, http.MethodPost, "/api/v1/drafts/"+url.PathEscape(strings.TrimSpace(id))+"/submit", nil, nil, &resp)
	return resp, err
}

func (c *Client) getJSON(ctx context.Context, path string, query url.Values, into any) error {
	return c.sendJSON(ctx, http.MethodGet, path, query, nil, into)
}

func (c *Client) sendJSON(ctx context.Context, method, path string, query url.Values, body any, into any) error {
	if c == nil {
		return fmt.Errorf("client is required")
	}
	if c.BaseURL == nil {
		return fmt.Errorf("base URL is required")
	}

	rel := &url.URL{Path: strings.TrimLeft(path, "/")}
	if len(query) > 0 {
		rel.RawQuery = query.Encode()
	}
	endpoint := c.BaseURL.ResolveReference(rel)

	var requestBody io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encode request: %w", err)
		}
		requestBody = bytes.NewReader(payload)
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint.String(), requestBody)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", defaultUserAgent)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	res, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return decodeAPIError(res)
	}
	if err := json.NewDecoder(res.Body).Decode(into); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

func decodeAPIError(res *http.Response) error {
	payload := APIError{StatusCode: res.StatusCode}
	body, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("read error response: %w", err)
	}
	if len(body) == 0 {
		payload.Message = http.StatusText(res.StatusCode)
		return &payload
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		payload.Message = strings.TrimSpace(string(body))
		if payload.Message == "" {
			payload.Message = http.StatusText(res.StatusCode)
		}
		return &payload
	}
	if strings.TrimSpace(payload.Message) == "" {
		payload.Message = http.StatusText(res.StatusCode)
	}
	return &payload
}
