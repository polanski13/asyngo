package spec

type Reference struct {
	Ref string `json:"$ref" yaml:"$ref"`
}

func NewRef(path string) Reference {
	return Reference{Ref: path}
}

func ComponentSchemaRef(name string) string {
	return "#/components/schemas/" + name
}

func ComponentMessageRef(name string) string {
	return "#/components/messages/" + name
}

func ServerRef(name string) string {
	return "#/servers/" + name
}

func ChannelRef(name string) string {
	return "#/channels/" + name
}

func ChannelMessageRef(channel, message string) string {
	return "#/channels/" + channel + "/messages/" + message
}
