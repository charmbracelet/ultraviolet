package tv

import (
	"math"
)

// LayoutDirection represents the direction of layout flow.
type LayoutDirection int

const (
	// Horizontal layouts flow left to right.
	Horizontal LayoutDirection = iota
	// Vertical layouts flow top to bottom.
	Vertical
)

// SizeConstraint represents how a layout item should be sized.
type SizeConstraint interface {
	// Calculate returns the size in pixels given the available space.
	Calculate(available int) int
	// IsFlexible returns true if this constraint can grow/shrink.
	IsFlexible() bool
}

// FixedSize represents a fixed size constraint.
type FixedSize int

func (f FixedSize) Calculate(available int) int {
	return int(f)
}

func (f FixedSize) IsFlexible() bool {
	return false
}

// RatioSize represents a proportional size constraint (0.0 to 1.0).
type RatioSize float64

func (r RatioSize) Calculate(available int) int {
	if r < 0 {
		r = 0
	} else if r > 1 {
		r = 1
	}
	return int(float64(available) * float64(r))
}

func (r RatioSize) IsFlexible() bool {
	return false
}

// FlexSize represents a flexible size constraint with optional grow/shrink factors.
type FlexSize struct {
	// Grow factor (default 1). Higher values take more space.
	Grow float64
	// Shrink factor (default 1). Higher values shrink more when space is limited.
	Shrink float64
	// Basis is the initial size before growing/shrinking.
	Basis SizeConstraint
}

func (f FlexSize) Calculate(available int) int {
	if f.Basis != nil {
		return f.Basis.Calculate(available)
	}
	return 0
}

func (f FlexSize) IsFlexible() bool {
	return true
}

// MinMaxSize wraps another constraint with minimum and maximum bounds.
type MinMaxSize struct {
	Constraint SizeConstraint
	Min        int
	Max        int
}

func (m MinMaxSize) Calculate(available int) int {
	size := m.Constraint.Calculate(available)
	if m.Min > 0 && size < m.Min {
		size = m.Min
	}
	if m.Max > 0 && size > m.Max {
		size = m.Max
	}
	return size
}

func (m MinMaxSize) IsFlexible() bool {
	return m.Constraint.IsFlexible()
}

// LayoutItem represents a single item in a layout.
type LayoutItem struct {
	// Width constraint for this item.
	Width SizeConstraint
	// Height constraint for this item.
	Height SizeConstraint
	// Margin around the item.
	Margin Spacing
	// Padding inside the item.
	Padding Spacing
}

// Spacing represents spacing values for all four sides.
type Spacing struct {
	Top, Right, Bottom, Left int
}

// Uniform creates spacing with the same value on all sides.
func Uniform(size int) Spacing {
	return Spacing{Top: size, Right: size, Bottom: size, Left: size}
}

// Horizontal creates spacing with horizontal and vertical values.
func HorizontalVertical(horizontal, vertical int) Spacing {
	return Spacing{Top: vertical, Right: horizontal, Bottom: vertical, Left: horizontal}
}

// Layout represents a container that arranges child items.
type Layout struct {
	// Direction of the layout flow.
	Direction LayoutDirection
	// Items in this layout.
	Items []LayoutItem
	// Gap between items.
	Gap int
	// Alignment of items in the cross axis.
	CrossAxisAlignment CrossAxisAlignment
	// Alignment of items in the main axis.
	MainAxisAlignment MainAxisAlignment
	// Whether items should wrap to new lines/columns.
	Wrap bool
}

// CrossAxisAlignment represents alignment perpendicular to the main axis.
type CrossAxisAlignment int

const (
	// CrossAxisStart aligns items to the start of the cross axis.
	CrossAxisStart CrossAxisAlignment = iota
	// CrossAxisCenter centers items on the cross axis.
	CrossAxisCenter
	// CrossAxisEnd aligns items to the end of the cross axis.
	CrossAxisEnd
	// CrossAxisStretch stretches items to fill the cross axis.
	CrossAxisStretch
)

// MainAxisAlignment represents alignment along the main axis.
type MainAxisAlignment int

const (
	// MainAxisStart aligns items to the start of the main axis.
	MainAxisStart MainAxisAlignment = iota
	// MainAxisCenter centers items on the main axis.
	MainAxisCenter
	// MainAxisEnd aligns items to the end of the main axis.
	MainAxisEnd
	// MainAxisSpaceBetween distributes items with space between them.
	MainAxisSpaceBetween
	// MainAxisSpaceAround distributes items with space around them.
	MainAxisSpaceAround
	// MainAxisSpaceEvenly distributes items with even spacing.
	MainAxisSpaceEvenly
)

