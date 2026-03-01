package schema

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/polanski13/asyngo/spec"
)

type TypeDef struct {
	File     *ast.File
	TypeSpec *ast.TypeSpec
	PkgPath  string
}

type PackageLookup interface {
	FindTypeSpec(typeName string, file *ast.File) (*TypeDef, error)
}

type Resolver struct {
	packages   PackageLookup
	parsed     map[string]*spec.Schema
	inProgress map[string]bool
	fp         *fieldProcessor
}

func NewResolver(packages PackageLookup) *Resolver {
	r := &Resolver{
		packages:   packages,
		parsed:     make(map[string]*spec.Schema),
		inProgress: make(map[string]bool),
	}
	r.fp = newFieldProcessor(r)
	return r
}

func (r *Resolver) ResolveExpr(
	expr ast.Expr,
	file *ast.File,
	components map[string]*spec.Schema,
) (*spec.SchemaRef, error) {
	switch t := expr.(type) {
	case *ast.Ident:
		return r.resolveIdent(t, file, components)
	case *ast.StarExpr:
		return r.resolvePointer(t, file, components)
	case *ast.ArrayType:
		return r.resolveArray(t, file, components)
	case *ast.MapType:
		return r.resolveMap(t, file, components)
	case *ast.SelectorExpr:
		return r.resolveSelector(t, file, components)
	case *ast.StructType:
		return r.resolveAnonymousStruct(t, file, components)
	case *ast.InterfaceType:
		return spec.NewInlineSchema(&spec.Schema{Type: "object"}), nil
	default:
		return nil, fmt.Errorf("%w: %T", ErrUnsupportedType, expr)
	}
}

func (r *Resolver) ResolveTypeName(
	typeName string,
	file *ast.File,
	components map[string]*spec.Schema,
) (*spec.SchemaRef, error) {
	if typ, format, ok := mapType(typeName); ok {
		s := &spec.Schema{Type: typ}
		if format != "" {
			s.Format = format
		}
		return spec.NewInlineSchema(s), nil
	}

	td, err := r.packages.FindTypeSpec(typeName, file)
	if err != nil {
		return nil, fmt.Errorf("resolving type %s: %w", typeName, err)
	}

	resolvedName := typeName
	if idx := strings.LastIndex(typeName, "."); idx >= 0 {
		resolvedName = typeName[idx+1:]
	}

	return r.resolveTypeDef(resolvedName, td, components)
}

func (r *Resolver) resolveIdent(
	ident *ast.Ident,
	file *ast.File,
	components map[string]*spec.Schema,
) (*spec.SchemaRef, error) {
	name := ident.Name

	if name == "any" {
		return spec.NewInlineSchema(&spec.Schema{Type: "object"}), nil
	}

	if typ, format, ok := mapType(name); ok {
		s := &spec.Schema{Type: typ}
		if format != "" {
			s.Format = format
		}
		return spec.NewInlineSchema(s), nil
	}

	td, err := r.packages.FindTypeSpec(name, file)
	if err != nil {
		return nil, fmt.Errorf("resolving type %s: %w", name, err)
	}

	return r.resolveTypeDef(name, td, components)
}

func (r *Resolver) resolvePointer(
	star *ast.StarExpr,
	file *ast.File,
	components map[string]*spec.Schema,
) (*spec.SchemaRef, error) {
	inner, err := r.ResolveExpr(star.X, file, components)
	if err != nil {
		return nil, err
	}
	if inner.Schema != nil {
		inner.Schema.Nullable = true
	}
	return inner, nil
}

func (r *Resolver) resolveArray(
	arr *ast.ArrayType,
	file *ast.File,
	components map[string]*spec.Schema,
) (*spec.SchemaRef, error) {
	items, err := r.ResolveExpr(arr.Elt, file, components)
	if err != nil {
		return nil, err
	}
	return spec.NewInlineSchema(&spec.Schema{
		Type:  "array",
		Items: items,
	}), nil
}

func (r *Resolver) resolveMap(
	m *ast.MapType,
	file *ast.File,
	components map[string]*spec.Schema,
) (*spec.SchemaRef, error) {
	valSchema, err := r.ResolveExpr(m.Value, file, components)
	if err != nil {
		return nil, err
	}
	return spec.NewInlineSchema(&spec.Schema{
		Type:                 "object",
		AdditionalProperties: valSchema,
	}), nil
}

