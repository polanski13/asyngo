package spec

import (
	"testing"
)

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
