//go:build e2e

package main

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"os"
	"strconv"

	"cloud.google.com/go/storage"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// gcsBucketName matches the hardcoded bucket name publishPairings/
// publishSchoolsStatus write to in routes.go.
const gcsBucketName = "duda_pairings"
const gcsProjectID = "e2e-test-project"

var gcsContainer testcontainers.Container

// startFakeGCS runs fsouza/fake-gcs-server and points the Go storage client
// at it via STORAGE_EMULATOR_HOST, which cloud.google.com/go/storage honors
// natively (it also disables auth automatically when this is set) - so
// getStorageClient in routes.go needs no changes at all to work against the
// emulator. GCS doesn't auto-create buckets on first write, so the bucket
// routes.go publishes to is created here once, up front.
func startFakeGCS(ctx context.Context) error {
	// fake-gcs-server needs to be told its own externally-reachable address
	// via FAKE_GCS_EXTERNAL_URL/FAKE_GCS_PUBLIC_HOST at startup: without it,
	// uploads "succeed" (Close returns no error) but reads 404, because the
	// Go client's Reader hits the XML-style "/{bucket}/{object}" route,
	// which fake-gcs-server only serves for the host it's been told is
	// public (defaults to storage.googleapis.com). That means the host port
	// has to be known before the container starts, so it's reserved here
	// rather than left to testcontainers' usual dynamic allocation.
	hostPort, err := reserveFreePort()
	if err != nil {
		return fmt.Errorf("reserve host port for fake-gcs-server: %w", err)
	}
	externalURL := fmt.Sprintf("http://localhost:%d", hostPort)
	publicHost := fmt.Sprintf("localhost:%d", hostPort)

	containerPort := network.MustParsePort("4443/tcp")
	ctr, err := testcontainers.Run(ctx, "fsouza/fake-gcs-server:latest",
		testcontainers.WithExposedPorts("4443/tcp"),
		testcontainers.WithEnv(map[string]string{
			"FAKE_GCS_SCHEME":       "http",
			"FAKE_GCS_BACKEND":      "memory",
			"FAKE_GCS_EXTERNAL_URL": externalURL,
			"FAKE_GCS_PUBLIC_HOST":  publicHost,
		}),
		testcontainers.WithHostConfigModifier(func(hc *container.HostConfig) {
			hc.PortBindings = network.PortMap{
				containerPort: {{HostIP: netip.IPv4Unspecified(), HostPort: strconv.Itoa(hostPort)}},
			}
		}),
		testcontainers.WithWaitStrategy(wait.ForHTTP("/_internal/healthcheck").WithPort("4443/tcp")),
	)
	if err != nil {
		return fmt.Errorf("start fake-gcs-server: %w", err)
	}
	gcsContainer = ctr

	if err := os.Setenv("STORAGE_EMULATOR_HOST", externalURL); err != nil {
		return fmt.Errorf("set STORAGE_EMULATOR_HOST: %w", err)
	}

	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("storage.NewClient against emulator: %w", err)
	}
	defer client.Close()

	if err := client.Bucket(gcsBucketName).Create(ctx, gcsProjectID, nil); err != nil {
		return fmt.Errorf("create bucket %s: %w", gcsBucketName, err)
	}
	return nil
}

// reserveFreePort finds an available TCP port on the host by briefly
// binding to port 0 and reading back what the OS assigned. There's a small
// unavoidable race between closing this listener and the container binding
// the same port, but it's the standard pattern for this and fine for tests.
func reserveFreePort() (int, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}
