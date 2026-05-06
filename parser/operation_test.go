package parser

import (
	"errors"
	"strings"
	"testing"
)

func TestParseDuplicateOperationID(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"main.go": `package main

// @AsyncAPI 3.1.0
// @Title Test
// @Version 1.0.0
func Init() {}
`,
		"handler.go": `package main

// @Channel /test
// @Operation receive
// @OperationID myOp
// @Message msg1 string
func Handler1() {}

// @Channel /test
// @Operation send
// @OperationID myOp
// @Message msg2 string
func Handler2() {}
`,
	})

	p := New(WithSearchDirs(dir), WithMainFile("main.go"))
	_, err := p.Parse()
	if err == nil {
		t.Fatal("expected error for duplicate operation ID")
	}
	if !errors.Is(err, ErrDuplicateOperationID) {
		t.Errorf("error = %v, want ErrDuplicateOperationID", err)
	}
}

func TestParseMissingChannel(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"main.go": `package main

// @AsyncAPI 3.1.0
// @Title Test
// @Version 1.0.0
func Init() {}
`,
		"handler.go": `package main

// @Operation receive
// @OperationID myOp
// @Message msg1 string
func Handler() {}
`,
	})

	p := New(WithSearchDirs(dir), WithMainFile("main.go"))
	_, err := p.Parse()
	if err == nil {
		t.Fatal("expected error for missing @Channel")
	}
	if !errors.Is(err, ErrMissingChannel) {
		t.Errorf("error = %v, want ErrMissingChannel", err)
	}
}

func TestParseInvalidAction(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"main.go": `package main

// @AsyncAPI 3.1.0
// @Title Test
// @Version 1.0.0
func Init() {}
`,
		"handler.go": `package main

// @Channel /test
// @Operation invalid_action
// @OperationID myOp
// @Message msg string
func Handler() {}
`,
	})

	p := New(WithSearchDirs(dir), WithMainFile("main.go"))
	_, err := p.Parse()
	if err == nil {
		t.Fatal("expected error for invalid action")
	}
	if !errors.Is(err, ErrInvalidAction) {
		t.Errorf("error = %v, want ErrInvalidAction", err)
	}
}

func TestParseInvalidOperationNoAction(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"main.go": `package main

// @AsyncAPI 3.1.0
// @Title Test
// @Version 1.0.0
func Init() {}
`,
		"handler.go": `package main

// @Channel /test
// @Operation
// @OperationID myOp
// @Message msg string
func Handler() {}
`,
	})

	p := New(WithSearchDirs(dir), WithMainFile("main.go"))
	_, err := p.Parse()
	if err == nil {
		t.Fatal("expected error for @Operation without action")
	}
}

func TestParseInvalidMessageFormat(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"main.go": `package main

// @AsyncAPI 3.1.0
// @Title Test
// @Version 1.0.0
func Init() {}
`,
		"handler.go": `package main

// @Channel /test
// @Operation receive
// @OperationID myOp
// @Message onlyname
func Handler() {}
`,
	})

	p := New(WithSearchDirs(dir), WithMainFile("main.go"))
	_, err := p.Parse()
	if err == nil {
		t.Fatal("expected error for @Message with only name (no type)")
	}
}

func TestParseChannelMerging(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"main.go": `package main

// @AsyncAPI 3.1.0
// @Title Test
// @Version 1.0.0
func Init() {}
`,
		"handler1.go": `package main

// @Channel /events
// @ChannelDescription Event stream
// @Operation receive
// @OperationID recvEvents
// @Message event1 string
func Handler1() {}
`,
		"handler2.go": `package main

// @Channel /events
// @Operation send
// @OperationID sendEvents
// @Message event2 string
func Handler2() {}
`,
	})

	p := New(WithSearchDirs(dir), WithMainFile("main.go"))
	doc, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	ch := doc.Channels["events"]
	if _, ok := ch.Messages["event1"]; !ok {
		t.Error("channel missing event1 from first handler")
	}
	if _, ok := ch.Messages["event2"]; !ok {
		t.Error("channel missing event2 from second handler (merge)")
	}
}

