package spec

import "encoding/json"

type Message struct {
	Name          string           `json:"name,omitempty" yaml:"name,omitempty"`
	Title         string           `json:"title,omitempty" yaml:"title,omitempty"`
	Summary       string           `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description   string           `json:"description,omitempty" yaml:"description,omitempty"`
	ContentType   string           `json:"contentType,omitempty" yaml:"contentType,omitempty"`
	Payload       *SchemaRef       `json:"payload,omitempty" yaml:"payload,omitempty"`
	Headers       *SchemaRef       `json:"headers,omitempty" yaml:"headers,omitempty"`
	CorrelationID *CorrelationID   `json:"correlationId,omitempty" yaml:"correlationId,omitempty"`
	Tags          []Tag            `json:"tags,omitempty" yaml:"tags,omitempty"`
	ExternalDocs  *ExternalDocs    `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
	Bindings      map[string]any   `json:"bindings,omitempty" yaml:"bindings,omitempty"`
	Examples      []MessageExample `json:"examples,omitempty" yaml:"examples,omitempty"`
	Traits        []Reference      `json:"traits,omitempty" yaml:"traits,omitempty"`
}

type MessageExample struct {
	Name    string         `json:"name,omitempty" yaml:"name,omitempty"`
	Summary string         `json:"summary,omitempty" yaml:"summary,omitempty"`
	Headers map[string]any `json:"headers,omitempty" yaml:"headers,omitempty"`
	Payload map[string]any `json:"payload,omitempty" yaml:"payload,omitempty"`
}

type CorrelationID struct {
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Location    string `json:"location" yaml:"location"`
}

type MessageRef struct {
	Ref string `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	*Message
}

func (mr MessageRef) MarshalJSON() ([]byte, error) {
	if mr.Ref != "" {
		return json.Marshal(struct {
			Ref string `json:"$ref"`
		}{Ref: mr.Ref})
	}
	if mr.Message != nil {
		return json.Marshal(*mr.Message)
	}
	return []byte("{}"), nil
}
