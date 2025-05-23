# Advanced Layout System for TV

This document describes the new advanced layout system for the TV terminal library. The new system provides flexible, powerful layout capabilities while maintaining backward compatibility with the existing simple split functions.

## Overview

The new layout system supports:

- **Flexible Layouts**: Flexbox-style layouts with grow/shrink factors
- **Ratio-based Layouts**: Proportional sizing using ratios
- **Fixed Layouts**: Absolute pixel-based sizing
- **Grid Layouts**: CSS Grid-style 2D layouts
- **Spacing and Alignment**: Margins, padding, gaps, and alignment options
- **Responsive Design**: Layouts that adapt to different screen sizes

## Core Concepts

### Size Constraints

Size constraints define how layout items should be sized:

```go
// Fixed size - always the same number of pixels
width := tv.FixedSize(100)

// Ratio size - percentage of available space (0.0 to 1.0)
width := tv.RatioSize(0.3) // 30% of available width

// Flexible size - grows/shrinks based on available space
width := tv.FlexSize{
    Grow:   1,    // How much to grow (default 1)
    Shrink: 1,    // How much to shrink (default 1)
    Basis:  nil,  // Initial size before flex (optional)
}

// Min/max constraints - bounds for other constraints
width := tv.MinMaxSize{
    Constraint: tv.FlexSize{Grow: 1},
    Min:        50,  // Minimum 50 pixels
    Max:        200, // Maximum 200 pixels
}
```

### Layout Items

Layout items represent individual elements in a layout:

```go
item := tv.LayoutItem{
    Width:   tv.FixedSize(100),
    Height:  tv.FixedSize(50),
    Margin:  tv.Uniform(5),           // 5px margin on all sides
    Padding: tv.HorizontalVertical(10, 5), // 10px horizontal, 5px vertical
}

// Helper function for creating items
item := tv.Item(tv.FixedSize(100), tv.FixedSize(50))
```

### Spacing

Spacing can be applied as margins and padding:

```go
// Uniform spacing (same on all sides)
spacing := tv.Uniform(10)

// Different horizontal and vertical spacing
spacing := tv.HorizontalVertical(20, 15) // 20px horizontal, 15px vertical

// Custom spacing for each side
spacing := tv.Spacing{
    Top:    10,
    Right:  20,
    Bottom: 10,
    Left:   20,
}
```

## Linear Layouts

Linear layouts arrange items in a single direction (horizontal or vertical).

### Basic Linear Layout

```go
// Horizontal layout
layout := tv.NewHorizontalLayout(
    tv.Item(tv.FixedSize(100), tv.FixedSize(50)),
    tv.Item(tv.FixedSize(150), tv.FixedSize(50)),
)

// Vertical layout
layout := tv.NewVerticalLayout(
    tv.Item(tv.FixedSize(100), tv.FixedSize(50)),
    tv.Item(tv.FixedSize(100), tv.FixedSize(75)),
)

// Calculate layout for a given area
area := tv.Rect(0, 0, 300, 100)
result := layout.Calculate(area)

// Access individual item areas
for i, itemArea := range result.Areas {
    fmt.Printf("Item %d: %+v\n", i, itemArea)
}
```

### Layout with Gaps and Alignment

```go
layout := &tv.Layout{
    Direction: tv.Horizontal,
    Items: []tv.LayoutItem{
        tv.Item(tv.FixedSize(100), tv.FixedSize(50)),
        tv.Item(tv.FixedSize(100), tv.FixedSize(50)),
    },
    Gap: 20, // 20px gap between items
    
    // Cross-axis alignment (perpendicular to direction)
    CrossAxisAlignment: tv.CrossAxisCenter,
    
    // Main-axis alignment (along the direction)
    MainAxisAlignment: tv.MainAxisSpaceBetween,
}
```

### Alignment Options

**Cross-axis alignment** (perpendicular to layout direction):
- `CrossAxisStart`: Align to start
- `CrossAxisCenter`: Center alignment
- `CrossAxisEnd`: Align to end
- `CrossAxisStretch`: Stretch to fill

**Main-axis alignment** (along layout direction):
- `MainAxisStart`: Align to start
- `MainAxisCenter`: Center alignment
- `MainAxisEnd`: Align to end
- `MainAxisSpaceBetween`: Space between items
- `MainAxisSpaceAround`: Space around items
- `MainAxisSpaceEvenly`: Even spacing