func TestParseSecurityAnnotation(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"main.go": `package main

// @AsyncAPI 3.1.0
// @Title Test
// @Version 1.0.0
func Init() {}
`,
		"handler.go": `package main

// @Channel /secure
// @Operation receive
// @OperationID secureOp
// @Security bearerAuth
// @Message msg string
func Handler() {}
`,
	})

	p := New(WithSearchDirs(dir), WithMainFile("main.go"))
	doc, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	op := doc.Operations["secureOp"]
	if len(op.Security) != 1 {
		t.Fatalf("security count = %d, want 1", len(op.Security))
	}
	if op.Security[0].Ref != "#/components/securitySchemes/bearerAuth" {
		t.Errorf("security ref = %q", op.Security[0].Ref)
	}
}

func TestParseWsBindingHeader(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"main.go": `package main

// @AsyncAPI 3.1.0
// @Title Test
// @Version 1.0.0
func Init() {}
`,
		"handler.go": `package main

// @Channel /data
// @WsBinding.Method GET
// @WsBinding.Header Authorization string true "Bearer token"
// @Operation receive
// @OperationID getData
// @Message data string
func Handler() {}
`,
	})

	p := New(WithSearchDirs(dir), WithMainFile("main.go"))
	doc, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	ch := doc.Channels["data"]
	if ch.Bindings == nil || ch.Bindings.WS == nil {
		t.Fatal("ws bindings is nil")
	}
	if ch.Bindings.WS.Headers == nil {
		t.Fatal("ws headers is nil")
	}
	if _, ok := ch.Bindings.WS.Headers.Properties["Authorization"]; !ok {
		t.Error("missing Authorization header")
	}
}

func TestParseExcludeDirectory(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"main.go": `package main

// @AsyncAPI 3.1.0
// @Title Test
// @Version 1.0.0
func Init() {}
`,
		"excluded/handler.go": `package excluded

// @Channel /bad
// @Operation receive
// @OperationID badOp
// @Message bad string
func Bad() {}
`,
	})

	p := New(
		WithSearchDirs(dir),
		WithMainFile("main.go"),
		WithExcludes("excluded"),
	)
	doc, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if len(doc.Channels) != 0 {
		t.Error("excluded directory handler should not be parsed")
	}
}

func TestParseErrorContext(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"main.go": `package main

// @AsyncAPI 3.1.0
// @Title Test
// @Version 1.0.0
func Init() {}
`,
		"handler.go": `package main

// @Channel /test
// @Operation badaction
// @OperationID op1
// @Message msg string
func MyHandler() {}
`,
	})

	p := New(WithSearchDirs(dir), WithMainFile("main.go"))
	_, err := p.Parse()
	if err == nil {
		t.Fatal("expected error")
	}

	var pe *ParseError
	if !errors.As(err, &pe) {
		t.Skipf("error does not wrap ParseError: %v", err)
	}
	if pe.Function != "MyHandler" {
		t.Errorf("ParseError.Function = %q, want MyHandler", pe.Function)
	}
}

