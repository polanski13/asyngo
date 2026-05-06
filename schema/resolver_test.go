package schema

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/polanski13/asyngo/spec"
)

type testPackageLookup struct {
	files map[string]*ast.File
	types map[string]*TypeDef
}

func newTestLookup(src string) (*testPackageLookup, *ast.File) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	lookup := &testPackageLookup{
		files: map[string]*ast.File{"test.go": f},
		types: make(map[string]*TypeDef),
	}

	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, s := range genDecl.Specs {
			ts, ok := s.(*ast.TypeSpec)
			if !ok {
				continue
			}
			if ts.Doc == nil && genDecl.Doc != nil {
				ts.Doc = genDecl.Doc
			}
			lookup.types[ts.Name.Name] = &TypeDef{
				File:     f,
				TypeSpec: ts,
				PkgPath:  "test",
			}
		}
	}

	return lookup, f
}

func (tl *testPackageLookup) FindTypeSpec(typeName string, file *ast.File) (*TypeDef, error) {
	if td, ok := tl.types[typeName]; ok {
		return td, nil
	}
	return nil, fmt.Errorf("%w: %s", ErrUnresolvedType, typeName)
}

func TestResolvePrimitives(t *testing.T) {
	t.Parallel()
	src := `package test
type Dummy struct{}
`
	lookup, file := newTestLookup(src)
	r := NewResolver(lookup)
	components := make(map[string]*spec.Schema)

	tests := []struct {
		name       string
		expr       ast.Expr
		wantType   string
		wantFormat string
	}{
		{"string", &ast.Ident{Name: "string"}, "string", ""},
		{"int", &ast.Ident{Name: "int"}, "integer", ""},
		{"int64", &ast.Ident{Name: "int64"}, "integer", "int64"},
		{"float64", &ast.Ident{Name: "float64"}, "number", "double"},
		{"bool", &ast.Ident{Name: "bool"}, "boolean", ""},
		{"any", &ast.Ident{Name: "any"}, "object", ""},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ref, err := r.ResolveExpr(tt.expr, file, components)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if ref.Ref != "" {
				t.Fatalf("expected inline schema, got $ref: %s", ref.Ref)
			}
			if ref.Schema.Type != tt.wantType {
				t.Errorf("type = %q, want %q", ref.Schema.Type, tt.wantType)
			}
			if ref.Schema.Format != tt.wantFormat {
				t.Errorf("format = %q, want %q", ref.Schema.Format, tt.wantFormat)
			}
		})
	}
}

func TestResolveStruct(t *testing.T) {
	src := `package test
import "time"

type User struct {
	Name  string    ` + "`json:\"name\" validate:\"required\"`" + `
	Email string    ` + "`json:\"email\" example:\"user@example.com\"`" + `
	Age   int       ` + "`json:\"age\" minimum:\"0\" maximum:\"150\"`" + `
	CreatedAt time.Time ` + "`json:\"createdAt\"`" + `
}
`
	lookup, file := newTestLookup(src)
	r := NewResolver(lookup)
	components := make(map[string]*spec.Schema)

	ref, err := r.ResolveTypeName("User", file, components)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ref.Ref != "#/components/schemas/User" {
		t.Errorf("ref = %q, want #/components/schemas/User", ref.Ref)
	}

	schema, ok := components["User"]
	if !ok {
		t.Fatal("User schema not registered in components")
	}

	if schema.Type != "object" {
		t.Errorf("type = %q, want object", schema.Type)
	}

	if _, ok := schema.Properties["name"]; !ok {
		t.Error("missing 'name' property")
	}
	if _, ok := schema.Properties["email"]; !ok {
		t.Error("missing 'email' property")
	}
	if _, ok := schema.Properties["age"]; !ok {
		t.Error("missing 'age' property")
	}

	nameProp := schema.Properties["name"]
	if nameProp.Schema.Type != "string" {
		t.Errorf("name.type = %q", nameProp.Schema.Type)
	}

	emailProp := schema.Properties["email"]
	if emailProp.Schema.Example != "user@example.com" {
		t.Errorf("email.example = %v", emailProp.Schema.Example)
	}

	ageProp := schema.Properties["age"]
	if ageProp.Schema.Minimum == nil || *ageProp.Schema.Minimum != 0 {
		t.Errorf("age.minimum = %v", ageProp.Schema.Minimum)
	}
	if ageProp.Schema.Maximum == nil || *ageProp.Schema.Maximum != 150 {
		t.Errorf("age.maximum = %v", ageProp.Schema.Maximum)
	}

	if len(schema.Required) != 1 || schema.Required[0] != "name" {
		t.Errorf("required = %v, want [name]", schema.Required)
	}
}