## Flexible Layouts

Flexible layouts automatically distribute available space among items.

### Basic Flex Layout

```go
layout := tv.NewFlexLayout(tv.Horizontal,
    tv.Item(tv.FixedSize(100), nil),        // Fixed 100px width
    tv.FlexItem(1, 1, nil),                 // Takes remaining space
    tv.Item(tv.FixedSize(100), nil),        // Fixed 100px width
)
```

### Advanced Flex Layout

```go
layout := &tv.Layout{
    Direction: tv.Horizontal,
    Items: []tv.LayoutItem{
        // Fixed sidebar
        tv.Item(tv.FixedSize(200), nil),
        
        // Flexible content area
        tv.Item(tv.FlexSize{Grow: 2}, nil), // Grows twice as much
        
        // Flexible sidebar
        tv.Item(tv.FlexSize{Grow: 1}, nil), // Grows normally
    },
}
```

### Flex with Basis

```go
// Flex item with initial size
flexItem := tv.Item(
    tv.FlexSize{
        Grow:   1,
        Shrink: 1,
        Basis:  tv.FixedSize(200), // Start at 200px, then grow/shrink
    },
    nil,
)
```

## Grid Layouts

Grid layouts arrange items in a 2D grid structure.

### Basic Grid

```go
grid := tv.NewGrid(3, // 3 columns
    tv.GridItemAt(1, 1, tv.Item(nil, nil)), // Column 1, Row 1
    tv.GridItemAt(2, 1, tv.Item(nil, nil)), // Column 2, Row 1
    tv.GridItemAt(3, 1, tv.Item(nil, nil)), // Column 3, Row 1
    tv.GridItemAt(1, 2, tv.Item(nil, nil)), // Column 1, Row 2
)

area := tv.Rect(0, 0, 300, 200)
result := grid.Calculate(area)
```

### Grid with Gaps

```go
grid := &tv.Grid{
    Columns:   3,
    ColumnGap: 10, // 10px gap between columns
    RowGap:    20, // 20px gap between rows
    Items: []tv.GridItem{
        tv.GridItemAt(1, 1, tv.Item(nil, nil)),
        tv.GridItemAt(2, 1, tv.Item(nil, nil)),
        tv.GridItemAt(3, 1, tv.Item(nil, nil)),
    },
}
```

### Grid with Spanning

```go
grid := tv.NewGrid(4,
    // Item spanning 2 columns and 1 row
    tv.GridItemSpan(1, 1, 2, 1, tv.Item(nil, nil)),
    
    // Regular items
    tv.GridItemAt(3, 1, tv.Item(nil, nil)),
    tv.GridItemAt(4, 1, tv.Item(nil, nil)),
    
    // Item spanning 1 column and 2 rows
    tv.GridItemSpan(1, 2, 1, 2, tv.Item(nil, nil)),
)
```

### Auto-placement Grid

```go
// Items without explicit positions are auto-placed
grid := tv.NewGrid(3,
    tv.GridItemAt(0, 0, tv.Item(nil, nil)), // Auto-placed
    tv.GridItemAt(0, 0, tv.Item(nil, nil)), // Auto-placed
    tv.GridItemAt(0, 0, tv.Item(nil, nil)), // Auto-placed
)
```

## Practical Examples

### Dashboard Layout

```go
func createDashboard(area tv.Rectangle) tv.LayoutResult {
    // Main vertical layout: header, body, footer
    mainLayout := tv.NewVerticalLayout(
        tv.Item(nil, tv.FixedSize(60)),  // Header
        tv.Item(nil, tv.FlexSize{Grow: 1}), // Body
        tv.Item(nil, tv.FixedSize(30)),  // Footer
    )
    mainResult := mainLayout.Calculate(area)
    
    // Body horizontal layout: sidebar, content
    bodyLayout := tv.NewHorizontalLayout(
        tv.Item(tv.FixedSize(250), nil), // Sidebar
        tv.Item(tv.FlexSize{Grow: 1}, nil), // Content
    )
    bodyLayout.Gap = 10
    bodyResult := bodyLayout.Calculate(mainResult.Areas[1])
    
    // Content grid for widgets
    contentGrid := tv.NewGrid(3,
        tv.GridItemSpan(1, 1, 2, 1, tv.Item(nil, tv.FixedSize(200))), // Large widget
        tv.GridItemAt(3, 1, tv.Item(nil, tv.FixedSize(200))),
        tv.GridItemAt(1, 2, tv.Item(nil, tv.FixedSize(150))),
        tv.GridItemAt(2, 2, tv.Item(nil, tv.FixedSize(150))),
        tv.GridItemAt(3, 2, tv.Item(nil, tv.FixedSize(150))),
    )
    contentGrid.ColumnGap = 15
    contentGrid.RowGap = 15
    
    return contentGrid.Calculate(bodyResult.Areas[1])
}
```