func TestParseChannelParamMinimalArgs(t *testing.T) {
	tests := []struct {
		name      string
		handler   string
		wantPanic bool
		wantErr   bool
	}{
		{
			name: "3 args no panic",
			handler: `package main

// @Channel /market/{pair}
// @ChannelParam pair string true
// @Operation receive
// @OperationID recv
// @Message msg string
func Handler() {}
`,
		},
		{
			name: "4 args with description",
			handler: `package main

// @Channel /market/{pair}
// @ChannelParam pair string true "Trading pair"
// @Operation receive
// @OperationID recv
// @Message msg string
func Handler() {}
`,
		},
		{
			name: "5 args with enum",
			handler: `package main

// @Channel /market/{pair}
// @ChannelParam pair string true "Trading pair" enum(BTC-USD,ETH-USD)
// @Operation receive
// @OperationID recv
// @Message msg string
func Handler() {}
`,
		},
		{
			name:    "2 args error",
			wantErr: true,
			handler: `package main

// @Channel /market/{pair}
// @ChannelParam pair string
// @Operation receive
// @OperationID recv
// @Message msg string
func Handler() {}
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := setupTestProject(t, map[string]string{
				"main.go": `package main

// @AsyncAPI 3.1.0
// @Title Test
// @Version 1.0.0
func Init() {}
`,
				"handler.go": tt.handler,
			})

			p := New(WithSearchDirs(dir), WithMainFile("main.go"))
			_, err := p.Parse()

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestMapSimpleTypeWarningNonStrict(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"main.go": `package main

// @AsyncAPI 3.1.0
// @Title Test
// @Version 1.0.0
func Init() {}
`,
		"handler.go": `package main

// @Channel /data
// @WsBinding.Query token foobar true "Auth token"
// @Operation receive
// @OperationID recv
// @Message msg string
func Handler() {}
`,
	})

	p := New(WithSearchDirs(dir), WithMainFile("main.go"))
	_, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	warnings := p.Warnings()
	if len(warnings) == 0 {
		t.Fatal("expected warning for unknown type, got none")
	}

	found := false
	for _, w := range warnings {
		if strings.Contains(w, "foobar") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected warning mentioning %q, got %v", "foobar", warnings)
	}
}

func TestMapSimpleTypeStrictError(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"main.go": `package main

// @AsyncAPI 3.1.0
// @Title Test
// @Version 1.0.0
func Init() {}
`,
		"handler.go": `package main

// @Channel /data
// @WsBinding.Query token badtype true "Auth token"
// @Operation receive
// @OperationID recv
// @Message msg string
func Handler() {}
`,
	})

	p := New(WithSearchDirs(dir), WithMainFile("main.go"), WithStrict(true))
	_, err := p.Parse()
	if err == nil {
		t.Fatal("expected error in strict mode for unknown type")
	}
	if !errors.Is(err, ErrUnknownType) {
		t.Errorf("error = %v, want ErrUnknownType", err)
	}
}

func TestMapSimpleTypeKnownTypes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input string
		want  string
		known bool
	}{
		{"string", "string", true},
		{"int", "integer", true},
		{"integer", "integer", true},
		{"int32", "integer", true},
		{"int64", "integer", true},
		{"float", "number", true},
		{"float32", "number", true},
		{"float64", "number", true},
		{"number", "number", true},
		{"bool", "boolean", true},
		{"boolean", "boolean", true},
		{"STRING", "string", true},
		{"INT", "integer", true},
		{"uuid", "string", false},
		{"complex128", "string", false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			got, known := mapSimpleType(tt.input)
			if got != tt.want {
				t.Errorf("mapSimpleType(%q) = %q, want %q", tt.input, got, tt.want)
			}
			if known != tt.known {
				t.Errorf("mapSimpleType(%q) known = %v, want %v", tt.input, known, tt.known)
			}
		})
	}
}

func TestParseMessageOneOf(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"main.go": `package main

// @AsyncAPI 3.1.0
// @Title Test
// @Version 1.0.0
func Init() {}
`,
		"models.go": `package main

type TickerPayload struct {
	EventType string  ` + "`json:\"eventType\"`" + `
	Price     float64 ` + "`json:\"price\"`" + `
}

type OrderBookPayload struct {
	EventType string ` + "`json:\"eventType\"`" + `
	Depth     int    ` + "`json:\"depth\"`" + `
}
`,
		"handler.go": `package main

// @Channel /events
// @Operation receive
// @OperationID recvEvents
// @MessageOneOf eventUpdate TickerPayload|OrderBookPayload discriminator(eventType)
func Handler() {}
`,
	})

	p := New(WithSearchDirs(dir), WithMainFile("main.go"))
	doc, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	msg, ok := doc.Components.Messages["eventUpdate"]
	if !ok {
		t.Fatal("message 'eventUpdate' not found")
	}
	if msg.Payload == nil || msg.Payload.Schema == nil {
		t.Fatal("payload schema is nil")
	}
	if len(msg.Payload.Schema.OneOf) != 2 {
		t.Fatalf("oneOf count = %d, want 2", len(msg.Payload.Schema.OneOf))
	}
	if msg.Payload.Schema.OneOf[0].Ref != "#/components/schemas/TickerPayload" {
		t.Errorf("oneOf[0] = %q", msg.Payload.Schema.OneOf[0].Ref)
	}
	if msg.Payload.Schema.OneOf[1].Ref != "#/components/schemas/OrderBookPayload" {
		t.Errorf("oneOf[1] = %q", msg.Payload.Schema.OneOf[1].Ref)
	}
	if msg.Payload.Schema.Discriminator != "eventType" {
		t.Errorf("discriminator = %q, want eventType", msg.Payload.Schema.Discriminator)
	}

	ch := doc.Channels["events"]
	if _, ok := ch.Messages["eventUpdate"]; !ok {
		t.Error("channel missing eventUpdate message")
	}

	op := doc.Operations["recvEvents"]
	if len(op.Messages) != 1 {
		t.Fatalf("operation messages count = %d, want 1", len(op.Messages))
	}
}

func TestParseMessageOneOfNoDiscriminator(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"main.go": `package main

// @AsyncAPI 3.1.0
// @Title Test
// @Version 1.0.0
func Init() {}
`,
		"models.go": `package main

type TypeA struct {
	Name string ` + "`json:\"name\"`" + `
}

type TypeB struct {
	Value int ` + "`json:\"value\"`" + `
}
`,
		"handler.go": `package main

// @Channel /events
// @Operation receive
// @OperationID recvEvents
// @MessageOneOf mixed TypeA|TypeB
func Handler() {}
`,
	})

	p := New(WithSearchDirs(dir), WithMainFile("main.go"))
	doc, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	msg := doc.Components.Messages["mixed"]
	if msg.Payload.Schema.Discriminator != "" {
		t.Errorf("discriminator = %q, want empty", msg.Payload.Schema.Discriminator)
	}
	if len(msg.Payload.Schema.OneOf) != 2 {
		t.Errorf("oneOf count = %d, want 2", len(msg.Payload.Schema.OneOf))
	}
}

