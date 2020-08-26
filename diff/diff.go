package diff

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/olekukonko/tablewriter"
	log "github.com/sirupsen/logrus"
)

const (
	headerAPIKey       = "X-Api-Key"
	headerUserDma      = "X-User-Dma"
	headerToken        = "X-Access-Token"
	headerCacheControl = "Cache-Control"
)

type Summary struct {
	Count         int
	Passed        int
	Failed        int
	FailedRows    []int
	FailedRowsStr string
	Time          time.Duration
	Issues        map[string]int
}

type Config struct {
	BeforeBasePath  string
	AfterBasePath   string
	FixtureFilePath string
	AccessToken     string
	IgnoreFields    map[string]struct{}
	Rows            map[int]struct{}
	Retry           map[int]struct{}
	LogLevel        string
	Threads         int
}

// Cmp will compare the before and after
func Cmp(ctx context.Context, c Config) error {
	start := time.Now()

	if err := setLoglevel(c.LogLevel); err != nil {
		return err
	}

	// gen tests
	tChan, err := generateTests(c)
	if err != nil {
		return err
	}

	// init assertion workers
	client := newRetriableHTTPClient(&http.Client{}, c.Retry)
	cs := make([]<-chan result, c.Threads)
	for i := 0; i < c.Threads; i++ {
		cs[i] = assert(ctx, client, tChan, c.IgnoreFields)
	}

	// compute results
	sum := Summary{
		Issues: map[string]int{},
	}
	results := merge(cs...)
	for r := range results {
		sum.Count++

		if len(r.Diffs) > 0 {
			_ = tpl.ExecuteTemplate(os.Stdout, "curl", r.e)
			sum.FailedRows = append(sum.FailedRows, r.e.Row)
			table := tablewriter.NewWriter(os.Stdout)
			table.SetAutoFormatHeaders(false)
			table.SetHeader([]string{"Field", "Diff"})
			table.SetBorder(false)
			for _, v := range r.Diffs {
				sum.Issues[v.Field]++
				table.Append([]string{v.Field, v.Delta})
			}
			table.Render()
			fmt.Printf("\n\n")
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

	sum.Failed = sum.Count - sum.Passed
	sum.FailedRowsStr = strings.Join(failedRows, ",")
	sum.Time = time.Since(start)

	_ = tpl.ExecuteTemplate(os.Stdout, "summary", sum)
	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoFormatHeaders(false)
	table.SetHeader([]string{"Field", "Issues"})
	table.SetBorder(false)
	table.AppendBulk(sumTable)
	table.Render()

	return nil
}

func assert(ctx context.Context, client httpClient, tests <-chan test, ignore map[string]struct{}) <-chan result {
	results := make(chan result)

	go func() {
		for t := range tests {
			r, err := exec(ctx, client, t, ignore)
			if err != nil {
				log.Error(err)
				continue
			}
			results <- r
		}

		close(results)
	}()

	return results
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
