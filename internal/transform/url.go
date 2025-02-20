package transform

import (
	"fmt"
	"net/url"
)

func resolveRef(base *url.URL, link string) (string, error) {
	ref, err := url.Parse(link)
	if err != nil {
		return "", fmt.Errorf("parse URL from %s: %w", link, err)
	}
	return base.ResolveReference(ref).String(), nil
}
