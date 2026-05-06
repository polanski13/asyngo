package parser

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/polanski13/asyngo/spec"
)

func TestParseBasicTestdata(t *testing.T) {
	testdataDir, err := filepath.Abs("../testdata/basic")
	if err != nil {
		t.Fatal(err)
	}

	p := New(
		WithSearchDirs(testdataDir),
		WithMainFile("main.go"),
	)

	doc, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if doc.AsyncAPI != "3.1.0" {
		t.Errorf("AsyncAPI = %q, want %q", doc.AsyncAPI, "3.1.0")
	}
	if doc.Info.Title != "Trading Platform WebSocket API" {
		t.Errorf("Title = %q, want %q", doc.Info.Title, "Trading Platform WebSocket API")
	}
	if doc.Info.Version != "1.0.0" {
		t.Errorf("Version = %q, want %q", doc.Info.Version, "1.0.0")
	}
	if doc.Info.Description != "Real-time trading data streaming via WebSocket" {
		t.Errorf("Description = %q", doc.Info.Description)
	}
	if doc.DefaultContentType != "application/json" {
		t.Errorf("DefaultContentType = %q", doc.DefaultContentType)
	}

	if doc.Info.Contact == nil {
		t.Fatal("Contact is nil")
	}
	if doc.Info.Contact.Name != "Platform Team" {
		t.Errorf("Contact.Name = %q", doc.Info.Contact.Name)
	}
	if doc.Info.Contact.Email != "platform@example.com" {
		t.Errorf("Contact.Email = %q", doc.Info.Contact.Email)
	}

	if doc.Info.License == nil {
		t.Fatal("License is nil")
	}
	if doc.Info.License.Name != "MIT" {
		t.Errorf("License.Name = %q", doc.Info.License.Name)
	}

	if len(doc.Servers) != 2 {
		t.Fatalf("len(Servers) = %d, want 2", len(doc.Servers))
	}
	prod, ok := doc.Servers["production"]
	if !ok {
		t.Fatal("production server not found")
	}
	if prod.Host != "ws.trading.example.com" {
		t.Errorf("production.Host = %q", prod.Host)
	}
	if prod.Protocol != "wss" {
		t.Errorf("production.Protocol = %q", prod.Protocol)
	}
	if prod.Pathname != "/v1" {
		t.Errorf("production.Pathname = %q", prod.Pathname)
	}

	if len(doc.Channels) == 0 {
		t.Fatal("no channels found")
	}

	ch, ok := doc.Channels["marketPair"]
	if !ok {
		t.Fatalf("channel 'marketPair' not found, got keys: %v", channelKeys(doc))
	}
	if ch.Address != "/market/{pair}" {
		t.Errorf("channel address = %q", ch.Address)
	}
	if ch.Description != "Real-time market data stream for a trading pair" {
		t.Errorf("channel description = %q", ch.Description)
	}

	expectedMsgs := []string{"tickerUpdate", "orderBookSnapshot", "subscribe", "subscriptionAck"}
	for _, name := range expectedMsgs {
		if _, ok := ch.Messages[name]; !ok {
			t.Errorf("channel message %q not found", name)
		}
	}

	if ch.Bindings == nil || ch.Bindings.WS == nil {
		t.Fatal("channel ws binding is nil")
	}
	if ch.Bindings.WS.Method != "GET" {
		t.Errorf("ws binding method = %q", ch.Bindings.WS.Method)
	}
	if ch.Bindings.WS.Query == nil {
		t.Fatal("ws query binding is nil")
	}
	if _, ok := ch.Bindings.WS.Query.Properties["token"]; !ok {
		t.Error("ws query 'token' not found")
	}

	if len(doc.Operations) != 2 {
		t.Fatalf("len(Operations) = %d, want 2", len(doc.Operations))
	}

	recvOp, ok := doc.Operations["receiveMarketUpdate"]
	if !ok {
		t.Fatal("operation 'receiveMarketUpdate' not found")
	}
	if recvOp.Action != "receive" {
		t.Errorf("receiveMarketUpdate action = %q", recvOp.Action)
	}
	if len(recvOp.Messages) != 2 {
		t.Errorf("receiveMarketUpdate messages count = %d, want 2", len(recvOp.Messages))
	}

	sendOp, ok := doc.Operations["sendSubscription"]
	if !ok {
		t.Fatal("operation 'sendSubscription' not found")
	}
	if sendOp.Action != "send" {
		t.Errorf("sendSubscription action = %q", sendOp.Action)
	}
	if sendOp.Reply == nil {
		t.Fatal("sendSubscription reply is nil")
	}
	if len(sendOp.Reply.Messages) != 1 {
		t.Errorf("sendSubscription reply messages = %d, want 1", len(sendOp.Reply.Messages))
	}

	expectedSchemas := []string{"TickerPayload", "OrderBookPayload", "PriceLevel", "SubscribeRequest", "SubscriptionAck"}
	for _, name := range expectedSchemas {
		if _, ok := doc.Components.Schemas[name]; !ok {
			t.Errorf("schema %q not found in components", name)
		}
	}

	ticker := doc.Components.Schemas["TickerPayload"]
	if ticker.Type != "object" {
		t.Errorf("TickerPayload.Type = %q", ticker.Type)
	}
	if len(ticker.Properties) == 0 {
		t.Fatal("TickerPayload has no properties")
	}
	if _, ok := ticker.Properties["symbol"]; !ok {
		t.Error("TickerPayload missing 'symbol' property")
	}

	symbolProp := ticker.Properties["symbol"]
	if symbolProp.Schema == nil {
		t.Fatal("symbol schema is nil")
	}
	if symbolProp.Schema.Type != "string" {
		t.Errorf("symbol.Type = %q", symbolProp.Schema.Type)
	}
	if symbolProp.Schema.Example != "BTC-USD" {
		t.Errorf("symbol.Example = %v", symbolProp.Schema.Example)
	}

	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		t.Fatalf("JSON marshal error: %v", err)
	}

	outDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(outDir, "asyncapi.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}

	t.Logf("Generated AsyncAPI spec (%d bytes)", len(data))
}

