package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var imageExtensions = map[string]string{
	"image/png":  "png",
	"image/jpeg": "jpg",
	"image/jxl":  "jxl",
	"image/webp": "webp",
}

func WriteImage(dir string, data []byte, mediaType string) (string, error) {
	ext, ok := imageExtensions[strings.ToLower(strings.TrimSpace(mediaType))]
	if !ok {
		return "", fmt.Errorf("cannot store media type %q", mediaType)
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create output directory: %w", err)
	}

	stamp := time.Now().UTC().Format("20060102-150405")

	tmp, err := os.CreateTemp(dir, "novel-image-"+stamp+"-*.tmp")
	if err != nil {
		return "", fmt.Errorf("create output file: %w", err)
	}

	name := tmp.Name()
	defer func() {
		_ = os.Remove(name)
	}()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()

		return "", fmt.Errorf("write output file: %w", err)
	}

	if err := tmp.Chmod(0o644); err != nil {
		_ = tmp.Close()

		return "", fmt.Errorf("set output permissions: %w", err)
	}

	if err := tmp.Close(); err != nil {
		return "", fmt.Errorf("close output file: %w", err)
	}

	final := filepath.Join(dir, strings.TrimSuffix(filepath.Base(name), ".tmp")+"."+ext)

	if err := os.Rename(name, final); err != nil {
		return "", fmt.Errorf("store output file: %w", err)
	}

	name = ""

	return final, nil
}
