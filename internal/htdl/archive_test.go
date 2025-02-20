package htdl_test

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/danielrenes/bee"
	"github.com/danielrenes/htdl/internal/htdl"
	"github.com/danielrenes/htdl/internal/html"
)

func TestArchive(t *testing.T) {
	bee := bee.New(t)
	expected := strings.TrimSpace(`
<!DOCTYPE html>
<html>
    <head>
        <title>index</title>
        <style>
            .subtitle {
                font-size: 1.5rem;
            }
            @font-face {
                font-family: 'MyFont';
                src: url(data:font/ttf;base64,%s) format('truetype');
            }
        </style>
    </head>
    <body>
        <div id="target">
            <h2 class="subtitle">abc</h2>
            <img src="data:image/png;base64,%s"></img>
        </div>
    </body>
</html>
`)
	fs := http.FileServer(http.Dir("testdata/server"))
	srv := http.Server{Addr: ":8000", Handler: fs}
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			t.Logf("run server: %v", err)
		}
	}()
	b64Font, err := os.ReadFile("testdata/base64/font.b64")
	bee.Nil(err)
	b64Image, err := os.ReadFile("testdata/base64/img.b64")
	bee.Nil(err)
	outDir := filepath.Join(os.TempDir(), "out")
	err = os.MkdirAll(outDir, 0755)
	bee.Nil(err)
	err = htdl.Archive(outDir, "http://localhost:8000/index.html")
	bee.Nil(err)
	entries, err := os.ReadDir(outDir)
	bee.Nil(err)
	bee.Equal(len(entries), 1)
	data, err := os.ReadFile(filepath.Join(outDir, "index.html"))
	bee.Nil(err)
	bee.Equal(renderHTML(bee, string(data)), renderHTML(bee, fmt.Sprintf(expected, b64Font, b64Image)))
	err = os.RemoveAll(outDir)
	bee.Nil(err)
	err = srv.Close()
	bee.Nil(err)
}

func renderHTML(bee *bee.Bee, s string) string {
	root, err := html.Parse(strings.NewReader(s))
	bee.Nil(err)
	html := root.RenderString()
	lineStartSpaces := regexp.MustCompile("\\n\\s*")
	lineEndSpaces := regexp.MustCompile("\\s*\\n")
	spacesBetweenTags := regexp.MustCompile("(?s)>\\s*<")
	html = lineStartSpaces.ReplaceAllString(html, "\n")
	html = lineEndSpaces.ReplaceAllString(html, "\n")
	html = spacesBetweenTags.ReplaceAllString(html, "><")
	return html
}
