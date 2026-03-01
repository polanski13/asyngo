package spec

type Action string

const (
	ActionSend    Action = "send"
	ActionReceive Action = "receive"
)

type Operation struct {
	Action       Action          `json:"action" yaml:"action"`
	Channel      Reference       `json:"channel" yaml:"channel"`
	Title        string          `json:"title,omitempty" yaml:"title,omitempty"`
	Summary      string          `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description  string          `json:"description,omitempty" yaml:"description,omitempty"`
	Messages     []Reference     `json:"messages,omitempty" yaml:"messages,omitempty"`
	Reply        *OperationReply `json:"reply,omitempty" yaml:"reply,omitempty"`
	Security     []Reference     `json:"security,omitempty" yaml:"security,omitempty"`
	Tags         []Tag           `json:"tags,omitempty" yaml:"tags,omitempty"`
	ExternalDocs *ExternalDocs   `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
	Bindings     map[string]any  `json:"bindings,omitempty" yaml:"bindings,omitempty"`
	Traits       []Reference     `json:"traits,omitempty" yaml:"traits,omitempty"`
}

type OperationReply struct {
	Channel  *Reference    `json:"channel,omitempty" yaml:"channel,omitempty"`
	Messages []Reference   `json:"messages,omitempty" yaml:"messages,omitempty"`
	Address  *ReplyAddress `json:"address,omitempty" yaml:"address,omitempty"`
}

type ReplyAddress struct {
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Location    string `json:"location" yaml:"location"`
}
