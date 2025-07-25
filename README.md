# Ultraviolet

<img width="400" alt="Charm Ultraviolet" src="https://github.com/user-attachments/assets/3484e4b0-3741-4e8c-bebf-9ea51f5bb49c" />

<p>
    <a href="https://github.com/charmbracelet/ultraviolet/releases"><img src="https://img.shields.io/github/release/charmbracelet/ultraviolet.svg" alt="Latest Release"></a>
    <a href="https://pkg.go.dev/github.com/charmbracelet/ultraviolet?tab=doc"><img src="https://godoc.org/github.com/charmbracelet/ultraviolet?status.svg" alt="GoDoc"></a>
    <a href="https://github.com/charmbracelet/ultraviolet/actions"><img src="https://github.com/charmbracelet/ultraviolet/actions/workflows/build.yml/badge.svg" alt="Build Status"></a>
</p>

> [!CAUTION]
> This project is in very early development and may change significantly at any moment. Expect no API guarantees as of now.

Ultraviolet is a set of primitives for manipulating terminal emulators with a
focus on terminal user interfaces (TUIs). It provides a set of tools and
abstractions for interactive terminal applications that can handle user input
and display dynamic, cell-based content.

Ultraviolet is the secret power behind the wonder and majesty of the Charm’s terminal user interface libraries. It is the result of many years of research, development, collaboration and ingenuity.

_So mote it be_.

## Inspiration

Ultraviolet is inspired by the need for a modern, efficient, and ergonomic way
to build terminal applications in Go. It draws inspiration from various
libraries like [ncurses](https://invisible-island.net/ncurses/). It was born
out of the desire to create a library that simplifies terminal manipulation
while providing powerful abstractions for rendering and input handling in
a platform-agnostic way.

### Is it a replacement for Bubble Tea and Lip Gloss?

Simply put, no. Ultraviolet is not a replacement for existing libraries like
[Bubble Tea](https://github.com/charmbracelet/bubbletea) and [Lip Gloss](https://github.com/charmbracelet/lipgloss). Instead, it serves as a
foundation for both of these libraries and others like them, providing the
underlying primitives and abstractions needed to build text-based user
interfaces and applications.

### Features

Ultraviolet is designed to be a versatile and powerful library for building
text-based user interfaces with a focus on terminal emulators. It provides a
set of primitive building blocks that can be used to create interactive
applications, handle user input, and render dynamic content in a terminal
environment with the most efficient and ergonomic way possible.

This includes features like:

- Cell-based rendering model for efficient terminal content updates
- Dynamic content updates using a cell-based diffing algorithm
- Interactive input handling with support for multiple input sources
- Cross-platform compatibility with a consistent API
- Extensible architecture for building custom terminal user interfaces

## Usage

To use Ultraviolet, you can get it with:

```bash
go get github.com/charmbracelet/ultraviolet
```

Then import it in your Go code:

```go
import "github.com/charmbracelet/ultraviolet"
```

## Tutorial

You can find a simple tutorial on how to create a UV application that displays
"Hello, World!" on the screen in the [TUTORIAL.md](./TUTORIAL.md) file.

## Whatcha think?

We’d love to hear your thoughts on this project. Feel free to drop us a note!

- [Twitter](https://twitter.com/charmcli)
- [The Fediverse](https://mastodon.social/@charmcli)
- [Discord](https://charm.sh/chat)

## License

[MIT](./LICENSE)

---

Part of [Charm](https://charm.land).

<a href="https://charm.sh/"><img alt="The Charm logo" width="400" src="https://stuff.charm.sh/charm-banner-next.jpg" /></a>

Charm热爱开源 • Charm loves open source • نحنُ نحب المصادر المفتوحة
