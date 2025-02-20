package transform

import (
	"log/slog"
	"slices"
	"strings"

	"github.com/danielrenes/htdl/internal/html"
)

func InlineImages() Transformer {
	return TransformerFunc(func(node *html.Node, ctx *TransformerContext) error {
		nodes := make([]*html.Node, 0)
		nodes = append(nodes, node.FindAll(html.IsTag("img"))...)
		nodes = append(nodes, findSourceTagsWithImageType(node)...)
		for _, n := range nodes {
			if err := inlineImage(n); err != nil {
				return err
			}
		}
		return nil
	})
}

func inlineImage(node *html.Node) error {
	src, ok := getSource(node)
	if !ok {
		return nil
	}
	slog.Debug("Inline image", slog.String("src", src))
	newSrc, err := downloadAndBase64Encode(src)
	if err != nil {
		return err
	}
	node.DeleteAttr("src")
	node.DeleteAttr("srcset")
	node.SetAttr("src", newSrc)
	return nil
}

func findSourceTagsWithImageType(node *html.Node) []*html.Node {
	sources := node.FindAll(html.IsTag("source"))
	return slices.DeleteFunc(sources, func(source *html.Node) bool {
		if v, ok := source.GetAttr("type"); ok && strings.HasPrefix(v, "image/") {
			return false
		}
		src, ok := getSource(source)
		if !ok {
			return false
		}
		idx := strings.LastIndex(src, ".")
		if idx < 0 {
			return true
		}
		ext := src[idx+1:]
		if idx := strings.Index(ext, "?"); idx > 0 {
			ext = ext[:idx]
		}
		return ext != "jpg" && ext != "jpeg" && ext != "png" && ext != "svg"
	})
}

func getSource(node *html.Node) (string, bool) {
	if src, ok := node.GetAttr("src"); ok {
		return src, true
	}
	if srcset, ok := node.GetAttr("srcset"); ok {
		srcs := strings.Split(srcset, ",")
		src := strings.TrimSpace(srcs[len(srcs)-1])
		if idx := strings.Index(src, " "); idx > 0 {
			src = strings.TrimSpace(src[:idx])
		}
		return src, true
	}
	return "", false
}
