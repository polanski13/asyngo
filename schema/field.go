package schema

import (
	"go/ast"
	"reflect"
	"strconv"
	"strings"

	"github.com/polanski13/asyngo/spec"
)

type fieldProcessor struct {
	resolver *Resolver
}

func newFieldProcessor(r *Resolver) *fieldProcessor {
	return &fieldProcessor{resolver: r}
}

func (fp *fieldProcessor) processField(
	field *ast.Field,
	file *ast.File,
	components map[string]*spec.Schema,
) (name string, prop *spec.SchemaRef, required bool, skip bool, err error) {
	tag := extractStructTag(field)

	if tag.Get("asyncapiignore") == "true" {
		return "", nil, false, true, nil
	}

	name = fieldJSONName(field, tag)
	if name == "-" || name == "" {
		return "", nil, false, true, nil
	}

	prop, err = fp.resolver.ResolveExpr(field.Type, file, components)
	if err != nil {
		return "", nil, false, false, err
	}

	applyTags(prop, tag)

	required = isRequired(tag)

	if field.Doc != nil {
		desc := extractDescription(field.Doc)
		if desc != "" && prop.Schema != nil {
			prop.Schema.Description = desc
		}
	}

	return name, prop, required, false, nil
}

func fieldJSONName(field *ast.Field, tag reflect.StructTag) string {
	jsonTag := tag.Get("json")
	if jsonTag == "" {
		if len(field.Names) > 0 {
			return field.Names[0].Name
		}
		return ""
	}
	parts := strings.Split(jsonTag, ",")
	name := parts[0]
	if name == "" && len(field.Names) > 0 {
		return field.Names[0].Name
	}
	return name
}

func extractStructTag(field *ast.Field) reflect.StructTag {
	if field.Tag == nil {
		return ""
	}
	raw := field.Tag.Value
	raw = strings.Trim(raw, "`")
	return reflect.StructTag(raw)
}

func extractDescription(doc *ast.CommentGroup) string {
	if doc == nil {
		return ""
	}
	var parts []string
	for _, c := range doc.List {
		text := strings.TrimPrefix(c.Text, "//")
		text = strings.TrimPrefix(text, " ")
		if !strings.HasPrefix(text, "@") {
			parts = append(parts, text)
		}
	}
	return strings.TrimSpace(strings.Join(parts, " "))
}

func isRequired(tag reflect.StructTag) bool {
	for _, key := range []string{"validate", "binding"} {
		v := tag.Get(key)
		for _, rule := range strings.Split(v, ",") {
			if rule == "required" {
				return true
			}
		}
	}
	return false
}

func applyTags(prop *spec.SchemaRef, tag reflect.StructTag) {
	if prop.Ref != "" || prop.Schema == nil {
		return
	}
	s := prop.Schema

	if v := tag.Get("example"); v != "" {
		s.Example = parseTagValue(s.Type, v)
	}
	if v := tag.Get("default"); v != "" {
		s.Default = parseTagValue(s.Type, v)
	}
	if v := tag.Get("enum"); v != "" {
		parts := strings.Split(v, ",")
		enums := make([]any, len(parts))
		for i, p := range parts {
			enums[i] = parseTagValue(s.Type, strings.TrimSpace(p))
		}
		s.Enum = enums
	}
	if v := tag.Get("minimum"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			s.Minimum = &f
		}
	}
	if v := tag.Get("maximum"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			s.Maximum = &f
		}
	}
	if v := tag.Get("format"); v != "" {
		s.Format = v
	}
	if v := tag.Get("pattern"); v != "" {
		s.Pattern = v
	}

	applyValidationTag(s, tag.Get("validate"))
	applyValidationTag(s, tag.Get("binding"))
}

func applyValidationTag(s *spec.Schema, validate string) {
	if validate == "" {
		return
	}
	for _, rule := range strings.Split(validate, ",") {
		rule = strings.TrimSpace(rule)
		if strings.HasPrefix(rule, "min=") {
			if v, err := strconv.ParseFloat(strings.TrimPrefix(rule, "min="), 64); err == nil {
				if s.Type == "string" {
					i := int64(v)
					s.MinLength = &i
				} else if s.Type == "array" {
					i := int64(v)
					s.MinItems = &i
				} else {
					s.Minimum = &v
				}
			}
		}
		if strings.HasPrefix(rule, "max=") {
			if v, err := strconv.ParseFloat(strings.TrimPrefix(rule, "max="), 64); err == nil {
				if s.Type == "string" {
					i := int64(v)
					s.MaxLength = &i
				} else if s.Type == "array" {
					i := int64(v)
					s.MaxItems = &i
				} else {
					s.Maximum = &v
				}
			}
		}
		if strings.HasPrefix(rule, "oneof=") {
			vals := strings.Fields(strings.TrimPrefix(rule, "oneof="))
			enums := make([]any, len(vals))
			for i, v := range vals {
				enums[i] = parseTagValue(s.Type, v)
			}
			s.Enum = enums
		}
	}
}

func parseTagValue(schemaType, val string) any {
	switch schemaType {
	case "integer":
		if i, err := strconv.ParseInt(val, 10, 64); err == nil {
			return i
		}
	case "number":
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
	case "boolean":
		if b, err := strconv.ParseBool(val); err == nil {
			return b
		}
	}
	return val
}
