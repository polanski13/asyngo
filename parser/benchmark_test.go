package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkTokenizeArgs(b *testing.B) {
	inputs := []struct {
		name  string
		input string
	}{
		{"simple", "a b c"},
		{"quoted", `name string true "long description here" enum(x,y,z)`},
		{"long", `field1 field2 field3 field4 "multi word value" enum(a,b,c,d,e) example(1,2,3)`},
	}

	for _, tt := range inputs {
		b.Run(tt.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				tokenizeArgs(tt.input)
			}
		})
	}
}

func BenchmarkParseAnnotationLine(b *testing.B) {
	inputs := []struct {
		name  string
		input string
	}{
		{"simple", "@Title My API"},
		{"server", `@Server production wss://ws.example.com /v1 "Production endpoint"`},
		{"param", `@ChannelParam pair string true "Trading pair" enum(BTC-USD,ETH-USD)`},
	}

	for _, tt := range inputs {
		b.Run(tt.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				parseAnnotationLine(tt.input)
			}
		})
	}
}

func BenchmarkParse(b *testing.B) {
	testdataDir, err := filepath.Abs("../testdata/basic")
	if err != nil {
		b.Fatal(err)
	}
	if _, err := os.Stat(testdataDir); err != nil {
		b.Skipf("testdata not found: %v", err)
	}

	for i := 0; i < b.N; i++ {
		p := New(WithSearchDirs(testdataDir), WithMainFile("main.go"))
		_, err := p.Parse()
		if err != nil {
			b.Fatal(err)
		}
	}
}
