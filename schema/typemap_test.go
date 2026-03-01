package schema

import "testing"

func TestIsPrimitive(t *testing.T) {
	primitives := []string{"bool", "string", "int", "int8", "int16", "int32", "int64", "uint", "float32", "float64", "byte", "rune"}
	for _, p := range primitives {
		if !isPrimitive(p) {
			t.Errorf("IsPrimitive(%q) = false, want true", p)
		}
	}

	nonPrimitives := []string{"MyStruct", "time.Time", "any", "interface{}"}
	for _, p := range nonPrimitives {
		if isPrimitive(p) {
			t.Errorf("IsPrimitive(%q) = true, want false", p)
		}
	}
}

func TestMapType(t *testing.T) {
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
		t.Run(tt.input, func(t *testing.T) {
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
