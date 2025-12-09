module github.com/agatticelli/trading-cli

go 1.25.1

replace github.com/agatticelli/calculator-go => ../calculator-go

replace github.com/agatticelli/trading-go => ../trading-go

replace github.com/agatticelli/strategy-go => ../strategy-go

replace github.com/agatticelli/intent-go => ../intent-go

require (
	github.com/agatticelli/calculator-go v0.0.0-00010101000000-000000000000
	github.com/agatticelli/intent-go v0.1.0
	github.com/agatticelli/strategy-go v0.1.0
	github.com/agatticelli/trading-go v0.1.0
	github.com/charmbracelet/lipgloss v1.1.0
	github.com/spf13/cobra v1.10.2
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/aymanbagabas/go-osc52/v2 v2.0.1 // indirect
	github.com/charmbracelet/colorprofile v0.2.3-0.20250311203215-f60798e515dc // indirect
	github.com/charmbracelet/x/ansi v0.8.0 // indirect
	github.com/charmbracelet/x/cellbuf v0.0.13-0.20250311204145-2c3ea96c31dd // indirect
	github.com/charmbracelet/x/term v0.2.1 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/lucasb-eyer/go-colorful v1.2.0 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/muesli/termenv v0.16.0 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/spf13/pflag v1.0.9 // indirect
	github.com/xo/terminfo v0.0.0-20220910002029-abceb7e1c41e // indirect
	golang.org/x/sys v0.30.0 // indirect
)
