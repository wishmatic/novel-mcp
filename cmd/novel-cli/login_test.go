package main

import (
	"errors"
	"strings"
	"testing"

	"github.com/wishmatic/novel-cli/internal/cli"
	"github.com/wishmatic/novel-cli/internal/credential"
)

func TestLoginStoresATrimmedKeyFromAReader(t *testing.T) {
	store := &fakeStore{}

	out, _, err := runCommand(t, cli.Deps{In: strings.NewReader("  secret-key\n"), Credentials: store}, "login")
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	if store.key != "secret-key" {
		t.Errorf("stored key = %q, want the trimmed key", store.key)
	}

	if !strings.Contains(out, "logged in") {
		t.Errorf("output = %q, want the login confirmation", out)
	}
}

func TestLoginRejectsAnEmptyKey(t *testing.T) {
	store := &fakeStore{}

	out, _, err := runCommand(t, cli.Deps{In: strings.NewReader("  \n"), Credentials: store}, "login")
	if !errors.Is(err, credential.ErrEmptyKey) {
		t.Fatalf("Execute() error = %v, want ErrEmptyKey", err)
	}

	if store.key != "" {
		t.Errorf("stored key = %q, want nothing stored", store.key)
	}

	if strings.Contains(out, "logged in") {
		t.Errorf("output = %q, want no login confirmation", out)
	}
}

func TestLoginRejectsAKeySpanningLines(t *testing.T) {
	_, _, err := runCommand(t, cli.Deps{
		In:          strings.NewReader("first-line\nsecond-line\n"),
		Credentials: &fakeStore{},
	}, "login")
	if err == nil || !strings.Contains(err.Error(), "single line") {
		t.Errorf("Execute() error = %v, want a single-line rejection", err)
	}
}

func TestLoginSurfacesAStoreFailure(t *testing.T) {
	_, _, err := runCommand(t, cli.Deps{
		In:          strings.NewReader("secret-key"),
		Credentials: &fakeStore{saveErr: errors.New("keyring is locked")},
	}, "login")
	if err == nil || !strings.Contains(err.Error(), "keyring is locked") {
		t.Errorf("Execute() error = %v, want the store failure", err)
	}
}

func TestLogoutDeletesTheStoredKey(t *testing.T) {
	store := &fakeStore{key: "secret-key"}

	out, _, err := runCommand(t, cli.Deps{Credentials: store}, "logout")
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	if store.key != "" {
		t.Errorf("stored key = %q, want it deleted", store.key)
	}

	if !strings.Contains(out, "logged out") {
		t.Errorf("output = %q, want the logout confirmation", out)
	}
}

func TestLogoutSucceedsWithNothingStored(t *testing.T) {
	out, _, err := runCommand(t, cli.Deps{Credentials: &fakeStore{}}, "logout")
	if err != nil {
		t.Fatalf("Execute() error: %v, want a logged-out success", err)
	}

	if !strings.Contains(out, "logged out") {
		t.Errorf("output = %q, want the logout confirmation", out)
	}
}

type fakeStore struct {
	key     string
	saveErr error
}

func (s *fakeStore) Load() (string, error) {
	if s.key == "" {
		return "", credential.ErrNotFound
	}

	return s.key, nil
}

func (s *fakeStore) Save(key string) error {
	if s.saveErr != nil {
		return s.saveErr
	}

	s.key = key

	return nil
}

func (s *fakeStore) Delete() error {
	s.key = ""

	return nil
}
