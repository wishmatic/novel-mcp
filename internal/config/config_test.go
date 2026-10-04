package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func unsetEnv(t *testing.T, key string) {
	t.Helper()

	prev, had := os.LookupEnv(key)

	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("unset %s: %v", key, err)
	}

	t.Cleanup(func() {
		if had {
			_ = os.Setenv(key, prev)

			return
		}

		_ = os.Unsetenv(key)
	})
}

func TestAddr(t *testing.T) {
	cfg := Config{Host: "127.0.0.1", Port: 9000}

	if got := cfg.Addr(); got != "127.0.0.1:9000" {
		t.Errorf("Addr() = %q, want 127.0.0.1:9000", got)
	}
}

func TestValidateAcceptsThePortBounds(t *testing.T) {
	for _, port := range []int{1, 65535} {
		cfg := Config{Port: port}

		if err := cfg.Validate(); err != nil {
			t.Errorf("Validate() error for port %d = %v, want it accepted", port, err)
		}
	}
}

func TestValidateRejectsOutOfRangePorts(t *testing.T) {
	for _, port := range []int{0, -1, 65536} {
		cfg := Config{Port: port}

		err := cfg.Validate()
		if err == nil || !strings.Contains(err.Error(), "PORT") {
			t.Errorf("Validate() error for port %d = %v, want it to name PORT", port, err)
		}
	}
}

func TestLoadReadsTheEnvironment(t *testing.T) {
	t.Setenv("HOST", "127.0.0.1")
	t.Setenv("PORT", "9100")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("API_KEY", "secret")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.Addr() != "127.0.0.1:9100" {
		t.Errorf("Addr() = %q, want 127.0.0.1:9100", cfg.Addr())
	}

	if cfg.LogLevel != "debug" || cfg.APIKey != "secret" {
		t.Errorf("Load() = %+v, want the configured level and key", cfg)
	}
}

func TestLoadAppliesTheDefaults(t *testing.T) {
	t.Setenv("HOST", "")
	t.Setenv("PORT", "")
	t.Setenv("LOG_LEVEL", "")
	t.Setenv("API_KEY", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.Host != "0.0.0.0" || cfg.Port != 8080 || cfg.LogLevel != "info" || cfg.APIKey != "" {
		t.Errorf("Load() = %+v, want the documented defaults", cfg)
	}
}

func TestLoadRejectsANonNumericPort(t *testing.T) {
	t.Setenv("PORT", "http")

	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want the malformed port rejected")
	}
}

func TestLoadOutputFormatDefaults(t *testing.T) {
	unsetEnv(t, "OUTPUT_FORMAT")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.DefaultOutput != "webp" {
		t.Errorf("OutputFormat = %q, want webp", cfg.DefaultOutput)
	}
}

func TestLoadOutputFormat(t *testing.T) {
	tests := map[string]string{
		"jxl":  "jxl",
		"jpeg": "jpeg",
		"JPEG": "JPEG",
	}

	for value, want := range tests {
		t.Run(value, func(t *testing.T) {
			t.Setenv("OUTPUT_FORMAT", value)

			cfg, err := Load()
			if err != nil {
				t.Fatalf("Load() error: %v", err)
			}

			if cfg.DefaultOutput != want {
				t.Errorf("OutputFormat = %q, want %q", cfg.DefaultOutput, want)
			}
		})
	}
}

func TestLoadFilesDefaults(t *testing.T) {
	unsetEnv(t, "PUBLIC_HOST")
	unsetEnv(t, "FILES_DIR")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.PublicHost != "" {
		t.Errorf("PublicHost = %q, want empty by default", cfg.PublicHost)
	}

	if cfg.FilesDir != "files" {
		t.Errorf("FilesDir = %q, want files by default", cfg.FilesDir)
	}
}

func TestLoadFiles(t *testing.T) {
	t.Setenv("PUBLIC_HOST", "https://novel.example.com")
	t.Setenv("FILES_DIR", filepath.Join("data", "files"))

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.PublicHost != "https://novel.example.com" {
		t.Errorf("PublicHost = %q, want https://novel.example.com", cfg.PublicHost)
	}

	if cfg.FilesDir != filepath.Join("data", "files") {
		t.Errorf("FilesDir = %q, want the configured path", cfg.FilesDir)
	}
}

func TestPublicBase(t *testing.T) {
	tests := []struct {
		name       string
		publicHost string
		want       string
		wantErr    bool
	}{
		{name: "unset", publicHost: "", wantErr: false},
		{name: "valid", publicHost: "https://novel.example.com", want: "https://novel.example.com"},
		{name: "valid with port", publicHost: "http://192.168.1.10:8080", want: "http://192.168.1.10:8080"},
		{name: "trailing slash trimmed", publicHost: "https://novel.example.com/", want: "https://novel.example.com"},
		{name: "missing scheme", publicHost: "novel.example.com", wantErr: true},
		{name: "non-http scheme", publicHost: "ftp://novel.example.com", wantErr: true},
		{name: "empty host", publicHost: "https://", wantErr: true},
		{name: "path component", publicHost: "https://novel.example.com/novel", wantErr: true},
		{name: "query component", publicHost: "https://novel.example.com?x=1", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{PublicHost: tt.publicHost}

			base, err := cfg.PublicBase()
			if tt.wantErr {
				if err == nil {
					t.Fatalf("PublicBase() error = nil, want an error")
				}

				return
			}

			if err != nil {
				t.Fatalf("PublicBase() error: %v", err)
			}

			if tt.publicHost == "" {
				if base != nil {
					t.Fatalf("PublicBase() = %v, want nil when PUBLIC_HOST is unset", base)
				}

				return
			}

			if base == nil || base.Host == "" {
				t.Fatalf("PublicBase() = %v, want a URL with a host", base)
			}

			if got := base.String(); got != tt.want {
				t.Errorf("PublicBase() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLoadImageURLMap(t *testing.T) {
	spec := "https://example.com=http://example:5080,https://chat.example.com/images/=/data/librechat-data"

	t.Setenv("IMAGE_URL_MAP", spec)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.ImageURLMap != spec {
		t.Errorf("ImageURLMap = %q, want %q", cfg.ImageURLMap, spec)
	}
}

func TestLoadImageURLMapDefaults(t *testing.T) {
	unsetEnv(t, "IMAGE_URL_MAP")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.ImageURLMap != "" {
		t.Errorf("ImageURLMap = %q, want empty by default", cfg.ImageURLMap)
	}
}

func TestLoadNovelAIAPIKey(t *testing.T) {
	t.Setenv("NOVELAI_API_KEY", "sk-test")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.NovelAIAPIKey != "sk-test" {
		t.Errorf("NovelAIAPIKey = %q, want sk-test", cfg.NovelAIAPIKey)
	}
}
