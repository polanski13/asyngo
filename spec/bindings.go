package spec

type ChannelBindings struct {
	WS *WebSocketChannelBinding `json:"ws,omitempty" yaml:"ws,omitempty"`
}

type WebSocketChannelBinding struct {
	Method         string  `json:"method,omitempty" yaml:"method,omitempty"`
	Query          *Schema `json:"query,omitempty" yaml:"query,omitempty"`
	Headers        *Schema `json:"headers,omitempty" yaml:"headers,omitempty"`
	BindingVersion string  `json:"bindingVersion,omitempty" yaml:"bindingVersion,omitempty"`
}
