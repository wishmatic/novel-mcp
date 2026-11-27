package credential

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zalando/go-keyring"
)

func TestFileStoreRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "novelai_api_key")
	store := NewFileStoreAt(path)

	if _, err := store.Load(); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Load() before Save error = %v, want ErrNotFound", err)
	}

	if err := store.Save("  secret-key  "); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	key, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if key != "secret-key" {
		t.Errorf("Load() = %q, want secret-key", key)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat(%s) error: %v", path, err)
	}

	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("key file mode = %o, want 600", perm)
	}

	dirInfo, err := os.Stat(filepath.Dir(path))
	if err != nil {
		t.Fatalf("Stat(%s) error: %v", filepath.Dir(path), err)
	}

	if perm := dirInfo.Mode().Perm(); perm != 0o700 {
		t.Errorf("config directory mode = %o, want 700", perm)
	}

	if err := store.Delete(); err != nil {
		t.Fatalf("Delete() error: %v", err)
	}

	if _, err := store.Load(); !errors.Is(err, ErrNotFound) {
		t.Errorf("Load() after Delete error = %v, want ErrNotFound", err)
	}

	if err := store.Delete(); err != nil {
		t.Errorf("second Delete() error: %v, want nil", err)
	}
}

func TestFileStoreRejectsAnEmptyKey(t *testing.T) {
	store := NewFileStoreAt(filepath.Join(t.TempDir(), "novelai_api_key"))

	if err := store.Save("   "); !errors.Is(err, ErrEmptyKey) {
		t.Errorf("Save() error = %v, want ErrEmptyKey", err)
	}
}

func TestFileStoreTreatsABlankFileAsMissing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "novelai_api_key")
	if err := os.WriteFile(path, []byte(" \n\t "), 0o600); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	if _, err := NewFileStoreAt(path).Load(); !errors.Is(err, ErrNotFound) {
		t.Errorf("Load() error = %v, want ErrNotFound", err)
	}
}

func TestAutoStoreFallsBackToTheFile(t *testing.T) {
	keyring.MockInitWithError(errors.New("keyring unavailable"))

	file := NewFileStoreAt(filepath.Join(t.TempDir(), "novelai_api_key"))
	store := newAutoStore(KeyringStore{}, file)

	if err := store.Save("fallback-key"); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	key, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if key != "fallback-key" {
		t.Errorf("Load() = %q, want fallback-key", key)
	}
}

func TestAutoStorePrefersTheKeyringAndRemovesTheFallback(t *testing.T) {
	keyring.MockInit()

	path := filepath.Join(t.TempDir(), "novelai_api_key")
	file := NewFileStoreAt(path)
	if err := file.Save("old-file-key"); err != nil {
		t.Fatalf("file Save() error: %v", err)
	}

	store := newAutoStore(KeyringStore{}, file)

	if err := store.Save("keyring-key"); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("file fallback still exists after keyring save: %v", err)
	}

	key, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if key != "keyring-key" {
		t.Errorf("Load() = %q, want keyring-key", key)
	}

	if err := store.Delete(); err != nil {
		t.Fatalf("Delete() error: %v", err)
	}

	if _, err := store.Load(); !errors.Is(err, ErrNotFound) {
		t.Errorf("Load() after Delete error = %v, want ErrNotFound", err)
	}
}

func TestAutoStoreLogsOutWhenTheKeyringIsUnavailableAndTheFileWasUsed(t *testing.T) {
	keyring.MockInitWithError(errors.New("keyring unavailable"))

	path := filepath.Join(t.TempDir(), "novelai_api_key")
	file := NewFileStoreAt(path)
	if err := file.Save("fallback-key"); err != nil {
		t.Fatalf("file Save() error: %v", err)
	}

	if err := newAutoStore(KeyringStore{}, file).Delete(); err != nil {
		t.Fatalf("Delete() error: %v", err)
	}

	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("file fallback still exists after Delete: %v", err)
	}
}

func TestAutoStoreReportsAKeyringDeleteFailureThatLeavesTheKey(t *testing.T) {
	keyring := &failingDeleteStore{key: "still-there"}
	file := NewFileStoreAt(filepath.Join(t.TempDir(), "novelai_api_key"))

	err := newAutoStore(keyring, file).Delete()
	if err == nil || !strings.Contains(err.Error(), "delete failed") {
		t.Errorf("Delete() error = %v, want the keyring failure", err)
	}
}

type failingDeleteStore struct {
	key string
}

func (s *failingDeleteStore) Load() (string, error) {
	return s.key, nil
}

func (s *failingDeleteStore) Save(string) error {
	return nil
}

func (s *failingDeleteStore) Delete() error {
	return errors.New("delete failed")
}

func TestAutoStoreReportsNotFoundWhenBothAreEmpty(t *testing.T) {
	keyring.MockInit()

	store := newAutoStore(KeyringStore{}, NewFileStoreAt(filepath.Join(t.TempDir(), "novelai_api_key")))

	if _, err := store.Load(); !errors.Is(err, ErrNotFound) {
		t.Errorf("Load() error = %v, want ErrNotFound", err)
	}
}

func TestAutoStoreUsesTheFileWhenTheKeyringErrors(t *testing.T) {
	keyring.MockInitWithError(errors.New("keyring unavailable"))

	file := NewFileStoreAt(filepath.Join(t.TempDir(), "novelai_api_key"))
	if err := file.Save("file-key"); err != nil {
		t.Fatalf("file Save() error: %v", err)
	}

	key, err := newAutoStore(KeyringStore{}, file).Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if key != "file-key" {
		t.Errorf("Load() = %q, want file-key", key)
	}
}

func TestOpenRejectsAnUnknownBackend(t *testing.T) {
	t.Setenv(envBackend, "carrier-pigeon")

	if _, err := Open(); err == nil {
		t.Error("Open() error = nil, want a rejection")
	}
}

func TestOpenSelectsTheFileBackend(t *testing.T) {
	t.Setenv(envBackend, BackendFile)

	store, err := Open()
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}

	if _, ok := store.(*FileStore); !ok {
		t.Errorf("Open() = %T, want *FileStore", store)
	}
}

func TestOpenSelectsTheKeyringBackend(t *testing.T) {
	t.Setenv(envBackend, BackendKeyring)

	store, err := Open()
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}

	if _, ok := store.(KeyringStore); !ok {
		t.Errorf("Open() = %T, want KeyringStore", store)
	}
}
