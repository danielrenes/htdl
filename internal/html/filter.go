package html

import (
	"slices"
	"strings"
)

type NodeFilter interface {
	Eval(node *Node) bool
}

type NodeFilterFunc func(node *Node) bool

func (f NodeFilterFunc) Eval(node *Node) bool {
	return f(node)
}

func Not(filter NodeFilter) NodeFilter {
	return NodeFilterFunc(func(node *Node) bool {
		return !filter.Eval(node)
	})
}

func IsTag(tag string) NodeFilter {
	return NodeFilterFunc(func(node *Node) bool {
		return node.Tag() == tag
	})
}

func HasID(id string) NodeFilter {
	return HasAttr("id", id)
}

func HasClass(class string) NodeFilter {
	return HasAttrFunc("class", func(s string) bool {
		classes := strings.Split(s, " ")
		return slices.Contains(classes, class)
	})
}

func HasAttr(name, value string) NodeFilter {
	return NodeFilterFunc(func(node *Node) bool {
		v, ok := node.GetAttr(name)
		if !ok {
			return false
		}
		return v == value
	})
}

func HasAttrFunc(name string, valueFunc func(string) bool) NodeFilter {
	return NodeFilterFunc(func(node *Node) bool {
		v, ok := node.GetAttr(name)
		if !ok {
			return false
		}
		return valueFunc(v)
	})
}
