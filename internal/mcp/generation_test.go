package mcp

import (
	"testing"
)

func TestNovelAIGenerateRequestMapsInput(t *testing.T) {
	initImage := []byte("\x89PNG\r\n\x1a\n")

	in := novelaiInput{
		generationInput: generationInput{
			Prompt:            "p",
			NegativePrompt:    "n",
			SamplingMethod:    "k_dpmpp_2m",
			SamplingSteps:     30,
			Width:             768,
			Height:            512,
			CFGScale:          6.5,
			Seed:              42,
			InitImageURL:      "https://example.com/a.png",
			DenoisingStrength: 0.6,
		},
		Model: "nai-diffusion-5-full",
		Noise: 0.1,
	}

	got := novelaiGenerateRequest(in, initImage)

	if string(got.InitImage) != string(initImage) {
		t.Errorf("init image = %v, want it copied", got.InitImage)
	}

	if got.Model != "nai-diffusion-5-full" || got.Prompt != "p" || got.NegativePrompt != "n" {
		t.Errorf("prompt fields = %+v, want them copied", got)
	}

	if got.Sampler != "k_dpmpp_2m" || got.Steps != 30 {
		t.Errorf("sampling fields = %+v, want them copied", got)
	}

	if got.Width != 768 || got.Height != 512 || got.Scale != 6.5 || got.Seed != 42 {
		t.Errorf("dimensions, scale, and seed = %+v, want them copied", got)
	}

	if got.Strength != 0.6 || got.Noise != 0.1 {
		t.Errorf("strength = %v, noise = %v, want 0.6 and 0.1", got.Strength, got.Noise)
	}
}

func TestNovelAIGenerateRequestWithoutInitImage(t *testing.T) {
	got := novelaiGenerateRequest(novelaiInput{generationInput: generationInput{Prompt: "p"}}, nil)

	if got.InitImage != nil {
		t.Errorf("init image = %v, want nil, as an unset URL means txt2img", got.InitImage)
	}
}
