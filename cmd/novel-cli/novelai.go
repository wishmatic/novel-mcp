package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wishmatic/novel-mcp/internal/cli"
	"github.com/wishmatic/novel-mcp/internal/format"
	"github.com/wishmatic/novel-mcp/internal/novelai"
)

const (
	defaultModel             = "nai-diffusion-5-full"
	defaultSteps             = 20
	defaultSeed              = -1
	defaultCFGScale          = 7.0
	defaultDenoisingStrength = 0.75
)

func newNovelAICommand(deps cli.Deps) *cobra.Command {
	var (
		model             string
		negativePrompt    string
		samplerName       string
		steps             int
		width             int
		height            int
		cfgScale          float64
		seed              int
		initImage         string
		initImageURL      string
		denoisingStrength float64
		noise             float64
		outputFormat      string
		out               string
	)

	cmd := &cobra.Command{
		Use:   "novelai <prompt...>",
		Short: "Generate an image through NovelAI",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			prompt, err := promptFromArgs(args)
			if err != nil {
				return err
			}

			if strings.TrimSpace(model) == "" {
				return errors.New("novelai: --model must not be empty")
			}

			client, err := cli.NovelClient(deps)
			if err != nil {
				return err
			}

			imageFormat, err := cli.OutputFormat(outputFormat)
			if err != nil {
				return err
			}

			initImageData, err := readInitImage(cmd, deps, initImage, initImageURL)
			if err != nil {
				return err
			}

			images, err := client.Generate(cmd.Context(), novelai.GenerateRequest{
				InitImage: initImageData,

				Model:          model,
				Prompt:         prompt,
				NegativePrompt: negativePrompt,
				Sampler:        samplerName,
				Steps:          steps,

				Width:  width,
				Height: height,
				Scale:  cfgScale,

				Strength: denoisingStrength,
				Noise:    noise,

				Seed: seed,
			})
			if err != nil {
				return fmt.Errorf("novelai: %w", err)
			}

			paths, err := writeImages(out, images, imageFormat)
			if err != nil {
				return fmt.Errorf("novelai: %w", err)
			}

			return cli.WritePaths(deps.Out, paths)
		},
	}

	cmd.Flags().StringVar(&model, "model", defaultModel,
		"NovelAI model id to generate with")
	cmd.Flags().StringVar(&negativePrompt, "negative-prompt", "",
		"things to avoid in the generated image")
	cmd.Flags().StringVar(&samplerName, "sampler-name", novelai.DefaultSampler, "NovelAI sampler")
	cmd.Flags().IntVar(&steps, "steps", defaultSteps, "number of sampling steps")
	cmd.Flags().IntVar(&width, "width", 0, "output width; defaults to 512 or the init image's width")
	cmd.Flags().IntVar(&height, "height", 0,
		"output height; defaults to 512 or the init image's height")
	cmd.Flags().Float64Var(&cfgScale, "cfg-scale", defaultCFGScale, "classifier-free guidance scale")
	cmd.Flags().IntVar(&seed, "seed", defaultSeed, "random seed; -1 uses a random seed")
	cmd.Flags().StringVar(&initImage, "init-image", "",
		"path to a local image to transform instead of generating from scratch")
	cmd.Flags().StringVar(&initImageURL, "init-image-url", "",
		"URL of an image to transform instead of generating from scratch")
	cmd.Flags().Float64Var(&denoisingStrength, "denoising-strength", defaultDenoisingStrength,
		"how much of the init image to change")
	cmd.Flags().Float64Var(&noise, "noise", 0, "extra image noise; only used with an init image")
	cmd.Flags().StringVar(&outputFormat, "format", format.Default.String(),
		"output format: png, jpeg, jxl, or webp")
	cmd.Flags().StringVar(&out, "out", ".", "directory for generated images")

	return cmd
}

func promptFromArgs(args []string) (string, error) {
	prompt := strings.TrimSpace(strings.Join(args, " "))
	if prompt == "" {
		return "", fmt.Errorf("prompt is empty")
	}

	return prompt, nil
}

func readInitImage(cmd *cobra.Command, deps cli.Deps, path, address string) ([]byte, error) {
	switch {
	case path != "" && address != "":
		return nil, errors.New("novelai: pass only one of --init-image or --init-image-url")
	case path != "":
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("novelai: read --init-image: %w", err)
		}

		return data, nil
	case address != "":
		resolver, err := cli.ImageResolver(deps)
		if err != nil {
			return nil, err
		}

		image, err := resolver.Resolve(cmd.Context(), address)
		if err != nil {
			return nil, fmt.Errorf("novelai: resolve init image: %w", err)
		}

		return image.Data, nil
	default:
		return nil, nil
	}
}

func writeImages(dir string, images [][]byte, imageFormat format.Format) ([]string, error) {
	paths := make([]string, 0, len(images))

	for _, image := range images {
		data, err := format.Convert(image, imageFormat)
		if err != nil {
			return nil, err
		}

		path, err := cli.WriteImage(dir, data, imageFormat.MediaType())
		if err != nil {
			return nil, err
		}

		paths = append(paths, path)
	}

	return paths, nil
}
