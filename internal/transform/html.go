package transform

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/danielrenes/htdl/internal/html"
)

func RemoveTags(tags ...string) Transformer {
	return TransformerFunc(func(node *html.Node, ctx *TransformerContext) error {
		for _, tag := range tags {
			node.RemoveAll(html.IsTag(tag))
		}
		return nil
	})
}

func ResolveLinks(baseURL *url.URL) Transformer {
	return TransformerFunc(func(node *html.Node, ctx *TransformerContext) error {
		targets := map[string]string{
			"link":   "href",
			"a":      "href",
			"script": "src",
			"img":    "src",
		}
		for tag, attr := range targets {
			if err := resolveLink(node, baseURL, tag, attr); err != nil {
				return err
			}
		}
		return nil
	})
}

func resolveLink(node *html.Node, baseURL *url.URL, tag string, attr string) error {
	for _, n := range node.FindAll(html.IsTag(tag)) {
		if v, ok := n.GetAttr(attr); ok {
			v, err := resolveRef(baseURL, v)
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
