package main

import (
	"flag"
	"fmt"
	"log/slog"
	"maps"
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
