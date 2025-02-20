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

type stage struct {
	name        string
	transformer transform.Transformer
}

func Archive(dir string, link string) error {
	slog.Info("Processing link", slog.String("link", link))
	htmlData, err := http.Download(link)
	if err != nil {
		return fmt.Errorf("download %s: %w", link, err)
	}
	htmlRoot, err := html.Parse(bytes.NewReader(htmlData))
	if err != nil {
		return err
	}
	baseURL, err := url.Parse(link)
	if err != nil {
		return fmt.Errorf("parse URL from %s: %w", link, err)
	}
	stages := []stage{
		{name: "resolve links", transformer: transform.ResolveLinks(baseURL)},
		{name: "inline styles", transformer: transform.InlineStyles(baseURL)},
		{name: "inline images", transformer: transform.InlineImages()},
		{name: "remove tags", transformer: transform.RemoveTags("style", "link", "script")},
		{name: "append inlined styles", transformer: transform.AppendInlinedStyles()},
	}
	ctx := transform.NewTransformerContext()
	for _, stage := range stages {
		if err := stage.transformer.Transform(htmlRoot, ctx); err != nil {
			return fmt.Errorf("%s: %w", stage.name, err)
		}
	}
	title, err := htmlRoot.Find(html.IsTag("title"))
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
