package layout

type Spacing interface{ isSpacing() }

type (
	SpacingSpace   int
	SpacingOverlap int
)

func (SpacingSpace) isSpacing()   {}
func (SpacingOverlap) isSpacing() {}