func TestResolvePointer(t *testing.T) {
	src := `package test
type Inner struct {
	Value string ` + "`json:\"value\"`" + `
}
type Outer struct {
	Ref *Inner ` + "`json:\"ref\"`" + `
}
`
	lookup, file := newTestLookup(src)
	r := NewResolver(lookup)
	components := make(map[string]*spec.Schema)

	_, err := r.ResolveTypeName("Outer", file, components)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := components["Inner"]; !ok {
		t.Error("Inner schema not resolved")
	}
	if _, ok := components["Outer"]; !ok {
		t.Error("Outer schema not resolved")
	}
}

func TestResolveSlice(t *testing.T) {
	src := `package test
type Item struct {
	ID string ` + "`json:\"id\"`" + `
}
type Collection struct {
	Items []Item ` + "`json:\"items\"`" + `
	Tags  []string ` + "`json:\"tags\"`" + `
}
`
	lookup, file := newTestLookup(src)
	r := NewResolver(lookup)
	components := make(map[string]*spec.Schema)

	_, err := r.ResolveTypeName("Collection", file, components)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	schema := components["Collection"]
	itemsProp := schema.Properties["items"]
	if itemsProp.Schema.Type != "array" {
		t.Errorf("items.type = %q, want array", itemsProp.Schema.Type)
	}
	if itemsProp.Schema.Items == nil {
		t.Fatal("items.items is nil")
	}
	if itemsProp.Schema.Items.Ref != "#/components/schemas/Item" {
		t.Errorf("items.items.$ref = %q", itemsProp.Schema.Items.Ref)
	}

	tagsProp := schema.Properties["tags"]
	if tagsProp.Schema.Type != "array" {
		t.Errorf("tags.type = %q, want array", tagsProp.Schema.Type)
	}
	if tagsProp.Schema.Items.Schema.Type != "string" {
		t.Errorf("tags.items.type = %q", tagsProp.Schema.Items.Schema.Type)
	}
}

func TestResolveMap(t *testing.T) {
	src := `package test
type Config struct {
	Labels map[string]string ` + "`json:\"labels\"`" + `
}
`
	lookup, file := newTestLookup(src)
	r := NewResolver(lookup)
	components := make(map[string]*spec.Schema)

	_, err := r.ResolveTypeName("Config", file, components)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	schema := components["Config"]
	labelsProp := schema.Properties["labels"]
	if labelsProp.Schema.Type != "object" {
		t.Errorf("labels.type = %q, want object", labelsProp.Schema.Type)
	}
	if labelsProp.Schema.AdditionalProperties == nil {
		t.Fatal("labels.additionalProperties is nil")
	}
	if labelsProp.Schema.AdditionalProperties.Schema.Type != "string" {
		t.Errorf("labels.additionalProperties.type = %q", labelsProp.Schema.AdditionalProperties.Schema.Type)
	}
}

func TestResolveTypeAlias(t *testing.T) {
	src := `package test
type UserID string
type Status int
type Wrapper struct {
	ID     UserID ` + "`json:\"id\"`" + `
	Status Status ` + "`json:\"status\"`" + `
}
`
	lookup, file := newTestLookup(src)
	r := NewResolver(lookup)
	components := make(map[string]*spec.Schema)

	_, err := r.ResolveTypeName("Wrapper", file, components)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	schema := components["Wrapper"]

	idProp := schema.Properties["id"]
	if idProp == nil {
		t.Fatal("id property is nil")
	}
	if idProp.Schema == nil {
		t.Skip("type alias resolves to $ref")
	}
	if idProp.Schema.Type != "string" {
		t.Errorf("id.type = %q, want string", idProp.Schema.Type)
	}

	statusProp := schema.Properties["status"]
	if statusProp == nil {
		t.Fatal("status property is nil")
	}
	if statusProp.Schema == nil {
		t.Skip("type alias resolves to $ref")
	}
	if statusProp.Schema.Type != "integer" {
		t.Errorf("status.type = %q, want integer", statusProp.Schema.Type)
	}
}

