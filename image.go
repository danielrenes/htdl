package main

import (
	"encoding/base64"
	"fmt"
	"log/slog"
	"slices"
	"strings"
)

func InlineImages(node *HTMLNode) error {
	sources := node.FindAll(IsTag("source"))
	sources = slices.DeleteFunc(sources, func(source *HTMLNode) bool {
		if v, ok := source.GetAttr("type"); ok && strings.HasPrefix(v, "image/") {
			return false
		}
		src, ok := source.GetAttr("src")
		if !ok {
			if srcset, ok := source.GetAttr("srcset"); ok {
				srcs := strings.Split(srcset, ",")
				src = strings.TrimSpace(srcs[len(srcs)-1])
				if idx := strings.Index(src, " "); idx > 0 {
					src = strings.TrimSpace(src[:idx])
				}
			}
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
	nodes := make([]*HTMLNode, 0)
	nodes = append(nodes, node.FindAll(IsTag("img"))...)
	nodes = append(nodes, sources...)
	for _, n := range nodes {
		src, ok := n.GetAttr("src")
		if !ok {
			srcset, ok := n.GetAttr("srcset")
			if !ok {
				continue
			}
			srcs := strings.Split(srcset, ",")
			src = strings.TrimSpace(srcs[len(srcs)-1])
			if idx := strings.Index(src, " "); idx > 0 {
				src = strings.TrimSpace(src[:idx])
			}
		}
		slog.Debug("Inline image", slog.String("src", src))
		rawData, err := Download(src)
		if err != nil {
			return err
		}
		b64Data := base64.StdEncoding.EncodeToString(rawData)
		idx := strings.LastIndex(src, ".")
		if idx < 0 {
			return fmt.Errorf("unknown extension: %s", src)
		}
		ext := src[idx+1:]
		if idx := strings.Index(ext, "?"); idx > 0 {
			ext = ext[:idx]
		}
		var mimeType string
		switch ext {
		case "jpg":
			mimeType = "jpeg"
		case "svg":
			mimeType = "svg+xml"
		default:
			mimeType = ext
		}
		newSrc := fmt.Sprintf("data:image/%s;base64,%s", mimeType, b64Data)
		n.DeleteAttr("src")
		n.DeleteAttr("srcset")
		n.SetAttr("src", newSrc)
	}
	return nil
}
