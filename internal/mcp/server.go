package mcp

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/wishmatic/novel-mcp/internal/format"
)

const version = "0.1.0"

func New(clients Clients) (*mcp.Server, error) {
	if clients.DefaultOutputFormat == "" {
		clients.DefaultOutputFormat = format.Default
	}

	srv := mcp.NewServer(&mcp.Implementation{
		Name:    "novel-mcp",
		Version: version,
	}, nil)

	registerTools(srv, &clients)

	return srv, nil
}

func registerTools(srv *mcp.Server, c *Clients) {
	if c.NovelAI != nil {
		registerNovelAI(srv, c)
	}
}