func TestResolveEmbeddedStruct(t *testing.T) {
	src := `package test
type Base struct {
	ID        string ` + "`json:\"id\" validate:\"required\"`" + `
	CreatedBy string ` + "`json:\"createdBy\"`" + `
}
type Extended struct {
	Base
	Extra string ` + "`json:\"extra\"`" + `
}
`
	lookup, file := newTestLookup(src)
	r := NewResolver(lookup)
	components := make(map[string]*spec.Schema)

	_, err := r.ResolveTypeName("Extended", file, components)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	schema := components["Extended"]
	if _, ok := schema.Properties["id"]; !ok {
		t.Error("embedded field 'id' not flattened into Extended")
	}
	if _, ok := schema.Properties["createdBy"]; !ok {
		t.Error("embedded field 'createdBy' not flattened into Extended")
	}
	if _, ok := schema.Properties["extra"]; !ok {
		t.Error("own field 'extra' missing")
	}
}

func TestResolveRecursiveType(t *testing.T) {
	src := `package test
type TreeNode struct {
	Value    string      ` + "`json:\"value\"`" + `
	Children []*TreeNode ` + "`json:\"children\"`" + `
}
`
	lookup, file := newTestLookup(src)
	r := NewResolver(lookup)
	components := make(map[string]*spec.Schema)

	ref, err := r.ResolveTypeName("TreeNode", file, components)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ref.Ref != "#/components/schemas/TreeNode" {
		t.Errorf("ref = %q", ref.Ref)
	}

	schema := components["TreeNode"]
	if schema == nil {
		t.Fatal("TreeNode not in components")
	}

	childrenProp := schema.Properties["children"]
	if childrenProp.Schema.Type != "array" {
		t.Errorf("children.type = %q", childrenProp.Schema.Type)
	}
	items := childrenProp.Schema.Items
	if items.Schema == nil || len(items.Schema.OneOf) != 2 {
		t.Fatalf("children.items: expected oneOf wrapper for *TreeNode, got %+v", items)
	}
	if items.Schema.OneOf[0].Ref != "#/components/schemas/TreeNode" {
		t.Errorf("children.items.oneOf[0].$ref = %q", items.Schema.OneOf[0].Ref)
	}
	if items.Schema.OneOf[1].Schema == nil || items.Schema.OneOf[1].Schema.Type != "null" {
		t.Errorf("children.items.oneOf[1] = %+v, want type=null", items.Schema.OneOf[1])
	}
}

func TestResolveEnumTag(t *testing.T) {
	src := `package test
type Status struct {
	State string ` + "`json:\"state\" enum:\"active,inactive,deleted\"`" + `
}
`
	lookup, file := newTestLookup(src)
	r := NewResolver(lookup)
	components := make(map[string]*spec.Schema)

	_, err := r.ResolveTypeName("Status", file, components)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	schema := components["Status"]
	stateProp := schema.Properties["state"]
	if len(stateProp.Schema.Enum) != 3 {
		t.Fatalf("enum count = %d, want 3", len(stateProp.Schema.Enum))
	}
	if stateProp.Schema.Enum[0] != "active" {
		t.Errorf("enum[0] = %v", stateProp.Schema.Enum[0])
	}
}

