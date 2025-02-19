package main

import (
	"fmt"
	"net/url"
)

func ResolveLink(baseURL *url.URL, link string) (string, error) {
	ref, err := url.Parse(link)
	if err != nil {
		return "", fmt.Errorf("parse URL from %s: %w", link, err)
	}
	return baseURL.ResolveReference(ref).String(), nil
}
