package spec

import (
	"strings"
	"testing"
)

func TestValidateBasicValid(t *testing.T) {
	doc := NewAsyncAPI()
	doc.Info.Title = "Test"
	doc.Info.Version = "1.0.0"

	errs := doc.ValidateBasic()
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestValidateBasicMissingFields(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*AsyncAPI)
		wantErr string
	}{
		{
			name:    "missing asyncapi version",
			setup:   func(doc *AsyncAPI) { doc.AsyncAPI = "" },
			wantErr: "asyncapi version is required",
		},
		{
			name:    "missing title",
			setup:   func(doc *AsyncAPI) { doc.Info.Title = "" },
			wantErr: "info.title is required",
		},
		{
			name:    "missing version",
			setup:   func(doc *AsyncAPI) { doc.Info.Version = "" },
			wantErr: "info.version is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := NewAsyncAPI()
			doc.Info.Title = "Test"
			doc.Info.Version = "1.0.0"
			tt.setup(doc)

			errs := doc.ValidateBasic()
			if len(errs) == 0 {
				t.Fatal("expected error")
			}
			found := false
			for _, err := range errs {
				if err.Error() == tt.wantErr {
					found = true
				}
			}
			if !found {
				t.Errorf("expected error %q, got %v", tt.wantErr, errs)
			}
		})
	}
}

func TestValidateBasicEmpty(t *testing.T) {
	doc := &AsyncAPI{}
	errs := doc.ValidateBasic()
	if len(errs) != 3 {
		t.Errorf("expected 3 errors, got %d: %v", len(errs), errs)
	}
}

func TestValidateIncludesBasic(t *testing.T) {
	doc := &AsyncAPI{}
	errs := doc.Validate()
	if len(errs) < 3 {
		t.Errorf("Validate should include basic errors, got %d: %v", len(errs), errs)
	}
}

