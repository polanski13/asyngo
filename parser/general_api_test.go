package parser

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func setupTestProject(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestParseGeneralAPI_MissingAnnotation(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"main.go": `package main
func main() {}
`,
	})

	p := New(WithSearchDirs(dir), WithMainFile("main.go"))
	_, err := p.Parse()
	if err == nil {
		t.Fatal("expected error for missing annotations")
	}
	if !errors.Is(err, ErrMissingAnnotation) {
		t.Errorf("error = %v, want ErrMissingAnnotation", err)
	}
}

func TestParseGeneralAPI_InvalidServer(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"main.go": `package main

// @AsyncAPI 3.1.0
// @Title Test
// @Version 1.0.0
// @Server production
func Init() {}
`,
	})

	p := New(WithSearchDirs(dir), WithMainFile("main.go"))
	_, err := p.Parse()
	if err == nil {
		t.Fatal("expected error for invalid @Server")
	}
}

func TestParseGeneralAPI_AllProtocols(t *testing.T) {
	tests := []struct {
		host      string
		wantProto string
		wantHost  string
	}{
		{"wss://ws.example.com", "wss", "ws.example.com"},
		{"ws://ws.example.com", "ws", "ws.example.com"},
		{"https://api.example.com", "https", "api.example.com"},
		{"http://localhost", "http", "localhost"},
		{"mqtt://broker.example.com", "mqtt", "broker.example.com"},
		{"amqp://mq.example.com", "amqp", "mq.example.com"},
		{"kafka://kafka.example.com", "kafka", "kafka.example.com"},
		{"bare.example.com", "ws", "bare.example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			host, proto := parseHostProtocol(tt.host)
			if host != tt.wantHost {
				t.Errorf("host = %q, want %q", host, tt.wantHost)
			}
			if proto != tt.wantProto {
				t.Errorf("protocol = %q, want %q", proto, tt.wantProto)
			}
		})
	}
}

func TestParseGeneralAPI_ExternalDocs(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"main.go": `package main

// @AsyncAPI 3.1.0
// @Title Test API
// @Version 1.0.0
// @ExternalDocs.Description API documentation
// @ExternalDocs.URL https://docs.example.com
// @TermsOfService https://example.com/tos
// @ID urn:example:api
func Init() {}
`,
	})

	p := New(WithSearchDirs(dir), WithMainFile("main.go"))
	doc, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if doc.Info.ExternalDocs == nil {
		t.Fatal("ExternalDocs is nil")
	}
	if doc.Info.ExternalDocs.Description != "API documentation" {
		t.Errorf("ExternalDocs.Description = %q", doc.Info.ExternalDocs.Description)
	}
	if doc.Info.ExternalDocs.URL != "https://docs.example.com" {
		t.Errorf("ExternalDocs.URL = %q", doc.Info.ExternalDocs.URL)
	}
	if doc.Info.TermsOfService != "https://example.com/tos" {
		t.Errorf("TermsOfService = %q", doc.Info.TermsOfService)
	}
	if doc.ID != "urn:example:api" {
		t.Errorf("ID = %q", doc.ID)
	}
}

func TestParseGeneralAPI_ServerNoPathname(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"main.go": `package main

// @AsyncAPI 3.1.0
// @Title Test
// @Version 1.0.0
// @Server dev ws://localhost:8080 - "Development server"
func Init() {}
`,
	})

	p := New(WithSearchDirs(dir), WithMainFile("main.go"))
	doc, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	srv := doc.Servers["dev"]
	if srv.Pathname != "" {
		t.Errorf("pathname = %q, want empty", srv.Pathname)
	}
	if srv.Description != "Development server" {
		t.Errorf("description = %q", srv.Description)
	}
}
