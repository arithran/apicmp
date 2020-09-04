package diff

import (
	"context"
	"encoding/csv"
	"io"
	"net/http"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

const headerParts = 2

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

			t := test{
				Row: cursor,
				Before: input{
					Method: http.MethodGet,
					Path:   c.BeforeBasePath + path,
					Headers: map[string]string{
						headerAPIKey:  apikey,
						headerUserDma: dma,
						headerToken:   c.AccessToken,
					},
				},
				After: input{
					Method: http.MethodGet,
					Path:   c.AfterBasePath + path,
					Headers: map[string]string{
						headerAPIKey:  apikey,
						headerUserDma: dma,
						headerToken:   c.AccessToken,
					},
				},
			}

			for _, h := range c.Headers {
				parts := strings.Split(h, ":")
				if len(parts) != headerParts {
					log.Errorf("skipping invalid header --header %s", h)
					continue
				}

				k := strings.TrimSpace(parts[0])
				v := strings.TrimSpace(parts[1])

				t.Before.Headers[k] = v
				t.After.Headers[k] = v
			}

			select {
			case out <- t:
			case <-ctx.Done():
				return
			}
		}
	}()

	return out, nil
}