func TestParseMultiPackageTestdata(t *testing.T) {
	testdataDir, err := filepath.Abs("../testdata/multipackage")
	if err != nil {
		t.Fatal(err)
	}

	p := New(
		WithSearchDirs(testdataDir),
		WithMainFile("main.go"),
	)

	doc, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if doc.Info.Title != "Multi-Package API" {
		t.Errorf("Title = %q", doc.Info.Title)
	}
	if doc.Info.Version != "2.0.0" {
		t.Errorf("Version = %q", doc.Info.Version)
	}

	if len(doc.Servers) != 1 {
		t.Fatalf("len(Servers) = %d, want 1", len(doc.Servers))
	}
	if _, ok := doc.Servers["production"]; !ok {
		t.Error("production server not found")
	}

	if len(doc.Channels) == 0 {
		t.Fatal("no channels found")
	}
	ch, ok := doc.Channels["eventsUserId"]
	if !ok {
		t.Fatalf("channel 'eventsUserId' not found, got keys: %v", channelKeys(doc))
	}
	if ch.Address != "/events/{userId}" {
		t.Errorf("channel address = %q", ch.Address)
	}

	if _, ok := ch.Messages["notification"]; !ok {
		t.Error("channel message 'notification' not found")
	}
	if _, ok := ch.Messages["errorResponse"]; !ok {
		t.Error("channel message 'errorResponse' not found")
	}

	if len(doc.Operations) != 2 {
		t.Fatalf("len(Operations) = %d, want 2", len(doc.Operations))
	}

	recvOp := doc.Operations["receiveEvents"]
	if recvOp.Action != spec.ActionReceive {
		t.Errorf("receiveEvents.action = %q", recvOp.Action)
	}
	sendOp := doc.Operations["sendAck"]
	if sendOp.Action != spec.ActionSend {
		t.Errorf("sendAck.action = %q", sendOp.Action)
	}

	notifMsg := doc.Components.Messages["notification"]
	if notifMsg == nil {
		t.Fatal("notification message not found in components")
	}
	if notifMsg.Payload == nil {
		t.Fatal("notification payload is nil")
	}
	if notifMsg.Payload.Ref != "#/components/schemas/Notification" {
		t.Errorf("notification payload ref = %q, want #/components/schemas/Notification", notifMsg.Payload.Ref)
	}

	errMsg := doc.Components.Messages["errorResponse"]
	if errMsg == nil {
		t.Fatal("errorResponse message not found in components")
	}
	if errMsg.Payload.Ref != "#/components/schemas/ErrorResponse" {
		t.Errorf("errorResponse payload ref = %q, want #/components/schemas/ErrorResponse", errMsg.Payload.Ref)
	}

	if _, ok := doc.Components.Schemas["Notification"]; !ok {
		t.Error("schema 'Notification' not found in components")
	}
	if _, ok := doc.Components.Schemas["ErrorResponse"]; !ok {
		t.Error("schema 'ErrorResponse' not found in components")
	}
	if _, ok := doc.Components.Schemas["Event"]; !ok {
		t.Error("schema 'Event' not found (embedded in Notification)")
	}

	notifSchema := doc.Components.Schemas["Notification"]
	if notifSchema == nil {
		t.Fatal("Notification schema is nil")
	}
	if _, ok := notifSchema.Properties["userId"]; !ok {
		t.Error("Notification missing 'userId' (own field)")
	}
	if _, ok := notifSchema.Properties["id"]; !ok {
		t.Error("Notification missing 'id' (embedded from Event)")
	}
	if _, ok := notifSchema.Properties["type"]; !ok {
		t.Error("Notification missing 'type' (embedded from Event)")
	}

	errSchema := doc.Components.Schemas["ErrorResponse"]
	if errSchema == nil {
		t.Fatal("ErrorResponse schema is nil")
	}
	if _, ok := errSchema.Properties["code"]; !ok {
		t.Error("ErrorResponse missing 'code'")
	}
	if _, ok := errSchema.Properties["message"]; !ok {
		t.Error("ErrorResponse missing 'message'")
	}

	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		t.Fatalf("JSON marshal error: %v", err)
	}
	t.Logf("Multi-package spec (%d bytes)", len(data))
}

