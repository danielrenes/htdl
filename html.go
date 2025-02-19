package main

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"slices"
	"strings"

	"golang.org/x/net/html"
)

type HTMLNode struct {
	node *html.Node
}

func NewHTMLNode(tag string, attrs map[string]string, text string) *HTMLNode {
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
	return &HTMLNode{node: node}
}

func ParseHTML(r io.Reader) (*HTMLNode, error) {
	root, err := html.Parse(r)
	if err != nil {
		return nil, fmt.Errorf("parse HTML: %w", err)
	}
	return &HTMLNode{root}, nil
}

func (n *HTMLNode) Render(w io.Writer) error {
	err := html.Render(w, n.node)
	if err != nil {
		return fmt.Errorf("render HTML: %w", err)
	}
	return nil
}

func (n *HTMLNode) RenderString() string {
	sb := &strings.Builder{}
	_ = n.Render(sb)
	return sb.String()
}

func (n *HTMLNode) Tag() string {
	return n.node.Data
}

func (n *HTMLNode) Text() string {
	if n.node.FirstChild == nil || n.node.FirstChild.Type != html.TextNode {
		return ""
	}
	return n.node.FirstChild.Data
}

type HTMLNodeFilter interface {
	Eval(*HTMLNode) bool
}

type HTMLNodeFilterFunc func(*HTMLNode) bool

func (f HTMLNodeFilterFunc) Eval(node *HTMLNode) bool {
	return f(node)
}

func (n *HTMLNode) eval(filters ...HTMLNodeFilter) bool {
	for _, filter := range filters {
		if !filter.Eval(n) {
			return false
		}
	}
	return true
}

func IsTag(tag string) HTMLNodeFilter {
	return HTMLNodeFilterFunc(func(node *HTMLNode) bool {
		return node.Tag() == tag
	})
}

func HasID(id string) HTMLNodeFilter {
	return HasAttr("id", id)
}

func HasClass(class string) HTMLNodeFilter {
	return HasAttrFunc("class", func(s string) bool {
		classes := strings.Split(s, " ")
		return slices.Contains(classes, class)
	})
}

func HasAttr(name, value string) HTMLNodeFilter {
	return HTMLNodeFilterFunc(func(node *HTMLNode) bool {
		v, ok := node.GetAttr(name)
		if !ok {
			return false
		}
		return v == value
	})
}

func HasAttrFunc(name string, valueFunc func(string) bool) HTMLNodeFilter {
	return HTMLNodeFilterFunc(func(node *HTMLNode) bool {
		v, ok := node.GetAttr(name)
		if !ok {
			return false
		}
		return valueFunc(v)
	})
}

func (n *HTMLNode) GetAttr(name string) (string, bool) {
	idx := slices.IndexFunc(n.node.Attr, func(attr html.Attribute) bool {
		return attr.Key == name
	})
	if idx < 0 {
		return "", false
	}
	return n.node.Attr[idx].Val, true
}

func (n *HTMLNode) SetAttr(name string, value string) {
	n.node.Attr = append(n.node.Attr, html.Attribute{Key: name, Val: value})
}

func (n *HTMLNode) DeleteAttr(name string) {
	n.node.Attr = slices.DeleteFunc(n.node.Attr, func(attr html.Attribute) bool {
		return attr.Key == name
	})
}

func (n *HTMLNode) Children() []HTMLNode {
	children := make([]HTMLNode, 0)
	for child := n.node.FirstChild; child != nil; child = child.NextSibling {
		children = append(children, HTMLNode{child})
	}
	return children
}

func (n *HTMLNode) ResolveLinks(baseURL *url.URL) error {
	resolveLink := func(tag string, attr string) error {
		for _, n := range n.FindAll(IsTag(tag)) {
			if v, ok := n.GetAttr(attr); ok {
				v, err := ResolveLink(baseURL, v)
				if err != nil {
					return err
				}
				if strings.HasPrefix(v, fmt.Sprintf("%s#", baseURL)) {
					v = strings.TrimPrefix(v, baseURL.String())
				}
				n.DeleteAttr(attr)
				n.SetAttr(attr, v)
			}
		}
		return nil
	}
	if err := resolveLink("link", "href"); err != nil {
		return err
	}
	if err := resolveLink("a", "href"); err != nil {
		return err
	}
	if err := resolveLink("script", "src"); err != nil {
		return err
	}
	if err := resolveLink("img", "src"); err != nil {
		return err
	}
	return nil
}

func (n *HTMLNode) Find(filters ...HTMLNodeFilter) (*HTMLNode, error) {
	res := n.FindAll(filters...)
	if len(res) == 0 {
		return nil, errors.New("no HTML nodes matching filters")
	}
	return res[0], nil
}

func (n *HTMLNode) FindAll(filters ...HTMLNodeFilter) []*HTMLNode {
	nodes := make([]*HTMLNode, 0)
	if n.eval(filters...) {
		nodes = append(nodes, n)
	}
	for _, child := range n.Children() {
		nodes = append(nodes, child.FindAll(filters...)...)
	}
	return nodes
}

func (n *HTMLNode) AddStyles(styles map[string]string) {
	style, ok := n.GetAttr("style")
	if ok {
		style = strings.TrimSpace(style)
		style = strings.TrimSuffix(style, ";")
		if len(style) != 0 {
			style = fmt.Sprintf("%s;", style)
		}
	}
	for k, v := range styles {
		style = fmt.Sprintf("%s%s: %s;", style, k, v)
	}
	n.DeleteAttr("style")
	n.SetAttr("style", style)
}

func (n *HTMLNode) Add(child *HTMLNode) {
	n.node.AppendChild(child.node)
}

func (n *HTMLNode) RemoveAll(filters ...HTMLNodeFilter) {
	for _, matchingNode := range n.FindAll(filters...) {
		parent := matchingNode.node.Parent
		if parent == nil {
			matchingNode.node = nil
		} else {
			parent.RemoveChild(matchingNode.node)
		}
	}
}
