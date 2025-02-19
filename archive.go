package main

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func Archive(dir string, link string) error {
	slog.Info("Processing link", slog.String("link", link))
	htmlData, err := Download(link)
	if err != nil {
		return fmt.Errorf("download %s: %w", link, err)
	}
	htmlRoot, err := ParseHTML(bytes.NewReader(htmlData))
	if err != nil {
		return err
	}
	baseURL, err := url.Parse(link)
	if err != nil {
		return fmt.Errorf("parse URL from %s: %w", link, err)
	}
	if err := htmlRoot.ResolveLinks(baseURL); err != nil {
		return fmt.Errorf("resolve links: %w", err)
	}
	styles, err := collectStyles(htmlRoot)
	if err != nil {
		return fmt.Errorf("collect styles: %w", err)
	}
	styles, err = InlineCSSLinks(baseURL, strings.NewReader(styles))
	if err != nil {
		return fmt.Errorf("inline CSS links: %w", err)
	}
	if err := InlineImages(htmlRoot); err != nil {
		return fmt.Errorf("inline images: %w", err)
	}
	htmlRoot.RemoveAll(IsTag("style"))
	htmlRoot.RemoveAll(IsTag("link"))
	htmlRoot.RemoveAll(IsTag("script"))
	head, err := htmlRoot.Find(IsTag("head"))
	if err != nil {
		return fmt.Errorf("find head element: %w", err)
	}
	head.Add(NewHTMLNode("style", nil, styles))
	title, err := head.Find(IsTag("title"))
	if err != nil {
		return fmt.Errorf("find title element: %w", err)
	}
	name := title.Text()
	name = filepath.Join(dir, fmt.Sprintf("%s.html", name))
	slog.Info("Writing file", slog.String("file", name))
	fp, err := os.Create(name)
	if err != nil {
		return fmt.Errorf("create %s: %w", name, err)
	}
	defer fp.Close()
	if err := htmlRoot.Render(fp); err != nil {
		return fmt.Errorf("render HTML to %s: %w", name, err)
	}
	return nil
}

func collectStyles(node *HTMLNode) (string, error) {
	sb := &strings.Builder{}
	for _, styleTag := range node.FindAll(IsTag("style")) {
		_, _ = fmt.Fprintln(sb, styleTag.Text())
	}
	for _, linkTag := range node.FindAll(IsTag("link"), HasAttr("rel", "stylesheet")) {
		if href, ok := linkTag.GetAttr("href"); ok {
			cssData, err := Download(href)
			if err != nil {
				return "", err
			}
			_, _ = fmt.Fprintln(sb, string(cssData))
		}
	}
	return sb.String(), nil
}
