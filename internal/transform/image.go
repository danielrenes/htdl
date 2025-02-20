package transform

import (
	"iter"
	"log/slog"
	"strings"

	"github.com/danielrenes/htdl/internal/html"
)

func InlineImages() Transformer {
	return TransformerFunc(func(node *html.Node, ctx *TransformerContext) error {
		iters := []iter.Seq[*html.Node]{
			node.FindAll(html.IsTag("img")),
			findSourceTagsWithImageType(node),
		}
		for _, iter := range iters {
			for n := range iter {
				if err := inlineImage(n); err != nil {
					return err
				}
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

func findSourceTagsWithImageType(node *html.Node) iter.Seq[*html.Node] {
	return func(yield func(*html.Node) bool) {
		for n := range node.FindAll(html.IsTag("source")) {
			if v, ok := n.GetAttr("type"); ok && strings.HasPrefix(v, "image/") {
				if !yield(n) {
					return
				}
				continue
			}
			src, ok := getSource(n)
			if !ok {
				continue
			}
			idx := strings.LastIndex(src, ".")
			if idx < 0 {
				continue
			}
			ext := src[idx+1:]
			if idx := strings.Index(ext, "?"); idx > 0 {
				ext = ext[:idx]
			}
			if ext != "jpg" && ext != "jpeg" && ext != "png" && ext != "svg" {
				continue
			}
			if !yield(n) {
				return
			}
		}
	}
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
