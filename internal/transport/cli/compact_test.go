package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestCompactMachineOutputPreservesEnvelope(t *testing.T) {
	t.Parallel()
	var pretty, compact, stderr bytes.Buffer
	if err := Execute(context.Background(), []string{"--json", "version"}, &pretty, &stderr); err != nil {
		t.Fatal(err)
	}
	if err := Execute(context.Background(), []string{"--json", "--compact", "version"}, &compact, &stderr); err != nil {
		t.Fatal(err)
	}
	var a, b any
	if err := json.Unmarshal(pretty.Bytes(), &a); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(compact.Bytes(), &b); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(a, b) || strings.Count(compact.String(), "\n") != 1 || compact.Len() >= pretty.Len() {
		t.Fatal("compact output changed data or retained line framing overhead")
	}
}
