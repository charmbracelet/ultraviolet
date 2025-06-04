package layout

// Direction defines the direction in which a component should be laid out.
type Direction bool

const (
	// Horizontal indicates that the component should be laid out horizontally.
	Horizontal Direction = false
	// Vertical indicates that the component should be laid out vertically.
	Vertical Direction = true
)
