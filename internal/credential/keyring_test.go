package credential

import (
	"errors"
	"testing"

	"github.com/zalando/go-keyring"
)

func TestKeyringStoreRoundTrip(t *testing.T) {
	keyring.MockInit()

	store := KeyringStore{}

	if err := store.Delete(); err != nil {
		t.Fatalf("Delete() before Save error: %v", err)
	}

	if _, err := store.Load(); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Load() before Save error = %v, want ErrNotFound", err)
	}

	if err := store.Save("  keyring-key  "); err != nil {
		t.Fatalf("Save() error: %v", err)
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

func TestKeyringStoreRejectsAnEmptyKey(t *testing.T) {
	keyring.MockInit()

	if err := (KeyringStore{}).Save(" "); !errors.Is(err, ErrEmptyKey) {
		t.Errorf("Save() error = %v, want ErrEmptyKey", err)
	}
}

func TestKeyringStoreReportsAnUnavailableKeyring(t *testing.T) {
	keyring.MockInitWithError(errors.New("keyring unavailable"))

	if _, err := (KeyringStore{}).Load(); err == nil {
		t.Error("Load() error = nil, want the keyring error")
	}
}
