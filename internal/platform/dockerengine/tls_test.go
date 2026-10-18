package dockerengine

import (
	"context"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/moby/moby/client"
)

func TestObservationPoolingPreservesVerifiedTLS(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("API-Version", "1.45")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	connection := Connection{Host: "tcp://" + server.Listener.Addr().String()}
	untrusted, err := connection.NewClient()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := untrusted.Ping(context.Background(), client.PingOptions{}); err == nil {
		t.Fatal("untrusted TLS certificate accepted")
	}
	_ = untrusted.Close()
	directory := t.TempDir()
	if err := os.Mkdir(filepath.Join(directory, "docker"), 0700); err != nil {
		t.Fatal(err)
	}
	certificate := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: server.Certificate().Raw})
	if err := os.WriteFile(filepath.Join(directory, "docker/ca.pem"), certificate, 0600); err != nil {
		t.Fatal(err)
	}
	connection.TLSPath = directory
	verified, err := connection.NewClient()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = verified.Close() }()
	if _, err := verified.Ping(context.Background(), client.PingOptions{}); err != nil {
		t.Fatalf("verified context unavailable: %v", err)
	}
}
