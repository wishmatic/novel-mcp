package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/wishmatic/novel-mcp/internal/cli"
	"github.com/wishmatic/novel-mcp/internal/credential"
)

const maxAPIKeyBytes = 4096

func newLoginCommand(deps cli.Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Store a NovelAI API key",
		Long: "Reads a NovelAI API key from STDIN and stores it in the OS keyring, with a " +
			"user-only file fallback when no keyring is available",
		Args: cobra.NoArgs,
		RunE: func(*cobra.Command, []string) error {
			return login(deps)
		},
	}
}

func login(deps cli.Deps) error {
	key, err := readAPIKey(deps)
	if err != nil {
		return err
	}

	store, err := cli.OpenCredentials(deps)
	if err != nil {
		return err
	}

	if err := store.Save(key); err != nil {
		return fmt.Errorf("store API key: %w", err)
	}

	fmt.Fprintln(deps.Out, "logged in")

	return nil
}

func readAPIKey(deps cli.Deps) (string, error) {
	if file, ok := deps.In.(*os.File); ok && term.IsTerminal(int(file.Fd())) {
		return readAPIKeyFromTerminal(deps, file)
	}

	return readAPIKeyFromReader(deps.In)
}

func readAPIKeyFromTerminal(deps cli.Deps, file *os.File) (string, error) {
	fmt.Fprint(deps.Err, "NovelAI API key: ")

	key, err := term.ReadPassword(int(file.Fd()))
	fmt.Fprintln(deps.Err)

	if err != nil {
		return "", fmt.Errorf("read API key: %w", err)
	}

	return normalizeAPIKey(key)
}

func readAPIKeyFromReader(r io.Reader) (string, error) {
	data, err := io.ReadAll(io.LimitReader(r, maxAPIKeyBytes+1))
	if err != nil {
		return "", fmt.Errorf("read API key: %w", err)
	}

	if len(data) > maxAPIKeyBytes {
		return "", fmt.Errorf("API key is longer than %d bytes", maxAPIKeyBytes)
	}

	return normalizeAPIKey(data)
}

func normalizeAPIKey(data []byte) (string, error) {
	key := strings.TrimSpace(string(data))
	if key == "" {
		return "", credential.ErrEmptyKey
	}

	if strings.ContainsAny(key, "\r\n") {
		return "", fmt.Errorf("API key must be a single line")
	}

	return key, nil
}
