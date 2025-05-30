package uv

// Component represents a drawable component on a [Screen].
type Component interface {
	// Draw renders the component on the screen for the given area.
	Draw(scr Screen, area Rectangle)
}
