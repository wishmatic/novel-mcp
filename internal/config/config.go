package config

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	Host string `env:"HOST" envDefault:"0.0.0.0"`
	Port int    `env:"PORT" envDefault:"8080"`

	LogLevel string `env:"LOG_LEVEL" envDefault:"info"`

	DefaultOutput string `env:"OUTPUT_FORMAT" envDefault:"webp"`

	APIKey string `env:"API_KEY"`

	ImageURLMap string `env:"IMAGE_URL_MAP"`

	NovelAIAPIKey string `env:"NOVELAI_API_KEY"`

	PublicHost string `env:"PUBLIC_HOST"`
	FilesDir   string `env:"FILES_DIR" envDefault:"files"`
}

func Load() (Config, error) {
	_ = godotenv.Load()

	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return Config{}, fmt.Errorf("parse environment: %w", err)
	}

	return cfg, nil
}

func (c Config) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// PublicBase normalises PUBLIC_HOST into the base URL every returned image link is built from.
func (c Config) PublicBase() (*url.URL, error) {
	if c.PublicHost == "" {
		return nil, nil
	}

	base, err := url.Parse(strings.TrimSuffix(c.PublicHost, "/"))
	if err != nil {
		return nil, fmt.Errorf("PUBLIC_HOST %q is not a valid URL: %w", c.PublicHost, err)
	}
	if base.Scheme != "http" && base.Scheme != "https" {
		return nil, fmt.Errorf("PUBLIC_HOST %q must use http or https", c.PublicHost)
	}
	if base.Host == "" {
		return nil, fmt.Errorf("PUBLIC_HOST %q must include a host", c.PublicHost)
	}
	if base.Path != "" || base.RawQuery != "" || base.Fragment != "" {
		return nil, fmt.Errorf(
			"PUBLIC_HOST %q must not include a path, query, or fragment", c.PublicHost,
		)
	}

	return base, nil
}

func (c Config) Validate() error {
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("PORT must be between 1 and 65535, got %d", c.Port)
	}

	return nil
}
