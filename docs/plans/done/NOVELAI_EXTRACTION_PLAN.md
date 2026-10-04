# NovelAI extraction into novel-mcp

## Goal

Take NovelAI image generation out of `neo-mcp` and stand it up here, so that this repo serves the `novelai`
tool on its own and a call to it behaves exactly as it did inside `neo-mcp`. `neo-mcp` keeps its other
backends.

## Scope

In scope:

- `internal/novelai`, plus the image pipeline the tool reads and writes through:
  `internal/format`, `internal/present`, `internal/resolve`, `internal/sourcemap`,
  `internal/store`, and `internal/utils`.
- The `internal/mcp` plumbing the tool shares: `Clients`, `generationInput`, `publishImages`, the
  inline budget, and the tool annotations.
- Configuration and wiring for `NOVELAI_API_KEY`, the image store, and the input URL resolver.
- `AGENTS.md`, `docs/`, the README, the Dockerfile, `.gitignore`, and `.env.example` made to describe
  this repo rather than `neo-mcp`.

Out of scope:

- `forge`, `nanogpt`, `bgkill`, `edit`, and `convert`. None of them are NovelAI, so this repo registers
  no tool for them and carries none of their code.

## Design

- `novelai` is registered only when `NOVELAI_API_KEY` is set, which is how `neo-mcp` gated it. With no
  key the server comes up serving no tools at all.
- The tool's description, input schema, annotations, outgoing NovelAI request, and call result are
  copied verbatim. The only differences from `neo-mcp` are the module path and the server's own
  `serverInfo` name, which each repo takes from itself.
- The image pipeline is copied whole rather than trimmed to the NovelAI path, so a stored URL, the
  attached inline copy, and the image host that serves it behave as they did there.
- `Clients` keeps only the fields the NovelAI path uses, and `registerTools` keeps only its one
  branch, so nothing here can register a tool whose backend is gone.
- The module path became `github.com/wishmatic/novel-mcp` and the server name `novel-mcp`: both were
  left over from the template's `go-mcp`, and the git remote is `wishmatic/novel-mcp`.

## Implementation units

### 1. Backend and pipeline packages

- Copy `internal/novelai` and its dependencies unchanged, rewriting only the module path.
- Keep their tests, which cover the NovelAI request shape, the msgpack and SSE stream decoding, the
  format conversions, the inline shrink, the URL resolver, and the store.

### 2. The `novelai` tool

- Copy `novelai.go`, `shared.go`, `generation.go`, `publish.go`, and `annotations.go` unchanged.
- Trim `Clients` and `registerTools` to the NovelAI backend.
- Keep the tool's tests, dropping the fixtures and expectations that belonged to the removed tools.

### 3. Configuration and wiring

- `Config` keeps `HOST`, `PORT`, `LOG_LEVEL`, `API_KEY`, `OUTPUT_FORMAT`, `PUBLIC_HOST`, `FILES_DIR`,
  `IMAGE_URL_MAP`, and `NOVELAI_API_KEY`, and loses `SD_URL` and `NANOGPT_PROVIDERS`.
- `server.New` builds the store, the source map, and the resolver, then the NovelAI client when the
  key is set, and mounts the store's `/i/*` route and the protected `/mcp`.
- The Dockerfile pre-creates `/data` as uid 65532 and works from it, so a bind mount keeps stored
  images.

## Acceptance criteria

1. [x] `go build ./...`, `go vet ./...`, `gofmt -l .`, and `go test ./... -race -count=1` are clean,
       which is what CI runs.
2. [x] `internal/{novelai,format,present,resolve,sourcemap,store,utils}` are byte-identical to
       `neo-mcp`'s copies once the module path is normalised.
3. [x] The `novelai` tool's `ListTools` JSON - name, description, input schema, and annotations - is
       byte-identical to `neo-mcp`'s.
4. [x] A call posts the same body to `/ai/generate-image-stream`, and returns the same text URL, image
       block, MIME type, audience, and `{count, urls}` structured output, with the same inline bytes.
5. [x] `NOVELAI_API_KEY` unset registers no tools; set, it registers `novelai` and nothing else, over
       the in-memory transport and over HTTP.
6. [x] `PUBLIC_HOST` and `FILES_DIR` are required of the server, `API_KEY` gates `/mcp` while
       `/healthz` and stored images stay open, and `OUTPUT_FORMAT` defaults to `webp` and is validated.
7. [x] No `NOVELAI_API_KEY`, init image bytes, or private side of `IMAGE_URL_MAP` reaches a log line.
8. [ ] The human check: one real call against NovelAI, confirming the image comes back attached at the
       size asked for and that the dimensions follow the init image on an img2img call.

## Verification

- [x] `go test ./... -race -count=1`; the `novelai` schema, handler, and log tests are `neo-mcp`'s own.
- [x] A throwaway `ListTools` and `CallTool` dump run in both modules and diffed, normalising the
       store's uuid. Only `serverInfo.name` and the module path differed.
- [ ] The human call in item 8 above.
