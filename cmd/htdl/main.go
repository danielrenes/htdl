package main

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/danielrenes/htdl/internal/htdl"
)

func run(args *args) error {
	logger := slog.New(NewSlogHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     args.LogLevel,
	}))
	slog.SetDefault(logger)
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get current working directory: %w", err)
	}
	errs := make([]error, 0)
	for _, link := range args.Links {
		if err := htdl.Archive(cwd, link); err != nil {
			slog.Warn(fmt.Sprintf("Error downloading %s", link), slog.String("error", err.Error()))
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func main() {
	args, err := parseArgs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		flag.Usage()
		os.Exit(1)
	}
	if err := run(args); err != nil {
		slog.Error("Error running main", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
