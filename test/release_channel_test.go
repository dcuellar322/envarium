package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestReleaseChannelClassification(t *testing.T) {
	t.Parallel()
	tests := []struct {
		tag     string
		channel string
		valid   bool
	}{
		{tag: "v1.0.0", channel: "stable", valid: true},
		{tag: "v1.1.0-alpha.1", channel: "alpha", valid: true},
		{tag: "v1.1.0-beta.2", channel: "beta", valid: true},
		{tag: "v1.0.0-beta.rc.1", channel: "beta", valid: true},
		{tag: "v1.2.0-nightly.20260716", channel: "nightly", valid: true},
		{tag: "1.0.0", valid: false},
		{tag: "v1.0", valid: false},
		{tag: "v1.0.0-rc.1", valid: false},
		{tag: "v1.0.0;echo-unsafe", valid: false},
	}
	for _, test := range tests {
		test := test
		t.Run(test.tag, func(t *testing.T) {
			t.Parallel()
			command := exec.Command("sh", "../scripts/release-channel.sh", test.tag)
			output, err := command.CombinedOutput()
			if !test.valid {
				if err == nil {
					t.Fatalf("accepted invalid release tag; output=%q", output)
				}
				return
			}
			if err != nil {
				t.Fatalf("classify release tag: %v: %s", err, output)
			}
			if strings.TrimSpace(string(output)) != test.channel {
				t.Fatalf("channel = %q, want %q", output, test.channel)
			}
		})
	}
}

func TestReleaseChannelWritesExactPrereleaseVersion(t *testing.T) {
	t.Parallel()
	outputPath := filepath.Join(t.TempDir(), "github-output")
	command := exec.Command("sh", "../scripts/release-channel.sh", "v1.0.0-beta.rc.1")
	command.Env = append(os.Environ(), "GITHUB_OUTPUT="+outputPath)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("classify prerelease: %v: %s", err, output)
	}
	output, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(output) != "tag=v1.0.0-beta.rc.1\nchannel=beta\nversion=1.0.0-beta.rc.1\n" {
		t.Fatalf("unexpected release metadata: %q", output)
	}
}
