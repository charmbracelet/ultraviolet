# HTML/CSS Box Model Implementation

This document describes the box model implementation in the `doc` package.

## Overview

The box model implementation follows the CSS 2.1 specification for block and inline layout as described in:
https://developer.mozilla.org/en-US/docs/Web/CSS/Guides/Display/Block_and_inline_layout

## Implementation Status

### ✅ Completed Features

#### 1. Box Model Spacing Properties

Added support for margin, padding, and border properties in `ComputedStyle`:

```go
type ComputedStyle struct {
    // Box model spacing (in cells)
    MarginTop, MarginRight, MarginBottom, MarginLeft       int
    PaddingTop, PaddingRight, PaddingBottom, PaddingLeft   int
    BorderTop, BorderRight, BorderBottom, BorderLeft       int
    // ... other properties
}
```

#### 2. CSS Property Parsing

Implemented parsing for box model properties from inline styles:

- **Individual properties**: `margin-top`, `margin-right`, `margin-bottom`, `margin-left`
- **Shorthand syntax**: `margin: 10px` (all sides), `margin: 5px 10px` (vertical horizontal), etc.
- **Same for**: `padding-*` and `border-*` properties
- **Unit support**: Accepts values with `px`, `em`, `rem`, `%` units (stripped for terminal cells)

Examples:
```html
<div style="margin: 10px; padding: 5px; border: 1px">Content</div>
<div style="margin: 10px 20px; padding: 5px 10px 15px">Content</div>
<div style="margin-top: 5px; padding-left: 10px">Content</div>
```

#### 3. Layout Calculations

Implemented box model layout calculations in `renderer.go`:

**Helper Functions:**
- `calculateContentWidth()` - Calculates content width from available width minus spacing
- `calculateTotalWidth()` - Calculates total width including all spacing
- `calculateTotalHeight()` - Calculates total height including all spacing  
- `applyBoxModelSpacing()` - Computes ContentRect, PaddingRect, BorderRect, MarginRect

**Layout Logic:**
- `layoutBlockBox()` - Updated to account for margin, border, padding when laying out block elements
- Children are positioned within the parent's content area
- Vertical margins are applied between stacked block boxes
- Content width is calculated accounting for horizontal spacing

#### 4. Display Types

Added `inline-block` display type:

```go
const (
    DisplayBlock       DisplayType = "block"
    DisplayInline      DisplayType = "inline"
    DisplayInlineBlock DisplayType = "inline-block"  // NEW
    DisplayNone        DisplayType = "none"
    DisplayFlex        DisplayType = "flex"
)
```

Inline-block boxes:
- Flow inline like inline boxes
- Establish block formatting context internally
- Treated as inline-level boxes in the box tree (`IsInline()` returns true)

#### 5. Box Tree Structure

Added `InlineBlockBox` type to the `BoxType` enum:

```go
const (
    BlockBox BoxType = iota
    InlineBox
    InlineBlockBox           // NEW
    AnonymousBlockBox
    AnonymousInlineBox
)
```

Updated box tree generation in `buildBoxTree()` to handle inline-block display values.

#### 6. Layout Box Structure

Enhanced `LayoutBox` to track different box model rectangles:

```go
type LayoutBox struct {
    Rect        uv.Rectangle  // Total box including margin
    ContentRect uv.Rectangle  // Content area only
    PaddingRect uv.Rectangle  // Content + padding
    BorderRect  uv.Rectangle  // Content + padding + border
    // ...
}
```

This allows proper rendering of borders, backgrounds, and content with correct spacing.

### 🚧 Not Yet Implemented

The following features are defined but not fully implemented:

1. **Margin Collapsing** - Adjacent vertical margins should collapse per CSS rules
2. **Border Styles** - Only border width is supported, not styles (solid, dashed, etc.)
3. **Border Colors** - Border colors are not yet rendered
4. **Box Sizing** - `box-sizing: border-box` property not implemented
5. **Background Rendering** - Backgrounds should extend through padding and border areas
6. **Inline Layout Spacing** - Inline elements don't yet fully respect padding/border
7. **Positioning** - `position`, `top`, `left`, etc. properties not implemented
8. **Float** - `float` and `clear` properties not implemented

## Usage Examples

### Basic Box Model

```html
<div style="margin: 10px; padding: 5px; border: 1px">
    This div has 10px margin on all sides, 5px padding, and 1px border.
</div>
```

### Different Spacing Per Side

```html
<div style="margin: 10px 20px 15px 25px; padding: 5px 10px">
    Margin: 10px top, 20px right, 15px bottom, 25px left
    Padding: 5px top/bottom, 10px left/right
</div>
```

### Inline-Block Elements

```html
<div>
    <span style="display: inline-block; margin: 5px; padding: 10px">
        This span flows inline but can have block-level spacing.
    </span>
</div>
```

## Testing

Comprehensive tests are provided in `box_model_spacing_test.go`:

- `TestParseSpacingValue` - Tests parsing of spacing values with units
- `TestParseSpacingShorthand` - Tests CSS shorthand syntax (1-4 values)
- `TestApplyStylePropertyMargin` - Tests margin property parsing
- `TestApplyStylePropertyPadding` - Tests padding property parsing
- `TestApplyStylePropertyBorder` - Tests border property parsing
- `TestDisplayInlineBlock` - Tests inline-block display type
- `TestCalculateContentWidth` - Tests width calculations with spacing
- `TestCalculateTotalWidth` - Tests total width calculations

Run tests with:
```bash
go test -v -run TestParseSpacing
```

## Implementation Notes

### Terminal Cells vs Pixels

All spacing values are in terminal cells, not pixels. CSS units like `px`, `em`, `rem`, and `%` are stripped during parsing:

- `margin: 5px` → 5 cells
- `padding: 10em` → 10 cells
- `border: 2rem` → 2 cells

### Negative Values

Negative spacing values are clamped to 0:

```go
if val < 0 {
    return 0
}
```

### Box Model Calculation

The box model follows the standard CSS box model:

```
+-- Margin (transparent) --+
|  +-- Border ------------+|
|  |  +-- Padding -----+  ||
|  |  | Content        |  ||
|  |  +----------------+  ||
|  +----------------------+|
+-------------------------+
```

Total width = margin-left + border-left + padding-left + content-width + padding-right + border-right + margin-right

Total height = margin-top + border-top + padding-top + content-height + padding-bottom + border-bottom + margin-bottom

## Future Enhancements

1. Implement margin collapsing for adjacent block boxes
2. Add border rendering with styles and colors
3. Implement `box-sizing: border-box` property
4. Support percentage-based spacing (relative to parent width)
5. Add min-width, max-width, min-height, max-height properties
6. Implement overflow handling (hidden, scroll, auto)
7. Add positioning (relative, absolute, fixed)
8. Implement float and clear properties

## References

- [MDN: Block and Inline Layout](https://developer.mozilla.org/en-US/docs/Web/CSS/Guides/Display/Block_and_inline_layout)
- [CSS 2.1 Box Model](https://www.w3.org/TR/CSS21/box.html)
- [CSS 2.1 Visual Formatting Model](https://www.w3.org/TR/CSS21/visuren.html)
