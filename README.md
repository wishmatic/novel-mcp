# Novel MCP

A thin MCP wrapper around NovelAI's image generation API.

## Features

- `novelai` generates an image synchronously: the call blocks until generation completes and returns
  the image.
  - Set `NOVELAI_API_KEY` to enable it; without that key no tools are registered.
  - Pass an optional `init_image_url` and it transforms that image instead of generating from
    scratch.
  - Without an init image, width and height default to 512. With one, they follow its size.
  - Calls may spend Anlas on the account the key belongs to.
- Not just PNG; WebP output by default, and JPEG and JXL are supported too: pass `format` per call for
  a different file type.
- Every call stores the image and returns its URL as text; pass `inline: true` to attach the image as
  an MCP image block as well, such that a vision-capable agent can see the returned result.
  - The attached copy is shrunk to a size budget the caller can raise or lower per call with
    `inline_max_edge` and `inline_max_bytes`; the stored image at its URL keeps its own size.
  - This also means that this isn't just an MCP; it also is a simple image hosting service!
- Private networking support; input URLs can be mapped before they are fetched, even to a local
  directory seen by the container.

## Usage

Deploy as a Docker image:

```sh
docker run -d \
  -p 8080:8080 \
  -e API_KEY=change-me \
  -e NOVELAI_API_KEY=sk-xxx \
  -e PUBLIC_HOST=http://192.168.1.10:8080 \
  -v /mnt/user/appdata/novel-mcp:/data \
  ghcr.io/wishmatic/novel-mcp:latest
```

The image runs `novel-cli serve`; the MCP endpoint is served at `/mcp`.

`/data` holds the image store, so bind-mount a host directory there to keep files across container
replacements; the container runs as uid 65532, so that directory must be writable by it.

All other configuration is optional but strongly recommended; see [.env.example](.env.example).

### Command line

`novel-cli` also generates images without the server:

```sh
go install github.com/wishmatic/novel-mcp/cmd/novel-cli@latest
```

Then store your NovelAI key once:

```sh
novel-cli login
```

And generate:

```sh
novel-cli novelai "a cat" --out ./images
novel-cli novelai "a cat" --init-image ./photo.png --out ./images
```

`--model` defaults to `nai-diffusion-5-full`. To transform an existing image instead of generating
from scratch, pass `--init-image` with a local path, or `--init-image-url` with an address the
resolver can read.

`novel-cli help` lists the commands, and `novel-cli serve` runs the MCP server locally
(`go run ./cmd/novel-cli serve`). Generated images are written to `--out` (default `.`), in `--format`
(WebP unless named), and are never hosted.

The NovelAI key is resolved from `NOVELAI_API_KEY` (environment or `.env`) first, then from what
`novel-cli login` stored: the OS keyring, with a user-only file fallback at
`$XDG_CONFIG_HOME/novel/novelai_api_key` or its platform equivalent. Set
`NOVEL_CREDENTIAL_BACKEND` to `keyring`, `file`, or `auto` (the default) to pick a backend, and use
`novel-cli logout` to delete the stored key from both. A key passed as an environment variable is
much easier to leak, even accidentally, than one in the keyring.

### Authentication

`API_KEY` is required on every `/mcp` request, sent as `Authorization: Bearer <API_KEY>`. Stored
images are served without authentication, so anyone with a URL can read one.

### Agent Model Knowledge

Your agent will need knowledge of NovelAI model ids and samplers, as no list is maintained here: the
agent supplies them. [docs/PROMPT.md](docs/PROMPT.md) is a template for the system prompt that gives
it that knowledge.

## Warnings

This MCP will be available and maintained so long as I use it, and is built for my own purposes.
Extending features via issue requests and PRs will be _considered_ but unless I find use out of it
myself, I probably won't work on those features.

This is also very bespoke to my use case. I recommend forking this and adjusting features to your
needs if it doesn't quite fit your own.

## License

Novel MCP is licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE).

This project is not affiliated with or endorsed by NovelAI. NovelAI is a trademark of its respective
owner and is used here only to describe compatibility.
