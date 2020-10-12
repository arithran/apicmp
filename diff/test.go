package diff

import (
	"context"
	"encoding/csv"
	"errors"
	"io"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

const (
	headerParts = 2
	unset       = -1
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
		Body    string
	}
)

type csvHelper struct {
	method  int
	path    int
	body    int
	headers map[string]int
}

func newCSVHelper(header []string) csvHelper {
	h := csvHelper{
		method:  unset,
		path:    unset,
		body:    unset,
		headers: make(map[string]int),
	}

	for k, v := range header {
		// strip BOM
		v = strings.Replace(v, "\ufeff", "", -1)

		switch v {
		case "method":
			h.method = k

		case "path":
			h.path = k

		case "body":
			h.body = k

		default:
			// anything else is a header
			h.headers[v] = k
		}
	}

	return h
}

func (h csvHelper) validate() error {
	if h.path == unset {
		return errors.New("csv file is missing 'path' column. please see instructions at https://github.com/arithran/apicmp")
	}

	return nil
}
func (h csvHelper) Method(row []string) string {
	if h.method != unset && len(row) > h.method {
		return row[h.method]
	}

	return "GET"
}
func (h csvHelper) Path(row []string) string {
	if h.path != unset && len(row) > h.path {
		return row[h.path]
	}

	return ""
}
func (h csvHelper) Body(row []string) string {
	if h.body != unset && len(row) > h.body {
		return row[h.body]
	}

	return ""
}
func (h csvHelper) Headers(row []string) map[string]string {
	hs := map[string]string{}

	// default headers
	hs["Content-Type"] = "application/json"

	for k, v := range h.headers {
		if len(row) > v {
			hs[k] = row[v]
		}
	}
	return hs
}

func (h csvHelper) totalFields() int {
	var count int
	if h.method != unset {
		count++
	}
	if h.path != unset {
		count++
	}
	if h.body != unset {
		count++
	}

	if length := len(h.headers); length != 0 {
		count += length
	}

	return count
}

func generateTests(ctx context.Context, c Config) (<-chan test, error) {
	// read csv file
	f, err := os.Open(c.FixtureFilePath)
	if err != nil {
		return nil, err
	}
	reader := csv.NewReader(f)
	reader.LazyQuotes = true

	// lets determine the shape of this CSV file based on the header
	header, err := reader.Read()
	if err != nil {
		return nil, err
	}
	h := newCSVHelper(header)
	if err := h.validate(); err != nil {
		return nil, err
	}

	totalFields := h.totalFields()

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

			if len(fields) != totalFields {
				log.Errorf("invalid row at #%d", cursor)
				continue
			}

			if len(c.Rows) > 0 {
				if _, ok := c.Rows[cursor]; !ok {
					continue
				}
			}

			t := test{
				Row: cursor,
				Before: input{
					Method:  h.Method(fields),
					Path:    buildURL(c.BeforeBasePath, h.Path(fields), c.QueryStrings),
					Headers: h.Headers(fields),
					Body:    h.Body(fields),
				},
				After: input{
					Method:  h.Method(fields),
					Path:    buildURL(c.AfterBasePath, h.Path(fields), c.QueryStrings),
					Headers: h.Headers(fields),
					Body:    h.Body(fields),
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
