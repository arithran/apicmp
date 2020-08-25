package main

import (
	"errors"
	"log"
	"os"

	"github.com/arithran/apicmp/diff"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "apicmp",
		Usage: "The apicmp command diffs API responses between NodeJS and Go services",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "token",
				EnvVars: []string{"X_API_TOKEN"},
			},
		},
		Before: func(c *cli.Context) error {
			if c.String("token") == "" {
				return errors.New("X_API_TOKEN required")
			}
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:  "diff",
				Usage: "apicmp diff",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "beforePath",
						Aliases: []string{"b"},
						Usage:   "Eg: http://localhost:4140/nodevideo_v1",
					},
					&cli.StringFlag{
						Name:    "afterPath",
						Aliases: []string{"a"},
						Usage:   "Eg: http://localhost:4140/video_v2",
					},
					&cli.StringFlag{
						Name:    "fixtureFile",
						Aliases: []string{"f"},
						Usage:   "Eg: ~/Downloads/fixture.csv",
					},
					&cli.StringFlag{
						Name:    "ignoreFields",
						Aliases: []string{"i"},
						Usage:   "Eg: @id,createdDate,ContentSKUs",
					},
					&cli.StringFlag{
						Name:    "rows",
						Aliases: []string{"r"},
						Usage:   "Eg: 1,7,12 (rerun failed or specific tests)",
					},
					&cli.StringFlag{
						Name:  "retry",
						Usage: "Eg: 424,500 (retry specific status codes once)",
					},
					&cli.StringFlag{
						Name:  "loglevel",
						Value: "debug",
						Usage: "Eg: debug",
					},
					&cli.StringFlag{
						Name:  "threads",
						Value: "4",
						Usage: "Eg: 6",
					},
				},
				Before: func(c *cli.Context) error {
					if c.String("beforePath") == "" {
						return errors.New("beforePath required")
					}
					if c.String("afterPath") == "" {
						return errors.New("afterPath required")
					}
					if c.String("fixtureFile") == "" {
						return errors.New("fixtureFile required")
					}
					return nil
				},
				Action: func(c *cli.Context) error {
					return diff.Cmp(c.Context, diff.Config{
						BeforeBasePath:  c.String("beforePath"),
						AfterBasePath:   c.String("afterPath"),
						FixtureFilePath: c.String("fixtureFile"),
						AccessToken:     c.String("token"),
						IgnoreFields:    c.String("ignoreFields"),
						Rows:            diff.Atoim(c.String("rows")),
						Retry:           diff.Atoim(c.String("retry")),
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
