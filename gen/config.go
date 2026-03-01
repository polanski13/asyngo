package gen

type Config struct {
	SearchDir   string
	MainAPIFile string
	OutputDir   string
	OutputTypes []string
	Excludes    string
	Strict      bool
}

func DefaultConfig() *Config {
	return &Config{
		SearchDir:   ".",
		MainAPIFile: "main.go",
		OutputDir:   "./docs",
		OutputTypes: []string{"json", "yaml"},
	}
}
