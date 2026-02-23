package layout

import uv "github.com/charmbracelet/ultraviolet"

// Padding defines the inset applied to a [Layout]'s outer area before solving.
type Padding struct {
	Top, Right, Bottom, Left int
}

func (p Padding) apply(area uv.Rectangle) uv.Rectangle {
	horizontal := p.Right + p.Left
	vertical := p.Top + p.Bottom

	if area.Dx() < horizontal || area.Dy() < vertical {
		return uv.Rectangle{}
	}

	return uv.Rect(
		area.Min.X+p.Left,
		area.Min.Y+p.Top,
		max(0, area.Dx()-horizontal),
		max(0, area.Dy()-vertical),
	)
}

// Pad builds a [Padding] value from a variable number of sides,
// following the same shorthand convention as CSS:
//   - 0 args: all sides zero.
//   - 1 arg: uniform on every side.
//   - 2 args: first is top/bottom, second is left/right.
//   - 4 args: top, right, bottom, left.
//
// Any other count causes a panic.
func Pad(sides ...int) Padding {
	switch len(sides) {
	case 0:
		return Padding{}

	case 1:
		side := sides[0]

		return Padding{Top: side, Right: side, Bottom: side, Left: side}

	case 2:
		vertical := sides[0]
		horizontal := sides[1]

		return Padding{
			Top:    vertical,
			Right:  horizontal,
			Bottom: vertical,
			Left:   horizontal,
		}

	case 4:
		return Padding{
			Top:    sides[0],
			Right:  sides[1],
			Bottom: sides[2],
			Left:   sides[3],
		}

	default:
		panic("layout.Pad: unexpected sides count")
	}
}