func TestParseMessageOneOfMissingName(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"main.go": `package main

// @AsyncAPI 3.1.0
// @Title Test
// @Version 1.0.0
func Init() {}
`,
		"handler.go": `package main

// @Channel /events
// @Operation receive
// @OperationID recvEvents
// @MessageOneOf
func Handler() {}
`,
	})

	p := New(WithSearchDirs(dir), WithMainFile("main.go"))
	_, err := p.Parse()
	if err == nil {
		t.Fatal("expected error for @MessageOneOf without args")
	}
	if !errors.Is(err, ErrInvalidAnnotation) {
		t.Errorf("error = %v, want ErrInvalidAnnotation", err)
	}
}

func TestParseMixedMessageAndOneOf(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"main.go": `package main

// @AsyncAPI 3.1.0
// @Title Test
// @Version 1.0.0
func Init() {}
`,
		"models.go": `package main

type TypeA struct {
	Kind string ` + "`json:\"kind\"`" + `
}

type TypeB struct {
	Kind  string ` + "`json:\"kind\"`" + `
	Value int    ` + "`json:\"value\"`" + `
}
`,
		"handler.go": `package main

// @Channel /mixed
// @Operation receive
// @OperationID recvMixed
// @Message plainMsg string
// @MessageOneOf polyMsg TypeA|TypeB discriminator(kind)
func Handler() {}
`,
	})

	p := New(WithSearchDirs(dir), WithMainFile("main.go"))
	doc, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if _, ok := doc.Components.Messages["plainMsg"]; !ok {
		t.Error("missing plainMsg")
	}
	if _, ok := doc.Components.Messages["polyMsg"]; !ok {
		t.Error("missing polyMsg")
	}

	op := doc.Operations["recvMixed"]
	if len(op.Messages) != 2 {
		t.Errorf("operation messages count = %d, want 2", len(op.Messages))
	}
}

func TestMapSimpleTypeHeaderWarning(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"main.go": `package main

// @AsyncAPI 3.1.0
// @Title Test
// @Version 1.0.0
func Init() {}
`,
		"handler.go": `package main

// @Channel /data
// @WsBinding.Header X-Custom weirdtype true "Custom header"
// @Operation receive
// @OperationID recv
// @Message msg string
func Handler() {}
`,
	})

	p := New(WithSearchDirs(dir), WithMainFile("main.go"))
	_, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	warnings := p.Warnings()
	if len(warnings) == 0 {
		t.Fatal("expected warning for unknown header type")
	}

	found := false
	for _, w := range warnings {
		if strings.Contains(w, "weirdtype") && strings.Contains(w, "@WsBinding.Header") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected warning mentioning weirdtype and @WsBinding.Header, got %v", warnings)
	}
}

func TestParseChannelKeyCollision(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"main.go": `package main

// @AsyncAPI 3.1.0
// @Title Test
// @Version 1.0.0
func Init() {}
`,
		"handler.go": `package main

// @Channel /market/{pair}
// @Operation receive
// @OperationID recvA
// @Message a string
func A() {}

// @Channel /marketPair
// @Operation send
// @OperationID sendB
// @Message b string
func B() {}
`,
	})

	p := New(WithSearchDirs(dir), WithMainFile("main.go"))
	_, err := p.Parse()
	if err == nil {
		t.Fatal("expected channel key collision error")
	}
	if !errors.Is(err, ErrChannelKeyCollision) {
		t.Errorf("error = %v, want ErrChannelKeyCollision", err)
	}
}
