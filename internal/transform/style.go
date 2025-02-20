package transform

import (
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
		for styleTag := range node.FindAll(html.IsTag("style")) {
			if !yield(styleTag.Text(), nil) {
				return
			}
		}
		for linkTag := range node.FindAll(html.IsTag("link"), html.HasAttr("rel", "stylesheet")) {
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
		search = []rune("url(")
		char   rune
		size   int
		match  int
		arg    strings.Builder
		sb     strings.Builder
	)
	for i := 0; i < len(style); i += size {
		char, size = utf8.DecodeRuneInString(style[i:])
		if char == unicode.ReplacementChar {
			continue
		}
		if match < len(search) && char == search[match] {
			_, _ = sb.WriteRune(char)
			match++
		} else if match == len(search) {
			if char != ')' {
				_, _ = arg.WriteRune(char)
			} else {
				link := strings.Trim(arg.String(), `'"`)
				if strings.HasPrefix(link, "data:") {
					_, _ = sb.WriteString(link)
				} else {
					url, err := resolveRef(baseURL, link)
					if err != nil {
						return "", err
					}
					newSrc, err := downloadAndBase64Encode(url)
					if err != nil {
						return "", err
					}
					_, _ = fmt.Fprintf(&sb, "%s)", newSrc)
				}
				match = 0
				arg.Reset()
			}
		} else {
			match = 0
			arg.Reset()
			_, _ = sb.WriteRune(char)
		}
	}
	return sb.String(), nil
}
