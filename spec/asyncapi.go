package spec

type AsyncAPI struct {
	AsyncAPI           string               `json:"asyncapi" yaml:"asyncapi"`
	ID                 string               `json:"id,omitempty" yaml:"id,omitempty"`
	Info               Info                 `json:"info" yaml:"info"`
	DefaultContentType string               `json:"defaultContentType,omitempty" yaml:"defaultContentType,omitempty"`
	Servers            map[string]Server    `json:"servers,omitempty" yaml:"servers,omitempty"`
	Channels           map[string]Channel   `json:"channels,omitempty" yaml:"channels,omitempty"`
	Operations         map[string]Operation `json:"operations,omitempty" yaml:"operations,omitempty"`
	Components         *Components          `json:"components,omitempty" yaml:"components,omitempty"`
}

func NewAsyncAPI() *AsyncAPI {
	return &AsyncAPI{
		AsyncAPI:   "3.0.0",
		Servers:    make(map[string]Server),
		Channels:   make(map[string]Channel),
		Operations: make(map[string]Operation),
		Components: NewComponents(),
	}
}
