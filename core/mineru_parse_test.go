package core

import (
	"encoding/json"
	"testing"
)

func TestMarshalRawPayloadsPreservesNonJSONBodies(t *testing.T) {
	got := marshalRawPayloads([]any{
		json.RawMessage(`{"ok":true}`),
		"<html>upstream unavailable</html>",
	})
	want := `[
  {
    "ok": true
  },
  "\u003chtml\u003eupstream unavailable\u003c/html\u003e"
]`
	if got != want {
		t.Fatalf("marshalRawPayloads() = %s", got)
	}
}