// LayoutResult represents the result of a layout calculation.
type LayoutResult struct {
	// Areas for each item in the layout.
	Areas []Rectangle
	// Total size used by the layout.
	Size Rectangle
}

// Calculate computes the layout for the given available area.
func (l *Layout) Calculate(area Rectangle) LayoutResult {
	if len(l.Items) == 0 {
		return LayoutResult{Size: area}
	}

	switch l.Direction {
	case Horizontal:
		return l.calculateHorizontal(area)
	case Vertical:
		return l.calculateVertical(area)
	default:
		return LayoutResult{Size: area}
	}
}

func (l *Layout) calculateHorizontal(area Rectangle) LayoutResult {
	availableWidth := area.Dx()
	availableHeight := area.Dy()

	// Calculate total gap space
	totalGap := l.Gap * (len(l.Items) - 1)
	if totalGap < 0 {
		totalGap = 0
	}

	// Calculate widths for each item
	widths := l.calculateSizes(availableWidth-totalGap, true)

	// Calculate heights for each item
	heights := make([]int, len(l.Items))
	for i, item := range l.Items {
		if item.Height != nil {
			heights[i] = item.Height.Calculate(availableHeight)
		} else {
			heights[i] = availableHeight
		}
	}

	// Position items
	areas := make([]Rectangle, len(l.Items))
	x := area.Min.X

	for i, item := range l.Items {
		width := widths[i]
		height := heights[i]

		// Apply margin
		marginWidth := item.Margin.Left + item.Margin.Right
		marginHeight := item.Margin.Top + item.Margin.Bottom

		// Calculate y position based on cross-axis alignment
		y := l.calculateCrossAxisPosition(area.Min.Y, availableHeight, height+marginHeight)

		// Create the area including margin
		itemArea := Rect(x+item.Margin.Left, y+item.Margin.Top,
			width-marginWidth, height-marginHeight)
		areas[i] = itemArea

		x += width + l.Gap
	}

	// Apply main axis alignment
	l.applyMainAxisAlignment(areas, area, true)

	return LayoutResult{
		Areas: areas,
		Size:  area,
	}
}

func (l *Layout) calculateVertical(area Rectangle) LayoutResult {
	availableWidth := area.Dx()
	availableHeight := area.Dy()

	// Calculate total gap space
	totalGap := l.Gap * (len(l.Items) - 1)
	if totalGap < 0 {
		totalGap = 0
	}

	// Calculate heights for each item
	heights := l.calculateSizes(availableHeight-totalGap, false)

	// Calculate widths for each item
	widths := make([]int, len(l.Items))
	for i, item := range l.Items {
		if item.Width != nil {
			widths[i] = item.Width.Calculate(availableWidth)
		} else {
			widths[i] = availableWidth
		}
	}

	// Position items
	areas := make([]Rectangle, len(l.Items))
	y := area.Min.Y

	for i, item := range l.Items {
		width := widths[i]
		height := heights[i]

		// Apply margin
		marginWidth := item.Margin.Left + item.Margin.Right
		marginHeight := item.Margin.Top + item.Margin.Bottom

		// Calculate x position based on cross-axis alignment
		x := l.calculateCrossAxisPosition(area.Min.X, availableWidth, width+marginWidth)

		// Create the area including margin
		itemArea := Rect(x+item.Margin.Left, y+item.Margin.Top,
			width-marginWidth, height-marginHeight)
		areas[i] = itemArea

		y += height + l.Gap
	}

	// Apply main axis alignment
	l.applyMainAxisAlignment(areas, area, false)

	return LayoutResult{
		Areas: areas,
		Size:  area,
	}
}

func (l *Layout) calculateSizes(available int, isWidth bool) []int {
	sizes := make([]int, len(l.Items))
	flexItems := make([]int, 0)
	usedSpace := 0

	// First pass: calculate fixed and ratio sizes
	for i, item := range l.Items {
		var constraint SizeConstraint
		if isWidth {
			constraint = item.Width
		} else {
			constraint = item.Height
		}

		if constraint == nil {
			// No constraint means flexible
			flexItems = append(flexItems, i)
			continue
		}

		if constraint.IsFlexible() {
			flexItems = append(flexItems, i)
			if flex, ok := constraint.(FlexSize); ok && flex.Basis != nil {
				sizes[i] = flex.Basis.Calculate(available)
				usedSpace += sizes[i]
			}
		} else {
			sizes[i] = constraint.Calculate(available)
			usedSpace += sizes[i]
		}
	}

	// Second pass: distribute remaining space to flexible items
	remainingSpace := available - usedSpace
	if remainingSpace > 0 && len(flexItems) > 0 {
		l.distributeFlex(sizes, flexItems, remainingSpace, isWidth)
	}

	return sizes
}

