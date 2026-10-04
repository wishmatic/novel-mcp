package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/zap"
)

type bearerRoundTripper struct {
	token string
	base  http.RoundTripper
}

func (t bearerRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	clone.Header.Set("Authorization", "Bearer "+t.token)

	return t.base.RoundTrip(clone)
}

func TestMCPOverHTTPListsTheNovelAITool(t *testing.T) {
	cfg := testConfig(t)
	cfg.NovelAIAPIKey = "sk-test"

	srv, err := New(cfg, zap.NewNop())
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	api := httptest.NewServer(srv.router)
	t.Cleanup(api.Close)

	session, err := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "v0"}, nil).Connect(
		context.Background(),
		&mcp.StreamableClientTransport{
			Endpoint: api.URL + "/mcp",
			HTTPClient: &http.Client{Transport: bearerRoundTripper{
				token: "server-key",
				base:  http.DefaultTransport,
			}},
			DisableStandaloneSSE: true,
		},
		nil,
	)
	if err != nil {
		t.Fatalf("Connect() error: %v", err)
	}

	t.Cleanup(func() { _ = session.Close() })

	tools, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools() error: %v", err)
	}

	if len(tools.Tools) != 1 || tools.Tools[0].Name != "novelai" {
		t.Errorf("tools = %v, want just novelai", tools.Tools)
	}
}

func TestMCPOverHTTPListsNoToolsWithoutANovelAIKey(t *testing.T) {
	srv, err := New(testConfig(t), zap.NewNop())
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	api := httptest.NewServer(srv.router)
	t.Cleanup(api.Close)

	session, err := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "v0"}, nil).Connect(
		context.Background(),
		&mcp.StreamableClientTransport{
			Endpoint: api.URL + "/mcp",
			HTTPClient: &http.Client{Transport: bearerRoundTripper{
				token: "server-key",
				base:  http.DefaultTransport,
			}},
			DisableStandaloneSSE: true,
		},
		nil,
	)
	if err != nil {
		t.Fatalf("Connect() error: %v", err)
	}

	t.Cleanup(func() { _ = session.Close() })

	tools, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools() error: %v", err)
	}

	if len(tools.Tools) != 0 {
		t.Errorf("tools = %v, want none without NOVELAI_API_KEY", tools.Tools)
	}
}

func TestMCPRejectsUnauthenticatedRequests(t *testing.T) {
	srv, err := New(testConfig(t), zap.NewNop())
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	api := httptest.NewServer(srv.router)
	t.Cleanup(api.Close)

	resp, err := http.Post(api.URL+"/mcp", "application/json", strings.NewReader("{}"))
	if err != nil {
		t.Fatalf("Post() error: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
}