func TestResolveOmitempty(t *testing.T) {
	src := `package test
type Response struct {
	Data  string ` + "`json:\"data\"`" + `
	Error string ` + "`json:\"error,omitempty\"`" + `
}
`
	lookup, file := newTestLookup(src)
	r := NewResolver(lookup)
	components := make(map[string]*spec.Schema)

	_, err := r.ResolveTypeName("Response", file, components)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	schema := components["Response"]
	if _, ok := schema.Properties["data"]; !ok {
		t.Error("data property missing")
	}
	if _, ok := schema.Properties["error"]; !ok {
		t.Error("error property missing (omitempty should not skip)")
	}
}

func TestResolveJSONDash(t *testing.T) {
	src := `package test
type Secret struct {
	Public  string ` + "`json:\"public\"`" + `
	Private string ` + "`json:\"-\"`" + `
}
`
	lookup, file := newTestLookup(src)
	r := NewResolver(lookup)
	components := make(map[string]*spec.Schema)

	_, err := r.ResolveTypeName("Secret", file, components)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	schema := components["Secret"]
	if _, ok := schema.Properties["public"]; !ok {
		t.Error("public property missing")
	}
	if _, ok := schema.Properties["Private"]; ok {
		t.Error("Private should be skipped (json:\"-\")")
	}
	if _, ok := schema.Properties["-"]; ok {
		t.Error("'-' should not be a property name")
	}
}

func TestResolveAsyncapiIgnore(t *testing.T) {
	src := `package test
type Filtered struct {
	Visible string ` + "`json:\"visible\"`" + `
	Hidden  string ` + "`json:\"hidden\" asyncapiignore:\"true\"`" + `
}
`
	lookup, file := newTestLookup(src)
	r := NewResolver(lookup)
	components := make(map[string]*spec.Schema)

	_, err := r.ResolveTypeName("Filtered", file, components)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	schema := components["Filtered"]
	if _, ok := schema.Properties["visible"]; !ok {
		t.Error("visible property missing")
	}
	if _, ok := schema.Properties["hidden"]; ok {
		t.Error("hidden should be skipped (asyncapiignore)")
	}
}

func TestResolveAnonymousStruct(t *testing.T) {
	src := `package test
type Wrapper struct {
	Meta struct {
		Version string ` + "`json:\"version\"`" + `
		Build   int    ` + "`json:\"build\"`" + `
	} ` + "`json:\"meta\"`" + `
}
`
	lookup, file := newTestLookup(src)
	r := NewResolver(lookup)
	components := make(map[string]*spec.Schema)

	_, err := r.ResolveTypeName("Wrapper", file, components)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	schema := components["Wrapper"]
	metaProp := schema.Properties["meta"]
	if metaProp == nil {
		t.Fatal("meta property is nil")
	}
	if metaProp.Schema == nil {
		t.Fatal("meta schema is nil (should be inline)")
	}
	if metaProp.Schema.Type != "object" {
		t.Errorf("meta.type = %q, want object", metaProp.Schema.Type)
	}
	if _, ok := metaProp.Schema.Properties["version"]; !ok {
		t.Error("meta missing 'version'")
	}
	if _, ok := metaProp.Schema.Properties["build"]; !ok {
		t.Error("meta missing 'build'")
	}
}

func TestResolveInterface(t *testing.T) {
	src := `package test
type Dummy struct{}
`
	lookup, file := newTestLookup(src)
	r := NewResolver(lookup)
	components := make(map[string]*spec.Schema)

	ref, err := r.ResolveExpr(&ast.InterfaceType{}, file, components)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ref.Schema == nil || ref.Schema.Type != "object" {
		t.Error("interface should resolve to object")
	}
}

func TestResolveUnsupportedType(t *testing.T) {
	src := `package test
type Dummy struct{}
`
	lookup, file := newTestLookup(src)
	r := NewResolver(lookup)
	components := make(map[string]*spec.Schema)

	_, err := r.ResolveExpr(&ast.ChanType{Dir: ast.SEND, Value: &ast.Ident{Name: "int"}}, file, components)
	if err == nil {
		t.Fatal("expected error for channel type")
	}
}

func TestResolveUnresolvedType(t *testing.T) {
	src := `package test
type Dummy struct{}
`
	lookup, file := newTestLookup(src)
	r := NewResolver(lookup)
	components := make(map[string]*spec.Schema)

	_, err := r.ResolveTypeName("NonExistent", file, components)
	if err == nil {
		t.Fatal("expected error for non-existent type")
	}
}

