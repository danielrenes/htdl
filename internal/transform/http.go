package transform

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/danielrenes/htdl/internal/http"
)

func downloadAndBase64Encode(link string) (string, error) {
	data, err := http.Download(link)
	if err != nil {
		return "", err
	}
	b64Data := base64.StdEncoding.EncodeToString(data)
	idx := strings.LastIndex(link, ".")
	if idx < 0 {
		return "", fmt.Errorf("extension not found: %s", link)
	}
	ext := link[idx+1:]
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
	src := fmt.Sprintf("data:%s/%s;base64,%s", dataType, mimeType, b64Data)
	return src, nil
}