func (l *Layout) distributeFlex(sizes []int, flexItems []int, remainingSpace int, isWidth bool) {
	totalGrow := 0.0

	// Calculate total grow factor
	for _, i := range flexItems {
		var constraint SizeConstraint
		if isWidth {
			constraint = l.Items[i].Width
		} else {
			constraint = l.Items[i].Height
		}

		if constraint == nil {
			totalGrow += 1.0 // Default grow factor
		} else if flex, ok := constraint.(FlexSize); ok {
			if flex.Grow > 0 {
				totalGrow += flex.Grow
			} else {
				totalGrow += 1.0
			}
		} else {
			totalGrow += 1.0
		}
	}

	// Distribute space proportionally
	for _, i := range flexItems {
		var constraint SizeConstraint
		if isWidth {
			constraint = l.Items[i].Width
		} else {
			constraint = l.Items[i].Height
		}

		grow := 1.0
		if constraint != nil {
			if flex, ok := constraint.(FlexSize); ok && flex.Grow > 0 {
				grow = flex.Grow
			}
		}

		additionalSpace := int(float64(remainingSpace) * (grow / totalGrow))
		sizes[i] += additionalSpace
	}
}

func (l *Layout) calculateCrossAxisPosition(start, available, itemSize int) int {
	switch l.CrossAxisAlignment {
	case CrossAxisStart:
		return start
	case CrossAxisCenter:
		return start + (available-itemSize)/2
	case CrossAxisEnd:
		return start + available - itemSize
	case CrossAxisStretch:
		return start
	default:
		return start
	}
}

func (l *Layout) applyMainAxisAlignment(areas []Rectangle, container Rectangle, isHorizontal bool) {
	if l.MainAxisAlignment == MainAxisStart {
		return // Already positioned correctly
	}

	var totalItemSize, availableSize int
	if isHorizontal {
		for _, area := range areas {
			totalItemSize += area.Dx()
		}
		totalItemSize += l.Gap * (len(areas) - 1)
		availableSize = container.Dx()
	} else {
		for _, area := range areas {
			totalItemSize += area.Dy()
		}
		totalItemSize += l.Gap * (len(areas) - 1)
		availableSize = container.Dy()
	}

	extraSpace := availableSize - totalItemSize
	if extraSpace <= 0 {
		return
	}

	var offset int
	var spacing int

	switch l.MainAxisAlignment {
	case MainAxisCenter:
		offset = extraSpace / 2
	case MainAxisEnd:
		offset = extraSpace
	case MainAxisSpaceBetween:
		if len(areas) > 1 {
			spacing = extraSpace / (len(areas) - 1)
		}
	case MainAxisSpaceAround:
		spacing = extraSpace / len(areas)
		offset = spacing / 2
	case MainAxisSpaceEvenly:
		spacing = extraSpace / (len(areas) + 1)
		offset = spacing
	}

	// Apply the adjustments
	for i := range areas {
		if isHorizontal {
			areas[i] = areas[i].Add(Pos(offset+i*spacing, 0))
		} else {
			areas[i] = areas[i].Add(Pos(0, offset+i*spacing))
		}
	}
}

// Grid represents a grid layout system.
type Grid struct {
	// Number of columns in the grid.
	Columns int
	// Number of rows in the grid (0 means auto).
	Rows int
	// Gap between grid items.
	ColumnGap int
	RowGap    int
	// Items in the grid.
	Items []GridItem
}

// GridItem represents an item in a grid layout.
type GridItem struct {
	// Grid position (1-based, 0 means auto).
	Column, Row int
	// How many columns/rows this item spans.
	ColumnSpan, RowSpan int
	// Layout item properties.
	LayoutItem
}

