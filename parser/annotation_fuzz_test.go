package parser

import "testing"

func FuzzTokenizeArgs(f *testing.F) {
	f.Add("")
	f.Add("simple")
	f.Add("a b c")
	f.Add(`"quoted string"`)
	f.Add(`a "multi word" b`)
	f.Add("enum(a,b,c)")
	f.Add(`name string true "description" enum(x,y)`)
	f.Add(`"hello`)
	f.Add(`enum(a,b`)
	f.Add(`""`)
	f.Add(`"enum(x)"`)
	f.Add("   ")

	f.Fuzz(func(t *testing.T, input string) {
		tokenizeArgs(input)
	})
}

func FuzzParseAnnotationLine(f *testing.F) {
	f.Add("@Title My API")
	f.Add("@Version 1.0.0")
	f.Add("@Server production wss://ws.example.com /v1")
	f.Add(`@ChannelParam pair string true "Trading pair" enum(BTC-USD,ETH-USD)`)
	f.Add("@")
	f.Add("@Name")
	f.Add("")
	f.Add("no-at-prefix")

	f.Fuzz(func(t *testing.T, input string) {
		parseAnnotationLine(input)
	})
}