### Responsive Layout

```go
func createResponsiveLayout(area tv.Rectangle) tv.LayoutResult {
    width := area.Dx()
    
    if width < 600 {
        // Mobile layout: single column
        return tv.NewVerticalLayout(
            tv.Item(nil, tv.FixedSize(100)), // Header
            tv.Item(nil, tv.FlexSize{Grow: 1}), // Content
            tv.Item(nil, tv.FixedSize(50)),  // Footer
        ).Calculate(area)
    } else if width < 1200 {
        // Tablet layout: sidebar below header
        mainLayout := tv.NewVerticalLayout(
            tv.Item(nil, tv.FixedSize(80)),  // Header
            tv.Item(nil, tv.FlexSize{Grow: 1}), // Body
            tv.Item(nil, tv.FixedSize(50)),  // Footer
        )
        mainResult := mainLayout.Calculate(area)
        
        bodyLayout := tv.NewVerticalLayout(
            tv.Item(nil, tv.FixedSize(200)), // Sidebar
            tv.Item(nil, tv.FlexSize{Grow: 1}), // Content
        )
        bodyLayout.Gap = 10
        bodyResult := bodyLayout.Calculate(mainResult.Areas[1])
        
        return tv.LayoutResult{
            Areas: append([]tv.Rectangle{mainResult.Areas[0]}, 
                   append(bodyResult.Areas, mainResult.Areas[2])...),
            Size: area,
        }
    } else {
        // Desktop layout: sidebar on left
        return createDashboard(area)
    }
}
```

## Migration from Old System

The new system maintains backward compatibility:

```go
// Old way
left, right := tv.SplitVertical(area, 0.3)

// New way (equivalent)
layout := tv.NewHorizontalLayout(
    tv.Item(tv.RatioSize(0.3), nil),
    tv.Item(tv.RatioSize(0.7), nil),
)
result := layout.Calculate(area)
left, right := result.Areas[0], result.Areas[1]

// Or using the updated functions (same API)
left, right := tv.SplitVertical(area, 0.3) // Uses new system internally
```

## Performance Considerations

- Layout calculations are O(n) where n is the number of items
- Grid layouts are O(n) for positioning, O(1) for cell size calculation
- Flex layouts require two passes: fixed sizing, then flex distribution
- Cache layout results when the area hasn't changed
- Use fixed sizes when possible for better performance

## Best Practices

1. **Use semantic layouts**: Choose the layout type that best represents your UI structure
2. **Prefer constraints over absolute sizes**: Use ratios and flex for responsive designs
3. **Group related items**: Use nested layouts for complex UIs
4. **Consider performance**: Cache layout results and avoid unnecessary recalculations
5. **Test different screen sizes**: Ensure your layouts work across various terminal sizes
6. **Use margins and padding consistently**: Establish a spacing system for your application

## Helper Functions

The system provides several helper functions for common patterns:

```go
// Create items
item := tv.Item(width, height)
flexItem := tv.FlexItem(grow, shrink, basis)

// Create layouts
hLayout := tv.NewHorizontalLayout(items...)
vLayout := tv.NewVerticalLayout(items...)
flexLayout := tv.NewFlexLayout(direction, items...)

// Create grids
grid := tv.NewGrid(columns, items...)

// Create grid items
gridItem := tv.GridItemAt(col, row, item)
spanItem := tv.GridItemSpan(col, row, colSpan, rowSpan, item)

// Create spacing
uniform := tv.Uniform(size)
hvSpacing := tv.HorizontalVertical(h, v)
```

This new layout system provides the flexibility and power needed for modern terminal applications while maintaining the simplicity and performance that TV is known for.