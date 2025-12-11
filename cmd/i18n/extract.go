package main

import (
	"context"

	"github.com/urfave/cli/v3"
)

func extract(extractor func(*cli.Command) error) *cli.Command {
	return &cli.Command{
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
			return extractor(command)
		},
	}
}
