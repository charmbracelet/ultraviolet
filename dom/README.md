# DOM Package

The `dom` package provides DOM-inspired primitives for building terminal user interfaces using Ultraviolet. It follows a declarative approach similar to HTML/CSS and [FTXUI](https://github.com/ArthurSonzogni/FTXUI), where UI components are composed as a tree of elements.

## Overview

Elements are the building blocks of a DOM-based TUI. They implement the `Element` interface and can be composed together to create complex layouts. The package provides:

- **Text Elements**: Text, Paragraph
- **Containers**: VBox (vertical), HBox (horizontal)
- **Box Model**: Unified container with borders, padding, scrolling, focus, and selection
- **Layout Helpers**: Separator, Spacer, Flex
- **Interactive Elements**: Button, Input, Checkbox, Window

## Core Concepts

### Element Interface

All DOM elements implement the `Element` interface:

```go
type Element interface {
    Render(scr uv.Screen, area uv.Rectangle)
    MinSize(scr uv.Screen) (width, height int)
}
```

### Box Model

The `Box` type is a unified container that provides common functionality for all elements. It follows a CSS-like box model with:

- **Content**: The child element
- **Padding**: Inner spacing between content and border
- **Border**: Optional border with customizable style
- **Scrolling**: Built-in scroll support (horizontal and vertical)
- **Focus**: Focus state management
- **Selection**: Selection state with customizable style

```go
box := dom.NewBox(content).
    WithBorder(dom.BorderStyleRounded()).
    WithPadding(1).
    WithFocus(true)

// Scroll the content
box.ScrollDown(5)
box.ScrollRight(10)
```

### Text Width Calculation

The package uses [displaywidth](https://github.com/clipperhouse/displaywidth) for accurate text width calculation, ensuring proper rendering of Unicode characters, emojis, and wide characters.

## Basic Usage

```go
import (
    uv "github.com/charmbracelet/ultraviolet"
    "github.com/charmbracelet/ultraviolet/dom"
)

// Create a simple UI
ui := dom.VBox(
    dom.Text("Hello, World!"),
    dom.Separator(),
    dom.HBox(
        dom.Button("OK"),
        dom.Spacer(2, 0),
        dom.Button("Cancel"),
    ),
)

// Render it
ui.Render(screen, area)
```

## Box Model Example

```go
// Create a scrollable box with border and padding
content := dom.VBox(
    dom.Text("Line 1"),
    dom.Text("Line 2"),
    dom.Text("Line 3"),
    // ... more lines
)

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
