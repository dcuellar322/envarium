package adapters

import (
	"net/url"
	"strings"
)

// remoteURL is display metadata, never a credential-bearing execution input.
func remoteURL(value string) string {
	if !strings.Contains(value, "://") {
		// Preserve conventional SSH scp syntax. Other userinfo may be a PAT.
		if at := strings.LastIndex(value, "@"); at >= 0 && !strings.HasPrefix(value, "git@") {
			value = value[at+1:]
		}
		if strings.ContainsAny(value, "?#") {
			return "[REDACTED]"
		}
		return value
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme == "" {
		return "[REDACTED]"
	}
	parsed.User = nil
	parsed.RawQuery, parsed.ForceQuery, parsed.Fragment, parsed.RawFragment = "", false, "", ""
	return parsed.String()
}