func (r *Resolver) resolveSelector(
	sel *ast.SelectorExpr,
	file *ast.File,
	components map[string]*spec.Schema,
) (*spec.SchemaRef, error) {
	pkgIdent, ok := sel.X.(*ast.Ident)
	if !ok {
		return nil, fmt.Errorf("unexpected selector expression")
	}

	fullName := pkgIdent.Name + "." + sel.Sel.Name

	if typ, format, ok := mapType(fullName); ok {
		s := &spec.Schema{Type: typ}
		if format != "" {
			s.Format = format
		}
		return spec.NewInlineSchema(s), nil
	}

	td, err := r.packages.FindTypeSpec(fullName, file)
	if err != nil {
		return nil, fmt.Errorf("resolving type %s: %w", fullName, err)
	}

	return r.resolveTypeDef(sel.Sel.Name, td, components)
}

func (r *Resolver) resolveAnonymousStruct(
	st *ast.StructType,
	file *ast.File,
	components map[string]*spec.Schema,
) (*spec.SchemaRef, error) {
	schema, err := r.resolveStructFields(st, file, components)
	if err != nil {
		return nil, err
	}
	return spec.NewInlineSchema(schema), nil
}

func (r *Resolver) resolveTypeDef(
	name string,
	td *TypeDef,
	components map[string]*spec.Schema,
) (*spec.SchemaRef, error) {
	if _, exists := r.parsed[name]; exists {
		return spec.NewSchemaRef(spec.ComponentSchemaRef(name)), nil
	}

	if r.inProgress[name] {
		return spec.NewSchemaRef(spec.ComponentSchemaRef(name)), nil
	}

	switch t := td.TypeSpec.Type.(type) {
	case *ast.StructType:
		r.inProgress[name] = true
		defer delete(r.inProgress, name)

		schema, err := r.resolveStructFields(t, td.File, components)
		if err != nil {
			return nil, fmt.Errorf("resolving struct %s: %w", name, err)
		}

		if td.TypeSpec.Doc != nil {
			desc := extractDescription(td.TypeSpec.Doc)
			if desc != "" {
				schema.Description = desc
			}
		}

		components[name] = schema
		r.parsed[name] = schema
		return spec.NewSchemaRef(spec.ComponentSchemaRef(name)), nil

	case *ast.Ident:
		return r.resolveIdent(t, td.File, components)

	case *ast.ArrayType:
		return r.resolveArray(t, td.File, components)

	case *ast.MapType:
		return r.resolveMap(t, td.File, components)

	case *ast.InterfaceType:
		return spec.NewInlineSchema(&spec.Schema{Type: "object"}), nil

	default:
		return nil, fmt.Errorf("%w: %s has underlying type %T", ErrUnsupportedType, name, td.TypeSpec.Type)
	}
}

func (r *Resolver) resolveStructFields(
	st *ast.StructType,
	file *ast.File,
	components map[string]*spec.Schema,
) (*spec.Schema, error) {
	schema := &spec.Schema{
		Type:       "object",
		Properties: make(map[string]*spec.SchemaRef),
	}

	if st.Fields == nil {
		return schema, nil
	}

	for _, field := range st.Fields.List {
		if len(field.Names) == 0 {
			if err := r.handleEmbeddedField(field, file, schema, components); err != nil {
				return nil, err
			}
			continue
		}

		name, prop, required, skip, err := r.fp.processField(field, file, components)
		if err != nil {
			return nil, err
		}
		if skip {
			continue
		}

		schema.Properties[name] = prop
		if required {
			schema.Required = append(schema.Required, name)
		}
	}

	return schema, nil
}

func (r *Resolver) handleEmbeddedField(
	field *ast.Field,
	file *ast.File,
	parent *spec.Schema,
	components map[string]*spec.Schema,
) error {
	tag := extractStructTag(field)
	jsonTag := tag.Get("json")
	if jsonTag != "" && jsonTag != "-" {
		name := strings.Split(jsonTag, ",")[0]
		if name != "" {
			prop, err := r.ResolveExpr(field.Type, file, components)
			if err != nil {
				return err
			}
			parent.Properties[name] = prop
			return nil
		}
	}

	resolved, err := r.ResolveExpr(field.Type, file, components)
	if err != nil {
		return err
	}

	if resolved.Ref != "" {
		refName := resolved.Ref
		refName = refName[strings.LastIndex(refName, "/")+1:]
		if embeddedSchema, ok := components[refName]; ok {
			for k, v := range embeddedSchema.Properties {
				if _, exists := parent.Properties[k]; !exists {
					parent.Properties[k] = v
				}
			}
			for _, req := range embeddedSchema.Required {
				parent.Required = append(parent.Required, req)
			}
		}
	} else if resolved.Schema != nil && resolved.Schema.Properties != nil {
		for k, v := range resolved.Schema.Properties {
			if _, exists := parent.Properties[k]; !exists {
				parent.Properties[k] = v
			}
		}
		for _, req := range resolved.Schema.Required {
			parent.Required = append(parent.Required, req)
		}
	}

	return nil
}
