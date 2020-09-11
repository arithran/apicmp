package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/arithran/apicmp/diff"
	"github.com/urfave/cli/v2"
)

var validMatches = map[string]struct{}{
	"exact":    {},
	"superset": {},
}

func main() {
	app := &cli.App{
		Name:  "apicmp",
		Usage: "The apicmp command diffs API responses between NodeJS and Go services",
		Commands: []*cli.Command{
			{
				Name:  "diff",
				Usage: "apicmp diff",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "before",
						Aliases: []string{"B"},
						Usage:   "https://api.example.com",
					},
					&cli.StringFlag{
						Name:    "after",
						Aliases: []string{"A"},
						Usage:   "https://qa-api.example.com",
					},
					&cli.StringFlag{
						Name:    "file",
						Aliases: []string{"F"},
						Usage:   "~/Downloads/fixtures.csv",
					},
					&cli.StringSliceFlag{
						Name:    "header",
						Aliases: []string{"H"},
						Usage:   "'Cache-Control: no-cache' ",
					},
					&cli.StringFlag{
						Name:    "ignore",
						Aliases: []string{"I"},
						Usage:   "createdAt,modifiedAt",
					},
					&cli.StringFlag{
						Name:    "rows",
						Aliases: []string{"R"},
						Usage:   "1,7,12 (Rerun failed or specific tests from file)",
					},
					&cli.StringFlag{
						Name:  "retry",
						Usage: "424,500 (HTTP status codes)",
					},
					&cli.StringFlag{
						Name:  "match",
						Value: "exact",
						Usage: "exact|superset",
					},
					&cli.StringFlag{
						Name:  "threads",
						Value: "4",
						Usage: "10",
					},
					&cli.StringFlag{
						Name:  "loglevel",
						Value: "debug",
						Usage: "info",
					},
				},
				Before: func(c *cli.Context) error {
					if c.String("before") == "" {
						return errors.New("before required")
					}
					if c.String("after") == "" {
						return errors.New("after required")
					}
					if c.String("file") == "" {
						return errors.New("file required")
					}
					if _, ok := validMatches[c.String("match")]; !ok {
						return errors.New("invalid --match flag")
					}
					return nil
				},
				Action: func(c *cli.Context) error {
					ctx, cancel := context.WithCancel(c.Context)

					go func() {
						done := make(chan os.Signal, 1)
						signal.Notify(done, os.Interrupt, syscall.SIGTERM)
						<-done
						cancel()
					}()

					return diff.Cmp(ctx, diff.Config{
						BeforeBasePath:  c.String("before"),
						AfterBasePath:   c.String("after"),
						FixtureFilePath: c.String("file"),
						Headers:         c.StringSlice("header"),
						IgnoreFields:    diff.Atoam(c.String("ignore")),
						Rows:            diff.Atoim(c.String("rows")),
						Retry:           diff.Atoim(c.String("retry")),
						Match:           c.String("match"),
						LogLevel:        c.String("loglevel"),
						Threads:         c.Int("threads"),
					})
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
