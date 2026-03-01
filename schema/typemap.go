package schema

type schemaType struct {
	Type   string
	Format string
}

var goTypeToSchema = map[string]schemaType{
	"bool":    {Type: "boolean"},
	"string":  {Type: "string"},
	"byte":    {Type: "integer", Format: "int32"},
	"rune":    {Type: "integer", Format: "int32"},
	"int":     {Type: "integer"},
	"int8":    {Type: "integer", Format: "int32"},
	"int16":   {Type: "integer", Format: "int32"},
	"int32":   {Type: "integer", Format: "int32"},
	"int64":   {Type: "integer", Format: "int64"},
	"uint":    {Type: "integer"},
	"uint8":   {Type: "integer", Format: "int32"},
	"uint16":  {Type: "integer", Format: "int32"},
	"uint32":  {Type: "integer", Format: "int32"},
	"uint64":  {Type: "integer", Format: "int64"},
	"float32": {Type: "number", Format: "float"},
	"float64": {Type: "number", Format: "double"},
}

var specialTypeToSchema = map[string]schemaType{
	"time.Time":            {Type: "string", Format: "date-time"},
	"time.Duration":        {Type: "string"},
	"json.RawMessage":      {Type: "object"},
	"uuid.UUID":            {Type: "string", Format: "uuid"},
	"decimal.Decimal":      {Type: "string"},
	"net.IP":               {Type: "string", Format: "ipv4"},
	"url.URL":              {Type: "string", Format: "uri"},
	"mail.Address":         {Type: "string", Format: "email"},
	"multipart.FileHeader": {Type: "string", Format: "binary"},
}

func isPrimitive(typeName string) bool {
	_, ok := goTypeToSchema[typeName]
	return ok
}

func mapType(typeName string) (typ, format string, ok bool) {
	if st, found := goTypeToSchema[typeName]; found {
		return st.Type, st.Format, true
	}
	if st, found := specialTypeToSchema[typeName]; found {
		return st.Type, st.Format, true
	}
	return "", "", false
}
