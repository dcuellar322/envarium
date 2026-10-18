package adapters

import (
	"sync"
	"testing"
)

func TestRedactorConcurrentRegistrationAndText(t *testing.T) {
	t.Parallel()
	r, err := NewRedactor(nil)
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for range 1000 {
			r.AddSecret("synthetic-secret")
		}
	}()
	go func() {
		defer wg.Done()
		for range 1000 {
			_, _ = r.RedactText("synthetic-secret")
		}
	}()
	wg.Wait()
	if got, _ := r.RedactText("synthetic-secret"); got != redactionMarker {
		t.Fatalf("redacted = %q", got)
	}
	if len(r.secrets) != 1 {
		t.Fatalf("retained %d duplicate secrets", len(r.secrets))
	}
}
