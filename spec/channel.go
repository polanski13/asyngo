package spec

type Channel struct {
	Address      string                `json:"address,omitempty" yaml:"address,omitempty"`
	Messages     map[string]MessageRef `json:"messages,omitempty" yaml:"messages,omitempty"`
	Title        string                `json:"title,omitempty" yaml:"title,omitempty"`
	Summary      string                `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description  string                `json:"description,omitempty" yaml:"description,omitempty"`
	Servers      []Reference           `json:"servers,omitempty" yaml:"servers,omitempty"`
	Parameters   map[string]Parameter  `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	Tags         []Tag                 `json:"tags,omitempty" yaml:"tags,omitempty"`
	ExternalDocs *ExternalDocs         `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
	Bindings     *ChannelBindings      `json:"bindings,omitempty" yaml:"bindings,omitempty"`
}

type Parameter struct {
	Description string   `json:"description,omitempty" yaml:"description,omitempty"`
	Enum        []string `json:"enum,omitempty" yaml:"enum,omitempty"`
	Default     string   `json:"default,omitempty" yaml:"default,omitempty"`
	Examples    []string `json:"examples,omitempty" yaml:"examples,omitempty"`
	Location    string   `json:"location,omitempty" yaml:"location,omitempty"`
}
