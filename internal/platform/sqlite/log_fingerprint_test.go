package sqlite

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"
	"time"

	runtime "envarium.dev/envarium/internal/runtime/domain"
)

func TestLogFingerprintSizeUsesTheExactDeduplicationBytes(t *testing.T) {
	t.Parallel()
	for _, message := range []string{"plain", "escaped\n\"message\"", strings.Repeat("🙂", 2_000)} {
		entry := runtime.LogEntry{ProjectID: "one", Message: message, Timestamp: time.Date(2026, 10, 17, 12, 0, 0, 0, time.UTC), Attributes: map[string]string{"b": "two", "a": "one"}}
		raw, err := json.Marshal(entry)
		if err != nil {
			t.Fatal(err)
		}
		digest, size, err := logDigest(entry)
		expected := sha256.Sum256(raw)
		if err != nil || digest != hex.EncodeToString(expected[:]) || size != len(raw) {
			t.Fatalf("fingerprint=%q size=%d want=%d err=%v", digest, size, len(raw), err)
		}
	}
}
