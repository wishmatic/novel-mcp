package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/wishmatic/novel-cli/internal/cli"
)

func newGenCommand(deps cli.Deps) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "gen",
		Aliases: []string{"generate"},
		Short:   "Generate images",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(newGenImageCommand(deps))

	return cmd
}

func promptFromArgs(args []string) (string, error) {
	prompt := strings.TrimSpace(strings.Join(args, " "))
	if prompt == "" {
		return "", fmt.Errorf("prompt is empty")
	}

	return prompt, nil
}
