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
	a := buildCLI()
	if err := a.Run(context.Background(), os.Args); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func buildCLI() *cli.Command {
	return &cli.Command{
		Name:    "i18n",
		Usage:   "i18n tool",
		Version: version,
		Commands: []*cli.Command{
			{
				Name:      "extract",
				Usage:     "Extract i18n messages from templates",
				ArgsUsage: " ",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "dir",
						Value: ".",
						Usage: "directory to scan for templates",
					},
					&cli.StringFlag{
						Name:  "out",
						Usage: "output JSON found i18n messages",
					},
					&cli.StringFlag{
						Name:  "gofile",
						Value: "gotext_stub.go",
						Usage: "synthetic Go file generated for gotext extract/update",
					},
					&cli.StringFlag{
						Name:  "pkg",
						Value: "main",
						Usage: "package name to use in generated Go file",
					},
					&cli.StringSliceFlag{
						Name:  "ext",
						Usage: "template extensions to consider",
						Value: []string{".html", ".htm", ".tmpl", ".gohtml", ".txt", ".tpl"},
					},
				},
				Action: func(_ context.Context, command *cli.Command) error {
					extractor := i18n.NewExtractor(
						command.String("dir"),
						command.String("out"),
						command.String("pkg"),
						command.String("gofile"),
						command.StringSlice("ext")...,
					)

					return extractor.Extract()
				},
			},
		},
	}
}
