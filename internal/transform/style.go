package transform

import (
	"encoding/base64"
	"fmt"
	"iter"
	"net/url"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/danielrenes/htdl/internal/html"
	"github.com/danielrenes/htdl/internal/http"
)

type inlineStylesKey struct{}

func InlineStyles(baseURL *url.URL) Transformer {
	return TransformerFunc(func(node *html.Node, ctx *TransformerContext) error {
		styles := strings.Builder{}
		for style, err := range iterStyles(node) {
			if err != nil {
				return err
			}
			style, err = inlineLinks(baseURL, style)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintln(&styles, style)
		}
		ctx.SetValue(inlineStylesKey{}, styles.String())
		return nil
	})
}

func AppendInlinedStyles() Transformer {
	return TransformerFunc(func(node *html.Node, ctx *TransformerContext) error {
		head, err := node.Find(html.IsTag("head"))
		if err != nil {
			return fmt.Errorf("find head element: %w", err)
		}
		styles := ctx.GetValue(inlineStylesKey{}).(string)
		head.AppendChild(html.NewNode("style", nil, styles))
		return nil
	})
}

func iterStyles(node *html.Node) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		for _, styleTag := range node.FindAll(html.IsTag("style")) {
			if !yield(styleTag.Text(), nil) {
				return
			}
		}
		for _, linkTag := range node.FindAll(html.IsTag("link"), html.HasAttr("rel", "stylesheet")) {
			if href, ok := linkTag.GetAttr("href"); ok {
				cssData, err := http.Download(href)
				if !yield(string(cssData), err) {
					return
				}
			}
		}
	}
}

func inlineLinks(baseURL *url.URL, style string) (string, error) {
	var (
		match int
		arg   strings.Builder
		sb    strings.Builder
	)
	for i, w := 0, 0; i < len(style); i += w {
		char, size := utf8.DecodeRuneInString(style[i:])
		if char == unicode.ReplacementChar {
			continue
		}
		if (match == 0 && char == 'u') ||
			(match == 1 && char == 'r') ||
			(match == 2 && char == 'l') ||
			(match == 3 && char == '(') {
			_, _ = sb.WriteRune(char)
			match++
		} else if match == 4 {
			if char != ')' {
				_, _ = arg.WriteRune(char)
			} else {
				link := arg.String()
				link = strings.Trim(link, `'"`)
				if strings.HasPrefix(link, "data:") {
					_, _ = sb.WriteString(link)
				} else {
					url, err := resolveRef(baseURL, link)
					if err != nil {
						return "", err
					}
					rawData, err := http.Download(url)
					if err != nil {
						return "", err
					}
					b64Data := base64.StdEncoding.EncodeToString(rawData)
					idx := strings.LastIndex(url, ".")
					if idx < 0 {
						return "", fmt.Errorf("unknown extension: %s", url)
					}
					ext := url[idx+1:]
					if idx := strings.Index(ext, "?"); idx > 0 {
						ext = ext[:idx]
					}
					var dataType string
					switch ext {
					case "png", "jpg", "jpeg", "svg":
						dataType = "image"
					case "otf", "ttf", "woff", "woff2":
						dataType = "font"
					default:
						dataType = "text"
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
					_, _ = fmt.Fprintf(&sb, "data:%s/%s;base64,%s)", dataType, mimeType, b64Data)
				}
				match = 0
				arg.Reset()
			}
		} else {
			match = 0
			_, _ = sb.WriteRune(char)
		}
		w = size
	}
	return sb.String(), nil
}
