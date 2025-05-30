package tv

import "github.com/charmbracelet/x/ansi"

// screenBuffer is a buffer that is disguised as a screen.
type screenBuffer struct {
	*Buffer
}

var _ Screen = screenBuffer{}

// Size implements Screen.
func (s screenBuffer) Size() Size {
	return Size{Width: s.Buffer.Width(), Height: s.Buffer.Height()}
}

// WidthMethod implements Screen.
func (s screenBuffer) WidthMethod() WidthMethod {
	return ansi.WcWidth
}
