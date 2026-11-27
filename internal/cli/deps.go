package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/wishmatic/novel-mcp/internal/config"
	"github.com/wishmatic/novel-mcp/internal/credential"
	"github.com/wishmatic/novel-mcp/internal/novelai"
	"github.com/wishmatic/novel-mcp/internal/resolve"
	"github.com/wishmatic/novel-mcp/internal/sourcemap"
)

type Deps struct {
	In          io.Reader
	Out         io.Writer
	Err         io.Writer
	Config      config.Config
	Credentials credential.Store
	Client      *novelai.Client
	Resolver    *resolve.Client
}

func DefaultDeps() Deps {
	return Deps{
		In:  os.Stdin,
		Out: os.Stdout,
		Err: os.Stderr,
	}
}

func WithDefaults(deps Deps) Deps {
	if deps.In == nil {
		deps.In = os.Stdin
	}
	if deps.Out == nil {
		deps.Out = os.Stdout
	}
	if deps.Err == nil {
		deps.Err = os.Stderr
	}

	return deps
}

func LoadConfig(deps Deps) (config.Config, error) {
	if deps.Config != (config.Config{}) {
		return deps.Config, nil
	}

	return config.Load()
}

func OpenCredentials(deps Deps) (credential.Store, error) {
	if deps.Credentials != nil {
		return deps.Credentials, nil
	}

	return credential.Open()
}

func NovelClient(deps Deps) (*novelai.Client, error) {
	if deps.Client != nil {
		return deps.Client, nil
	}

	cfg, err := LoadConfig(deps)
	if err != nil {
		return nil, err
	}

	key, err := APIKey(deps, cfg)
	if err != nil {
		return nil, err
	}

	return novelai.New(novelai.DefaultBaseURL, key), nil
}

// APIKey prefers the environment, which is what a service or CI job sets, over the key `login` stored.
func APIKey(deps Deps, cfg config.Config) (string, error) {
	if key := strings.TrimSpace(cfg.NovelAIAPIKey); key != "" {
		return key, nil
	}

	store, err := OpenCredentials(deps)
	if err != nil {
		return "", err
	}

	key, err := store.Load()
	if errors.Is(err, credential.ErrNotFound) {
		return "", fmt.Errorf("not logged in: run `novel-cli login` or set NOVELAI_API_KEY")
	}
	if err != nil {
		return "", fmt.Errorf("load API key: %w", err)
	}

	return key, nil
}

func ImageResolver(deps Deps) (*resolve.Client, error) {
	if deps.Resolver != nil {
		return deps.Resolver, nil
	}

	cfg, err := LoadConfig(deps)
	if err != nil {
		return nil, err
	}

	sources, err := sourcemap.Parse(cfg.ImageURLMap)
	if err != nil {
		return nil, fmt.Errorf("IMAGE_URL_MAP: %w", err)
	}

	return resolve.New(nil, "", sources)
}
