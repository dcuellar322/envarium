package dockerengine

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestDecodeContextAndPreserveEndpoint(t *testing.T) {
	t.Parallel()
	connection, err := DecodeContext([]byte(`[{"Name":"desktop-linux","Endpoints":{"docker":{"Host":"unix:///tmp/desktop.sock"}},"Storage":{"TLSPath":"/tmp/tls"}}]`))
	if err != nil {
		t.Fatal(err)
	}
	if connection.ContextName != "desktop-linux" || connection.Host != "unix:///tmp/desktop.sock" || connection.TLSPath != "/tmp/tls" {
		t.Fatalf("connection = %#v", connection)
	}
	engine, err := connection.NewClient()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = engine.Close() }()
	if engine.DaemonHost() != connection.Host {
		t.Fatalf("SDK selected %q instead of context endpoint %q", engine.DaemonHost(), connection.Host)
	}
}

func TestRefusesMalformedAndUnverifiedContexts(t *testing.T) {
	t.Parallel()
	for _, input := range []string{`not json`, `[]`, `[{},{}]`, `[{"Name":"broken","Endpoints":{}}]`} {
		if _, err := DecodeContext([]byte(input)); err == nil {
			t.Errorf("accepted malformed context %s", input)
		}
	}
	for _, connection := range []Connection{
		{ContextName: "remote", Host: "ssh://example.invalid"},
		{ContextName: "insecure", Host: "tcp://example.invalid:2376", SkipTLSVerify: true},
	} {
		if engine, err := connection.NewClient(); err == nil {
			_ = engine.Close()
			t.Errorf("accepted unverified context %#v", connection)
		}
	}
}

func TestDefaultClientUsesActiveContextAndHonorsCancellation(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix fake Docker command; context decoding is tested on every platform")
	}
	directory := t.TempDir()
	docker := filepath.Join(directory, "docker")
	// Use the test executable as the Docker CLI so this exercises command
	// selection without requiring Docker or a running daemon.
	if err := os.Symlink(os.Args[0], docker); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", directory+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("ENVARIUM_FAKE_DOCKER_CONTEXT", "1")
	t.Setenv("DOCKER_HOST", "unix:///tmp/wrong-context.sock")
	t.Setenv("DOCKER_CONTEXT", "desktop-linux")
	engine, err := NewClient(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = engine.Close() }()
	if engine.DaemonHost() != "unix:///tmp/envarium-context-test.sock" {
		t.Fatalf("default observer host = %q", engine.DaemonHost())
	}
	t.Setenv("DOCKER_CONTEXT", "")
	environmentEngine, err := NewClient(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = environmentEngine.Close() }()
	if environmentEngine.DaemonHost() != "unix:///tmp/wrong-context.sock" {
		t.Fatalf("DOCKER_HOST was ignored: %q", environmentEngine.DaemonHost())
	}
	t.Setenv("DOCKER_CONTEXT", "desktop-linux")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := NewClient(ctx); err == nil {
		t.Fatal("cancelled context lookup succeeded")
	}
}

func TestMain(m *testing.M) {
	if os.Getenv("ENVARIUM_FAKE_DOCKER_CONTEXT") == "1" && strings.Join(os.Args[1:], " ") == "context inspect" {
		_, _ = os.Stdout.WriteString(`[{"Name":"desktop-linux","Endpoints":{"docker":{"Host":"unix:///tmp/envarium-context-test.sock"}}}]`)
		os.Exit(0)
	}
	os.Exit(m.Run())
}
