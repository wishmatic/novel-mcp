package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/wishmatic/novel-mcp/internal/config"
	"github.com/wishmatic/novel-mcp/internal/format"
)

func TestOutputFormatUsesTheFlag(t *testing.T) {
	got, err := OutputFormat(Deps{Config: config.Config{DefaultOutput: "png"}}, "jpeg")
	if err != nil {
		t.Fatalf("OutputFormat() error: %v", err)
	}

	if got != format.JPEG {
		t.Errorf("OutputFormat() = %q, want jpeg", got)
	}
}

func TestOutputFormatUsesTheConfiguredDefault(t *testing.T) {
	got, err := OutputFormat(Deps{Config: config.Config{DefaultOutput: "jxl"}}, "")
	if err != nil {
		t.Fatalf("OutputFormat() error: %v", err)
	}

	if got != format.JXL {
		t.Errorf("OutputFormat() = %q, want jxl", got)
	}
}

func TestOutputFormatFallsBackToThePackageDefault(t *testing.T) {
	got, err := OutputFormat(Deps{Config: config.Config{Host: "localhost"}}, "")
	if err != nil {
		t.Fatalf("OutputFormat() error: %v", err)
	}

	if got != format.Default {
		t.Errorf("OutputFormat() = %q, want %q", got, format.Default)
	}
}

func TestOutputFormatRejectsAnUnknownName(t *testing.T) {
	_, err := OutputFormat(Deps{Config: config.Config{Host: "localhost"}}, "gif")
	if err == nil || !strings.Contains(err.Error(), "unknown format") {
		t.Errorf("OutputFormat() error = %v, want an unknown-format rejection", err)
	}
}

func TestWritePathsWritesEachPath(t *testing.T) {
	var buf bytes.Buffer

	if err := WritePaths(&buf, []string{"first.png", "second.png"}); err != nil {
		t.Fatalf("WritePaths() error: %v", err)
	}

	if buf.String() != "first.png\nsecond.png\n" {
		t.Errorf("output = %q, want one path per line", buf.String())
	}
}
