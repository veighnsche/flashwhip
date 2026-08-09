package tools

import (
	"bytes"
	"compress/gzip"
	"testing"
)

func TestDecompressGzip(t *testing.T) {
	rawText := "Hello Flashwhip Decompress Gzip Test!"
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	_, _ = gw.Write([]byte(rawText))
	_ = gw.Close()

	gzippedData := buf.Bytes()

	decompressed, err := decompressGzip(gzippedData)
	if err != nil {
		t.Fatalf("decompressGzip failed: %v", err)
	}

	if string(decompressed) != rawText {
		t.Errorf("decompressed = %q, want %q", string(decompressed), rawText)
	}
}

func TestDecompressGzip_PlainData(t *testing.T) {
	plainData := []byte("Plain uncompressed text")
	decompressed, err := decompressGzip(plainData)
	if err != nil {
		t.Fatalf("decompressGzip plain data failed: %v", err)
	}

	if string(decompressed) != string(plainData) {
		t.Errorf("decompressed plain data = %q, want %q", string(decompressed), string(plainData))
	}
}
