package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wishmatic/novel-cli/internal/cli"
)

func main() {
	if err := newRoot(cli.DefaultDeps()).Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newRoot(deps cli.Deps) *cobra.Command {
	deps = cli.WithDefaults(deps)

	root := &cobra.Command{
		Use:           "novel-cli",
		Short:         "Use NovelAI from the command line",
		Long:          "Generate images with NovelAI, or run the MCP server.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.SetIn(deps.In)
	root.SetOut(deps.Out)
	root.SetErr(deps.Err)

	root.AddCommand(
		newServeCommand(deps),
		newNovelAICommand(deps),
		newLoginCommand(deps),
		newLogoutCommand(deps),
	)

	return root
}
