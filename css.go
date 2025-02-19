package main

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io"
	"net/url"
	"strings"
	"unicode"
)

func InlineCSSLinks(baseURL *url.URL, r io.Reader) (string, error) {
	var (
		br   = bufio.NewReader(r)
		char rune
		size int
		err  error
		fn   int
		arg  strings.Builder
		sb   strings.Builder
	)
	for err == nil {
		char, size, err = br.ReadRune()
		if size == 1 && char == unicode.ReplacementChar {
			continue
		}
		if (fn == 0 && char == 'u') ||
			(fn == 1 && char == 'r') ||
			(fn == 2 && char == 'l') ||
			(fn == 3 && char == '(') {
			_, _ = sb.WriteRune(char)
			fn++
		} else if fn == 4 {
			if char != ')' {
				_, _ = arg.WriteRune(char)
			} else {
				link := arg.String()
				link = strings.Trim(link, `'"`)
				if strings.HasPrefix(link, "data:") {
					_, _ = sb.WriteString(link)
				} else {
					url, err := ResolveLink(baseURL, link)
					if err != nil {
						return "", err
					}
					rawData, err := Download(url)
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
				fn = 0
				arg.Reset()
			}
		} else {
			fn = 0
			_, _ = sb.WriteRune(char)
		}
	}
	if err == io.EOF {
		err = nil
	}
	return sb.String(), err
}
