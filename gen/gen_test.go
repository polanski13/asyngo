package gen

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/polanski13/asyngo/spec"
)

func testDoc() *spec.AsyncAPI {
	doc := spec.NewAsyncAPI()
	doc.Info = spec.Info{
		Title:   "Test API",
		Version: "1.0.0",
	}
	doc.Servers["production"] = spec.Server{
		Host:     "ws.example.com",
		Protocol: "wss",
	}
	doc.Channels["testChannel"] = spec.Channel{
		Address:     "/test",
		Description: "A test channel",
	}
	doc.Components.Schemas["TestPayload"] = &spec.Schema{
		Type: "object",
		Properties: map[string]*spec.SchemaRef{
			"id": spec.NewInlineSchema(&spec.Schema{Type: "string"}),
		},
	}
	return doc
}

func TestWriteJSON(t *testing.T) {
	dir := t.TempDir()
	cfg := &Config{OutputDir: dir}
	doc := testDoc()

	if err := writeJSON(cfg, doc); err != nil {
		t.Fatalf("writeJSON: %v", err)
	}

	path := filepath.Join(dir, "asyncapi.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}

	var parsed spec.AsyncAPI
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}

	if parsed.AsyncAPI != "3.0.0" {
		t.Errorf("asyncapi = %q, want 3.0.0", parsed.AsyncAPI)
	}
	if parsed.Info.Title != "Test API" {
		t.Errorf("title = %q, want Test API", parsed.Info.Title)
	}
	if parsed.Info.Version != "1.0.0" {
		t.Errorf("version = %q, want 1.0.0", parsed.Info.Version)
	}

	if !strings.Contains(string(data), "\"asyncapi\"") {
		t.Error("JSON output missing asyncapi key")
	}
	if !strings.Contains(string(data), "  ") {
		t.Error("JSON output not indented")
	}
}

func TestWriteYAML(t *testing.T) {
	dir := t.TempDir()
	cfg := &Config{OutputDir: dir}
	doc := testDoc()

	if err := writeYAML(cfg, doc); err != nil {
		t.Fatalf("writeYAML: %v", err)
	}

	path := filepath.Join(dir, "asyncapi.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "asyncapi:") {
		t.Error("YAML output missing asyncapi key")
	}
	if !strings.Contains(content, "title: Test API") {
		t.Error("YAML output missing title")
	}
	if !strings.Contains(content, "version: 1.0.0") {
		t.Error("YAML output missing version")
	}
	if strings.Contains(content, "{") {
		t.Error("YAML output contains JSON-like braces")
	}
}

func TestWriteGo(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "docs")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		t.Fatal(err)
	}

	cfg := &Config{OutputDir: outDir}
	doc := testDoc()

	if err := writeGo(cfg, doc); err != nil {
		t.Fatalf("writeGo: %v", err)
	}

	path := filepath.Join(outDir, "docs.go")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "package docs") {
		t.Error("Go file missing package declaration")
	}
	if !strings.Contains(content, "go:embed asyncapi.json") {
		t.Error("Go file missing go:embed directive")
	}
	if !strings.Contains(content, `Title = "Test API"`) {
		t.Error("Go file missing Title constant")
	}
	if !strings.Contains(content, `Version = "1.0.0"`) {
		t.Error("Go file missing Version constant")
	}
	if !strings.Contains(content, "DocJSON") {
		t.Error("Go file missing DocJSON variable")
	}
}

func TestWriteGoPackageName(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "apidocs")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		t.Fatal(err)
	}

	cfg := &Config{OutputDir: outDir}
	doc := testDoc()

	if err := writeGo(cfg, doc); err != nil {
		t.Fatalf("writeGo: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(outDir, "docs.go"))
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(data), "package apidocs") {
		t.Error("package name should match output directory name")
	}
}

func TestJsonToYAML(t *testing.T) {
	input := `{"name":"test","count":42,"nested":{"key":"value"}}`
	yamlData, err := jsonToYAML([]byte(input))
	if err != nil {
		t.Fatalf("jsonToYAML: %v", err)
	}

	content := string(yamlData)
	if !strings.Contains(content, "name: test") {
		t.Error("YAML missing name field")
	}
	if !strings.Contains(content, "count: 42") {
		t.Error("YAML missing count field")
	}
	if !strings.Contains(content, "key: value") {
		t.Error("YAML missing nested key")
	}
}

