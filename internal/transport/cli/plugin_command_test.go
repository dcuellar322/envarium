package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"envarium.dev/envarium/internal/bootstrap"
	"envarium.dev/envarium/internal/platform/localipc"
	"envarium.dev/envarium/internal/transport/contract/generated"
	"envarium.dev/envarium/internal/transport/httpclient"
)

func TestPluginCLISetupThroughLocalTransport(t *testing.T) {
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	dataDir := t.TempDir()
	packageDir := filepath.Join(dataDir, "plugins", "fixture-inspector")
	if err := os.MkdirAll(packageDir, 0o700); err != nil {
		t.Fatal(err)
	}
	executable := "envarium-fixture-plugin"
	if runtime.GOOS == "windows" {
		executable += ".exe"
	}
	build := exec.CommandContext(t.Context(), "go", "build", "-trimpath", "-o", filepath.Join(packageDir, executable), "./examples/plugins/fixture")
	build.Dir = root
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build reviewed plugin: %v\n%s", err, output)
	}
	manifest, err := os.ReadFile(filepath.Join(root, "examples", "plugins", "fixture", "plugin.json"))
	if err != nil {
		t.Fatal(err)
	}
	manifest = bytes.ReplaceAll(manifest, []byte(`"envarium-fixture-plugin"`), []byte(`"`+executable+`"`))
	manifestPath := filepath.Join(packageDir, "plugin.json")
	if err := os.WriteFile(manifestPath, manifest, 0o600); err != nil {
		t.Fatal(err)
	}
	config, err := bootstrap.DefaultConfig()
	if err != nil {
		t.Fatal(err)
	}
	config.DataDir, config.HTTPAddr = dataDir, "127.0.0.1:0"
	config.IPCAddr = localipc.DefaultAddress(dataDir)
	config.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	ctx, cancel := context.WithCancel(t.Context())
	finished := make(chan error, 1)
	go func() { finished <- bootstrap.RunDaemon(ctx, config) }()
	t.Cleanup(func() {
		cancel()
		select {
		case err := <-finished:
			if err != nil && !errors.Is(err, context.Canceled) {
				t.Errorf("stop test daemon: %v", err)
			}
		case <-time.After(10 * time.Second):
			t.Error("test daemon did not stop")
		}
	})
	client, err := httpclient.NewIPC(config.IPCAddr)
	if err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(10 * time.Second)
	for {
		if _, err := client.System(t.Context()); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("test daemon did not become ready")
		}
		time.Sleep(20 * time.Millisecond)
	}
	run := func(wantFailure bool, args ...string) json.RawMessage {
		t.Helper()
		var output bytes.Buffer
		full := append([]string{"--data-dir", dataDir, "--ipc-address", config.IPCAddr, "--json"}, args...)
		err := Execute(t.Context(), full, &output, &output)
		if wantFailure {
			if err == nil {
				t.Fatalf("%v unexpectedly succeeded: %s", args, output.String())
			}
			if classifyError(err).Code == "IDEMPOTENCY_KEY_INVALID" {
				t.Fatalf("%v failed at the transport boundary instead of the plugin policy: %v", args, err)
			}
			return nil
		}
		if err != nil {
			t.Fatalf("%v: %v", args, err)
		}
		var envelope struct {
			Data json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal(output.Bytes(), &envelope); err != nil {
			t.Fatal(err)
		}
		return envelope.Data
	}
	var plugins []generated.PluginRegistration
	if err := json.Unmarshal(run(false, "plugin", "refresh"), &plugins); err != nil || len(plugins) != 1 {
		t.Fatalf("refresh plugins: %v, %v", plugins, err)
	}
	id, fingerprint := plugins[0].Id, plugins[0].Fingerprint
	run(true, "plugin", "health", id)
	run(true, "plugin", "trust", id, "--fingerprint", fingerprint)
	run(false, "plugin", "trust", id, "--fingerprint", fingerprint, "--yes")
	run(false, "plugin", "enable", id, "--scope", "project.metadata.read", "--scope", "project.files.read", "--scope", "project.operate", "--yes")
	run(false, "plugin", "health", id)
	var proposal generated.ManifestProposal
	if err := json.Unmarshal(run(false, "add", filepath.Join(root, "test", "fixtures", "node-single-process"), "--allow-outside-roots"), &proposal); err != nil {
		t.Fatal(err)
	}
	run(false, "project", "trust", proposal.ProjectId, "--yes")
	run(false, "plugin", "inspect", id, proposal.ProjectId)
	var operation generated.Operation
	if err := json.Unmarshal(run(false, "plugin", "run", id, proposal.ProjectId, "fixture.echo", "--input", `{"reviewed":true}`, "--yes"), &operation); err != nil {
		t.Fatal(err)
	}
	deadline = time.Now().Add(10 * time.Second)
	for {
		if err := json.Unmarshal(run(false, "operation", "get", operation.Id), &operation); err != nil {
			t.Fatal(err)
		}
		if operation.State == generated.OperationStateSucceeded {
			break
		}
		if operation.State == generated.OperationStateFailed || time.Now().After(deadline) {
			t.Fatalf("plugin operation did not succeed: %#v", operation)
		}
		time.Sleep(20 * time.Millisecond)
	}
	run(false, "plugin", "disable", id)
	run(true, "plugin", "inspect", id, proposal.ProjectId)
	if err := os.WriteFile(manifestPath, append(manifest, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
	run(false, "plugin", "refresh")
	run(true, "plugin", "enable", id, "--scope", "project.metadata.read", "--yes")
}
