//go:build e2e

package main

import (
	"context"
	"io"
	"testing"

	"github.com/a-h/templ"
)

// TestFakeGCSUploadRoundTrip proves the emulator wiring actually works by
// driving the real, unmodified getStorageClient and uploadComponent
// functions from routes.go - not a reimplementation of them - end to end:
// write an object through the app's own upload path, then read it back.
func TestFakeGCSUploadRoundTrip(t *testing.T) {
	ctx := context.Background()

	storageClient, err := getStorageClient(ctx)
	if err != nil {
		t.Fatalf("getStorageClient: %v", err)
	}
	defer storageClient.Close()

	const objectName = "e2e-smoke-test.txt"
	const content = "hello from the e2e smoke test"
	component := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := w.Write([]byte(content))
		return err
	})

	if err := uploadComponent(ctx, storageClient, component, gcsBucketName, objectName); err != nil {
		t.Fatalf("uploadComponent: %v", err)
	}

	reader, err := storageClient.Bucket(gcsBucketName).Object(objectName).NewReader(ctx)
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}
	defer reader.Close()

	got, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if string(got) != content {
		t.Fatalf("object content = %q, want %q", got, content)
	}
}
