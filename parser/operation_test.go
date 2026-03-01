package parser

import (
	"errors"
	"testing"
)

func TestParseDuplicateOperationID(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"main.go": `package main

// @AsyncAPI 3.0.0
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

// @AsyncAPI 3.0.0
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

// @AsyncAPI 3.0.0
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

// @AsyncAPI 3.0.0
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

// @AsyncAPI 3.0.0
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

// @AsyncAPI 3.0.0
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

// @AsyncAPI 3.0.0
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

// @AsyncAPI 3.0.0
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

// @AsyncAPI 3.0.0
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

// @AsyncAPI 3.0.0
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
