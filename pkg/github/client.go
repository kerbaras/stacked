// Package github provides GitHub API operations for managing pull requests.
package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	gogh "github.com/cli/go-gh/v2/pkg/api"
)

// Client defines the GitHub operations needed by stacked.
type Client interface {
	CreatePR(ctx context.Context, input CreatePRInput) (*PR, error)
	UpdatePR(ctx context.Context, owner, repo string, number int, input UpdatePRInput) error
	GetPR(ctx context.Context, owner, repo string, number int) (*PR, error)
	UpdatePRBase(ctx context.Context, owner, repo string, number int, base string) error
	IsMerged(ctx context.Context, owner, repo string, number int) (bool, error)
}

// PR represents a GitHub pull request.
type PR struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	Body   string `json:"body"`
	State  string `json:"state"`
	Merged bool   `json:"merged"`
	HTMLURL string `json:"html_url"`
	Base   struct {
		Ref string `json:"ref"`
	} `json:"base"`
	Head struct {
		Ref string `json:"ref"`
	} `json:"head"`
}

// CreatePRInput contains fields for creating a PR.
type CreatePRInput struct {
	Owner string
	Repo  string
	Title string
	Body  string
	Head  string
	Base  string
}

// UpdatePRInput contains fields for updating a PR.
type UpdatePRInput struct {
	Title string `json:"title,omitempty"`
	Body  string `json:"body,omitempty"`
}

// NewClient creates a Client using go-gh auth, falling back to environment variables.
func NewClient() (Client, string, string, error) {
	// Try go-gh first for auth token
	opts := gogh.ClientOptions{}
	_, err := gogh.NewRESTClient(opts)
	var token string
	if err == nil {
		// go-gh manages the token internally; we use it for requests
		return &ghClient{}, "", "", fmt.Errorf("use NewClientWithToken for direct HTTP; go-gh integration requires repo context")
	}

	// Fall back to env vars
	token = os.Getenv("GITHUB_TOKEN")
	if token == "" {
		token = os.Getenv("GH_TOKEN")
	}
	if token == "" {
		return nil, "", "", fmt.Errorf("no GitHub token found; run `gh auth login` or set GITHUB_TOKEN")
	}

	return &httpClient{token: token}, "", "", nil
}

// NewClientForRepo creates a Client with the given owner/repo, auto-detecting auth.
func NewClientForRepo(owner, repo string) (Client, error) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		token = os.Getenv("GH_TOKEN")
	}

	if token == "" {
		// Try go-gh
		opts := gogh.ClientOptions{}
		rest, err := gogh.NewRESTClient(opts)
		if err != nil {
			return nil, fmt.Errorf("no GitHub token found; run `gh auth login` or set GITHUB_TOKEN")
		}
		return &goghClient{rest: rest, owner: owner, repo: repo}, nil
	}

	return &httpClient{token: token}, nil
}

// httpClient implements Client using direct HTTP calls with a token.
type httpClient struct {
	token string
}

func (c *httpClient) CreatePR(ctx context.Context, input CreatePRInput) (*PR, error) {
	body := map[string]string{
		"title": input.Title,
		"body":  input.Body,
		"head":  input.Head,
		"base":  input.Base,
	}
	var pr PR
	err := c.do(ctx, "POST", fmt.Sprintf("/repos/%s/%s/pulls", input.Owner, input.Repo), body, &pr)
	return &pr, err
}

func (c *httpClient) UpdatePR(ctx context.Context, owner, repo string, number int, input UpdatePRInput) error {
	return c.do(ctx, "PATCH", fmt.Sprintf("/repos/%s/%s/pulls/%d", owner, repo, number), input, nil)
}

func (c *httpClient) GetPR(ctx context.Context, owner, repo string, number int) (*PR, error) {
	var pr PR
	err := c.do(ctx, "GET", fmt.Sprintf("/repos/%s/%s/pulls/%d", owner, repo, number), nil, &pr)
	return &pr, err
}

func (c *httpClient) UpdatePRBase(ctx context.Context, owner, repo string, number int, base string) error {
	body := map[string]string{"base": base}
	return c.do(ctx, "PATCH", fmt.Sprintf("/repos/%s/%s/pulls/%d", owner, repo, number), body, nil)
}

func (c *httpClient) IsMerged(ctx context.Context, owner, repo string, number int) (bool, error) {
	pr, err := c.GetPR(ctx, owner, repo, number)
	if err != nil {
		return false, err
	}
	return pr.Merged || pr.State == "closed", nil
}

const apiBase = "https://api.github.com"

func (c *httpClient) do(ctx context.Context, method, path string, reqBody any, result any) error {
	var bodyReader io.Reader
	if reqBody != nil {
		data, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, apiBase+path, bodyReader)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("github api request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("github api %s %s: %d %s", method, path, resp.StatusCode, string(body))
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}

	return nil
}

// ghClient is a placeholder — not used directly.
type ghClient struct{}

func (c *ghClient) CreatePR(ctx context.Context, input CreatePRInput) (*PR, error) {
	return nil, fmt.Errorf("not implemented")
}
func (c *ghClient) UpdatePR(ctx context.Context, owner, repo string, number int, input UpdatePRInput) error {
	return fmt.Errorf("not implemented")
}
func (c *ghClient) GetPR(ctx context.Context, owner, repo string, number int) (*PR, error) {
	return nil, fmt.Errorf("not implemented")
}
func (c *ghClient) UpdatePRBase(ctx context.Context, owner, repo string, number int, base string) error {
	return fmt.Errorf("not implemented")
}
func (c *ghClient) IsMerged(ctx context.Context, owner, repo string, number int) (bool, error) {
	return false, fmt.Errorf("not implemented")
}

// goghClient implements Client using go-gh's REST client.
type goghClient struct {
	rest  *gogh.RESTClient
	owner string
	repo  string
}

func (c *goghClient) CreatePR(ctx context.Context, input CreatePRInput) (*PR, error) {
	body := map[string]string{
		"title": input.Title,
		"body":  input.Body,
		"head":  input.Head,
		"base":  input.Base,
	}
	var pr PR
	err := c.rest.Post(fmt.Sprintf("repos/%s/%s/pulls", input.Owner, input.Repo), jsonReader(body), &pr)
	return &pr, err
}

func (c *goghClient) UpdatePR(ctx context.Context, owner, repo string, number int, input UpdatePRInput) error {
	return c.rest.Patch(fmt.Sprintf("repos/%s/%s/pulls/%d", owner, repo, number), jsonReader(input), nil)
}

func (c *goghClient) GetPR(ctx context.Context, owner, repo string, number int) (*PR, error) {
	var pr PR
	err := c.rest.Get(fmt.Sprintf("repos/%s/%s/pulls/%d", owner, repo, number), &pr)
	return &pr, err
}

func (c *goghClient) UpdatePRBase(ctx context.Context, owner, repo string, number int, base string) error {
	body := map[string]string{"base": base}
	return c.rest.Patch(fmt.Sprintf("repos/%s/%s/pulls/%d", owner, repo, number), jsonReader(body), nil)
}

func jsonReader(v any) io.Reader {
	data, _ := json.Marshal(v)
	return bytes.NewReader(data)
}

func (c *goghClient) IsMerged(ctx context.Context, owner, repo string, number int) (bool, error) {
	pr, err := c.GetPR(ctx, owner, repo, number)
	if err != nil {
		return false, err
	}
	return pr.Merged || pr.State == "closed", nil
}
