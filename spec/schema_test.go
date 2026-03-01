package spec

import (
	"encoding/json"
	"testing"
)

func TestSchemaRefMarshalJSON_Ref(t *testing.T) {
	sr := SchemaRef{Ref: "#/components/schemas/User"}
	data, err := json.Marshal(sr)
	if err != nil {
		t.Fatal(err)
	}
	expected := `{"$ref":"#/components/schemas/User"}`
	if string(data) != expected {
		t.Errorf("got %s, want %s", data, expected)
	}
}

func TestSchemaRefMarshalJSON_Inline(t *testing.T) {
	sr := SchemaRef{Schema: &Schema{Type: "string", Format: "date-time"}}
	data, err := json.Marshal(sr)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	if m["type"] != "string" {
		t.Errorf("type = %v", m["type"])
	}
	if m["format"] != "date-time" {
		t.Errorf("format = %v", m["format"])
	}
	if _, ok := m["$ref"]; ok {
		t.Error("unexpected $ref in inline schema")
	}
}

func TestMessageRefMarshalJSON_Ref(t *testing.T) {
	mr := MessageRef{Ref: "#/components/messages/ticker"}
	data, err := json.Marshal(mr)
	if err != nil {
		t.Fatal(err)
	}
	expected := `{"$ref":"#/components/messages/ticker"}`
	if string(data) != expected {
		t.Errorf("got %s, want %s", data, expected)
	}
}

func TestMessageRefMarshalJSON_Inline(t *testing.T) {
	mr := MessageRef{Message: &Message{Name: "test", ContentType: "application/json"}}
	data, err := json.Marshal(mr)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	if m["name"] != "test" {
		t.Errorf("name = %v", m["name"])
	}
	if m["contentType"] != "application/json" {
		t.Errorf("contentType = %v", m["contentType"])
	}
}

func TestReferenceHelpers(t *testing.T) {
	tests := []struct {
		fn   func(string) string
		arg  string
		want string
	}{
		{ComponentSchemaRef, "User", "#/components/schemas/User"},
		{ComponentMessageRef, "ticker", "#/components/messages/ticker"},
		{ServerRef, "production", "#/servers/production"},
		{ChannelRef, "marketPair", "#/channels/marketPair"},
	}

	for _, tt := range tests {
		got := tt.fn(tt.arg)
		if got != tt.want {
			t.Errorf("fn(%q) = %q, want %q", tt.arg, got, tt.want)
		}
	}

	got := ChannelMessageRef("marketPair", "ticker")
	want := "#/channels/marketPair/messages/ticker"
	if got != want {
		t.Errorf("ChannelMessageRef = %q, want %q", got, want)
	}
}

func TestNewAsyncAPI(t *testing.T) {
	doc := NewAsyncAPI()
	if doc.AsyncAPI != "3.0.0" {
		t.Errorf("AsyncAPI = %q", doc.AsyncAPI)
	}
	if doc.Servers == nil {
		t.Error("Servers is nil")
	}
	if doc.Channels == nil {
		t.Error("Channels is nil")
	}
	if doc.Operations == nil {
		t.Error("Operations is nil")
	}
	if doc.Components == nil {
		t.Error("Components is nil")
	}
}
