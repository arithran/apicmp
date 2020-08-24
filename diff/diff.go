package diff

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"text/template"

	"github.com/nsf/jsondiff"
	"github.com/olekukonko/tablewriter"
	log "github.com/sirupsen/logrus"
)

const curlTemplate = `
Before:
curl --location --request {{ .Before.Method }} '{{ .Before.Path }}' \{{range $k, $v := .Before.Headers}}
--header '{{$k}}: {{$v}}'{{end}}

After:
curl --location --request {{ .After.Method }} '{{ .After.Path }}' \{{range $k, $v := .After.Headers}}
--header '{{$k}}: {{$v}}'{{end}}

`

const (
	headerAPIKey  = "X-Api-Key"
	headerUserDma = "X-User-Dma"
	headerToken   = "X-Access-Token"
)

type Config struct {
	BeforeBasePath  string
	AfterBasePath   string
	FixtureFilePath string
	AccessToken     string
	IgnoreFields    string
	Rows            string
	LogLevel        string
	Threads         int
}

type Summary struct {
	Count      int
	Passed     int
	FailedRows []string
	Issues     map[string]int
}

// Cmp will compare the before and after
func Cmp(ctx context.Context, c Config) error {
	t := template.Must(template.New("curl").Parse(curlTemplate))

	// set loglevel
	l, err := log.ParseLevel(c.LogLevel)
	if err != nil {
		return err
	}
	log.SetLevel(l)

	// generate events from csv
	eChan, err := ReadEvents(c, parseRows(c.Rows)...)
	if err != nil {
		return err
	}

	// kickoff workers
	cs := make([]<-chan Diff, c.Threads)
	for i := 0; i < c.Threads; i++ {
		cs[i] = Test(ctx, c, &http.Client{}, eChan)
	}

	// ignore certain fields in diffs
	ignoreFields := strings.Split(c.IgnoreFields, ",")
	ignore := map[string]struct{}{}
	for _, v := range ignoreFields {
		ignore[v] = struct{}{}
	}

	// compare diffs
	sum := Summary{
		Issues: map[string]int{},
	}
	opts := jsondiff.DefaultConsoleOptions()
	opts.PrintTypes = false
	diffs := merge(cs...)
	for d := range diffs {
		sum.Count++
		if d.beforeCode != d.afterCode {
			_ = t.Execute(os.Stdout, d.e)
			log.Error("StatusCodes didn't match")
			fmt.Printf("\n\n")
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
			_ = t.Execute(os.Stdout, d.e)
			table := tablewriter.NewWriter(os.Stdout)
			table.SetAutoFormatHeaders(false)
			table.SetHeader([]string{"Field", "Diff"})
			table.SetBorder(false)
			table.AppendBulk(delta)
			table.Render()
			fmt.Printf("\n\n")

			sum.FailedRows = append(sum.FailedRows, strconv.Itoa(d.e.Row))
		} else {
			sum.Passed++
		}
	}

	sumTable := [][]string{}
	for k, v := range sum.Issues {
		sumTable = append(sumTable, []string{k, strconv.Itoa(v)})
	}
	sort.Sort(sortDelta(sumTable))
	fmt.Printf("Summary:\n Total Tests:%d\n Passed:%d\n Failed:%d\n Failed Rows:%s\n", sum.Count, sum.Passed, sum.Count-sum.Passed, strings.Join(sum.FailedRows, ","))
	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoFormatHeaders(false)
	table.SetHeader([]string{"Field", "Issues"})
	table.SetBorder(false)
	table.AppendBulk(sumTable)
	table.Render()

	return nil
}

type Event struct {
	Row    int
	DMA    string
	APIKey string
	Path   string
	Before input
	After  input
}

type input struct {
	Method  string
	Path    string
	Headers map[string]string
}

func ReadEvents(c Config, selectedRows ...int) (<-chan Event, error) {
	// selectedRows is useful when re-running failed tests
	var mSelectedRows map[int]struct{}
	if len(selectedRows) > 0 {
		mSelectedRows = make(map[int]struct{})
		for _, v := range selectedRows {
			mSelectedRows[v] = struct{}{}
		}
	}

	f, err := os.Open(c.FixtureFilePath)
	if err != nil {
		return nil, err
	}

	out := make(chan Event)
	go func() {
		reader := csv.NewReader(f)
		reader.LazyQuotes = true
		_, _ = reader.Read() // discard header

		cursor := 0
		for {
			cursor++

			row, err := reader.Read()
			if err != nil {
				if err == io.EOF {
					break
				}
				log.Error(err)
				continue
			}

			if len(mSelectedRows) > 0 {
				if _, ok := mSelectedRows[cursor]; !ok {
					continue
				}
			}

			dma := row[0]
			apikey := row[1]
			path := row[2]
			out <- Event{
				Row:    cursor,
				DMA:    row[0],
				APIKey: row[1],
				Path:   row[2],
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

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Diff struct {
	e          Event
	beforeCode string
	beforeBody map[string]json.RawMessage
	afterCode  string
	afterBody  map[string]json.RawMessage
}

func Test(ctx context.Context, c Config, client httpClient, in <-chan Event) <-chan Diff {
	out := make(chan Diff)

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
			trace(before, bResp)

			aResp, err := client.Do(after)
			if err != nil {
				log.Error(err)
				continue
			}
			trace(after, aResp)

			// unmashall & return
			d := Diff{e: e}

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

func merge(cs ...<-chan Diff) <-chan Diff {
	var wg sync.WaitGroup
	out := make(chan Diff)

	output := func(c <-chan Diff) {
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

func trace(req *http.Request, resp *http.Response) {
	reqStr, _ := httputil.DumpRequestOut(req, true)
	log.Tracef("---TRACE REQUEST---\n%s\n--- END ---\n\n", reqStr)

	respStr, _ := httputil.DumpResponse(resp, true)
	log.Tracef("---TRACE RESPONSE---\n%s\n--- END ---\n\n", respStr)
}
