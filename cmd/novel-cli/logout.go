package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/wishmatic/novel-cli/internal/cli"
)

func newLogoutCommand(deps cli.Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Delete the stored NovelAI API key",
		Args:  cobra.NoArgs,
		RunE: func(*cobra.Command, []string) error {
			return logout(deps)
		},
	}
}

func logout(deps cli.Deps) error {
	store, err := cli.OpenCredentials(deps)
	if err != nil {
		return err
	}

	if err := store.Delete(); err != nil {
		return fmt.Errorf("delete API key: %w", err)
	}

	fmt.Fprintln(deps.Out, "logged out")

	return nil
}