func TestResolveEmptyStruct(t *testing.T) {
	src := `package test
type Empty struct{}
`
	lookup, file := newTestLookup(src)
	r := NewResolver(lookup)
	components := make(map[string]*spec.Schema)

	_, err := r.ResolveTypeName("Empty", file, components)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	schema := components["Empty"]
	if schema.Type != "object" {
		t.Errorf("type = %q, want object", schema.Type)
	}
	if len(schema.Properties) != 0 {
		t.Errorf("properties count = %d, want 0", len(schema.Properties))
	}
}

func TestResolveValidationTags(t *testing.T) {
	src := `package test
type Form struct {
	Name    string   ` + "`json:\"name\" validate:\"required,min=2,max=100\"`" + `
	Score   float64  ` + "`json:\"score\" validate:\"min=0,max=100\"`" + `
	Tags    []string ` + "`json:\"tags\" validate:\"min=1,max=10\"`" + `
	Status  string   ` + "`json:\"status\" validate:\"oneof=active inactive\"`" + `
	Level   int      ` + "`json:\"level\" binding:\"required\"`" + `
}
`
	lookup, file := newTestLookup(src)
	r := NewResolver(lookup)
	components := make(map[string]*spec.Schema)

	_, err := r.ResolveTypeName("Form", file, components)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	schema := components["Form"]

	nameProp := schema.Properties["name"]
	if nameProp.Schema.MinLength == nil || *nameProp.Schema.MinLength != 2 {
		t.Errorf("name.minLength = %v, want 2", nameProp.Schema.MinLength)
	}
	if nameProp.Schema.MaxLength == nil || *nameProp.Schema.MaxLength != 100 {
		t.Errorf("name.maxLength = %v, want 100", nameProp.Schema.MaxLength)
	}

	scoreProp := schema.Properties["score"]
	if scoreProp.Schema.Minimum == nil || *scoreProp.Schema.Minimum != 0 {
		t.Errorf("score.minimum = %v, want 0", scoreProp.Schema.Minimum)
	}
	if scoreProp.Schema.Maximum == nil || *scoreProp.Schema.Maximum != 100 {
		t.Errorf("score.maximum = %v, want 100", scoreProp.Schema.Maximum)
	}

	tagsProp := schema.Properties["tags"]
	if tagsProp.Schema.MinItems == nil || *tagsProp.Schema.MinItems != 1 {
		t.Errorf("tags.minItems = %v, want 1", tagsProp.Schema.MinItems)
	}
	if tagsProp.Schema.MaxItems == nil || *tagsProp.Schema.MaxItems != 10 {
		t.Errorf("tags.maxItems = %v, want 10", tagsProp.Schema.MaxItems)
	}

	statusProp := schema.Properties["status"]
	if len(statusProp.Schema.Enum) != 2 {
		t.Fatalf("status.enum count = %d, want 2", len(statusProp.Schema.Enum))
	}
	if statusProp.Schema.Enum[0] != "active" {
		t.Errorf("status.enum[0] = %v", statusProp.Schema.Enum[0])
	}

	if len(schema.Required) != 2 {
		t.Errorf("required count = %d, want 2 (name, level)", len(schema.Required))
	}
}

func TestResolveFormatPatternDefault(t *testing.T) {
	src := `package test
type Config struct {
	Email   string ` + "`json:\"email\" format:\"email\"`" + `
	Phone   string ` + "`json:\"phone\" pattern:\"^\\\\+[0-9]{10,15}$\"`" + `
	Mode    string ` + "`json:\"mode\" default:\"auto\"`" + `
}
`
	lookup, file := newTestLookup(src)
	r := NewResolver(lookup)
	components := make(map[string]*spec.Schema)

	_, err := r.ResolveTypeName("Config", file, components)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	schema := components["Config"]

	emailProp := schema.Properties["email"]
	if emailProp.Schema.Format != "email" {
		t.Errorf("email.format = %q, want email", emailProp.Schema.Format)
	}

	phoneProp := schema.Properties["phone"]
	if phoneProp.Schema.Pattern == "" {
		t.Error("phone.pattern is empty")
	}

	modeProp := schema.Properties["mode"]
	if modeProp.Schema.Default != "auto" {
		t.Errorf("mode.default = %v, want auto", modeProp.Schema.Default)
	}
}

