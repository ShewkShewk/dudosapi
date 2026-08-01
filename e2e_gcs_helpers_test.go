//go:build e2e

package main

import (
	"context"
	"io"
	"testing"

	"cloud.google.com/go/storage"
)

// readGcsBlob reads an object's full content out of gcsBucketName, failing
// the test on any error.
func readGcsBlob(t *testing.T, ctx context.Context, client *storage.Client, name string) string {
	t.Helper()
	r, err := client.Bucket(gcsBucketName).Object(name).NewReader(ctx)
	if err != nil {
		t.Fatalf("read object %s: %v", name, err)
	}
	defer r.Close()
	content, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read object %s: %v", name, err)
	}
	return string(content)
}
