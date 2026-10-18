package adapters

import "testing"

func TestRemoteURLNeverReturnsCredentials(t *testing.T) {
	t.Parallel()
	for _, test := range []struct{ input, want string }{
		{"https://alice:synthetic-password@example.com/org/repo", "https://example.com/org/repo"},
		{"https://synthetic-token@example.com/org/repo?token=private#private", "https://example.com/org/repo"},
		{"https://user%40name:p%40ss@example.com/repo", "https://example.com/repo"},
		{"ssh://git@example.com/org/repo", "ssh://example.com/org/repo"},
		{"git@example.com:org/repo.git", "git@example.com:org/repo.git"},
		{"synthetic-token@example.com:org/repo.git", "example.com:org/repo.git"},
		{"https://bad%xx:secret@example.com/repo", "[REDACTED]"},
		{"https://example.com/org/repo", "https://example.com/org/repo"},
	} {
		if got := remoteURL(test.input); got != test.want {
			t.Errorf("sanitized remote differs: got %q want %q", got, test.want)
		}
	}
}