func TestResolveFieldWithoutJSONTag(t *testing.T) {
	src := `package test
type NoTag struct {
	Exported   string
	unexported string
}
`
	lookup, file := newTestLookup(src)
	r := NewResolver(lookup)
	components := make(map[string]*spec.Schema)

	_, err := r.ResolveTypeName("NoTag", file, components)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	schema := components["NoTag"]
	if _, ok := schema.Properties["Exported"]; !ok {
		t.Error("field without json tag should use Go name")
	}
}

func TestResolveEmbeddedWithJSONTag(t *testing.T) {
	src := `package test
type Inner struct {
	Value string ` + "`json:\"value\"`" + `
}
type Outer struct {
	Inner ` + "`json:\"inner\"`" + `
	Name  string ` + "`json:\"name\"`" + `
}
`
	lookup, file := newTestLookup(src)
	r := NewResolver(lookup)
	components := make(map[string]*spec.Schema)

	_, err := r.ResolveTypeName("Outer", file, components)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	schema := components["Outer"]
	if _, ok := schema.Properties["inner"]; !ok {
		t.Error("embedded with json tag should be a regular property")
	}
	if _, ok := schema.Properties["value"]; ok {
		t.Error("embedded with json tag should NOT flatten")
	}
}

func TestResolveTypeNameWithDot(t *testing.T) {
	src := `package test
type Notification struct {
	Text string ` + "`json:\"text\"`" + `
}
`
	lookup, file := newTestLookup(src)
	lookup.types["pkg.Notification"] = lookup.types["Notification"]
	r := NewResolver(lookup)
	components := make(map[string]*spec.Schema)

	ref, err := r.ResolveTypeName("pkg.Notification", file, components)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ref.Ref != "#/components/schemas/Notification" {
		t.Errorf("ref = %q, want #/components/schemas/Notification", ref.Ref)
	}
}

func TestResolvePointerEmbedding(t *testing.T) {
	src := `package test
type Base struct {
	ID string ` + "`json:\"id\"`" + `
}
type WithPointerEmbed struct {
	*Base
	Name string ` + "`json:\"name\"`" + `
}
`
	lookup, file := newTestLookup(src)
	r := NewResolver(lookup)
	components := make(map[string]*spec.Schema)

	_, err := r.ResolveTypeName("WithPointerEmbed", file, components)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	schema := components["WithPointerEmbed"]
	if _, ok := schema.Properties["id"]; !ok {
		t.Error("pointer-embedded field 'id' not flattened")
	}
	if _, ok := schema.Properties["name"]; !ok {
		t.Error("own field 'name' missing")
	}
}

func TestResolveFixedLengthArray(t *testing.T) {
	src := `package test
type Matrix struct {
	Row [3]int ` + "`json:\"row\"`" + `
}
`
	lookup, file := newTestLookup(src)
	r := NewResolver(lookup)
	components := make(map[string]*spec.Schema)

	_, err := r.ResolveTypeName("Matrix", file, components)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	schema := components["Matrix"]
	rowProp := schema.Properties["row"]
	if rowProp.Schema.Type != "array" {
		t.Errorf("row.type = %q, want array", rowProp.Schema.Type)
	}
}

