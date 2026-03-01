package gen

type Config struct {
	SearchDirs  []string
	MainAPIFile string
	OutputDir   string
	OutputTypes []string
	Excludes    []string
	Strict      bool
}

func DefaultConfig() *Config {
	return &Config{
		SearchDirs:  []string{"."},
		MainAPIFile: "main.go",
		OutputDir:   "./docs",
		OutputTypes: []string{"json", "yaml"},
	}
}
