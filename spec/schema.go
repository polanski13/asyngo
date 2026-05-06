package spec

import "encoding/json"

type Schema struct {
	Type                 string                `json:"type,omitempty" yaml:"type,omitempty"`
	Format               string                `json:"format,omitempty" yaml:"format,omitempty"`
	Description          string                `json:"description,omitempty" yaml:"description,omitempty"`
	Properties           map[string]*SchemaRef `json:"properties,omitempty" yaml:"properties,omitempty"`
	Required             []string              `json:"required,omitempty" yaml:"required,omitempty"`
	Items                *SchemaRef            `json:"items,omitempty" yaml:"items,omitempty"`
	Enum                 []any                 `json:"enum,omitempty" yaml:"enum,omitempty"`
	Default              any                   `json:"default,omitempty" yaml:"default,omitempty"`
	Example              any                   `json:"example,omitempty" yaml:"example,omitempty"`
	Minimum              *float64              `json:"minimum,omitempty" yaml:"minimum,omitempty"`
	Maximum              *float64              `json:"maximum,omitempty" yaml:"maximum,omitempty"`
	MinLength            *int64                `json:"minLength,omitempty" yaml:"minLength,omitempty"`
	MaxLength            *int64                `json:"maxLength,omitempty" yaml:"maxLength,omitempty"`
	Pattern              string                `json:"pattern,omitempty" yaml:"pattern,omitempty"`
	MinItems             *int64                `json:"minItems,omitempty" yaml:"minItems,omitempty"`
	MaxItems             *int64                `json:"maxItems,omitempty" yaml:"maxItems,omitempty"`
	UniqueItems          bool                  `json:"uniqueItems,omitempty" yaml:"uniqueItems,omitempty"`
	AllOf                []*SchemaRef          `json:"allOf,omitempty" yaml:"allOf,omitempty"`
	OneOf                []*SchemaRef          `json:"oneOf,omitempty" yaml:"oneOf,omitempty"`
	Discriminator        string                `json:"discriminator,omitempty" yaml:"discriminator,omitempty"`
	AnyOf                []*SchemaRef          `json:"anyOf,omitempty" yaml:"anyOf,omitempty"`
	Not                  *SchemaRef            `json:"not,omitempty" yaml:"not,omitempty"`
	ReadOnly             bool                  `json:"readOnly,omitempty" yaml:"readOnly,omitempty"`
	WriteOnly            bool                  `json:"writeOnly,omitempty" yaml:"writeOnly,omitempty"`
	Deprecated           bool                  `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
	AdditionalProperties *SchemaRef            `json:"additionalProperties,omitempty" yaml:"additionalProperties,omitempty"`
}

type SchemaRef struct {
	Ref string `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	*Schema
}

func (sr SchemaRef) MarshalJSON() ([]byte, error) {
	if sr.Ref != "" {
		return json.Marshal(struct {
			Ref string `json:"$ref"`
		}{Ref: sr.Ref})
	}
	if sr.Schema != nil {
		return json.Marshal(*sr.Schema)
	}
	return []byte("{}"), nil
}

func NewSchemaRef(ref string) *SchemaRef {
	return &SchemaRef{Ref: ref}
}

func NewInlineSchema(s *Schema) *SchemaRef {
	return &SchemaRef{Schema: s}
}
