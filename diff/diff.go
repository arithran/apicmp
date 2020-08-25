package diff

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/nsf/jsondiff"
	"github.com/olekukonko/tablewriter"
	log "github.com/sirupsen/logrus"
)

const (
	headerAPIKey  = "X-Api-Key"
	headerUserDma = "X-User-Dma"
	headerToken   = "X-Access-Token"
)

type Summary struct {
	Count      int
	Passed     int
	FailedRows []int
	Issues     map[string]int
}

type Config struct {
	BeforeBasePath  string
	AfterBasePath   string
	FixtureFilePath string
	AccessToken     string
	IgnoreFields    string
	Rows            map[int]struct{}
	Retry           map[int]struct{}
	LogLevel        string
	Threads         int
}

// Cmp will compare the before and after
func Cmp(ctx context.Context, c Config) error {
	if err := setLoglevel(c.LogLevel); err != nil {
		return err
	}

	// gen tests
	tChan, err := generateTests(c, c.Rows)
	if err != nil {
		return err
	}

	// init assertion workers
	client := newRetriableHTTPClient(&http.Client{}, c.Retry)
	cs := make([]<-chan result, c.Threads)
	for i := 0; i < c.Threads; i++ {
		cs[i] = assert(ctx, client, tChan)
	}

	// compute results
	sum := Summary{
		Issues: map[string]int{},
	}
	ignore := Atoam(c.IgnoreFields)
	opts := jsondiff.DefaultConsoleOptions()
	opts.PrintTypes = false
	diffs := merge(cs...)
	for d := range diffs {
		sum.Count++
		if d.beforeCode != d.afterCode {
			_ = curlTpl.Execute(os.Stdout, d.e)
			log.Errorf("StatusCodes didn't match,\n before: %s\n after : %s", d.beforeCode, d.afterCode)
			fmt.Printf("\n\n")
			sum.FailedRows = append(sum.FailedRows, d.e.Row)
			continue
		}

		delta := [][]string{}
		for k, v := range d.beforeBody {
			if _, ok := ignore[k]; ok {
				continue
			}

			result, diff := jsondiff.Compare(v, d.afterBody[k], &opts)
			if result != jsondiff.FullMatch {
				delta = append(delta, []string{k, cleanDiff(diff)})
				sum.Issues[k]++
			}
		}

		if len(delta) > 0 {
			sort.Sort(sortDelta(delta))
			_ = curlTpl.Execute(os.Stdout, d.e)
			table := tablewriter.NewWriter(os.Stdout)
			table.SetAutoFormatHeaders(false)
			table.SetHeader([]string{"Field", "Diff"})
			table.SetBorder(false)
			table.AppendBulk(delta)
			table.Render()
			fmt.Printf("\n\n")

			sum.FailedRows = append(sum.FailedRows, d.e.Row)
		} else {
			sum.Passed++
		}
	}

	sumTable := [][]string{}
	for k, v := range sum.Issues {
		sumTable = append(sumTable, []string{k, strconv.Itoa(v)})
	}
	sort.Sort(sortDelta(sumTable))
	sort.Ints(sum.FailedRows)
	failedRows := make([]string, len(sum.FailedRows))
	for k, v := range sum.FailedRows {
		failedRows[k] = strconv.Itoa(v)
	}

	fmt.Printf("Summary:\n Total Tests:%d\n Passed:%d\n Failed:%d\n Failed Rows:%s\n", sum.Count, sum.Passed, sum.Count-sum.Passed, strings.Join(failedRows, ","))
	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoFormatHeaders(false)
	table.SetHeader([]string{"Field", "Issues"})
	table.SetBorder(false)
	table.AppendBulk(sumTable)
	table.Render()

	return nil
}

type (
	test struct {
		Row    int
		Before args
		After  args
	}

	args struct {
		Method  string
		Path    string
		Headers map[string]string
	}
)

func generateTests(c Config, rows map[int]struct{}) (<-chan test, error) {
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

			if len(rows) > 0 {
				if _, ok := rows[cursor]; !ok {
					continue
				}
			}

			dma := fields[0]
			apikey := fields[1]
			path := fields[2]
			out <- test{
				Row: cursor,
				Before: args{
					Method: http.MethodGet,
					Path:   c.BeforeBasePath + path,
					Headers: map[string]string{
						headerAPIKey:  apikey,
						headerUserDma: dma,
						headerToken:   c.AccessToken,
					},
				},
				After: args{
					Method: http.MethodGet,
					Path:   c.AfterBasePath + path,
					Headers: map[string]string{
						headerAPIKey:                      apikey,
						headerUserDma:                     dma,
						headerToken:                       c.AccessToken,
						"X-Feature-V1getvideobyidenabled": "true",
					},
				},
			}
		}

		f.Close()
		close(out)
	}()

	return out, nil
}

type result struct {
	e          test
	beforeCode string
	beforeBody map[string]json.RawMessage
	afterCode  string
	afterBody  map[string]json.RawMessage
}

func assert(ctx context.Context, client httpClient, in <-chan test) <-chan result {
	out := make(chan result)

	go func() {
		for e := range in {
			// create requests
			before, err := http.NewRequestWithContext(ctx, e.Before.Method, e.Before.Path, nil)
			if err != nil {
				log.Error(err)
				continue
			}
			for k, v := range e.Before.Headers {
				before.Header.Add(k, v)
			}

			after, err := http.NewRequestWithContext(ctx, e.After.Method, e.After.Path, nil)
			if err != nil {
				log.Error(err)
				continue
			}
			for k, v := range e.After.Headers {
				after.Header.Add(k, v)
			}

			// make requests
			bResp, err := client.Do(before)
			if err != nil {
				log.Error(err)
				continue
			}
			httpTrace(before, bResp)

			aResp, err := client.Do(after)
			if err != nil {
				log.Error(err)
				continue
			}
			httpTrace(after, aResp)

			// unmashall & return
			d := result{e: e}

			d.beforeCode = bResp.Status
			err = json.NewDecoder(bResp.Body).Decode(&d.beforeBody)
			if err != nil {
				log.Error(err)
				continue
			}
			bResp.Body.Close()

			d.afterCode = aResp.Status
			err = json.NewDecoder(aResp.Body).Decode(&d.afterBody)
			if err != nil {
				log.Error(err)
				continue
			}
			aResp.Body.Close()

			out <- d
		}

		close(out)
	}()

	return out
}

func merge(cs ...<-chan result) <-chan result {
	var wg sync.WaitGroup
	out := make(chan result)

	output := func(c <-chan result) {
		for n := range c {
			out <- n
		}
		wg.Done()
	}
	wg.Add(len(cs))
	for _, c := range cs {
		go output(c)
	}

	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}
