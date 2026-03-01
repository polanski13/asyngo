package schema

import "testing"

func TestMapType(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input      string
		wantType   string
		wantFormat string
		wantOK     bool
	}{
		{"string", "string", "", true},
		{"int", "integer", "", true},
		{"int64", "integer", "int64", true},
		{"float64", "number", "double", true},
		{"float32", "number", "float", true},
		{"bool", "boolean", "", true},
		{"time.Time", "string", "date-time", true},
		{"uuid.UUID", "string", "uuid", true},
		{"json.RawMessage", "object", "", true},
		{"MyCustomType", "", "", false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			typ, format, ok := mapType(tt.input)
			if ok != tt.wantOK {
				t.Fatalf("MapType(%q) ok = %v, want %v", tt.input, ok, tt.wantOK)
			}
			if typ != tt.wantType {
				t.Errorf("MapType(%q) type = %q, want %q", tt.input, typ, tt.wantType)
			}
			if format != tt.wantFormat {
				t.Errorf("MapType(%q) format = %q, want %q", tt.input, format, tt.wantFormat)
			}
		})
	}
}