func TestJsonToYAMLInvalid(t *testing.T) {
	_, err := jsonToYAML([]byte("{invalid"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestNewGen(t *testing.T) {
	g := New()
	if g == nil {
		t.Fatal("New() returned nil")
	}

	types := []string{"json", "yaml", "yml", "go"}
	for _, typ := range types {
		if _, ok := g.writers[typ]; !ok {
			t.Errorf("missing writer for %q", typ)
		}
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.SearchDir != "." {
		t.Errorf("SearchDir = %q, want \".\"", cfg.SearchDir)
	}
	if cfg.MainAPIFile != "main.go" {
		t.Errorf("MainAPIFile = %q, want \"main.go\"", cfg.MainAPIFile)
	}
	if cfg.OutputDir != "./docs" {
		t.Errorf("OutputDir = %q, want \"./docs\"", cfg.OutputDir)
	}
	if len(cfg.OutputTypes) != 2 {
		t.Errorf("OutputTypes len = %d, want 2", len(cfg.OutputTypes))
	}
}

func TestBuildInvalidSearchDir(t *testing.T) {
	g := New()
	cfg := &Config{
		SearchDir: "/nonexistent/path/that/does/not/exist",
		OutputDir: t.TempDir(),
	}

	err := g.Build(cfg)
	if err == nil {
		t.Fatal("expected error for nonexistent search dir")
	}
	if !strings.Contains(err.Error(), "search directory") {
		t.Errorf("error = %q, want to contain 'search directory'", err.Error())
	}
}

func TestBuildUnknownOutputType(t *testing.T) {
	dir := t.TempDir()
	cfg := &Config{
		SearchDir:   "../testdata/basic",
		MainAPIFile: "main.go",
		OutputDir:   dir,
		OutputTypes: []string{"xml", "json"},
	}

	g := New()
	err := g.Build(cfg)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	jsonPath := filepath.Join(dir, "asyncapi.json")
	if _, err := os.Stat(jsonPath); err != nil {
		t.Error("json output should still be created")
	}

	xmlPath := filepath.Join(dir, "asyncapi.xml")
	if _, err := os.Stat(xmlPath); !os.IsNotExist(err) {
		t.Error("xml output should not be created for unknown type")
	}
}

func TestBuildWithTestdata(t *testing.T) {
	dir := t.TempDir()
	g := New()

	cfg := &Config{
		SearchDir:   "../testdata/basic",
		MainAPIFile: "main.go",
		OutputDir:   dir,
		OutputTypes: []string{"json", "yaml"},
	}

	if err := g.Build(cfg); err != nil {
		t.Fatalf("Build: %v", err)
	}

	jsonPath := filepath.Join(dir, "asyncapi.json")
	if _, err := os.Stat(jsonPath); err != nil {
		t.Error("asyncapi.json not created")
	}

	yamlPath := filepath.Join(dir, "asyncapi.yaml")
	if _, err := os.Stat(yamlPath); err != nil {
		t.Error("asyncapi.yaml not created")
	}

	jsonData, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatal(err)
	}

	var doc spec.AsyncAPI
	if err := json.Unmarshal(jsonData, &doc); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	if doc.AsyncAPI != "3.0.0" {
		t.Errorf("asyncapi = %q", doc.AsyncAPI)
	}
	if doc.Info.Title == "" {
		t.Error("info.title is empty")
	}
	if len(doc.Channels) == 0 {
		t.Error("no channels in output")
	}
	if len(doc.Operations) == 0 {
		t.Error("no operations in output")
	}
}

func TestBuildWithGoOutput(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "docs")
	g := New()

	cfg := &Config{
		SearchDir:   "../testdata/basic",
		MainAPIFile: "main.go",
		OutputDir:   outDir,
		OutputTypes: []string{"json", "go"},
	}

	if err := g.Build(cfg); err != nil {
		t.Fatalf("Build: %v", err)
	}

	goPath := filepath.Join(outDir, "docs.go")
	data, err := os.ReadFile(goPath)
	if err != nil {
		t.Fatalf("read docs.go: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "package docs") {
		t.Error("docs.go missing package declaration")
	}
	if !strings.Contains(content, "go:embed") {
		t.Error("docs.go missing go:embed")
	}
}

func TestBuildMultipleSearchDirs(t *testing.T) {
	dir := t.TempDir()
	g := New()

	cfg := &Config{
		SearchDir:   "../testdata/basic, ../testdata/basic",
		MainAPIFile: "main.go",
		OutputDir:   dir,
		OutputTypes: []string{"json"},
	}

	err := g.Build(cfg)
	if err != nil {
		t.Fatalf("Build with multiple dirs: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "asyncapi.json")); err != nil {
		t.Error("asyncapi.json not created")
	}
}

func TestBuildExcludes(t *testing.T) {
	dir := t.TempDir()
	g := New()

	cfg := &Config{
		SearchDir:   "../testdata/basic",
		MainAPIFile: "main.go",
		OutputDir:   dir,
		OutputTypes: []string{"json"},
		Excludes:    "vendor,node_modules",
	}

	if err := g.Build(cfg); err != nil {
		t.Fatalf("Build with excludes: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "asyncapi.json")); err != nil {
		t.Error("asyncapi.json not created")
	}
}

func TestWriteJSONRoundtrip(t *testing.T) {
	dir := t.TempDir()
	doc := testDoc()
	doc.Components.Messages["testMsg"] = &spec.Message{
		Name:    "testMsg",
		Summary: "A test message",
		Payload: spec.NewSchemaRef(spec.ComponentSchemaRef("TestPayload")),
	}
	doc.Operations["receiveTest"] = spec.Operation{
		Action:  spec.ActionReceive,
		Channel: spec.Reference{Ref: spec.ChannelRef("testChannel")},
	}

	cfg := &Config{OutputDir: dir}
	if err := writeJSON(cfg, doc); err != nil {
		t.Fatalf("writeJSON: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "asyncapi.json"))
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(data), `"$ref"`) {
		t.Error("JSON output should contain $ref for message payload")
	}
	if !strings.Contains(string(data), "#/components/schemas/TestPayload") {
		t.Error("JSON output should contain schema $ref")
	}
}

func TestWriteYAMLWithRef(t *testing.T) {
	dir := t.TempDir()
	doc := testDoc()
	doc.Operations["receiveTest"] = spec.Operation{
		Action:  spec.ActionReceive,
		Channel: spec.Reference{Ref: spec.ChannelRef("testChannel")},
	}

	cfg := &Config{OutputDir: dir}
	if err := writeYAML(cfg, doc); err != nil {
		t.Fatalf("writeYAML: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "asyncapi.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	if !strings.Contains(content, "#/channels/testChannel") {
		t.Error("YAML should contain channel $ref")
	}
}
