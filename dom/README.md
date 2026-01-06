# DOM Package

The `dom` package provides a Document Object Model (DOM) implementation for building terminal user interfaces with Ultraviolet. It follows the core concepts of the Web DOM ([MDN Web Docs](https://developer.mozilla.org/en-US/docs/Web/API/Document_Object_Model)), adapted for terminal environments.

## Overview

Like the Web DOM, this package organizes UI as a tree of elements that can be composed declaratively. It follows familiar concepts:

- **Element nodes** form a tree structure (like DOM nodes)
- **Text nodes** for inline content (analogous to DOM Text)
- **Block-level containers** (analogous to HTMLDivElement)  
- **CSS box model** with content, padding, and borders
- **Flexbox-style layout** for containers

This makes the API familiar to developers with HTML/CSS/DOM experience.

## Core Element Types

### TextNode (Inline Element)

Like DOM Text nodes and HTML `<span>`, TextNode is an inline element that flows with content and only breaks on explicit newlines.

```go
text := dom.Text("Hello, World!")
styled := dom.Text("Bold text").WithStyle(uv.Style{Attrs: uv.AttrBold})
```

**Analogous to:**
- DOM: Text node
- HTML: `<span>` or text content
- CSS: `display: inline`

### Box (Block-Level Container)

Like DOM's HTMLDivElement and HTML `<div>`, Box is a block-level container that follows the CSS box model.

```go
box := dom.NewBox(content).
    WithBorder(dom.BorderStyleRounded()).
    WithPadding(1).
    WithFocus(true)

// Scroll the content (like CSS overflow)
box.ScrollDown(5)
box.ScrollRight(10)
```

**Analogous to:**
- DOM: HTMLDivElement
- HTML: `<div>`
- CSS: `display: block` with box model (padding, border)

**Box Model Support:**
- **Content**: The child element
- **Padding**: Inner spacing between content and border
- **Border**: Optional border with customizable style
- **Scrolling**: Built-in overflow support (horizontal and vertical)
- **Focus**: Focus state management
- **Selection**: Selection state with customizable style

## Layout Containers

### VBox (Vertical Flexbox)

Stacks children vertically, similar to CSS flexbox with `flex-direction: column`.

```go
ui := dom.VBox(
    dom.Text("Header"),
    dom.Text("Content"),
    dom.Text("Footer"),
)
```

**Analogous to:**
- CSS: `display: flex; flex-direction: column`

### HBox (Horizontal Flexbox)

Arranges children horizontally, similar to CSS flexbox with `flex-direction: row`.

```go
ui := dom.HBox(
    dom.Text("Left"),
    dom.Text("Center"),
    dom.Text("Right"),
)
```

**Analogous to:**
- CSS: `display: flex; flex-direction: row`

**Flex Behavior:**
- Children with `MinSize` of 0 are flexible (like `flex: 1`)
- Children with fixed `MinSize` use exactly that much space
- Remaining space is distributed equally among flexible children

## Border Styles

```go
dom.BorderStyleNormal()   // Standard box-drawing characters
dom.BorderStyleRounded()  // Rounded corners
dom.BorderStyleDouble()   // Double-line borders
dom.BorderStyleThick()    // Thick line borders
```

## Layout Helpers

```go
dom.Separator()          // Horizontal line (like HTML <hr>)
dom.SeparatorVertical()  // Vertical line
dom.Spacer(width, height) // Fixed-size spacing
dom.Flex()              // Flexible spacing (flex: 1)
dom.Center(child)       // Center child element
```

## Example Usage

### Simple Layout

```go
import (
    uv "github.com/charmbracelet/ultraviolet"
    "github.com/charmbracelet/ultraviolet/dom"
)

// Create a simple UI tree
ui := dom.VBox(
    dom.Text("Hello, World!"),
    dom.Separator(),
    dom.HBox(
        dom.Text("Left"),
        dom.Flex(),
        dom.Text("Right"),
    ),
)

// Render it
ui.Render(screen, area)
```

### Box Model with Scrolling

```go
// Create scrollable content
content := dom.VBox(
    dom.Text("Line 1"),
    dom.Text("Line 2"),
    dom.Text("Line 3"),
    // ... more lines
)

// Wrap in a box with border and padding
box := dom.NewBox(content).
    WithBorder(dom.BorderStyleRounded()).
    WithPadding(1).
    WithFocus(true)

// Render the box
box.Render(screen, area)

// Handle keyboard input for scrolling
switch key {
case "up":
    box.ScrollUp(1)
case "down":
    box.ScrollDown(1)
}
```

### Complex Nested Layout

```go
ui := dom.VBox(
    // Header
    dom.NewBox(
        dom.Text("Application Title"),
    ).WithBorder(dom.BorderStyleNormal()),
    
    // Main content area
    dom.HBox(
        // Left sidebar
        dom.VBox(
            dom.Text("Menu"),
            dom.Separator(),
            dom.Text("Item 1"),
            dom.Text("Item 2"),
        ),
        
        dom.SeparatorVertical(),
        
        // Right content
        dom.NewBox(
            dom.Paragraph("Your content here..."),
        ).WithPadding(1),
    ),
    
    dom.Separator(),
    
    // Footer
    dom.Center(dom.Text("Press q to quit")),
)
```

## Design Philosophy

This package follows core DOM concepts:

1. **Tree Structure**: Elements compose into a tree, just like DOM nodes
2. **Declarative**: Build UIs by describing what you want, not how to render it
3. **Web Standards**: Use familiar concepts (box model, flexbox, inline/block)
4. **Simple**: Two core element types (TextNode and Box) cover most use cases
5. **Composable**: Everything implements Element and can be nested arbitrarily

## Differences from Web DOM

While inspired by the Web DOM, this is a terminal UI library with some differences:

- **No CSS files**: Styling is done through Go code
- **Explicit rendering**: Call `Render()` rather than automatic updates
- **Limited text styling**: Terminal capabilities (foreground/background colors, attributes)
- **No DOM events**: Use terminal events directly
- **Manual focus**: Focus management is explicit

## Text Width Calculation

The package uses [displaywidth](https://github.com/clipperhouse/displaywidth) for accurate text width calculation, ensuring proper rendering of:
- Unicode characters and grapheme clusters
- Emoji and other wide characters
- Combining characters
- Zero-width characters
    box.ScrollDown(1)
}
```

## Elements

### Box

Unified container with borders, padding, scrolling, and state management:

```go
// Basic box
box := dom.NewBox(dom.Text("Content"))

// With border
box.WithBorder(dom.BorderStyleNormal())

// With padding
box.WithPadding(1) // All sides
// Or individual sides
box.PaddingTop = 2
box.PaddingLeft = 4

// With focus
box.WithFocus(true)

// With selection
box.WithSelection(true)

// Scrolling
box.ScrollDown(10)
box.ScrollUp(5)
box.ScrollLeft(3)
box.ScrollRight(7)
```

**Border Styles:**
- `BorderStyleNormal()` - Standard box-drawing characters
- `BorderStyleRounded()` - Rounded corners
- `BorderStyleDouble()` - Double-line borders
- `BorderStyleThick()` - Thick line borders

### Text Elements

#### Text
Renders single or multi-line text:
```go
dom.Text("Hello, World!")
dom.Styled("Styled text", uv.Style{Fg: ansi.Red})
```

#### Paragraph
Wraps text to fit available width:
```go
dom.Paragraph("This is a long paragraph that will wrap...")
dom.ParagraphStyled("Styled paragraph", style)
```

### Containers

#### VBox
Stacks elements vertically:
```go
dom.VBox(
    dom.Text("First"),
    dom.Text("Second"),
    dom.Text("Third"),
)
```

#### HBox
Arranges elements horizontally:
```go
dom.HBox(
    dom.Text("Left"),
    dom.Separator(),
    dom.Text("Right"),
)
```

### Decorators

#### Border
Adds a border around an element:
```go
dom.Border(dom.Text("Bordered text"))
dom.BorderStyled(child, uv.RoundedBorder(), style)
```

#### Padding
Adds padding around an element:
```go
dom.Padding(child, top, right, bottom, left)
dom.PaddingAll(child, 1) // uniform padding
```

#### Center
Centers an element within its area:
```go
dom.Center(dom.Text("Centered"))
```

### Layout Helpers

#### Separator
Creates horizontal or vertical separator lines:
```go
dom.Separator()                              // horizontal
dom.SeparatorVertical()                      // vertical
dom.SeparatorStyled("─", style)              // custom
```

#### Spacer
Creates fixed-size empty space:
```go
dom.Spacer(width, height)
```

#### Flex
Creates flexible space that expands to fill available space:
```go
dom.HBox(
    dom.Text("Left"),
    dom.Flex(),
    dom.Text("Right"),
)
```

### Interactive Elements

#### Button
Creates a clickable button:
```go
dom.Button("Click Me")
dom.ButtonStyled("Custom", style)
```

#### Input
Creates a text input field:
```go
dom.Input(20, "Placeholder...")
dom.InputStyled(20, "Placeholder...", style)
```

#### Checkbox
Creates a checkbox:
```go
dom.Checkbox("Option 1", true)  // checked
dom.Checkbox("Option 2", false) // unchecked
```

#### Window
Creates a titled window/panel:
```go
dom.Window("Title", content)
dom.WindowStyled("Title", content, style)
```

## Complex Example

```go
ui := dom.Window("Application",
    dom.VBox(
        dom.PaddingAll(
            dom.VBox(
                dom.Styled("Welcome!", uv.Style{Attrs: uv.AttrBold}),
                dom.Spacer(0, 1),
                dom.Paragraph("This is a demonstration of the DOM package."),
                dom.Spacer(0, 1),
                dom.Separator(),
                dom.Spacer(0, 1),
                dom.HBox(
                    dom.VBox(
                        dom.Text("Options:"),
                        dom.Checkbox("Option 1", true),
                        dom.Checkbox("Option 2", false),
                    ),
                    dom.Spacer(2, 0),
                    dom.SeparatorVertical(),
                    dom.Spacer(2, 0),
                    dom.VBox(
                        dom.Text("Actions:"),
                        dom.HBox(
                            dom.Button("Submit"),
                            dom.Spacer(1, 0),
                            dom.Button("Cancel"),
                        ),
                    ),
                ),
                dom.Spacer(0, 1),
                dom.Separator(),
                dom.Spacer(0, 1),
                dom.Center(
                    dom.Border(
                        dom.PaddingAll(dom.Text("Centered Box"), 1),
                    ),
                ),
            ),
            1,
        ),
    ),
)
```

## Layout Algorithm

### VBox and HBox

Containers distribute space among their children:

1. Calculate minimum size requirements for each child
2. Children with fixed sizes (MinSize > 0) get their required space
3. Remaining space is distributed equally among flexible children (MinSize = 0)
4. Each child is rendered in its allocated area

### Flex and Spacer

- **Flex**: Has MinSize = 0, expands to fill available space
- **Spacer**: Has fixed MinSize, takes up exactly that much space

## Design Philosophy

The DOM package is designed with these principles:

- **Declarative**: Describe what you want, not how to build it
- **Composable**: Elements can be nested arbitrarily
- **Flexible**: Mix fixed and flexible sizing
- **Intuitive**: API similar to HTML and CSS concepts
- **Type-safe**: Strong typing with Go interfaces

## Performance

The package is optimized for terminal rendering:

- Efficient layout calculations
- Minimal screen updates
- Proper Unicode handling with displaywidth
- No unnecessary allocations in hot paths

## See Also

- [Ultraviolet](https://github.com/charmbracelet/ultraviolet) - The underlying TUI library
- [FTXUI](https://github.com/ArthurSonzogni/FTXUI) - API inspiration
- [displaywidth](https://github.com/clipperhouse/displaywidth) - Text width calculation
