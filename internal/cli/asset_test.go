package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteImageWritesTheFile(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "output")

	path, err := WriteImage(dir, []byte("image bytes"), "image/png")
	if err != nil {
		t.Fatalf("WriteImage() error: %v", err)
	}

	if filepath.Dir(path) != dir {
		t.Errorf("path = %s, want it under %s", path, dir)
	}
	if !strings.HasPrefix(filepath.Base(path), "novel-image-") {
		t.Errorf("base name = %s, want the novel-image prefix", filepath.Base(path))
	}
	if filepath.Ext(path) != ".png" {
		t.Errorf("extension = %s, want .png", filepath.Ext(path))
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error: %v", err)
	}
	if string(data) != "image bytes" {
		t.Errorf("data = %q, want image bytes", data)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat() error: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o644 {
		t.Errorf("mode = %o, want 644", perm)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir() error: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("dir contains %d entries, want only the written file", len(entries))
	}
}

func TestWriteImageUsesDistinctNames(t *testing.T) {
	dir := t.TempDir()

	first, err := WriteImage(dir, []byte("first"), "image/png")
	if err != nil {
		t.Fatalf("WriteImage() error: %v", err)
	}

	second, err := WriteImage(dir, []byte("second"), "image/png")
	if err != nil {
		t.Fatalf("WriteImage() error: %v", err)
	}

	if first == second {
		t.Errorf("both writes produced %s, want distinct names", first)
	}
}

func TestWriteImageRejectsAnUnknownMediaType(t *testing.T) {
	_, err := WriteImage(t.TempDir(), []byte("data"), "application/pdf")
	if err == nil || !strings.Contains(err.Error(), "cannot store") {
		t.Errorf("WriteImage() error = %v, want a media-type rejection", err)
	}
}

func TestWriteImageMapsTheMediaTypeToTheExtension(t *testing.T) {
	tests := map[string]string{
		"image/png":  ".png",
		"image/jpeg": ".jpg",
		"image/jxl":  ".jxl",
		"image/webp": ".webp",
	}

	for mediaType, ext := range tests {
		path, err := WriteImage(t.TempDir(), []byte("data"), mediaType)
		if err != nil {
			t.Fatalf("WriteImage(%s) error: %v", mediaType, err)
		}

		if filepath.Ext(path) != ext {
			t.Errorf("WriteImage(%s) extension = %s, want %s", mediaType, filepath.Ext(path), ext)
		}
	}
}
