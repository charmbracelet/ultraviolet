package dom

import (
	uv "github.com/charmbracelet/ultraviolet"
)

// vbox represents a vertical box container that stacks elements vertically.
type vbox struct {
	children []Element
}

// VBox creates a vertical box container that stacks elements from top to bottom.
func VBox(children ...Element) Element {
	return &vbox{children: children}
}

// Render implements the Element interface.
func (v *vbox) Render(scr uv.Screen, area uv.Rectangle) {
	if len(v.children) == 0 || area.Dy() <= 0 {
		return
	}

	// Calculate total minimum height needed
	totalMinHeight := 0
	flexibleChildren := 0

	for _, child := range v.children {
		_, h := child.MinSize(scr)
		if h == 0 {
			flexibleChildren++
		} else {
			totalMinHeight += h
		}
	}

	// Calculate height per flexible child
	remainingHeight := area.Dy() - totalMinHeight
	heightPerFlexible := 0
	if flexibleChildren > 0 && remainingHeight > 0 {
		heightPerFlexible = remainingHeight / flexibleChildren
	}

	// Render children
	y := area.Min.Y
	flexibleIndex := 0
	for _, child := range v.children {
		if y >= area.Max.Y {
			break
		}

		_, minH := child.MinSize(scr)
		childHeight := minH
		if minH == 0 {
			childHeight = heightPerFlexible
			flexibleIndex++
			// Give remaining height to last flexible child
			if flexibleIndex == flexibleChildren {
				childHeight = area.Max.Y - y
			}
		}

		if childHeight > 0 {
			childArea := uv.Rect(area.Min.X, y, area.Dx(), childHeight)
			child.Render(scr, childArea)
			y += childHeight
		}
	}
}

// MinSize implements the Element interface.
func (v *vbox) MinSize(scr uv.Screen) (width, height int) {
	for _, child := range v.children {
		w, h := child.MinSize(scr)
		if w > width {
			width = w
		}
		height += h
	}
	return width, height
}

// hbox represents a horizontal box container that arranges elements horizontally.
type hbox struct {
	children []Element
}

// HBox creates a horizontal box container that arranges elements from left to right.
func HBox(children ...Element) Element {
	return &hbox{children: children}
}

// Render implements the Element interface.
func (h *hbox) Render(scr uv.Screen, area uv.Rectangle) {
	if len(h.children) == 0 || area.Dx() <= 0 {
		return
	}

	// Calculate total minimum width needed
	totalMinWidth := 0
	flexibleChildren := 0

	for _, child := range h.children {
		w, _ := child.MinSize(scr)
		if w == 0 {
			flexibleChildren++
		} else {
			totalMinWidth += w
		}
	}

	// Calculate width per flexible child
	remainingWidth := area.Dx() - totalMinWidth
	widthPerFlexible := 0
	if flexibleChildren > 0 && remainingWidth > 0 {
		widthPerFlexible = remainingWidth / flexibleChildren
	}

	// Render children
	x := area.Min.X
	flexibleIndex := 0
	for _, child := range h.children {
		if x >= area.Max.X {
			break
		}

		minW, _ := child.MinSize(scr)
		childWidth := minW
		if minW == 0 {
			childWidth = widthPerFlexible
			flexibleIndex++
			// Give remaining width to last flexible child
			if flexibleIndex == flexibleChildren {
				childWidth = area.Max.X - x
			}
		}

		if childWidth > 0 {
			childArea := uv.Rect(x, area.Min.Y, childWidth, area.Dy())
			child.Render(scr, childArea)
			x += childWidth
		}
	}
}

// MinSize implements the Element interface.
func (h *hbox) MinSize(scr uv.Screen) (width, height int) {
	for _, child := range h.children {
		w, h := child.MinSize(scr)
		width += w
		if h > height {
			height = h
		}
	}
	return width, height
}

// flex represents a flexible space that expands to fill available space.
type flex struct {
	grow int
}

// Flex creates a flexible spacer that expands to fill available space.
// The grow parameter determines how much space this element should take
// relative to other flex elements (default: 1).
func Flex(grow ...int) Element {
	g := 1
	if len(grow) > 0 && grow[0] > 0 {
		g = grow[0]
	}
	return &flex{grow: g}
}

// Render implements the Element interface.
func (f *flex) Render(scr uv.Screen, area uv.Rectangle) {
	// Flex doesn't render anything, it just takes up space
}

// MinSize implements the Element interface.
func (f *flex) MinSize(scr uv.Screen) (width, height int) {
	return 0, 0
}

// spacer represents a fixed-size empty space.
type spacer struct {
	width  int
	height int
}

// Spacer creates a fixed-size empty space with the given dimensions.
func Spacer(width, height int) Element {
	return &spacer{width: width, height: height}
}

// Render implements the Element interface.
func (s *spacer) Render(scr uv.Screen, area uv.Rectangle) {
	// Spacer doesn't render anything, it just takes up space
}

// MinSize implements the Element interface.
func (s *spacer) MinSize(scr uv.Screen) (width, height int) {
	return s.width, s.height
}