// Calculate computes the grid layout for the given available area.
func (g *Grid) Calculate(area Rectangle) LayoutResult {
	if len(g.Items) == 0 || g.Columns <= 0 {
		return LayoutResult{Size: area}
	}

	// Determine number of rows
	rows := g.Rows
	if rows <= 0 {
		rows = int(math.Ceil(float64(len(g.Items)) / float64(g.Columns)))
	}

	// Calculate cell dimensions
	totalColumnGap := g.ColumnGap * (g.Columns - 1)
	totalRowGap := g.RowGap * (rows - 1)

	cellWidth := (area.Dx() - totalColumnGap) / g.Columns
	cellHeight := (area.Dy() - totalRowGap) / rows

	areas := make([]Rectangle, len(g.Items))

	for i, item := range g.Items {
		// Determine grid position
		col := item.Column
		row := item.Row

		if col <= 0 || row <= 0 {
			// Auto-placement
			autoCol := i % g.Columns
			autoRow := i / g.Columns
			if col <= 0 {
				col = autoCol + 1
			}
			if row <= 0 {
				row = autoRow + 1
			}
		}

		// Convert to 0-based
		col--
		row--

		// Calculate spans
		colSpan := item.ColumnSpan
		rowSpan := item.RowSpan
		if colSpan <= 0 {
			colSpan = 1
		}
		if rowSpan <= 0 {
			rowSpan = 1
		}

		// Calculate position and size
		x := area.Min.X + col*(cellWidth+g.ColumnGap)
		y := area.Min.Y + row*(cellHeight+g.RowGap)

		width := cellWidth*colSpan + g.ColumnGap*(colSpan-1)
		height := cellHeight*rowSpan + g.RowGap*(rowSpan-1)

		// Apply margin
		x += item.Margin.Left
		y += item.Margin.Top
		width -= item.Margin.Left + item.Margin.Right
		height -= item.Margin.Top + item.Margin.Bottom

		areas[i] = Rect(x, y, width, height)
	}

	return LayoutResult{
		Areas: areas,
		Size:  area,
	}
}

// Helper functions for creating common layouts

// NewHorizontalLayout creates a new horizontal layout.
func NewHorizontalLayout(items ...LayoutItem) *Layout {
	return &Layout{
		Direction: Horizontal,
		Items:     items,
	}
}

// NewVerticalLayout creates a new vertical layout.
func NewVerticalLayout(items ...LayoutItem) *Layout {
	return &Layout{
		Direction: Vertical,
		Items:     items,
	}
}

// NewFlexLayout creates a flexible layout with the given direction.
func NewFlexLayout(direction LayoutDirection, items ...LayoutItem) *Layout {
	return &Layout{
		Direction:          direction,
		Items:              items,
		CrossAxisAlignment: CrossAxisStretch,
	}
}

// NewGrid creates a new grid layout.
func NewGrid(columns int, items ...GridItem) *Grid {
	return &Grid{
		Columns: columns,
		Items:   items,
	}
}

// Item creates a layout item with the given constraints.
func Item(width, height SizeConstraint) LayoutItem {
	return LayoutItem{
		Width:  width,
		Height: height,
	}
}

// FlexItem creates a flexible layout item.
func FlexItem(grow, shrink float64, basis SizeConstraint) LayoutItem {
	return LayoutItem{
		Width: FlexSize{
			Grow:   grow,
			Shrink: shrink,
			Basis:  basis,
		},
		Height: FlexSize{
			Grow:   grow,
			Shrink: shrink,
			Basis:  basis,
		},
	}
}

// GridItemAt creates a grid item at the specified position.
func GridItemAt(col, row int, item LayoutItem) GridItem {
	return GridItem{
		Column:     col,
		Row:        row,
		ColumnSpan: 1,
		RowSpan:    1,
		LayoutItem: item,
	}
}

// GridItemSpan creates a grid item that spans multiple cells.
func GridItemSpan(col, row, colSpan, rowSpan int, item LayoutItem) GridItem {
	return GridItem{
		Column:     col,
		Row:        row,
		ColumnSpan: colSpan,
		RowSpan:    rowSpan,
		LayoutItem: item,
	}
}

// Convenience functions for backward compatibility

// SplitVertical splits a rectangle into two rectangles vertically using the new layout system.
func SplitVertical(r Rectangle, ratio float64) (Rectangle, Rectangle) {
	layout := NewHorizontalLayout(
		Item(RatioSize(ratio), nil),
		Item(RatioSize(1-ratio), nil),
	)
	result := layout.Calculate(r)
	if len(result.Areas) >= 2 {
		return result.Areas[0], result.Areas[1]
	}
	return r, Rect(0, 0, 0, 0)
}

// SplitHorizontal splits a rectangle into two rectangles horizontally using the new layout system.
func SplitHorizontal(r Rectangle, ratio float64) (Rectangle, Rectangle) {
	layout := NewVerticalLayout(
		Item(nil, RatioSize(ratio)),
		Item(nil, RatioSize(1-ratio)),
	)
	result := layout.Calculate(r)
	if len(result.Areas) >= 2 {
		return result.Areas[0], result.Areas[1]
	}
	return r, Rect(0, 0, 0, 0)
}

