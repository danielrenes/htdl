package main

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"maps"
	"os"
	"slices"
	"strings"
)

type args struct {
	LogLevel slog.Level
	Links    []string
}

func parseArgs() (*args, error) {
	logLevels := make(map[string]slog.Level, 0)
	for _, lvl := range []slog.Level{slog.LevelDebug, slog.LevelInfo, slog.LevelWarn, slog.LevelError} {
		logLevels[strings.ToLower(lvl.String())] = lvl
	}
	logLevel := flag.String(
		"log-level",
		strings.ToLower(slog.LevelInfo.String()),
		fmt.Sprintf("The log level. Choices: %v", slices.Collect(maps.Keys(logLevels))),
	)
	flag.Parse()
	args := args{}
	if lvl, ok := logLevels[*logLevel]; ok {
		args.LogLevel = lvl
	} else {
		return nil, fmt.Errorf("invalid log level %s", *logLevel)
	}
	args.Links = flag.Args()
	return &args, nil
}

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
		if err := Archive(cwd, link); err != nil {
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
