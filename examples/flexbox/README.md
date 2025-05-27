# Flexbox Layout Demo

This example demonstrates the comprehensive flexbox capabilities of the TV layout system.

## Features Demonstrated

### 1. Basic Horizontal Flex
- Three items with equal flex grow factors (1:1:1)
- Shows how items automatically distribute available space

### 2. Proportional Flex Growth
- Items with different grow factors (1:2:3)
- Demonstrates how grow factors control space distribution

### 3. Fixed + Flex Layout
- Common pattern: fixed sidebars with flexible content area
- Shows mixing fixed and flexible sizing

### 4. Flex with Basis
- Items with initial size (basis) that then grow
- Demonstrates how basis affects final sizing

### 5. Vertical Flex Layout
- Vertical flexbox with different grow factors
- Shows flexbox works in both directions

### 6. Cross-Axis Alignment
- Horizontal flex with center cross-axis alignment
- Items of different heights aligned to center

### 7. Main-Axis Alignment
- Fixed-size items with space-between alignment
- Shows how items are distributed along main axis

### 8. Space Around Alignment
- Fixed-size items with space-around alignment
- Equal space around each item

### 9. Complex Nested Layout
- Multiple nested flexbox containers
- Demonstrates building complex layouts

### 10. Margin and Padding Demo
- Flex items with various margin and padding configurations
- Shows how spacing affects layout

## Controls

- **← →** or **h l**: Navigate between demos
- **Space**: Next demo
- **1-9**: Jump to specific demo
- **r**: Reset to first demo
- **q** or **Ctrl+C**: Quit

## Key Concepts Demonstrated

### Flex Grow
```go
tv.FlexSize{Grow: 2} // Takes twice as much space as Grow: 1
```

### Flex with Basis
```go
tv.FlexSize{
    Grow: 1,
    Basis: tv.FixedSize(100), // Start at 100px, then grow
}
```

### Cross-Axis Alignment
```go
layout.CrossAxisAlignment = tv.CrossAxisCenter // Center items perpendicular to main axis
```

### Main-Axis Alignment
```go
layout.MainAxisAlignment = tv.MainAxisSpaceBetween // Distribute space between items
```

### Spacing
```go
item.Margin = tv.Uniform(5)                    // 5px on all sides
item.Margin = tv.HorizontalVertical(10, 5)     // 10px horizontal, 5px vertical
item.Margin = tv.Spacing{Top: 2, Right: 8, Bottom: 2, Left: 8} // Custom spacing
```

## Running the Demo

```bash
cd examples/flexbox
go run main.go
```

The demo is fully interactive and responsive - resize your terminal to see how the layouts adapt to different screen sizes.

## Layout System Benefits

1. **Intuitive**: Similar to CSS Flexbox, familiar to web developers
2. **Powerful**: Handles complex layouts with simple declarations
3. **Responsive**: Automatically adapts to different screen sizes
4. **Composable**: Nest layouts to build complex UIs
5. **Performance**: Efficient O(n) layout calculations