func TestResolveIntegerExampleTag(t *testing.T) {
	src := `package test
type Stats struct {
	Count   int     ` + "`json:\"count\" example:\"42\"`" + `
	Rate    float64 ` + "`json:\"rate\" example:\"3.14\"`" + `
	Active  bool    ` + "`json:\"active\" example:\"true\"`" + `
}
`
	lookup, file := newTestLookup(src)
	r := NewResolver(lookup)
	components := make(map[string]*spec.Schema)

	_, err := r.ResolveTypeName("Stats", file, components)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	schema := components["Stats"]

	countProp := schema.Properties["count"]
	if countProp.Schema.Example != int64(42) {
		t.Errorf("count.example = %v (%T), want 42", countProp.Schema.Example, countProp.Schema.Example)
	}

	rateProp := schema.Properties["rate"]
	if rateProp.Schema.Example != 3.14 {
		t.Errorf("rate.example = %v (%T), want 3.14", rateProp.Schema.Example, rateProp.Schema.Example)
	}

	activeProp := schema.Properties["active"]
	if activeProp.Schema.Example != true {
		t.Errorf("active.example = %v (%T), want true", activeProp.Schema.Example, activeProp.Schema.Example)
	}
}

func TestResolveDescriptionFromComment(t *testing.T) {
	src := `package test

// UserProfile represents a user's public profile.
type UserProfile struct {
	Name string ` + "`json:\"name\"`" + `
}
`
	lookup, file := newTestLookup(src)
	r := NewResolver(lookup)
	components := make(map[string]*spec.Schema)

	_, err := r.ResolveTypeName("UserProfile", file, components)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	schema := components["UserProfile"]
	if schema.Description != "UserProfile represents a user's public profile." {
		t.Errorf("description = %q", schema.Description)
	}
}

func TestResolveMapWithComplexValue(t *testing.T) {
	src := `package test
type Item struct {
	ID string ` + "`json:\"id\"`" + `
}
type Catalog struct {
	Items map[string]Item ` + "`json:\"items\"`" + `
}
`
	lookup, file := newTestLookup(src)
	r := NewResolver(lookup)
	components := make(map[string]*spec.Schema)

	_, err := r.ResolveTypeName("Catalog", file, components)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	schema := components["Catalog"]
	itemsProp := schema.Properties["items"]
	if itemsProp.Schema.Type != "object" {
		t.Errorf("items.type = %q, want object", itemsProp.Schema.Type)
	}
	if itemsProp.Schema.AdditionalProperties == nil {
		t.Fatal("items.additionalProperties is nil")
	}
	if itemsProp.Schema.AdditionalProperties.Ref != "#/components/schemas/Item" {
		t.Errorf("items.additionalProperties.$ref = %q", itemsProp.Schema.AdditionalProperties.Ref)
	}
}

func TestResolveTypeDefInterface(t *testing.T) {
	src := `package test
type Handler interface {
	Handle()
}
`
	lookup, file := newTestLookup(src)
	r := NewResolver(lookup)
	components := make(map[string]*spec.Schema)

	ref, err := r.ResolveTypeName("Handler", file, components)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ref.Schema == nil || ref.Schema.Type != "object" {
		t.Error("interface typedef should resolve to object")
	}
}

func TestResolveTypeDefArray(t *testing.T) {
	src := `package test
type StringList []string
`
	lookup, file := newTestLookup(src)
	r := NewResolver(lookup)
	components := make(map[string]*spec.Schema)

	ref, err := r.ResolveTypeName("StringList", file, components)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ref.Schema == nil || ref.Schema.Type != "array" {
		t.Errorf("StringList should resolve to array, got %v", ref)
	}
}

func TestResolveTypeDefMap(t *testing.T) {
	src := `package test
type Metadata map[string]string
`
	lookup, file := newTestLookup(src)
	r := NewResolver(lookup)
	components := make(map[string]*spec.Schema)

	ref, err := r.ResolveTypeName("Metadata", file, components)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ref.Schema == nil || ref.Schema.Type != "object" {
		t.Errorf("Metadata should resolve to object, got %v", ref)
	}
	if ref.Schema.AdditionalProperties == nil {
		t.Error("Metadata should have additionalProperties")
	}
}

