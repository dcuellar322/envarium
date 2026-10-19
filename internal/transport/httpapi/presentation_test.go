package httpapi

import (
	"strings"
	"testing"
)

func TestPresentationLabelsRejectArbitraryCallerData(t *testing.T) {
	for _, value := range []string{"native:dashboard", "native:projects", "browser:ports", "browser:other"} {
		if presentationLabel(value) != value {
			t.Fatalf("valid label rejected: %q", value)
		}
	}
	for _, value := range []string{"", "native:/secret/path", "browser:dashboard?bootstrap=secret", "native:logs\nsecret", "agent:dashboard", strings.Repeat("x", 33)} {
		if presentationLabel(value) != "" {
			t.Fatal("untrusted metadata accepted")
		}
	}
}
