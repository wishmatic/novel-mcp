package credential

import (
	"errors"
	"fmt"
	"strings"

	"github.com/zalando/go-keyring"
)

const (
	keyringService = "novel-mcp"
	keyringUser    = "novelai-api-key"
)

type KeyringStore struct{}

func (KeyringStore) Load() (string, error) {
	key, err := keyring.Get(keyringService, keyringUser)
	if errors.Is(err, keyring.ErrNotFound) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("read API key from keyring: %w", err)
	}

	key = strings.TrimSpace(key)
	if key == "" {
		return "", ErrNotFound
	}

	return key, nil
}

func (KeyringStore) Save(key string) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return ErrEmptyKey
	}

	if err := keyring.Set(keyringService, keyringUser, key); err != nil {
		return fmt.Errorf("store API key in keyring: %w", err)
	}

	return nil
}

func (KeyringStore) Delete() error {
	err := keyring.Delete(keyringService, keyringUser)
	if errors.Is(err, keyring.ErrNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("delete API key from keyring: %w", err)
	}

	return nil
}
