package codex

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	agents "envarium.dev/envarium/internal/agents/application"
)

func TestDisabledInstalledProviderNeverDispatches(t *testing.T) {
	directory := t.TempDir()
	name := "codex"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	if err := os.WriteFile(filepath.Join(directory, name), []byte("fixture"), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", directory)
	provider := NewProposalProvider(ProposalConfig{Disabled: true})
	if provider.Descriptor(context.Background()).Available {
		t.Fatal("disabled provider advertised as available")
	}
	if _, err := provider.ProposeManifest(context.Background(), agents.ProviderRequest{}); !errors.Is(err, agents.ErrProviderUnavailable) {
		t.Fatal(err)
	}
	if _, err := provider.Diagnose(context.Background(), agents.ProviderRequest{}); !errors.Is(err, agents.ErrProviderUnavailable) {
		t.Fatal(err)
	}
}
