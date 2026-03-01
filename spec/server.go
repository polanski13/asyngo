package spec

type Server struct {
	Host            string                    `json:"host" yaml:"host"`
	Protocol        string                    `json:"protocol" yaml:"protocol"`
	ProtocolVersion string                    `json:"protocolVersion,omitempty" yaml:"protocolVersion,omitempty"`
	Pathname        string                    `json:"pathname,omitempty" yaml:"pathname,omitempty"`
	Description     string                    `json:"description,omitempty" yaml:"description,omitempty"`
	Title           string                    `json:"title,omitempty" yaml:"title,omitempty"`
	Summary         string                    `json:"summary,omitempty" yaml:"summary,omitempty"`
	Variables       map[string]ServerVariable `json:"variables,omitempty" yaml:"variables,omitempty"`
	Security        []Reference               `json:"security,omitempty" yaml:"security,omitempty"`
	Tags            []Tag                     `json:"tags,omitempty" yaml:"tags,omitempty"`
	ExternalDocs    *ExternalDocs             `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
	Bindings        map[string]any            `json:"bindings,omitempty" yaml:"bindings,omitempty"`
}

type ServerVariable struct {
	Enum        []string `json:"enum,omitempty" yaml:"enum,omitempty"`
	Default     string   `json:"default,omitempty" yaml:"default,omitempty"`
	Description string   `json:"description,omitempty" yaml:"description,omitempty"`
	Examples    []string `json:"examples,omitempty" yaml:"examples,omitempty"`
}
