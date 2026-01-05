package dom

import (
	"strings"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
	"github.com/clipperhouse/displaywidth"
)

// button represents a clickable button element.
type button struct {
	label    string
	style    uv.Style
	active   bool
	onSelect func()
}

// Button creates a button with the given label.
func Button(label string) Element {
	return &button{
		label: label,
		style: uv.Style{
			Fg: ansi.White,
			Bg: ansi.Blue,
		},
	}
}

// ButtonStyled creates a button with custom styling.
func ButtonStyled(label string, style uv.Style) Element {
	return &button{
		label: label,
		style: style,
	}
}

// Render implements the Element interface.
func (b *button) Render(scr uv.Screen, area uv.Rectangle) {
	if area.Dx() <= 0 || area.Dy() <= 0 {
		return
	}

	// Render button with padding
	text := " " + b.label + " "
	x := area.Min.X
	y := area.Min.Y

	gr := displaywidth.StringGraphemes(text)
	for gr.Next() {
		if x >= area.Max.X {
			break
		}

		grapheme := string(gr.Value())
		width := gr.Width()

		cell := &uv.Cell{
			Content: grapheme,
			Width:   width,
			Style:   b.style,
		}
		scr.SetCell(x, y, cell)
		x += width
	}
}

// MinSize implements the Element interface.
func (b *button) MinSize(scr uv.Screen) (width, height int) {
	// Button has padding of 1 on each side
	text := " " + b.label + " "
	return displaywidth.String(text), 1
}

// input represents a text input field.
type input struct {
	placeholder string
	value       string
	style       uv.Style
	width       int
}

// Input creates a text input field with the given width.
func Input(width int, placeholder string) Element {
	return &input{
		placeholder: placeholder,
		width:       width,
		style: uv.Style{
			Fg: ansi.White,
			Bg: ansi.Black,
		},
	}
}

// InputStyled creates a text input with custom styling.
func InputStyled(width int, placeholder string, style uv.Style) Element {
	return &input{
		placeholder: placeholder,
		width:       width,
		style:       style,
	}
}

// Render implements the Element interface.
func (i *input) Render(scr uv.Screen, area uv.Rectangle) {
	if area.Dx() <= 0 || area.Dy() <= 0 {
		return
	}

	width := i.width
	if width > area.Dx() {
		width = area.Dx()
	}

	text := i.value
	if text == "" {
		text = i.placeholder
	}

	// Truncate or pad to width
	textWidth := displaywidth.String(text)
	if textWidth > width {
		// Truncate safely using graphemes
		gr := displaywidth.StringGraphemes(text)
		var truncated strings.Builder
		currentWidth := 0
		for gr.Next() {
			w := gr.Width()
			if currentWidth+w > width {
				break
			}
			truncated.WriteString(string(gr.Value()))
			currentWidth += w
		}
		text = truncated.String()
		textWidth = currentWidth
	}
	
	if textWidth < width {
		// Pad with spaces
		text += strings.Repeat(" ", width-textWidth)
	}

	x := area.Min.X
	y := area.Min.Y

	gr := displaywidth.StringGraphemes(text)
	for gr.Next() {
		if x >= area.Min.X+width || x >= area.Max.X {
			break
		}

		grapheme := string(gr.Value())
		gWidth := gr.Width()

		cell := &uv.Cell{
			Content: grapheme,
			Width:   gWidth,
			Style:   i.style,
		}
		scr.SetCell(x, y, cell)
		x += gWidth
	}
}

// MinSize implements the Element interface.
func (i *input) MinSize(scr uv.Screen) (width, height int) {
	return i.width, 1
}

// checkbox represents a checkbox element.
type checkbox struct {
	label   string
	checked bool
	style   uv.Style
}

// Checkbox creates a checkbox with the given label.
func Checkbox(label string, checked bool) Element {
	return &checkbox{
		label:   label,
		checked: checked,
	}
}

// CheckboxStyled creates a checkbox with custom styling.
func CheckboxStyled(label string, checked bool, style uv.Style) Element {
	return &checkbox{
		label:   label,
		checked: checked,
		style:   style,
	}
}

// Render implements the Element interface.
func (c *checkbox) Render(scr uv.Screen, area uv.Rectangle) {
	if area.Dx() <= 0 || area.Dy() <= 0 {
		return
	}

	checkmark := "[ ]"
	if c.checked {
		checkmark = "[✓]"
	}

	text := checkmark + " " + c.label
	x := area.Min.X
	y := area.Min.Y

	gr := displaywidth.StringGraphemes(text)
	for gr.Next() {
		if x >= area.Max.X {
			break
		}

		grapheme := string(gr.Value())
		width := gr.Width()

		cell := &uv.Cell{
			Content: grapheme,
			Width:   width,
			Style:   c.style,
		}
		scr.SetCell(x, y, cell)
		x += width
	}
}

// MinSize implements the Element interface.
func (c *checkbox) MinSize(scr uv.Screen) (width, height int) {
	checkmark := "[ ]"
	if c.checked {
		checkmark = "[✓]"
	}
	text := checkmark + " " + c.label
	return displaywidth.String(text), 1
}

// window represents a titled window/panel.
type window struct {
	title string
	child Element
	style uv.Style
}

// Window creates a window with a title bar and content.
func Window(title string, child Element) Element {
	return &window{
		title: title,
		child: child,
	}
}

// WindowStyled creates a window with custom styling.
func WindowStyled(title string, child Element, style uv.Style) Element {
	return &window{
		title: title,
		child: child,
		style: style,
	}
}

// Render implements the Element interface.
func (w *window) Render(scr uv.Screen, area uv.Rectangle) {
	if area.Dx() < 2 || area.Dy() < 3 {
		return
	}

	// Draw border
	border := uv.RoundedBorder().Style(w.style)
	border.Draw(scr, area)

	// Draw title in the top border
	if w.title != "" && area.Dx() > 4 {
		titleText := " " + w.title + " "
		titleWidth := displaywidth.String(titleText)
		maxWidth := area.Dx() - 2

		if titleWidth > maxWidth {
			// Truncate safely using graphemes
			gr := displaywidth.StringGraphemes(titleText)
			var truncated strings.Builder
			currentWidth := 0
			for gr.Next() {
				w := gr.Width()
				if currentWidth+w > maxWidth {
					break
				}
				truncated.WriteString(string(gr.Value()))
				currentWidth += w
			}
			titleText = truncated.String()
		}

		x := area.Min.X + 2
		y := area.Min.Y

		gr := displaywidth.StringGraphemes(titleText)
		for gr.Next() {
			if x >= area.Max.X-1 {
				break
			}

			grapheme := string(gr.Value())
			width := gr.Width()

			cell := &uv.Cell{
				Content: grapheme,
				Width:   width,
				Style:   w.style,
			}
			scr.SetCell(x, y, cell)
			x += width
		}
	}

	// Render child in inner area
	innerArea := uv.Rect(
		area.Min.X+1,
		area.Min.Y+1,
		area.Dx()-2,
		area.Dy()-2,
	)

	if w.child != nil && innerArea.Dx() > 0 && innerArea.Dy() > 0 {
		w.child.Render(scr, innerArea)
	}
}

// MinSize implements the Element interface.
func (w *window) MinSize(scr uv.Screen) (width, height int) {
	if w.child != nil {
		width, height = w.child.MinSize(scr)
	}

	// Account for border (2) and title
	width += 2
	height += 2

	// Ensure minimum width for title
	titleWidth := displaywidth.String(w.title) + 4
	if titleWidth > width {
		width = titleWidth
	}

	return width, height
}
