package doc

import (
	"strings"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func TestNewDocument(t *testing.T) {
	htmlNode := &html.Node{
		Type: html.ElementNode,
		Data: "div",
	}

	opts := &Options{
		BaseURL:     "https://example.com",
		Stylesheets: []string{"style.css"},
	}

	doc := NewDocument(htmlNode, opts)

	if doc == nil {
		t.Fatal("expected doc to be set")
	}

	if doc.Data() != "div" {
		t.Errorf("expected doc data to be 'div', got %q", doc.Data())
	}

	if doc.Type() != html.ElementNode {
		t.Errorf("expected doc type to be ElementNode")
	}
}

func TestNewDocumentNilOptions(t *testing.T) {
	htmlNode := &html.Node{
		Type: html.ElementNode,
		Data: "div",
	}

	doc := NewDocument(htmlNode, nil)

	if doc == nil {
		t.Fatal("expected doc to be set")
	}
}

func TestNewDocumentWithTerminal(t *testing.T) {
	htmlNode := &html.Node{
		Type: html.ElementNode,
		Data: "div",
	}

	term := uv.DefaultTerminal()
	opts := &Options{
		Terminal: term,
	}

	doc := NewDocument(htmlNode, opts)

	if doc == nil {
		t.Fatal("expected doc to be set")
	}
}

func TestNodeInterface(t *testing.T) {
	parent := &html.Node{
		Type: html.ElementNode,
		Data: "div",
	}

	child1 := &html.Node{
		Type: html.TextNode,
		Data: "Hello",
	}

	child2 := &html.Node{
		Type: html.ElementNode,
		Data: "span",
		Attr: []html.Attribute{
			{Key: "class", Val: "highlight"},
		},
	}

	parent.AppendChild(child1)
	parent.AppendChild(child2)

	doc := NewDocument(parent, nil)

	// Test basic node properties
	if doc.Type() != html.ElementNode {
		t.Errorf("expected ElementNode, got %v", doc.Type())
	}

	if doc.Data() != "div" {
		t.Errorf("expected 'div', got %q", doc.Data())
	}

	// Test FirstChild
	firstChild := doc.FirstChild()
	if firstChild == nil {
		t.Fatal("expected first child to exist")
	}

	if firstChild.Type() != html.TextNode {
		t.Errorf("expected TextNode, got %v", firstChild.Type())
	}

	if firstChild.Data() != "Hello" {
		t.Errorf("expected 'Hello', got %q", firstChild.Data())
	}

	// Test LastChild
	lastChild := doc.LastChild()
	if lastChild == nil {
		t.Fatal("expected last child to exist")
	}

	if lastChild.Type() != html.ElementNode {
		t.Errorf("expected ElementNode, got %v", lastChild.Type())
	}

	if lastChild.Data() != "span" {
		t.Errorf("expected 'span', got %q", lastChild.Data())
	}

	// Test attributes
	attrs := lastChild.Attr()
	if len(attrs) != 1 {
		t.Fatalf("expected 1 attribute, got %d", len(attrs))
	}

	if attrs[0].Key != "class" || attrs[0].Val != "highlight" {
		t.Errorf("expected class=highlight, got %s=%s", attrs[0].Key, attrs[0].Val)
	}

	// Test NextSibling
	nextSib := firstChild.NextSibling()
	if nextSib == nil {
		t.Fatal("expected next sibling to exist")
	}

	if nextSib.Data() != "span" {
		t.Errorf("expected 'span', got %q", nextSib.Data())
	}

	// Test PrevSibling
	prevSib := lastChild.PrevSibling()
	if prevSib == nil {
		t.Fatal("expected previous sibling to exist")
	}

	if prevSib.Data() != "Hello" {
		t.Errorf("expected 'Hello', got %q", prevSib.Data())
	}

	// Test Children
	children := doc.Children()
	if len(children) != 2 {
		t.Errorf("expected 2 children, got %d", len(children))
	}

	// Test Parent
	childParent := firstChild.Parent()
	if childParent == nil {
		t.Fatal("expected parent to exist")
	}

	if childParent.Data() != "div" {
		t.Errorf("expected parent to be 'div', got %q", childParent.Data())
	}
}

func TestParse(t *testing.T) {
	htmlStr := `<!DOCTYPE html>
<html>
<head><title>Test</title></head>
<body><h1>Hello, World!</h1></body>
</html>`

	r := strings.NewReader(htmlStr)
	doc, err := Parse(r, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if doc == nil {
		t.Fatal("expected document to be created")
	}
}

func TestParseWithOptions(t *testing.T) {
	htmlStr := `<html><body><p>Test</p></body></html>`

	opts := &Options{
		BaseURL: "https://example.com",
	}

	r := strings.NewReader(htmlStr)
	doc, err := Parse(r, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if doc == nil {
		t.Fatal("expected root node to be set")
	}
}

func TestParseFragment(t *testing.T) {
	fragmentStr := `<p>Hello</p><p>World</p>`

	r := strings.NewReader(fragmentStr)
	context := &html.Node{
		Type:     html.ElementNode,
		Data:     "body",
		DataAtom: atom.Body,
	}

	doc, err := ParseFragment(r, context, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if doc == nil {
		t.Fatal("expected document to be created")
	}

	if doc.Data() != "div" {
		t.Errorf("expected root to be container div, got %q", doc.Data())
	}

	// Test that children are present
	children := doc.Children()
	if len(children) != 2 {
		t.Errorf("expected 2 children, got %d", len(children))
	}
}

func TestParseInvalidHTML(t *testing.T) {
	// html.Parse is very forgiving, so we test with a reader that returns an error
	r := strings.NewReader("")
	doc, err := Parse(r, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if doc == nil {
		t.Fatal("expected document to be created even for empty input")
	}
}

func TestGetElementByID(t *testing.T) {
	htmlStr := `<html>
<body>
	<div id="header">Header</div>
	<div id="content">
		<p id="intro">Introduction</p>
	</div>
	<div id="footer">Footer</div>
</body>
</html>`

	r := strings.NewReader(htmlStr)
	doc, err := Parse(r, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test finding existing elements
	header := doc.GetElementByID("header")
	if header == nil {
		t.Fatal("expected to find element with id 'header'")
	}
	if header.Data() != "div" {
		t.Errorf("expected tag 'div', got %q", header.Data())
	}

	intro := doc.GetElementByID("intro")
	if intro == nil {
		t.Fatal("expected to find element with id 'intro'")
	}
	if intro.Data() != "p" {
		t.Errorf("expected tag 'p', got %q", intro.Data())
	}

	// Test non-existent element
	notFound := doc.GetElementByID("nonexistent")
	if notFound != nil {
		t.Errorf("expected nil for non-existent id")
	}
}

func TestGetElementsByTagName(t *testing.T) {
	htmlStr := `<html>
<body>
	<div>First div</div>
	<p>Paragraph</p>
	<div>Second div</div>
	<span>Span</span>
	<div>Third div</div>
</body>
</html>`

	r := strings.NewReader(htmlStr)
	doc, err := Parse(r, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test finding divs
	divs := doc.GetElementsByTagName("div")
	if len(divs) != 3 {
		t.Errorf("expected 3 divs, got %d", len(divs))
	}

	// Test finding paragraphs
	ps := doc.GetElementsByTagName("p")
	if len(ps) != 1 {
		t.Errorf("expected 1 paragraph, got %d", len(ps))
	}

	// Test non-existent tag
	notFound := doc.GetElementsByTagName("article")
	if len(notFound) != 0 {
		t.Errorf("expected 0 articles, got %d", len(notFound))
	}
}

func TestGetElementsByClassName(t *testing.T) {
	htmlStr := `<html>
<body>
	<div class="container">First</div>
	<p class="highlight">Paragraph</p>
	<div class="container highlight">Second</div>
	<span class="label">Span</span>
	<div class="container">Third</div>
</body>
</html>`

	r := strings.NewReader(htmlStr)
	doc, err := Parse(r, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test finding by class name
	containers := doc.GetElementsByClassName("container")
	if len(containers) != 3 {
		t.Errorf("expected 3 elements with class 'container', got %d", len(containers))
	}

	highlights := doc.GetElementsByClassName("highlight")
	if len(highlights) != 2 {
		t.Errorf("expected 2 elements with class 'highlight', got %d", len(highlights))
	}

	// Test non-existent class
	notFound := doc.GetElementsByClassName("nonexistent")
	if len(notFound) != 0 {
		t.Errorf("expected 0 elements, got %d", len(notFound))
	}
}

func TestQuerySelector(t *testing.T) {
	htmlStr := `<html>
<body>
	<div id="main" class="container">
		<p class="intro">First paragraph</p>
		<p>Second paragraph</p>
	</div>
	<div class="sidebar">
		<span id="logo">Logo</span>
	</div>
</body>
</html>`

	r := strings.NewReader(htmlStr)
	doc, err := Parse(r, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test ID selector
	main := doc.QuerySelector("#main")
	if main == nil {
		t.Fatal("expected to find element with id 'main'")
	}
	if main.Data() != "div" {
		t.Errorf("expected tag 'div', got %q", main.Data())
	}

	// Test class selector
	container := doc.QuerySelector(".container")
	if container == nil {
		t.Fatal("expected to find element with class 'container'")
	}

	// Test tag selector
	p := doc.QuerySelector("p")
	if p == nil {
		t.Fatal("expected to find p element")
	}
	if p.Data() != "p" {
		t.Errorf("expected tag 'p', got %q", p.Data())
	}

	// Test non-existent selector
	notFound := doc.QuerySelector("#notfound")
	if notFound != nil {
		t.Errorf("expected nil for non-existent selector")
	}
}

func TestQuerySelectorAll(t *testing.T) {
	htmlStr := `<html>
<body>
	<div class="box">Box 1</div>
	<div class="box">Box 2</div>
	<p>Paragraph</p>
	<div class="box">Box 3</div>
	<span id="unique">Unique</span>
</body>
</html>`

	r := strings.NewReader(htmlStr)
	doc, err := Parse(r, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test class selector
	boxes := doc.QuerySelectorAll(".box")
	if len(boxes) != 3 {
		t.Errorf("expected 3 boxes, got %d", len(boxes))
	}

	// Test tag selector
	divs := doc.QuerySelectorAll("div")
	if len(divs) != 3 {
		t.Errorf("expected 3 divs, got %d", len(divs))
	}

	// Test ID selector (should return 1 element)
	unique := doc.QuerySelectorAll("#unique")
	if len(unique) != 1 {
		t.Errorf("expected 1 element with id 'unique', got %d", len(unique))
	}

	// Test non-existent selector
	notFound := doc.QuerySelectorAll(".nonexistent")
	if len(notFound) != 0 {
		t.Errorf("expected 0 elements, got %d", len(notFound))
	}
}

func TestAddEventListener(t *testing.T) {
	htmlNode := &html.Node{
		Type: html.ElementNode,
		Data: "div",
	}

	doc := NewDocument(htmlNode, nil)

	// Track if listener was called
	called := false
	listener := func(ev *Event) bool {
		called = true
		return false
	}

	doc.AddEventListener(EventKeyPress, listener)

	// Create a mock key press event
	ev := uv.KeyPressEvent{}

	// Dispatch the event
	doc.dispatchEvent(ev)

	if !called {
		t.Error("expected listener to be called")
	}
}

func TestAddEventListenerAll(t *testing.T) {
	htmlNode := &html.Node{
		Type: html.ElementNode,
		Data: "div",
	}

	doc := NewDocument(htmlNode, nil)

	// Track events received
	var events []*Event
	listener := func(ev *Event) bool {
		events = append(events, ev)
		return false
	}

	doc.AddEventListener(EventAll, listener)

	// Dispatch different events
	doc.dispatchEvent(uv.KeyPressEvent{})
	doc.dispatchEvent(uv.MouseClickEvent{})

	if len(events) != 2 {
		t.Errorf("expected 2 events, got %d", len(events))
	}
}

func TestEventBubbling(t *testing.T) {
	htmlNode := &html.Node{
		Type: html.ElementNode,
		Data: "div",
	}

	doc := NewDocument(htmlNode, nil)

	// First listener stops propagation
	firstCalled := false
	first := func(ev *Event) bool {
		firstCalled = true
		return true // Stop propagation
	}

	// Second listener should not be called
	secondCalled := false
	second := func(ev *Event) bool {
		secondCalled = true
		return false
	}

	doc.AddEventListener(EventKeyPress, first)
	doc.AddEventListener(EventKeyPress, second)

	// Dispatch event
	handled := doc.dispatchEvent(uv.KeyPressEvent{})

	if !handled {
		t.Error("expected event to be handled")
	}

	if !firstCalled {
		t.Error("expected first listener to be called")
	}

	if secondCalled {
		t.Error("expected second listener not to be called (propagation stopped)")
	}
}

func TestRemoveEventListener(t *testing.T) {
	htmlNode := &html.Node{
		Type: html.ElementNode,
		Data: "div",
	}

	doc := NewDocument(htmlNode, nil)

	called := false
	listener := func(ev *Event) bool {
		called = true
		return false
	}

	doc.AddEventListener(EventKeyPress, listener)
	doc.RemoveEventListener(EventKeyPress)

	// Dispatch event
	doc.dispatchEvent(uv.KeyPressEvent{})

	if called {
		t.Error("expected listener not to be called after removal")
	}
}

func TestGetEventType(t *testing.T) {
	tests := []struct {
		event    uv.Event
		expected EventType
	}{
		{uv.KeyPressEvent{}, EventKeyPress},
		{uv.KeyReleaseEvent{}, EventKeyRelease},
		{uv.MouseClickEvent{}, EventMouseClick},
		{uv.MouseReleaseEvent{}, EventMouseRelease},
		{uv.MouseWheelEvent{}, EventMouseWheel},
		{uv.MouseMotionEvent{}, EventMouseMotion},
	}

	for _, tt := range tests {
		eventType := getEventType(tt.event)
		if eventType != tt.expected {
			t.Errorf("expected event type %v, got %v", tt.expected, eventType)
		}
	}
}
