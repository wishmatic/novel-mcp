package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"image"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/vmihailenco/msgpack/v5"

	"github.com/wishmatic/novel-mcp/internal/cli"
	"github.com/wishmatic/novel-mcp/internal/format"
	"github.com/wishmatic/novel-mcp/internal/novelai"
	"github.com/wishmatic/novel-mcp/internal/resolve"
)

func pngImage(t *testing.T, width, height int) []byte {
	t.Helper()

	encoded, err := format.Encode(image.NewNRGBA(image.Rect(0, 0, width, height)), format.PNG)
	if err != nil {
		t.Fatalf("encode png: %v", err)
	}

	return encoded
}

func encodeFrame(t *testing.T, event map[string]any) []byte {
	t.Helper()

	payload, err := msgpack.Marshal(event)
	if err != nil {
		t.Fatalf("marshal frame: %v", err)
	}

	frame := make([]byte, 4+len(payload))
	binary.BigEndian.PutUint32(frame, uint32(len(payload)))
	copy(frame[4:], payload)

	return frame
}

func finalFrame(t *testing.T, image []byte) []byte {
	t.Helper()

	return encodeFrame(t, map[string]any{"event_type": "final", "image": image})
}

func runCommand(t *testing.T, deps cli.Deps, args ...string) (string, string, error) {
	t.Helper()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	deps.Out = out
	deps.Err = errOut

	command := newRoot(deps)
	command.SetArgs(args)

	err := command.Execute()

	return out.String(), errOut.String(), err
}

func newClientAt(t *testing.T, baseURL string) *novelai.Client {
	t.Helper()

	return novelai.New(baseURL, "test-key")
}

func newResolver(t *testing.T) *resolve.Client {
	t.Helper()

	resolver, err := resolve.New(nil, "", nil)
	if err != nil {
		t.Fatalf("resolve.New() error: %v", err)
	}

	return resolver
}

type capturedRequest struct {
	Path          string
	Authorization string
	ContentType   string
	Body          []byte
}

func captureServer(
	t *testing.T, status int, body []byte,
) (*httptest.Server, <-chan capturedRequest) {
	t.Helper()

	requests := make(chan capturedRequest, 16)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read request body: %v", err)
		}

		requests <- capturedRequest{
			Path:          r.URL.Path,
			Authorization: r.Header.Get("Authorization"),
			ContentType:   r.Header.Get("Content-Type"),
			Body:          raw,
		}

		w.WriteHeader(status)
		_, _ = w.Write(body)
	}))
	t.Cleanup(server.Close)

	return server, requests
}

func nextRequest(t *testing.T, requests <-chan capturedRequest) capturedRequest {
	t.Helper()

	select {
	case request := <-requests:
		return request
	default:
		t.Fatal("upstream received no request")

		return capturedRequest{}
	}
}

func (r capturedRequest) jsonBody(t *testing.T) map[string]any {
	t.Helper()

	var body map[string]any
	if err := json.Unmarshal(r.Body, &body); err != nil {
		t.Fatalf("decode request body %q: %v", r.Body, err)
	}

	return body
}

func paramsOf(t *testing.T, body map[string]any) map[string]any {
	t.Helper()

	params, ok := body["parameters"].(map[string]any)
	if !ok {
		t.Fatalf("parameters = %T, want a map", body["parameters"])
	}

	return params
}

func numberField(t *testing.T, values map[string]any, key string) float64 {
	t.Helper()

	value, ok := values[key].(float64)
	if !ok {
		t.Fatalf("%s = %T, want a number", key, values[key])
	}

	return value
}

func assertNoRequest(t *testing.T, requests <-chan capturedRequest) {
	t.Helper()

	select {
	case request := <-requests:
		t.Fatalf("upstream received %s, want no request", request.Path)
	default:
	}
}

func assertNoFiles(t *testing.T, dir string) {
	t.Helper()

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir() error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("output directory contains %d files, want none", len(entries))
	}
}
