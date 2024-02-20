package cborquery

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/fxamacker/cbor/v2"
)

type keyValue struct {
	key   string
	value interface{}
}

// A NodeType is the type of a Node.
type NodeType uint

const (
	// DocumentNode is a document object that, as the root of the document tree,
	// provides access to the entire XML document.
	DocumentNode NodeType = iota
	// ElementNode is an element.
	ElementNode
	// TextNode is the text content of a node.
	TextNode
)

// A Node consists of a NodeType and some Name (tag name for
// element nodes, content for text) and are part of a tree of Nodes.
type Node struct {
	Parent, PrevSibling, NextSibling, FirstChild, LastChild *Node

	Type  NodeType
	Name  string
	value interface{}
	level int
}

// Gets the object value.
func (n *Node) Value() interface{} {
	return n.value
}

// ChildNodes gets all child nodes of the node.
func (n *Node) ChildNodes() []*Node {
	var a []*Node
	for nn := n.FirstChild; nn != nil; nn = nn.NextSibling {
		a = append(a, nn)
	}
	return a
}

// InnerText will gets the value of the node and all its child nodes.
//
// Deprecated: Use Value() to get object value.
func (n *Node) InnerText() string {
	var output func(*strings.Builder, *Node)
	output = func(b *strings.Builder, n *Node) {
		if n.Type == TextNode {
			b.WriteString(fmt.Sprintf("%v", n.value))
			return
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			output(b, child)
		}
	}
	var b strings.Builder
	output(&b, n)
	return b.String()
}

func outputXML(buf *bytes.Buffer, n *Node) {
	if n.Type == TextNode {
		buf.WriteString(fmt.Sprintf("%v", n.value))
		return
	}

	buf.WriteString("<" + n.Name + ">")
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		outputXML(buf, child)
	}
	buf.WriteString("</" + n.Name + ">")
}

// OutputXML prints the XML string.
func (n *Node) OutputXML() string {
	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0"?>`)
	buf.WriteString("<root>")
	for n := n.FirstChild; n != nil; n = n.NextSibling {
		outputXML(&buf, n)
	}
	buf.WriteString("</root>")
	return buf.String()
}

// SelectElement finds the first of child elements with the specified name.
func (n *Node) SelectElement(name string) *Node {
	for nn := n.FirstChild; nn != nil; nn = nn.NextSibling {
		if nn.Name == name {
			return nn
		}
	}
	return nil
}

// LoadURL loads the document from the specified URL.
func LoadURL(url string) (*Node, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return Parse(resp.Body)
}

// Parse CBOR document.
func Parse(r io.Reader) (*Node, error) {
	buf, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var msg interface{}
	if err := cbor.Unmarshal(buf, &msg); err != nil {
		return nil, err
	}
	doc := &Node{Type: DocumentNode}
	err = parseValue(doc, msg, "", 1)
	return doc, err
}

func parseValue(parent *Node, msg interface{}, parentName string, level int) error {
	if parentName == "" {
		parentName = "element"
	}

	switch v := msg.(type) {
	case []interface{}:
		// Array
		for _, vv := range v {
			n := &Node{Name: parentName, Type: ElementNode, level: level, value: vv}
			addNode(parent, n)
			if err := parseValue(n, vv, "", level+1); err != nil {
				return err
			}
		}
	case map[interface{}]interface{}:
		// Object
		// Sort the elements by name as the map access is randomized and this
		// will break at least tests
		elements := make([]keyValue, 0, len(v))
		for k, val := range v {
			key, err := toNodeName(k)
			if err != nil {
				return fmt.Errorf("object key %v: %w", k, err)
			}
			elements = append(elements, keyValue{key, val})
		}
		sort.Slice(elements, func(i, j int) bool {
			return elements[i].key < elements[j].key
		})

		for _, el := range elements {
			key, val := el.key, el.value
			if _, isArray := val.([]interface{}); isArray {
				if err := parseValue(parent, val, key, level+1); err != nil {
					return err
				}
			} else {
				n := &Node{Name: key, Type: ElementNode, level: level, value: val}
				if err := parseValue(n, val, parent.Name, level+1); err != nil {
					return err
				}
				addNode(parent, n)
			}
		}
	default:
		// Elementary types
		n := &Node{Name: fmt.Sprintf("%v", v), Type: TextNode, level: level, value: v}
		addNode(parent, n)
	}
	return nil
}

func addNode(top, n *Node) {
	if n.level == top.level {
		top.NextSibling = n
		n.PrevSibling = top
		n.Parent = top.Parent
		if top.Parent != nil {
			top.Parent.LastChild = n
		}
	} else if n.level > top.level {
		n.Parent = top
		if top.FirstChild == nil {
			top.FirstChild = n
			top.LastChild = n
		} else {
			t := top.LastChild
			t.NextSibling = n
			n.PrevSibling = t
			top.LastChild = n
		}
	}
}
