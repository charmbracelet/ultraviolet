package styledstring

import (
	"github.com/charmbracelet/uv"
)

// StyledString is a styled string component that can be rendered to a screen.
type StyledString struct{ *uv.StyledString }

// New creates a new [StyledString].
func New(str string) StyledString {
	return StyledString{uv.NewStyledString(str)}
}

var _ uv.Drawable = StyledString{}
