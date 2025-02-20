package html_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/danielrenes/bee"
	"github.com/danielrenes/htdl/internal/html"
)

func TestParse(t *testing.T) {
	bee := bee.New(t)
	s := `<div class="container"><img src="logo.png"/><p>Some text</p></div>`
	root, err := html.Parse(strings.NewReader(s))
	bee.Nil(err)
	div, err := root.Find(html.IsTag("div"))
	bee.Nil(err)
	bee.Equal(div.Tag(), "div")
	attrEqual(bee, div, "class", "container")
	bee.Equal(div.Text(), "")
	bee.Equal(len(div.Children()), 2)
	img := div.Children()[0]
	bee.Equal(img.Tag(), "img")
	attrEqual(bee, img, "src", "logo.png")
	bee.Equal(img.Text(), "")
	bee.Equal(len(img.Children()), 0)
	p := div.Children()[1]
	bee.Equal(p.Tag(), "p")
	bee.Equal(p.Text(), "Some text")
	bee.Equal(len(p.Children()), 1)
}

func TestFind(t *testing.T) {
	bee := bee.New(t)
	s := `<div><a href="a.img"/><a href="b.img"/></div>`
	div, err := html.Parse(strings.NewReader(s))
	bee.Nil(err)
	a, err := div.Find(html.IsTag("a"), html.HasAttr("href", "a.img"))
	bee.Nil(err)
	bee.Equal(a.Tag(), "a")
	attrEqual(bee, a, "href", "a.img")
}

func TestFindAll(t *testing.T) {
	bee := bee.New(t)
	s := `<div><a href="a.img"/><a href="b.img"/></div>`
	div, err := html.Parse(strings.NewReader(s))
	bee.Nil(err)
	as := slices.Collect(div.FindAll(html.IsTag("a")))
	bee.Equal(len(as), 2)
	bee.Equal(as[0].Tag(), "a")
	attrEqual(bee, as[0], "href", "a.img")
	bee.Equal(as[1].Tag(), "a")
	attrEqual(bee, as[1], "href", "b.img")
}

func TestAppendChild(t *testing.T) {
	bee := bee.New(t)
	div := html.NewNode("div", nil, "")
	section := html.NewNode("section", nil, "")
	h1 := html.NewNode("h1", nil, "Hi!")
	div.AppendChild(section)
	section.AppendChild(h1)
	bee.Equal(len(div.Children()), 1)
	bee.Equal(div.Children()[0].Tag(), "section")
	bee.Equal(len(div.Children()[0].Children()), 1)
	bee.Equal(div.Children()[0].Children()[0].Tag(), "h1")
	bee.Equal(div.Children()[0].Children()[0].Text(), "Hi!")
	bee.Equal(len(div.Children()[0].Children()[0].Children()), 1)
}

func TestRemoveAll(t *testing.T) {
	bee := bee.New(t)
	s := `<div><section><a href="a.img"/><a href="b.img"/><section></div>`
	root, err := html.Parse(strings.NewReader(s))
	bee.Nil(err)
	div, err := root.Find(html.IsTag("div"))
	bee.Nil(err)
	div.RemoveAll(html.IsTag("a"))
	bee.Equal(len(div.Children()), 1)
	section := div.Children()[0]
	bee.Equal(section.Tag(), "section")
	bee.Equal(len(section.Children()), 0)
}

func attrEqual(bee *bee.Bee, node *html.Node, name, value string) {
	attr, ok := node.GetAttr(name)
	bee.True(ok)
	bee.Equal(attr, value)
}
