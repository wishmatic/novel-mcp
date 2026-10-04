package mcp

import (
	"slices"
	"strings"
	"testing"

	"github.com/wishmatic/novel-mcp/internal/novelai"
)

func TestNewRegistersTools(t *testing.T) {
	srv, err := New(Clients{Log: zapNop()})
	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}

	if srv == nil {
		t.Fatal("New() returned nil server")
	}
}

func TestServerInfo(t *testing.T) {
	srv, err := New(Clients{Log: zapNop()})
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	info := connectSession(t, srv).InitializeResult().ServerInfo
	if info.Name != "novel-mcp" || info.Version != version {
		t.Errorf("server info = %+v, want novel-mcp %s", info, version)
	}
}

func TestToolRegistration(t *testing.T) {
	tests := []struct {
		name    string
		novelai *novelai.Client
		want    []string
	}{
		{
			name:    "novelai only",
			novelai: novelai.New("http://example.com", "sk"),
			want:    []string{"novelai"},
		},
		{
			name: "no backend",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, err := New(Clients{
				Log:     zapNop(),
				NovelAI: tt.novelai,
			})
			if err != nil {
				t.Fatalf("New() error: %v", err)
			}

			got := toolNames(t, srv)
			want := append([]string(nil), tt.want...)

			slices.Sort(got)
			slices.Sort(want)

			if !slices.Equal(got, want) {
				t.Errorf("tools = %v, want %v", got, want)
			}
		})
	}
}

func TestGenerationDescriptionsDocumentTheInitImage(t *testing.T) {
	srv, err := New(Clients{
		Log:     zapNop(),
		NovelAI: novelai.New("http://example.com", "sk"),
	})
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	for _, name := range []string{"novelai"} {
		desc := toolByName(t, srv, name).Description

		for _, want := range []string{"init_image_url", "default to 512"} {
			if !strings.Contains(desc, want) {
				t.Errorf("%s description %q does not mention %q", name, desc, want)
			}
		}
	}
}
