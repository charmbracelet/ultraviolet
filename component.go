package tv

// Component represents a displayable component on a [Buffer].
type Component interface {
	// RenderComponent renders the component on the screen within the given
	// area.
	RenderComponent(buf *Buffer, area Rectangle) error
}
