package dockerengine

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeStoredContext(t *testing.T, directory, name, host string, skip bool) string {
	t.Helper()
	sum := sha256.Sum256([]byte(name))
	id := hex.EncodeToString(sum[:])
	path := filepath.Join(directory, "contexts", "meta", id, "meta.json")
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	doc := map[string]any{"Name": name, "Endpoints": map[string]any{"docker": map[string]any{"Host": host, "SkipTLSVerify": skip}}}
	contents, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, contents, 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestContextStoreSelectionAndImmediateChanges(t *testing.T) {
	root := t.TempDir()
	t.Setenv("DOCKER_CONFIG", root)
	t.Setenv("DOCKER_CONTEXT", "")
	t.Setenv("DOCKER_HOST", "")
	a := writeStoredContext(t, root, "one", "unix:///tmp/one.sock", false)
	writeStoredContext(t, root, "two", "unix:///tmp/two.sock", false)
	config := filepath.Join(root, "config.json")
	write := func(v string) {
		t.Helper()
		if err := os.WriteFile(config, []byte(v), 0600); err != nil {
			t.Fatal(err)
		}
	}
	write(`{"currentContext":"one","auths":{"ignored":{"auth":"not-decoded"}}}`)
	got, err := ReadConnection(context.Background(), "", "")
	if err != nil || got.ContextName != "one" || got.Host != "unix:///tmp/one.sock" {
		t.Fatalf("selection: %v %v", got, err)
	}
	write(`{"currentContext":"two"}`)
	got, err = ReadConnection(context.Background(), "", "")
	if err != nil || got.ContextName != "two" {
		t.Fatalf("context switch: %v %v", got, err)
	}
	t.Setenv("DOCKER_HOST", "unix:///tmp/env.sock")
	got, err = ReadConnection(context.Background(), "", "")
	if err != nil || !got.FromEnv {
		t.Fatalf("host precedence: %v %v", got, err)
	}
	t.Setenv("DOCKER_CONTEXT", "one")
	got, err = ReadConnection(context.Background(), "", "")
	if err != nil || got.ContextName != "one" {
		t.Fatalf("context precedence: %v %v", got, err)
	}
	got, err = ReadConnection(context.Background(), "two", "")
	if err != nil || got.ContextName != "two" {
		t.Fatalf("explicit context: %v %v", got, err)
	}
	if err := os.WriteFile(a, []byte(`{`), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadConnection(context.Background(), "one", ""); err == nil {
		t.Fatal("corrupt context reused old endpoint")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := ReadConnection(ctx, "default", ""); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

func TestContextStoreFailureBoundsAndTLS(t *testing.T) {
	root := t.TempDir()
	t.Setenv("DOCKER_CONFIG", root)
	t.Setenv("DOCKER_CONTEXT", "")
	t.Setenv("DOCKER_HOST", "")
	got, err := ReadConnection(context.Background(), "", "")
	if err != nil || !got.FromEnv {
		t.Fatalf("missing config default: %v %v", got, err)
	}
	if _, err := ReadConnection(context.Background(), "missing", ""); err == nil {
		t.Fatal("missing context accepted")
	}
	path := writeStoredContext(t, root, "tls", "tcp://example.invalid:2376", true)
	got, err = ReadConnection(context.Background(), "tls", "")
	if err != nil {
		t.Fatal(err)
	}
	if engine, err := got.NewClient(); err == nil {
		_ = engine.Close()
		t.Fatal("insecure TLS accepted")
	}
	if !strings.Contains(got.TLSPath, filepath.Join("contexts", "tls")) {
		t.Fatal("TLS material path lost")
	}
	if err := os.WriteFile(path, []byte(strings.Repeat(" ", 1<<20+1)), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadConnection(context.Background(), "tls", ""); err == nil {
		t.Fatal("oversized metadata accepted")
	}
	t.Setenv("DOCKER_CONFIG", "relative")
	writeStoredContext(t, filepath.Join(root, "relative"), "local", "unix:///tmp/local.sock", false)
	got, err = ReadConnection(context.Background(), "local", root)
	if err != nil || got.ContextName != "local" {
		t.Fatalf("relative config: %v %v", got, err)
	}
}
