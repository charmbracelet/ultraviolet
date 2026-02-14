package layout

// Spacing represents the spacing between segments in a layout.
//
// The [Spacing] enum is used to define the spacing between segments in a layout. It can represent
// either positive spacing (space between segments) or negative spacing (overlap between segments).
type Spacing interface{ isSpacing() }

type (
	// SpacingSpace represents positive spacing between segments.
	// The value indicates the number of cells.
	SpacingSpace int

	// SpacingOverlap represents negative spacing, causing overlap between segments.
	// The value indicates the number of overlapping cells.
	SpacingOverlap int
)

func (SpacingSpace) isSpacing()   {}
func (SpacingOverlap) isSpacing() {}
