package spec

type Info struct {
	Title          string        `json:"title" yaml:"title"`
	Version        string        `json:"version" yaml:"version"`
	Description    string        `json:"description,omitempty" yaml:"description,omitempty"`
	TermsOfService string        `json:"termsOfService,omitempty" yaml:"termsOfService,omitempty"`
	Contact        *Contact      `json:"contact,omitempty" yaml:"contact,omitempty"`
	License        *License      `json:"license,omitempty" yaml:"license,omitempty"`
	Tags           []Tag         `json:"tags,omitempty" yaml:"tags,omitempty"`
	ExternalDocs   *ExternalDocs `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
}

type Contact struct {
	Name  string `json:"name,omitempty" yaml:"name,omitempty"`
	URL   string `json:"url,omitempty" yaml:"url,omitempty"`
	Email string `json:"email,omitempty" yaml:"email,omitempty"`
}

type License struct {
	Name string `json:"name" yaml:"name"`
	URL  string `json:"url,omitempty" yaml:"url,omitempty"`
}
