module github.com/charmbracelet/uv

go 1.24.3

replace github.com/charmbracelet/x/exp/toner => ../x/exp/toner

require (
	github.com/charmbracelet/colorprofile v0.3.1
	github.com/charmbracelet/x/ansi v0.9.3-0.20250602153603-fb931ed90413
	github.com/charmbracelet/x/exp/toner v0.0.0-00010101000000-000000000000
	github.com/charmbracelet/x/term v0.2.1
	github.com/charmbracelet/x/termios v0.1.1
	github.com/charmbracelet/x/windows v0.2.1
	github.com/muesli/cancelreader v0.2.2
	github.com/rivo/uniseg v0.4.7
	github.com/xo/terminfo v0.0.0-20220910002029-abceb7e1c41e
	golang.org/x/sync v0.13.0
	golang.org/x/sys v0.32.0
)

require golang.org/x/exp v0.0.0-20231006140011-7918f672742d // indirect

require (
	github.com/charmbracelet/x/exp/charmtone v0.0.0-20250602192518-9e722df69bbb // indirect
	github.com/lucasb-eyer/go-colorful v1.2.0
	github.com/mattn/go-runewidth v0.0.16 // indirect
)
