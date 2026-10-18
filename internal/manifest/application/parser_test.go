package application

import (
	"envarium.dev/envarium/internal/manifest/domain"
	"strings"
	"testing"
)

func TestParseYAMLRejectsUnknownFields(t *testing.T) {
	t.Parallel()
	_, err := ParseYAML([]byte(`
schemaVersion: envarium.dev/v1
kind: Project
metadata:
  id: example
  name: Example
  unreviewedCommand: curl example.invalid
repository:
  root: .
`))
	if err == nil || !strings.Contains(err.Error(), "field unreviewedCommand not found") {
		t.Fatalf("ParseYAML() error = %v, want strict unknown-field error", err)
	}
}

func TestParseYAMLRejectsImplicitShellAndDependencyCycles(t *testing.T) {
	t.Parallel()
	_, err := ParseYAML([]byte(`
schemaVersion: envarium.dev/v1
kind: Project
metadata: {id: unsafe, name: Unsafe}
repository: {root: .}
runtime:
  driver: process
  process:
    processes:
      - id: api
        command: [sh, -c, echo unsafe]
services:
  - id: api
    source: {process: api}
    dependencies: [worker]
  - id: worker
    source: {process: api}
    dependencies: [api]
`))
	if err == nil || !strings.Contains(err.Error(), "without shell: true") || !strings.Contains(err.Error(), "dependency cycle") {
		t.Fatalf("ParseYAML() error = %v", err)
	}
}

func TestParseYAMLAcceptsKeychainReferencesWithoutSecretValues(t *testing.T) {
	t.Parallel()
	manifest, err := ParseYAML([]byte(`
schemaVersion: envarium.dev/v1
kind: Project
metadata: {id: safe, name: Safe}
repository: {root: .}
runtime:
  driver: process
  process:
    secrets:
      API_TOKEN: {provider: keychain, key: envarium-test, account: developer}
    processes:
      - id: api
        command: [uv, run, app.py]
services:
  - id: api
    source: {process: api}
`))
	if err != nil {
		t.Fatal(err)
	}
	if manifest.Runtime.Process.Secrets["API_TOKEN"].Key != "envarium-test" {
		t.Fatalf("secret reference = %#v", manifest.Runtime.Process.Secrets)
	}
}

func TestHealthPortReferencesPreserveLoopbackValidation(t *testing.T) {
	ports := []domain.Port{{ID: "http", Host: 19001}}
	for _, value := range []string{"http://127.0.0.1:${ports.http}/", "http://[::1]:${ports.http}/"} {
		if !loopbackURL(healthPortReferences(value, ports)) {
			t.Fatalf("declared loopback health port rejected: %s", value)
		}
	}
	for _, value := range []string{"http://example.com:${ports.http}/", "http://127.0.0.1:${ports.missing}/", "http://${ports.http}.example.com/"} {
		if loopbackURL(healthPortReferences(value, ports)) {
			t.Fatalf("unsafe or unknown target accepted: %s", value)
		}
	}
	if !loopbackAddress(healthPortReferences("127.0.0.1:${ports.http}", ports)) || loopbackAddress("127.0.0.1:${ports.missing}") {
		t.Fatal("TCP port validation incorrect")
	}
}
