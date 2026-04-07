package goddgs

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/andybalholm/brotli"
	"github.com/klauspost/compress/zstd"
)

func TestDecompressResponse_Gzip(t *testing.T) {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	_, _ = zw.Write([]byte("hello gzip"))
	_ = zw.Close()

	resp := &http.Response{
		Header: http.Header{"Content-Encoding": []string{"gzip"}, "Content-Length": []string{"999"}},
		Body:   io.NopCloser(bytes.NewReader(buf.Bytes())),
	}
	decompressResponse(resp)
	got, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(got) != "hello gzip" {
		t.Fatalf("got=%q", string(got))
	}
	if err := resp.Body.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	if resp.Header.Get("Content-Encoding") != "" {
		t.Fatalf("content-encoding not cleared: %q", resp.Header.Get("Content-Encoding"))
	}
}

func TestDecompressResponse_Brotli(t *testing.T) {
	var buf bytes.Buffer
	bw := brotli.NewWriter(&buf)
	_, _ = bw.Write([]byte("hello br"))
	_ = bw.Close()

	resp := &http.Response{
		Header: http.Header{"Content-Encoding": []string{"br"}},
		Body:   io.NopCloser(bytes.NewReader(buf.Bytes())),
	}
	decompressResponse(resp)
	got, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(got) != "hello br" {
		t.Fatalf("got=%q", string(got))
	}
	_ = resp.Body.Close()
}

func TestDecompressResponse_Zstd(t *testing.T) {
	var buf bytes.Buffer
	zw, err := zstd.NewWriter(&buf)
	if err != nil {
		t.Fatalf("new zstd writer: %v", err)
	}
	_, _ = zw.Write([]byte("hello zstd"))
	zw.Close()

	resp := &http.Response{
		Header: http.Header{"Content-Encoding": []string{"zstd"}},
		Body:   io.NopCloser(bytes.NewReader(buf.Bytes())),
	}
	decompressResponse(resp)
	got, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(got) != "hello zstd" {
		t.Fatalf("got=%q", string(got))
	}
	_ = resp.Body.Close()
}

func TestDecompressResponse_CommaSeparatedHeader(t *testing.T) {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	_, _ = zw.Write([]byte("combo"))
	_ = zw.Close()

	resp := &http.Response{
		Header: http.Header{"Content-Encoding": []string{"gzip, br"}},
		Body:   io.NopCloser(bytes.NewReader(buf.Bytes())),
	}
	decompressResponse(resp)
	got, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(got) != "combo" {
		t.Fatalf("got=%q", string(got))
	}
	_ = resp.Body.Close()
}

func TestDecompressResponse_Noop(t *testing.T) {
	body := []byte("plain text")
	resp := &http.Response{
		Header: http.Header{},
		Body:   io.NopCloser(bytes.NewReader(body)),
	}
	decompressResponse(resp)
	got, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(got) != "plain text" {
		t.Fatalf("got=%q", string(got))
	}
	_ = resp.Body.Close()
}

func TestDecompressResponse_AlreadyUncompressed(t *testing.T) {
	resp := &http.Response{
		Header:       http.Header{"Content-Encoding": []string{"gzip"}},
		Body:         io.NopCloser(strings.NewReader("raw")),
		Uncompressed: true,
	}
	decompressResponse(resp)
	got, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(got) != "raw" {
		t.Fatalf("got=%q", string(got))
	}
	_ = resp.Body.Close()
}
