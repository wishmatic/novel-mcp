package cli

import (
	"os"
	"strings"
	"testing"

	"github.com/wishmatic/novel-cli/internal/config"
	"github.com/wishmatic/novel-cli/internal/credential"
	"github.com/wishmatic/novel-cli/internal/novelai"
	"github.com/wishmatic/novel-cli/internal/resolve"
)

func TestWithDefaultsFillsTheStreams(t *testing.T) {
	deps := WithDefaults(Deps{})

	if deps.In != os.Stdin {
		t.Error("In is not os.Stdin")
	}
	if deps.Out != os.Stdout {
		t.Error("Out is not os.Stdout")
	}
	if deps.Err != os.Stderr {
		t.Error("Err is not os.Stderr")
	}
}

func TestWithDefaultsKeepsTheProvidedStreams(t *testing.T) {
	out := &strings.Builder{}
	deps := WithDefaults(Deps{Out: out})

	if deps.Out != out {
		t.Error("Out is not the provided writer")
	}
	if deps.In != os.Stdin {
		t.Error("In is not os.Stdin")
	}
}

func TestLoadConfigUsesTheInjectedConfig(t *testing.T) {
	cfg := config.Config{NovelAIAPIKey: "injected-key"}

	got, err := LoadConfig(Deps{Config: cfg})
	if err != nil {
		t.Fatalf("LoadConfig() error: %v", err)
	}

	if got != cfg {
		t.Errorf("LoadConfig() = %+v, want the injected config", got)
	}
}

func TestLoadConfigReadsDefaultsFromTheEnvironment(t *testing.T) {
	t.Setenv("PORT", "")

	got, err := LoadConfig(Deps{})
	if err != nil {
		t.Fatalf("LoadConfig() error: %v", err)
	}

	if got.Port != 8080 {
		t.Errorf("Port = %d, want the 8080 default", got.Port)
	}
}

func TestNovelClientUsesAnInjectedClient(t *testing.T) {
	client := novelai.New("http://example.com", "test-key")

	got, err := NovelClient(Deps{Client: client})
	if err != nil {
		t.Fatalf("NovelClient() error: %v", err)
	}

	if got != client {
		t.Error("NovelClient() did not return the injected client")
	}
}

func TestNovelClientReportsAMissingKey(t *testing.T) {
	t.Setenv("NOVELAI_API_KEY", "")
	t.Setenv("PORT", "")

	_, err := NovelClient(Deps{Config: config.Config{Host: "localhost"}, Credentials: emptyStore{}})
	if err == nil || !strings.Contains(err.Error(), "novel-cli login") {
		t.Errorf("NovelClient() error = %v, want it to name novel-cli login", err)
	}
}

func TestAPIKeyPrefersTheEnvironment(t *testing.T) {
	key, err := APIKey(Deps{Credentials: &fakeStore{key: "stored-key"}}, config.Config{NovelAIAPIKey: "env-key"})
	if err != nil {
		t.Fatalf("APIKey() error: %v", err)
	}

	if key != "env-key" {
		t.Errorf("APIKey() = %q, want the environment key", key)
	}
}

func TestAPIKeyFallsBackToTheStore(t *testing.T) {
	key, err := APIKey(Deps{Credentials: &fakeStore{key: "stored-key"}}, config.Config{})
	if err != nil {
		t.Fatalf("APIKey() error: %v", err)
	}

	if key != "stored-key" {
		t.Errorf("APIKey() = %q, want the stored key", key)
	}
}

func TestNovelClientBuildsFromConfig(t *testing.T) {
	client, err := NovelClient(Deps{Config: config.Config{NovelAIAPIKey: "config-key"}})
	if err != nil {
		t.Fatalf("NovelClient() error: %v", err)
	}

	if client == nil {
		t.Error("NovelClient() = nil, want a client")
	}
}

func TestImageResolverUsesAnInjectedResolver(t *testing.T) {
	resolver, err := resolve.New(nil, "", nil)
	if err != nil {
		t.Fatalf("resolve.New() error: %v", err)
	}

	got, err := ImageResolver(Deps{Resolver: resolver})
	if err != nil {
		t.Fatalf("ImageResolver() error: %v", err)
	}

	if got != resolver {
		t.Error("ImageResolver() did not return the injected resolver")
	}
}

func TestImageResolverRejectsAMalformedMap(t *testing.T) {
	_, err := ImageResolver(Deps{Config: config.Config{ImageURLMap: "not a pair"}})
	if err == nil || !strings.Contains(err.Error(), "IMAGE_URL_MAP") {
		t.Errorf("ImageResolver() error = %v, want it to name IMAGE_URL_MAP", err)
	}
}

type emptyStore struct{}

func (emptyStore) Load() (string, error) {
	return "", credential.ErrNotFound
}

func (emptyStore) Save(string) error {
	return nil
}

func (emptyStore) Delete() error {
	return nil
}

type fakeStore struct {
	key string
}

func (s *fakeStore) Load() (string, error) {
	return s.key, nil
}

func (s *fakeStore) Save(key string) error {
	s.key = key

	return nil
}

func (s *fakeStore) Delete() error {
	s.key = ""

	return nil
}
