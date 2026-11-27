package mcp

import (
	"context"
	"fmt"

	"github.com/wishmatic/novel-cli/internal/format"
	"github.com/wishmatic/novel-cli/internal/novelai"
	"github.com/wishmatic/novel-cli/internal/resolve"
	"github.com/wishmatic/novel-cli/internal/store"
	"go.uber.org/zap"
)

type Clients struct {
	Log *zap.Logger

	NovelAI  *novelai.Client
	Store    *store.Client
	Resolver *resolve.Client
}

func (c *Clients) outputFormat(name string) (format.Format, error) {
	if name == "" {
		return format.Default, nil
	}

	return format.Parse(name)
}

func (c *Clients) generationFailure(ctx context.Context, tool string, err error) error {
	if ctxErr := ctx.Err(); ctxErr != nil {
		c.Log.Warn(tool+" aborted: request context cancelled before completion",
			zap.Error(err),
			zap.String("ctx_err", ctxErr.Error()),
		)
	} else {
		c.Log.Error(tool+" generation failed", zap.Error(err))
	}

	return fmt.Errorf("%s: %w", tool, err)
}
