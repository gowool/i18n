package main

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/gowool/i18n"
)

var version = "v0.0.1"

func main() {
	a := buildCLI(func(command *cli.Command) error {
		return i18n.NewExtractor(
			command.String("dir"),
			command.String("out"),
			command.String("pkg"),
			command.String("gofile"),
			command.StringSlice("ext")...,
		).Extract()
	})

	if err := a.Run(context.Background(), os.Args); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func buildCLI(extractor func(*cli.Command) error) *cli.Command {
	return &cli.Command{
		Name:     "i18n",
		Usage:    "i18n tool",
		Version:  version,
		Commands: []*cli.Command{extract(extractor)},
	}
}
