package credential

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrNotFound = errors.New("no stored NovelAI API key")
	ErrEmptyKey = errors.New("NovelAI API key is empty")
)

type Store interface {
	Load() (string, error)
	Save(key string) error
	Delete() error
}

const envBackend = "NOVEL_CREDENTIAL_BACKEND"

const (
	BackendAuto    = "auto"
	BackendKeyring = "keyring"
	BackendFile    = "file"
)

func Open() (Store, error) {
	backend := strings.ToLower(strings.TrimSpace(os.Getenv(envBackend)))

	switch backend {
	case "", BackendAuto:
		file, err := NewFileStore()
		if err != nil {
			return nil, err
		}

		return newAutoStore(KeyringStore{}, file), nil
	case BackendKeyring:
		return KeyringStore{}, nil
	case BackendFile:
		return NewFileStore()
	default:
		return nil, fmt.Errorf("%s must be %s, %s, or %s, got %q",
			envBackend, BackendAuto, BackendKeyring, BackendFile, backend)
	}
}

type FileStore struct {
	path string
}

func NewFileStore() (*FileStore, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("find config directory: %w", err)
	}

	return NewFileStoreAt(filepath.Join(dir, "novel", "novelai_api_key")), nil
}

func NewFileStoreAt(path string) *FileStore {
	return &FileStore{path: path}
}

func (s *FileStore) Load() (string, error) {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("read API key: %w", err)
	}

	key := strings.TrimSpace(string(data))
	if key == "" {
		return "", ErrNotFound
	}

	return key, nil
}

func (s *FileStore) Save(key string) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return ErrEmptyKey
	}

	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	tmp, err := os.CreateTemp(dir, ".novelai_api_key-*")
	if err != nil {
		return fmt.Errorf("create temporary API key file: %w", err)
	}

	name := tmp.Name()
	defer func() {
		_ = os.Remove(name)
	}()

	if err := tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()

		return fmt.Errorf("set API key permissions: %w", err)
	}

	if _, err := tmp.WriteString(key); err != nil {
		_ = tmp.Close()

		return fmt.Errorf("write API key: %w", err)
	}

	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close API key file: %w", err)
	}

	if err := os.Rename(name, s.path); err != nil {
		return fmt.Errorf("store API key: %w", err)
	}

	name = ""

	return nil
}

func (s *FileStore) Delete() error {
	err := os.Remove(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("delete API key: %w", err)
	}

	return nil
}

type autoStore struct {
	keyring Store
	file    Store
}

func newAutoStore(keyring, file Store) *autoStore {
	return &autoStore{keyring: keyring, file: file}
}

func (s *autoStore) Load() (string, error) {
	key, keyringErr := s.keyring.Load()
	if keyringErr == nil {
		return key, nil
	}

	fileKey, fileErr := s.file.Load()
	if fileErr == nil {
		return fileKey, nil
	}

	if errors.Is(keyringErr, ErrNotFound) && errors.Is(fileErr, ErrNotFound) {
		return "", ErrNotFound
	}

	if !errors.Is(fileErr, ErrNotFound) {
		return "", fileErr
	}

	return "", fmt.Errorf("%w; keyring unavailable: %v", ErrNotFound, keyringErr)
}

func (s *autoStore) Save(key string) error {
	if err := s.keyring.Save(key); err == nil {
		if err := s.file.Delete(); err != nil {
			return fmt.Errorf("remove file fallback after keyring save: %w", err)
		}

		return nil
	}

	return s.file.Save(key)
}

func (s *autoStore) Delete() error {
	fileErr := s.file.Delete()
	keyringErr := s.keyring.Delete()

	if fileErr != nil {
		return fileErr
	}

	if keyringErr == nil || errors.Is(keyringErr, ErrNotFound) {
		return nil
	}

	if key, err := s.keyring.Load(); err == nil && key != "" {
		return keyringErr
	}

	return nil
}
