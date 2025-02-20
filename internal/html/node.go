package html

import (
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"

	"golang.org/x/net/html"
)

type Node struct {
	node *html.Node
}

func NewNode(tag string, attrs map[string]string, text string) *Node {
	attr := make([]html.Attribute, 0, len(attrs))
	for key, value := range attrs {
		attr = append(attr, html.Attribute{Key: key, Val: value})
	}
	node := &html.Node{
		Type: html.ElementNode,
		Data: tag,
		Attr: attr,
	}
	if len(text) > 0 {
		node.FirstChild = &html.Node{Type: html.TextNode, Data: text}
	}
	return &Node{node: node}
}

func Parse(r io.Reader) (*Node, error) {
	root, err := html.Parse(r)
	if err != nil {
		return nil, fmt.Errorf("parse HTML: %w", err)
	}
	return &Node{root}, nil
}

func (n *Node) Render(w io.Writer) error {
	err := html.Render(w, n.node)
	if err != nil {
		return fmt.Errorf("render HTML: %w", err)
	}
	return nil
}

func (n *Node) RenderString() string {
	sb := &strings.Builder{}
	_ = n.Render(sb)
	return sb.String()
}

func (n *Node) Tag() string {
	return n.node.Data
}

func (n *Node) Text() string {
	if n.node.FirstChild == nil || n.node.FirstChild.Type != html.TextNode {
		return ""
	}
	return n.node.FirstChild.Data
}

func (n *Node) GetAttr(name string) (string, bool) {
	idx := slices.IndexFunc(n.node.Attr, func(attr html.Attribute) bool {
		return attr.Key == name
	})
	if idx < 0 {
		return "", false
	}
	return n.node.Attr[idx].Val, true
}

func (n *Node) SetAttr(name string, value string) {
	n.node.Attr = append(n.node.Attr, html.Attribute{Key: name, Val: value})
}

func (n *Node) DeleteAttr(name string) {
	n.node.Attr = slices.DeleteFunc(n.node.Attr, func(attr html.Attribute) bool {
		return attr.Key == name
	})
}

func (n *Node) Children() []Node {
	children := make([]Node, 0)
	for child := n.node.FirstChild; child != nil; child = child.NextSibling {
		children = append(children, Node{child})
	}
	return children
}

func (n *Node) AppendChild(child *Node) {
	n.node.AppendChild(child.node)
}

func (n *Node) RemoveAll(filters ...NodeFilter) {
	for _, matchingNode := range n.FindAll(filters...) {
		parent := matchingNode.node.Parent
		if parent == nil {
			matchingNode.node = nil
		} else {
			parent.RemoveChild(matchingNode.node)
		}
	}
}

func (n *Node) Find(filters ...NodeFilter) (*Node, error) {
	res := n.FindAll(filters...)
	if len(res) == 0 {
		return nil, errors.New("no HTML nodes matching filters")
	}
	return res[0], nil
}

func (n *Node) FindAll(filters ...NodeFilter) []*Node {
	nodes := make([]*Node, 0)
	if n.eval(filters...) {
		nodes = append(nodes, n)
	}
	for _, child := range n.Children() {
		nodes = append(nodes, child.FindAll(filters...)...)
	}
	return nodes
}

func (n *Node) eval(filters ...NodeFilter) bool {
	for _, filter := range filters {
		if !filter.Eval(n) {
			return false
		}
	}
	return true
}
