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

// Space returns a new [Spacing] based on the given number.
//
// For positive and zero values it returns [SpacingSpace],
// for negative values it returns [SpacingOverlap] with modulus of a value.
func Space(n int) Spacing {
	if n < 0 {
		return SpacingOverlap(-n)
	}

	return SpacingSpace(n)
}
