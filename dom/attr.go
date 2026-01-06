package dom

import (
	"image/color"
)

// Attr represents an attribute on an Element.
// See: https://developer.mozilla.org/en-US/docs/Web/API/Attr
type Attr struct {
	// Name is the attribute name.
	Name string

	// Value is the attribute value.
	Value string
}

// NewAttr creates a new attribute with the given name and value.
func NewAttr(name, value string) *Attr {
	return &Attr{
		Name:  name,
		Value: value,
	}
}

// NamedNodeMap represents a collection of Attr nodes.
// In Web DOM, this is used to store element attributes.
type NamedNodeMap struct {
	attrs []*Attr
}

// NewNamedNodeMap creates a new empty NamedNodeMap.
func NewNamedNodeMap() *NamedNodeMap {
	return &NamedNodeMap{
		attrs: make([]*Attr, 0),
	}
}

// GetNamedItem returns an attribute by name, or nil if not found.
func (m *NamedNodeMap) GetNamedItem(name string) *Attr {
	for _, attr := range m.attrs {
		if attr.Name == name {
			return attr
		}
	}
	return nil
}

// SetNamedItem adds or replaces an attribute.
// Returns the old attribute if one existed, or nil.
func (m *NamedNodeMap) SetNamedItem(attr *Attr) *Attr {
	if attr == nil {
		return nil
	}

	for i, a := range m.attrs {
		if a.Name == attr.Name {
			old := m.attrs[i]
			m.attrs[i] = attr
			return old
		}
	}

	m.attrs = append(m.attrs, attr)
	return nil
}

// RemoveNamedItem removes an attribute by name.
// Returns the removed attribute, or nil if not found.
func (m *NamedNodeMap) RemoveNamedItem(name string) *Attr {
	for i, attr := range m.attrs {
		if attr.Name == name {
			removed := m.attrs[i]
			m.attrs = append(m.attrs[:i], m.attrs[i+1:]...)
			return removed
		}
	}
	return nil
}

// Item returns the attribute at the specified index, or nil if out of bounds.
func (m *NamedNodeMap) Item(index int) *Attr {
	if index < 0 || index >= len(m.attrs) {
		return nil
	}
	return m.attrs[index]
}

// Length returns the number of attributes.
func (m *NamedNodeMap) Length() int {
	return len(m.attrs)
}

// Clone creates a copy of the NamedNodeMap.
func (m *NamedNodeMap) Clone() *NamedNodeMap {
	clone := NewNamedNodeMap()
	for _, attr := range m.attrs {
		clone.attrs = append(clone.attrs, &Attr{
			Name:  attr.Name,
			Value: attr.Value,
		})
	}
	return clone
}

// Style represents inline styles that can be applied to an element.
// This is a simplified version of Web DOM's CSSStyleDeclaration.
type Style struct {
	Foreground color.Color
	Background color.Color
	Bold       bool
	Italic     bool
	Underline  bool
	Reverse    bool
}

// NewStyle creates a new empty style.
func NewStyle() *Style {
	return &Style{}
}

// Clone creates a copy of the style.
func (s *Style) Clone() *Style {
	return &Style{
		Foreground: s.Foreground,
		Background: s.Background,
		Bold:       s.Bold,
		Italic:     s.Italic,
		Underline:  s.Underline,
		Reverse:    s.Reverse,
	}
}