func TestValidateValid(t *testing.T) {
	doc := NewAsyncAPI()
	doc.Info.Title = "Test"
	doc.Info.Version = "1.0.0"
	doc.Channels["test"] = Channel{Address: "/test"}
	doc.Operations["recv"] = Operation{
		Action:  ActionReceive,
		Channel: NewRef(ChannelRef("test")),
	}

	errs := doc.Validate()
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestValidateMissingRequired(t *testing.T) {
	doc := &AsyncAPI{}
	errs := doc.Validate()

	if len(errs) < 3 {
		t.Fatalf("expected at least 3 errors, got %d: %v", len(errs), errs)
	}

	found := map[string]bool{}
	for _, err := range errs {
		found[err.Error()] = true
	}
	if !found["asyncapi version is required"] {
		t.Error("missing asyncapi version error")
	}
	if !found["info.title is required"] {
		t.Error("missing title error")
	}
	if !found["info.version is required"] {
		t.Error("missing version error")
	}
}

func TestValidateBrokenChannelRef(t *testing.T) {
	doc := NewAsyncAPI()
	doc.Info.Title = "Test"
	doc.Info.Version = "1.0.0"
	doc.Operations["recv"] = Operation{
		Action:  ActionReceive,
		Channel: NewRef(ChannelRef("nonexistent")),
	}

	errs := doc.Validate()
	if len(errs) == 0 {
		t.Fatal("expected error for broken channel ref")
	}

	found := false
	for _, err := range errs {
		if err != nil {
			found = true
		}
	}
	if !found {
		t.Error("expected at least one validation error")
	}
}

func TestValidateBrokenSchemaRef(t *testing.T) {
	doc := NewAsyncAPI()
	doc.Info.Title = "Test"
	doc.Info.Version = "1.0.0"
	doc.Components.Messages["msg"] = &Message{
		Name:    "msg",
		Payload: NewSchemaRef(ComponentSchemaRef("Missing")),
	}

	errs := doc.Validate()
	if len(errs) == 0 {
		t.Fatal("expected error for broken schema ref")
	}
}

func TestValidateBrokenMessageRef(t *testing.T) {
	doc := NewAsyncAPI()
	doc.Info.Title = "Test"
	doc.Info.Version = "1.0.0"
	doc.Channels["test"] = Channel{
		Address: "/test",
		Messages: map[string]MessageRef{
			"msg": {Ref: ComponentMessageRef("nonexistent")},
		},
	}

	errs := doc.Validate()
	if len(errs) == 0 {
		t.Fatal("expected error for broken message ref")
	}
}

func TestValidateEmptyChannelAddress(t *testing.T) {
	doc := NewAsyncAPI()
	doc.Info.Title = "Test"
	doc.Info.Version = "1.0.0"
	doc.Channels["test"] = Channel{}

	errs := doc.Validate()
	found := false
	for _, err := range errs {
		if err.Error() == `channel "test": address is required` {
			found = true
		}
	}
	if !found {
		t.Errorf("expected channel address error, got %v", errs)
	}
}

func TestValidateOneOfBrokenRef(t *testing.T) {
	doc := NewAsyncAPI()
	doc.Info.Title = "Test"
	doc.Info.Version = "1.0.0"
	doc.Components.Messages["event"] = &Message{
		Name: "event",
		Payload: NewInlineSchema(&Schema{
			OneOf: []*SchemaRef{
				NewSchemaRef(ComponentSchemaRef("Missing")),
				NewSchemaRef(ComponentSchemaRef("AlsoMissing")),
			},
			Discriminator: "eventType",
		}),
	}

	errs := doc.Validate()
	if len(errs) < 2 {
		t.Fatalf("expected at least 2 errors for broken oneOf refs, got %d: %v", len(errs), errs)
	}
	for _, err := range errs {
		if err == nil {
			t.Error("nil error in list")
		}
	}
}

func TestValidateOneOfValidRef(t *testing.T) {
	doc := NewAsyncAPI()
	doc.Info.Title = "Test"
	doc.Info.Version = "1.0.0"
	doc.Components.Schemas["TickerPayload"] = &Schema{Type: "object"}
	doc.Components.Schemas["OrderBookPayload"] = &Schema{Type: "object"}
	doc.Components.Messages["event"] = &Message{
		Name: "event",
		Payload: NewInlineSchema(&Schema{
			OneOf: []*SchemaRef{
				NewSchemaRef(ComponentSchemaRef("TickerPayload")),
				NewSchemaRef(ComponentSchemaRef("OrderBookPayload")),
			},
			Discriminator: "eventType",
		}),
	}

	errs := doc.Validate()
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestValidateOperationMissingAction(t *testing.T) {
	doc := NewAsyncAPI()
	doc.Info.Title = "Test"
	doc.Info.Version = "1.0.0"
	doc.Channels["test"] = Channel{Address: "/test"}
	doc.Operations["op"] = Operation{
		Channel: NewRef(ChannelRef("test")),
	}

	errs := doc.Validate()
	found := false
	for _, err := range errs {
		if err.Error() == `operation "op": action is required` {
			found = true
		}
	}
	if !found {
		t.Errorf("expected action error, got %v", errs)
	}
}

func TestValidateOperationMessageRefBroken(t *testing.T) {
	doc := NewAsyncAPI()
	doc.Info.Title = "Test"
	doc.Info.Version = "1.0.0"
	doc.Channels["events"] = Channel{
		Address:  "/events",
		Messages: map[string]MessageRef{"good": {Ref: ComponentMessageRef("good")}},
	}
	doc.Components.Messages["good"] = &Message{Name: "good"}
	doc.Operations["recv"] = Operation{
		Action:  ActionReceive,
		Channel: NewRef(ChannelRef("events")),
		Messages: []Reference{
			NewRef(ChannelMessageRef("events", "missing")),
		},
	}

	errs := doc.Validate()
	if len(errs) == 0 {
		t.Fatal("expected error for missing message ref")
	}
	found := false
	for _, e := range errs {
		if strings.Contains(e.Error(), `message "missing" not found`) {
			found = true
		}
	}
	if !found {
		t.Errorf("expected missing-message error, got %v", errs)
	}
}

func TestValidateOperationReplyRefs(t *testing.T) {
	doc := NewAsyncAPI()
	doc.Info.Title = "Test"
	doc.Info.Version = "1.0.0"
	doc.Channels["main"] = Channel{
		Address:  "/main",
		Messages: map[string]MessageRef{"req": {Ref: ComponentMessageRef("req")}},
	}
	doc.Components.Messages["req"] = &Message{Name: "req"}
	doc.Operations["op"] = Operation{
		Action:  ActionSend,
		Channel: NewRef(ChannelRef("main")),
		Reply: &OperationReply{
			Channel:  &Reference{Ref: ChannelRef("nonexistent")},
			Messages: []Reference{NewRef(ChannelMessageRef("main", "ghost"))},
		},
	}

	errs := doc.Validate()
	if len(errs) < 2 {
		t.Fatalf("expected at least 2 errors (reply channel + reply message), got %d: %v", len(errs), errs)
	}
}
