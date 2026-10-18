package application

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"testing"

	"envarium.dev/envarium/internal/team/domain"
)

func TestEffectivePolicyRequiredListsAreArraysWithAndWithoutPolicy(t *testing.T) {
	for _, installed := range []bool{false, true} {
		name := "local-only"
		if installed {
			name = "deny-all-policy"
		}
		t.Run(name, func(t *testing.T) {
			service, _ := newTeamTestService(t)
			if installed {
				public, private, err := ed25519.GenerateKey(rand.Reader)
				if err != nil {
					t.Fatal(err)
				}
				if _, err := service.TrustPublisher(context.Background(), "Owner", base64.StdEncoding.EncodeToString(public), true, Actor{}); err != nil {
					t.Fatal(err)
				}
				installTeamBundle(t, service, signTeamBundle(t, private, domain.KindPolicyPack, "policy.deny", domain.PolicyPack{}))
			}
			policy, err := service.EffectivePolicy(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			raw, err := json.Marshal(policy)
			if err != nil {
				t.Fatal(err)
			}
			var fields map[string]json.RawMessage
			if err := json.Unmarshal(raw, &fields); err != nil {
				t.Fatal(err)
			}
			for _, key := range []string{"sourceBundleIds", "allowedRemoteCapabilities", "allowedRemoteActions", "allowedRegistryPublishers"} {
				if len(fields[key]) == 0 || fields[key][0] != '[' {
					t.Errorf("%s must be an array, got %s", key, fields[key])
				}
			}
			if installed && (len(policy.SourceBundleIDs) != 1 || len(policy.AllowedRemoteCapabilities) != 0 || policy.TelemetryAllowed) {
				t.Fatalf("deny-all policy was weakened: %#v", policy)
			}
			if !installed && len(policy.SourceBundleIDs) != 0 {
				t.Fatalf("invented installed policy: %#v", policy)
			}
		})
	}
}