func TestParseOneOfTestdata(t *testing.T) {
	testdataDir, err := filepath.Abs("../testdata/oneof")
	if err != nil {
		t.Fatal(err)
	}

	p := New(
		WithSearchDirs(testdataDir),
		WithMainFile("main.go"),
	)

	doc, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if doc.AsyncAPI != "3.1.0" {
		t.Errorf("AsyncAPI = %q", doc.AsyncAPI)
	}
	if doc.Info.Title != "Market Events API" {
		t.Errorf("Title = %q", doc.Info.Title)
	}
	if doc.Info.Version != "1.0.0" {
		t.Errorf("Version = %q", doc.Info.Version)
	}

	ch, ok := doc.Channels["eventsSymbol"]
	if !ok {
		t.Fatalf("channel 'eventsSymbol' not found, got keys: %v", channelKeys(doc))
	}
	if ch.Address != "/events/{symbol}" {
		t.Errorf("channel address = %q", ch.Address)
	}
	if _, ok := ch.Messages["eventUpdate"]; !ok {
		t.Error("channel missing 'eventUpdate' message")
	}

	op, ok := doc.Operations["receiveEvents"]
	if !ok {
		t.Fatal("operation 'receiveEvents' not found")
	}
	if op.Action != spec.ActionReceive {
		t.Errorf("action = %q", op.Action)
	}
	if len(op.Messages) != 1 {
		t.Errorf("operation messages count = %d, want 1", len(op.Messages))
	}

	msg, ok := doc.Components.Messages["eventUpdate"]
	if !ok {
		t.Fatal("message 'eventUpdate' not found")
	}
	if msg.Payload == nil || msg.Payload.Schema == nil {
		t.Fatal("eventUpdate payload schema is nil")
	}
	if len(msg.Payload.Schema.OneOf) != 3 {
		t.Fatalf("oneOf count = %d, want 3", len(msg.Payload.Schema.OneOf))
	}
	if msg.Payload.Schema.Discriminator == nil || msg.Payload.Schema.Discriminator.PropertyName != "eventType" {
		t.Errorf("discriminator = %+v, want propertyName=eventType", msg.Payload.Schema.Discriminator)
	}

	expectedRefs := []string{
		"#/components/schemas/TickerPayload",
		"#/components/schemas/OrderBookPayload",
		"#/components/schemas/TradePayload",
	}
	for i, want := range expectedRefs {
		if msg.Payload.Schema.OneOf[i].Ref != want {
			t.Errorf("oneOf[%d] = %q, want %q", i, msg.Payload.Schema.OneOf[i].Ref, want)
		}
	}

	expectedSchemas := []string{"TickerPayload", "OrderBookPayload", "TradePayload"}
	for _, name := range expectedSchemas {
		if _, ok := doc.Components.Schemas[name]; !ok {
			t.Errorf("schema %q not found in components", name)
		}
	}

	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		t.Fatalf("JSON marshal error: %v", err)
	}

	outDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(outDir, "asyncapi.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}

	t.Logf("OneOf spec (%d bytes)", len(data))
}

func channelKeys(doc *spec.AsyncAPI) []string {
	keys := make([]string, 0, len(doc.Channels))
	for k := range doc.Channels {
		keys = append(keys, k)
	}
	return keys
}
