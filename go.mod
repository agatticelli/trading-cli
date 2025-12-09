module github.com/agatticelli/trading-cli

go 1.25.1

replace github.com/agatticelli/trading-go => ../trading-go

replace github.com/agatticelli/strategy-go => ../strategy-go

replace github.com/agatticelli/intent-go => ../intent-go

require (
	github.com/agatticelli/intent-go v0.1.0 // indirect
	github.com/agatticelli/strategy-go v0.1.0 // indirect
	github.com/agatticelli/trading-go v0.1.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/cobra v1.10.2 // indirect
	github.com/spf13/pflag v1.0.9 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
