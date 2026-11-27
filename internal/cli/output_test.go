package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/wishmatic/novel-mcp/internal/format"
)

func TestOutputFormatUsesTheFlag(t *testing.T) {
	got, err := OutputFormat("jpeg")
	if err != nil {
		t.Fatalf("OutputFormat() error: %v", err)
	}

	if got != format.JPEG {
		t.Errorf("OutputFormat() = %q, want jpeg", got)
	}
}

func TestOutputFormatDefaultsToWebP(t *testing.T) {
	got, err := OutputFormat("")
	if err != nil {
		t.Fatalf("OutputFormat() error: %v", err)
	}

	if got != format.WebP {
		t.Errorf("OutputFormat() = %q, want webp", got)
	}
}

func TestOutputFormatRejectsAnUnknownName(t *testing.T) {
	_, err := OutputFormat("gif")
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
