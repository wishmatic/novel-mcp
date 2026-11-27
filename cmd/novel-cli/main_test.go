package main

import (
	"strings"
	"testing"

	"github.com/wishmatic/novel-cli/internal/cli"
)

func TestHelpDoesNotRequireCredentials(t *testing.T) {
	out, _, err := runCommand(t, cli.Deps{}, "help")
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	for _, command := range []string{"serve", "novelai", "login", "logout"} {
		if !strings.Contains(out, command) {
			t.Errorf("help output does not mention %q:\n%s", command, out)
		}
	}
}

func TestHelpForASubcommand(t *testing.T) {
	out, _, err := runCommand(t, cli.Deps{}, "help", "novelai")
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	if !strings.Contains(out, "novel-cli novelai") {
		t.Errorf("help output = %q, want the novelai usage", out)
	}
}