func TestResolveAlreadyParsedType(t *testing.T) {
	src := `package test
type Shared struct {
	ID string ` + "`json:\"id\"`" + `
}
type A struct {
	S Shared ` + "`json:\"s\"`" + `
}
type B struct {
	S Shared ` + "`json:\"s\"`" + `
}
`
	lookup, file := newTestLookup(src)
	r := NewResolver(lookup)
	components := make(map[string]*spec.Schema)

	_, err := r.ResolveTypeName("A", file, components)
	if err != nil {
		t.Fatalf("unexpected error resolving A: %v", err)
	}

	_, err = r.ResolveTypeName("B", file, components)
	if err != nil {
		t.Fatalf("unexpected error resolving B: %v", err)
	}

	if _, ok := components["Shared"]; !ok {
		t.Error("Shared schema should exist")
	}

	aSchema := components["A"]
	if aSchema.Properties["s"].Ref != "#/components/schemas/Shared" {
		t.Errorf("A.s ref = %q", aSchema.Properties["s"].Ref)
	}
	bSchema := components["B"]
	if bSchema.Properties["s"].Ref != "#/components/schemas/Shared" {
		t.Errorf("B.s ref = %q", bSchema.Properties["s"].Ref)
	}
}

func assertNullableUnion(t *testing.T, prop *spec.SchemaRef, expectInner func(*testing.T, *spec.SchemaRef)) {
	t.Helper()
	if prop == nil || prop.Schema == nil {
		t.Fatalf("expected inline schema with oneOf, got %+v", prop)
	}
	if len(prop.Schema.OneOf) != 2 {
		t.Fatalf("expected oneOf of length 2, got %d", len(prop.Schema.OneOf))
	}
	expectInner(t, prop.Schema.OneOf[0])
	null := prop.Schema.OneOf[1]
	if null == nil || null.Schema == nil || null.Schema.Type != "null" {
		t.Fatalf("expected second oneOf entry to be {type:null}, got %+v", null)
	}
}

func TestResolvePointerPrimitive(t *testing.T) {
	src := `package test
type Wrapper struct {
	Name *string ` + "`json:\"name\"`" + `
}
`
	lookup, file := newTestLookup(src)
	r := NewResolver(lookup)
	components := make(map[string]*spec.Schema)

	_, err := r.ResolveTypeName("Wrapper", file, components)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	prop := components["Wrapper"].Properties["name"]
	assertNullableUnion(t, prop, func(t *testing.T, inner *spec.SchemaRef) {
		t.Helper()
		if inner.Schema == nil || inner.Schema.Type != "string" {
			t.Errorf("expected first oneOf to be {type:string}, got %+v", inner)
		}
	})
}

func TestResolvePointerNamedStruct(t *testing.T) {
	src := `package test
type Inner struct {
	X int ` + "`json:\"x\"`" + `
}
type Outer struct {
	Ptr *Inner ` + "`json:\"ptr\"`" + `
}
`
	lookup, file := newTestLookup(src)
	r := NewResolver(lookup)
	components := make(map[string]*spec.Schema)

	_, err := r.ResolveTypeName("Outer", file, components)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	prop := components["Outer"].Properties["ptr"]
	assertNullableUnion(t, prop, func(t *testing.T, inner *spec.SchemaRef) {
		t.Helper()
		if inner.Ref != "#/components/schemas/Inner" {
			t.Errorf("expected first oneOf to be $ref Inner, got %+v", inner)
		}
	})
}

func TestResolveDoublePointerNotDoubleWrapped(t *testing.T) {
	src := `package test
type Inner struct {
	X int ` + "`json:\"x\"`" + `
}
type Outer struct {
	Ptr **Inner ` + "`json:\"ptr\"`" + `
}
`
	lookup, file := newTestLookup(src)
	r := NewResolver(lookup)
	components := make(map[string]*spec.Schema)

	_, err := r.ResolveTypeName("Outer", file, components)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	prop := components["Outer"].Properties["ptr"]
	if prop == nil || prop.Schema == nil {
		t.Fatalf("expected inline schema, got %+v", prop)
	}
	nullCount := 0
	for _, opt := range prop.Schema.OneOf {
		if opt != nil && opt.Schema != nil && opt.Schema.Type == "null" {
			nullCount++
		}
	}
	if nullCount != 1 {
		t.Errorf("expected exactly one null branch in oneOf, got %d (schema=%+v)", nullCount, prop.Schema)
	}
}
