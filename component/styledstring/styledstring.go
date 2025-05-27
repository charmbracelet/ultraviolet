package styledstring

import (
	"github.com/charmbracelet/tv"
	"github.com/charmbracelet/x/ansi"
)

// StyledString is a styled string component that can be rendered to a screen.
type StyledString struct{ *tv.StyledString }

// New creates a new [StyledString].
func New(method ansi.Method, str string) StyledString {
	return StyledString{tv.NewStyledString(method, str)}
}

var _ tv.Component = StyledString{}
