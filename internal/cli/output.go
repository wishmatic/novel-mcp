package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/wishmatic/novel-cli/internal/format"
)

func OutputFormat(name string) (format.Format, error) {
	if strings.TrimSpace(name) == "" {
		return format.Default, nil
	}

	return format.Parse(name)
}

func WritePaths(w io.Writer, paths []string) error {
	for _, path := range paths {
		if _, err := fmt.Fprintln(w, path); err != nil {
			return fmt.Errorf("write path: %w", err)
		}
	}

	return nil
}
