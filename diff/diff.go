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

	"github.com/nsf/jsondiff"
	"github.com/olekukonko/tablewriter"
	log "github.com/sirupsen/logrus"
)

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
	// set loglevel
	l, err := log.ParseLevel(c.LogLevel)
	if err != nil {
		return err
	}
	log.SetLevel(l)

	// generate events from csv
	eChan, err := ReadEvents(c.FixtureFilePath, parseRows(c.Rows)...)
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
			log.Errorf("StatusCodes didn't match\n Before:%s\n After:%s\n DMA: %s\n APIKey: %s\n Path: %s\n\n\n", d.beforeCode, d.afterCode, d.e.DMA, d.e.APIKey, d.e.Path)
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

			fmt.Printf("Request:\n DMA: %s\n APIKey: %s\n Path: %s\n", d.e.DMA, d.e.APIKey, d.e.Path)
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
}

func ReadEvents(csvFile string, selectedRows ...int) (<-chan Event, error) {
	// selectedRows is useful when re-running failed tests
	var mSelectedRows map[int]struct{}
	if len(selectedRows) > 0 {
		mSelectedRows = make(map[int]struct{})
		for _, v := range selectedRows {
			mSelectedRows[v] = struct{}{}
		}
	}

	f, err := os.Open(csvFile)
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

			out <- Event{
				Row:    cursor,
				DMA:    row[0],
				APIKey: row[1],
				Path:   row[2],
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
			before, err := http.NewRequestWithContext(ctx, "GET", c.BeforeBasePath+e.Path, nil)
			if err != nil {
				log.Error(err)
				continue
			}
			before.Header.Add(headerAPIKey, e.APIKey)
			before.Header.Add(headerUserDma, e.DMA)
			before.Header.Add(headerToken, c.AccessToken)

			after, err := http.NewRequestWithContext(ctx, "GET", c.AfterBasePath+e.Path, nil)
			if err != nil {
				log.Error(err)
				continue
			}
			after.Header.Add(headerAPIKey, e.APIKey)
			after.Header.Add(headerUserDma, e.DMA)
			after.Header.Add(headerToken, c.AccessToken)
			after.Header.Add("X-Feature-V1getvideobyidenabled", "true")

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
	// if resp.StatusCode >= 300 {
	reqStr, _ := httputil.DumpRequestOut(req, true)
	log.Tracef("---TRACE REQUEST---\n%s\n--- END ---\n\n", reqStr)

	respStr, _ := httputil.DumpResponse(resp, true)
	log.Tracef("---TRACE RESPONSE---\n%s\n--- END ---\n\n", respStr)

	// }
}
