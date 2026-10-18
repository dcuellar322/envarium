package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"envarium.dev/envarium/internal/transport/contract/generated"
)

func TestNativePresentationRetainsTrayAndNotificationInputs(t *testing.T) {
	t.Parallel()
	// Exercise the serialized contract consumed by Rust, including degraded
	// observations. Unavailable data must not turn into a healthy success state.
	var source desktopSnapshot
	if err := json.Unmarshal([]byte(`{
		"system":{"apiVersion":"v1","databaseSchemaVersion":18,"status":"ready","version":"1.0.0"},
		"host":{"docker":{"connected":true}},
		"projects":[
			{"project":{"id":"one","displayName":"App","primaryLocation":"/private/repo"},"runtime":{"state":"running","services":[{"id":"private-service"}]},"health":{"status":"degraded"},"warnings":[]},
			{"project":{"id":"two","displayName":"Unavailable"},"warnings":["runtime observation unavailable"]}
		],
		"workspaces":[{"id":"workspace","name":"Daily","recipes":[{"id":"private-recipe"}]}],
		"operations":[{"id":"operation","kind":"runtime.start","state":"failed","errorMessage":"private-operation-detail"}],
		"diagnosticNotifications":[{"id":"notice","title":"Service failed","detail":"Inspect health","occurrences":3}],
		"portConflictCount":2,"warnings":["host observation partial"]
	}`), &source); err != nil {
		t.Fatal(err)
	}
	result := nativePresentation(source)
	if result.System != source.System || result.Host != source.Host || result.PortConflictCount != 2 || len(result.Warnings) != 1 {
		t.Fatalf("system/host/conflict warning inputs changed: %#v", result)
	}
	if result.Projects[0].Project.ID != "one" || result.Projects[0].Project.Name != "App" || result.Projects[0].Runtime.State != "running" || result.Projects[0].Health.Status != "degraded" {
		t.Fatalf("tray status changed: %#v", result.Projects[0])
	}
	if result.Projects[1].Runtime != nil || result.Projects[1].Health != nil || len(result.Projects[1].Warnings) != 1 {
		t.Fatalf("unavailable observations changed: %#v", result.Projects[1])
	}
	if result.Workspaces[0].ID != "workspace" || result.Workspaces[0].Name != "Daily" || result.Operations[0].ID != "operation" || result.Operations[0].Kind != "runtime.start" || result.Operations[0].State != "failed" || result.Notifications[0] != (desktopPresentationNotice{"notice", "Service failed", "Inspect health", 3}) {
		t.Fatalf("tray actions or notification transitions changed: %#v", result)
	}
	raw, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	for _, unused := range []string{"private-repo", "/private/repo", "private-service", "private-recipe", "private-operation-detail", `"services"`, `"recipes"`, `"primaryLocation"`} {
		if strings.Contains(string(raw), unused) {
			t.Fatalf("native-only response retained unused data %q", unused)
		}
	}
}

func TestNativePresentationEmptyCollectionsRemainArrays(t *testing.T) {
	t.Parallel()
	result := nativePresentation(desktopSnapshot{Warnings: []string{}})
	if result.Projects == nil || result.Workspaces == nil || result.Operations == nil || result.Notifications == nil {
		t.Fatal("empty native collections must serialize as arrays")
	}
	if result.Host != (*generated.HostObservation)(nil) {
		t.Fatal("unavailable host must remain absent")
	}
}
