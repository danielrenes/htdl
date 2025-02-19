package main

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

func Download(link string) ([]byte, error) {
	slog.Debug("Downloading link", slog.String("link", link))
	resp, err := http.Get(link)
	if err != nil {
		return nil, fmt.Errorf("download %s: %w", link, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusTooManyRequests {
		slog.Debug("Too many requests, retrying in 1 second")
		time.Sleep(1 * time.Second)
		return Download(link)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get %s: %s", link, resp.Status)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response from %s: %w", link, err)
	}
	return data, nil
}
