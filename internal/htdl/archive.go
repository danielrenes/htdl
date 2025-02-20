package htdl

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"

	"github.com/danielrenes/htdl/internal/html"
	"github.com/danielrenes/htdl/internal/http"
	"github.com/danielrenes/htdl/internal/transform"
)

func Archive(dir string, link string) error {
	slog.Info("Processing link", slog.String("link", link))
	htmlRoot, err := downloadHTML(link)
	if err != nil {
		return err
	}
	baseURL, err := url.Parse(link)
	if err != nil {
		return fmt.Errorf("parse URL from %s: %w", link, err)
	}
	pipeline := transform.NewPipeline(
		transform.Named("resolve links", transform.ResolveLinks(baseURL)),
		transform.Named("inline styles", transform.InlineStyles(baseURL)),
		transform.Named("inline images", transform.InlineImages()),
		transform.Named("remove tags", transform.RemoveTags("style", "link", "script")),
		transform.Named("append inlined styles", transform.AppendInlinedStyles()),
	)
	if err := pipeline.Run(htmlRoot); err != nil {
		return err
	}
	title, err := getTitle(htmlRoot)
	if err != nil {
		return fmt.Errorf("find title element: %w", err)
	}
	path := filepath.Join(dir, fmt.Sprintf("%s.html", title))
	slog.Info("Writing file", slog.String("path", path))
	if err := saveFile(path, htmlRoot); err != nil {
		return err
	}
	return nil
}

func downloadHTML(link string) (*html.Node, error) {
	htmlData, err := http.Download(link)
	if err != nil {
		return nil, fmt.Errorf("download %s: %w", link, err)
	}
	htmlRoot, err := html.Parse(bytes.NewReader(htmlData))
	if err != nil {
		return nil, err
	}
	return htmlRoot, nil
}

func getTitle(node *html.Node) (string, error) {
	title, err := node.Find(html.IsTag("title"))
	if err != nil {
		return "", fmt.Errorf("find title element: %w", err)
	}
	return title.Text(), nil
}

func saveFile(path string, node *html.Node) error {
	fp, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	defer fp.Close()
	if err := node.Render(fp); err != nil {
		return fmt.Errorf("render HTML to %s: %w", path, err)
	}
	return nil
}
