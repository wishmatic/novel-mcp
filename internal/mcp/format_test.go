package mcp

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"strings"
	"testing"

	"github.com/wishmatic/novel-mcp/internal/format"
)

const testImageSize = 4

func testImagePNG(t *testing.T) []byte {
	t.Helper()

	return testImagePNGAt(t, testImageSize)
}

func testImagePNGAt(t *testing.T, size int) []byte {
	t.Helper()

	img := image.NewNRGBA(image.Rect(0, 0, size, size))

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			img.SetNRGBA(x, y, color.NRGBA{R: 180, G: 40, B: 10, A: 255})
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}

	return buf.Bytes()
}

func testImageJPEG(t *testing.T) []byte {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, testImageSize, testImageSize))

	for y := 0; y < testImageSize; y++ {
		for x := 0; x < testImageSize; x++ {
			img.Set(x, y, color.RGBA{R: 20, G: 160, B: 90, A: 255})
		}
	}

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatalf("encode jpeg: %v", err)
	}

	return buf.Bytes()
}

func TestOutputFormatResolver(t *testing.T) {
	tests := []struct {
		name    string
		flag    string
		want    format.Format
		wantErr bool
	}{
		{name: "unset falls back to webp", want: format.Default},
		{name: "flag overrides the default", flag: "png", want: format.PNG},
		{name: "flag alias", flag: "jpg", want: format.JPEG},
		{name: "invalid flag", flag: "gif", wantErr: true},
	}

	h := &Clients{Log: zapNop()}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := h.outputFormat(tt.flag)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("outputFormat(%q) error = nil, want an error", tt.flag)
				}

				return
			}

			if err != nil {
				t.Fatalf("outputFormat(%q) error: %v", tt.flag, err)
			}

			if got != tt.want {
				t.Errorf("outputFormat(%q) = %q, want %q", tt.flag, got, tt.want)
			}
		})
	}
}

func TestNovelAIRejectsInvalidFormatBeforeGenerating(t *testing.T) {
	h := &Clients{Log: zapNop()}

	_, _, err := h.novelai(context.Background(), nil, novelaiInput{
		generationInput: generationInput{Prompt: "a cat", Format: "gif"},
		Model:           "nai-diffusion-5-full",
	})
	if err == nil || !strings.HasPrefix(err.Error(), "novelai:") {
		t.Fatalf("error = %v, want a novelai: prefix", err)
	}
}
