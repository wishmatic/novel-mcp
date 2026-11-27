package mcp

import (
	"encoding/json"
	"fmt"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/wishmatic/novel-cli/internal/format"
	"github.com/wishmatic/novel-cli/internal/present"
)

// inlineInput is the sizing of the copy of an image that rides in a tool result, which every tool returning an image
// shares. The copy is what the caller sees in the result, and is attached only when IsInline asks for it; the image
// stored at its URL is a separate file.
type inlineInput struct {
	IsInline bool `json:"inline,omitempty" jsonschema:"attach the image itself to this result alongside its URL, in the user's and the assistant's audience, so a vision-capable model can see it; unset returns only the URL"`

	InlineMaxEdge int `json:"inline_max_edge,omitempty" jsonschema:"longest edge in pixels of the copy of the image attached to this result, which is shrunk to fit it; the image stored at its URL keeps its own size"`

	InlineMaxBytes int `json:"inline_max_bytes,omitempty" jsonschema:"byte limit for the copy of the image attached to this result, which is compressed and, where that is not enough, shrunk until it fits; the image stored at its URL keeps its own size"`
}

func (i inlineInput) inlineBudget() present.InlineBudget {
	return present.InlineBudget{MaxEdge: i.InlineMaxEdge, MaxBytes: i.InlineMaxBytes}
}

type formatInput struct {
	Format string `json:"format,omitempty" jsonschema:"output image format: png, jpeg, jxl (JPEG XL), or webp; defaults to the server's configured output format"`

	inlineInput
}

// generationOutput is the structured output shared by the image tools. URLs lists the stored images for structured
// clients; the call result's content carries each URL as text, with the image itself beside it when the call asked for
// it inline.
type generationOutput struct {
	Count int      `json:"count" jsonschema:"number of images generated"`
	URLs  []string `json:"urls" jsonschema:"URLs for the generated images"`
}

func setDefault(props map[string]*jsonschema.Schema, name string, value any) {
	raw, err := json.Marshal(value)
	if err != nil {
		panic(fmt.Sprintf("marshal default for %s: %v", name, err))
	}

	props[name].Default = raw
}

func setFormatSchema(s *jsonschema.Schema) {
	setDefault(s.Properties, "format", format.Default.String())
	setFormatEnum(s)
	setInlineDefaults(s)
}

// setInlineDefaults advertises the same budget the presentation layer falls back to for a caller that names none.
func setInlineDefaults(s *jsonschema.Schema) {
	setDefault(s.Properties, "inline_max_edge", present.DefaultInlineMaxEdge)
	setDefault(s.Properties, "inline_max_bytes", present.DefaultInlineMaxBytes)
}

func setFormatEnum(s *jsonschema.Schema) {
	names := make([]any, 0, len(format.Names()))
	for _, name := range format.Names() {
		names = append(names, name)
	}

	s.Properties["format"].Enum = names
}
