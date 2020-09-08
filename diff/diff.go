package diff

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/arithran/jsondiff"
	"github.com/olekukonko/tablewriter"
	log "github.com/sirupsen/logrus"
)

type Summary struct {
	Count         int
	Passed        int
	Failed        int
	FailedRows    []int
	FailedRowsStr string
	Time          time.Duration
	Issues        map[string][]int
}

type Config struct {
	BeforeBasePath  string
	AfterBasePath   string
	FixtureFilePath string
	Headers         []string
	IgnoreFields    map[string]struct{}
	Rows            map[int]struct{}
	Retry           map[int]struct{}
	Match           string
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
	tChan, err := generateTests(ctx, c)
	if err != nil {
		return err
	}

	// init assertion workers
	client := newRetriableHTTPClient(&http.Client{}, c.Retry)
	var wantMatch jsondiff.Difference
	switch c.Match {
	case "superset":
		wantMatch = jsondiff.SupersetMatch
	default:
		wantMatch = jsondiff.FullMatch
	}
	cs := make([]<-chan result, c.Threads)
	for i := 0; i < c.Threads; i++ {
		cs[i] = assert(ctx, client, tChan, c.IgnoreFields, wantMatch)
	}

	// compute results
	sum := Summary{
		Issues: map[string][]int{},
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
				sum.Issues[v.Field] = append(sum.Issues[v.Field], r.e.Row)
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
		sumTable = append(sumTable, []string{k, strconv.Itoa(len(v)), Istoa(v, ",")})
	}
	sort.Sort(sortDelta(sumTable))
	sort.Ints(sum.FailedRows)

	sum.Failed = sum.Count - sum.Passed
	sum.FailedRowsStr = Istoa(sum.FailedRows, ",")
	sum.Time = time.Since(start)

	_ = tpl.ExecuteTemplate(os.Stdout, "summary", sum)
	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoFormatHeaders(false)
	table.SetHeader([]string{"Field", "Issues", "Rows"})
	table.SetBorder(false)
	table.AppendBulk(sumTable)
	table.Render()

	return nil
}

func assert(ctx context.Context, client httpClient, tests <-chan test, ignore map[string]struct{}, wantMatch jsondiff.Difference) <-chan result {
	results := make(chan result)

	go func() {
		for t := range tests {
			r, err := exec(ctx, client, t, ignore, wantMatch)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					log.Infof("row:%d was canceled", t.Row)
				} else {
					log.Errorf("row:%d err:%v", t.Row, err)
				}

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
