package diff

import (
	"context"
	"encoding/csv"
	"io"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"
)

type (
	test struct {
		Row    int
		Before input
		After  input
	}
	input struct {
		Method  string
		Path    string
		Headers map[string]string
	}
)

func generateTests(ctx context.Context, c Config) (<-chan test, error) {
	// read csv file
	f, err := os.Open(c.FixtureFilePath)
	if err != nil {
		return nil, err
	}
	reader := csv.NewReader(f)
	reader.LazyQuotes = true
	_, _ = reader.Read() // discard header row

	// generate tests
	out := make(chan test)
	go func() {
		defer close(out)
		defer f.Close()
		cursor := 0

		for {
			cursor++

			fields, err := reader.Read()
			if err != nil {
				if err == io.EOF {
					break
				}
				log.Error(err)
				continue
			}

			if len(fields) != 3 {
				log.Errorf("invalid row at #%d", cursor)
				continue
			}

			if len(c.Rows) > 0 {
				if _, ok := c.Rows[cursor]; !ok {
					continue
				}
			}

			dma := fields[0]
			apikey := fields[1]
			path := fields[2]
			select {
			case out <- test{
				Row: cursor,
				Before: input{
					Method: http.MethodGet,
					Path:   c.BeforeBasePath + path,
					Headers: map[string]string{
						headerAPIKey:       apikey,
						headerUserDma:      dma,
						headerToken:        c.AccessToken,
						headerCacheControl: "no-cache",
					},
				},
				After: input{
					Method: http.MethodGet,
					Path:   c.AfterBasePath + path,
					Headers: map[string]string{
						headerAPIKey:                      apikey,
						headerUserDma:                     dma,
						headerToken:                       c.AccessToken,
						headerCacheControl:                "no-cache",
						"X-Feature-V1getvideobyidenabled": "true",
					},
				},
			}:
			case <-ctx.Done():
				return
			}
		}

	}()

	return out, nil
}
