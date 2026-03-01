package parser

import (
	"go/ast"
	"testing"
)

func TestParseAnnotationLine(t *testing.T) {
	tests := []struct {
		input    string
		wantName string
		wantArgs []string
		wantRaw  string
	}{
		{
			input:    "@Title My API",
			wantName: "Title",
			wantArgs: []string{"My", "API"},
			wantRaw:  "My API",
		},
		{
			input:    "@Version 1.0.0",
			wantName: "Version",
			wantArgs: []string{"1.0.0"},
			wantRaw:  "1.0.0",
		},
		{
			input:    "@Server production wss://ws.example.com /v1 \"Production endpoint\"",
			wantName: "Server",
			wantArgs: []string{"production", "wss://ws.example.com", "/v1", "Production endpoint"},
			wantRaw:  "production wss://ws.example.com /v1 \"Production endpoint\"",
		},
		{
			input:    "@ChannelParam pair string true \"Trading pair\" enum(BTC-USD,ETH-USD)",
			wantName: "ChannelParam",
			wantArgs: []string{"pair", "string", "true", "Trading pair", "enum(BTC-USD,ETH-USD)"},
			wantRaw:  "pair string true \"Trading pair\" enum(BTC-USD,ETH-USD)",
		},
		{
			input:    "@WsBinding.Method GET",
			wantName: "WsBinding.Method",
			wantArgs: []string{"GET"},
			wantRaw:  "GET",
		},
		{
			input:    "@Operation receive",
			wantName: "Operation",
			wantArgs: []string{"receive"},
			wantRaw:  "receive",
		},
		{
			input:    "@Message tickerUpdate TickerPayload",
			wantName: "Message",
			wantArgs: []string{"tickerUpdate", "TickerPayload"},
			wantRaw:  "tickerUpdate TickerPayload",
		},
		{
			input:    "@Tags market,realtime,v2",
			wantName: "Tags",
			wantArgs: []string{"market,realtime,v2"},
			wantRaw:  "market,realtime,v2",
		},
		{
			input:    "@AsyncAPI 3.0.0",
			wantName: "AsyncAPI",
			wantArgs: []string{"3.0.0"},
			wantRaw:  "3.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			ann := parseAnnotationLine(tt.input)
			if ann == nil {
				t.Fatal("parseAnnotationLine returned nil")
			}
			if ann.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", ann.Name, tt.wantName)
			}
			if ann.Raw != tt.wantRaw {
				t.Errorf("Raw = %q, want %q", ann.Raw, tt.wantRaw)
			}
			if len(ann.Args) != len(tt.wantArgs) {
				t.Fatalf("len(Args) = %d, want %d; got %v", len(ann.Args), len(tt.wantArgs), ann.Args)
			}
			for i, want := range tt.wantArgs {
				if ann.Args[i] != want {
					t.Errorf("Args[%d] = %q, want %q", i, ann.Args[i], want)
				}
			}
		})
	}
}

func TestAnnotationSet(t *testing.T) {
	comments := &ast.CommentGroup{
		List: []*ast.Comment{
			{Text: "// @Channel /market/{pair}"},
			{Text: "// @ChannelDescription Real-time market data"},
			{Text: "// @Operation receive"},
			{Text: "// @OperationID recvMarket"},
			{Text: "// @Message ticker TickerPayload"},
			{Text: "// @Message orderbook OrderBookPayload"},
		},
	}

	as := newAnnotationSet(comments)

	if !as.Has("Channel") {
		t.Error("expected Has(Channel) = true")
	}
	if !as.Has("channel") {
		t.Error("expected Has(channel) = true (case insensitive)")
	}
	if as.Has("NonExistent") {
		t.Error("expected Has(NonExistent) = false")
	}

	ch := as.GetOne("Channel")
	if ch == nil {
		t.Fatal("GetOne(Channel) returned nil")
	}
	if ch.Raw != "/market/{pair}" {
		t.Errorf("Channel.Raw = %q", ch.Raw)
	}

	msgs := as.Get("Message")
	if len(msgs) != 2 {
		t.Fatalf("len(Get(Message)) = %d, want 2", len(msgs))
	}
	if msgs[0].Args[0] != "ticker" {
		t.Errorf("first message name = %q", msgs[0].Args[0])
	}
	if msgs[1].Args[0] != "orderbook" {
		t.Errorf("second message name = %q", msgs[1].Args[0])
	}
}

func TestAnnotationSetContinuation(t *testing.T) {
	comments := &ast.CommentGroup{
		List: []*ast.Comment{
			{Text: "// @Description This is a long description"},
			{Text: "// that continues on the next line"},
			{Text: "// and even one more line"},
			{Text: "// @Tags market"},
		},
	}

	as := newAnnotationSet(comments)

	desc := as.GetOne("Description")
	if desc == nil {
		t.Fatal("Description is nil")
	}
	expected := "This is a long description that continues on the next line and even one more line"
	if desc.Raw != expected {
		t.Errorf("Description.Raw = %q, want %q", desc.Raw, expected)
	}
}

func TestTokenizeArgs(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"", nil},
		{"simple", []string{"simple"}},
		{"a b c", []string{"a", "b", "c"}},
		{`"quoted string"`, []string{"quoted string"}},
		{`a "multi word" b`, []string{"a", "multi word", "b"}},
		{"enum(a,b,c)", []string{"enum(a,b,c)"}},
		{`name string true "description" enum(x,y)`, []string{"name", "string", "true", "description", "enum(x,y)"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := tokenizeArgs(tt.input)
			if len(got) != len(tt.want) {
				t.Fatalf("tokenizeArgs(%q) = %v (len %d), want %v (len %d)", tt.input, got, len(got), tt.want, len(tt.want))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestParseAnnotationLineEmpty(t *testing.T) {
	ann := parseAnnotationLine("@")
	if ann != nil {
		t.Error("bare @ should return nil")
	}
}

func TestAnnotationSetEmpty(t *testing.T) {
	comments := &ast.CommentGroup{
		List: []*ast.Comment{
			{Text: "// Just a regular comment"},
		},
	}
	as := newAnnotationSet(comments)
	if as.Has("Channel") {
		t.Error("empty set should not have Channel")
	}
	if as.GetOne("Channel") != nil {
		t.Error("GetOne should return nil for empty set")
	}
	if len(as.Get("Channel")) != 0 {
		t.Error("Get should return empty slice for empty set")
	}
}

func TestAddressToKey(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"/market/{pair}", "marketPair"},
		{"/users", "users"},
		{"/api/v1/events", "apiV1Events"},
		{"/", "root"},
		{"/chat/{room}/messages", "chatRoomMessages"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := addressToKey(tt.input)
			if got != tt.want {
				t.Errorf("addressToKey(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
