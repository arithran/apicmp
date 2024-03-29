package diff

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/arithran/jsondiff"
	"github.com/itchyny/gojq"
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
	BeforeBasePath     string
	AfterBasePath      string
	FixtureFilePath    string
	Headers            []string
	QueryStrings       []string
	IgnoreQueryStrings *regexp.Regexp // regex to remove matched query strings
	IgnoreFields       map[string]struct{}
	Rows               map[int]struct{}
	Retry              map[int]struct{}
	Match              string
	LogLevel           string
	Threads            int
	PostmanFilePath    string
	Jq                 string
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

	// parse query
	var jq *gojq.Query
	if c.Jq != "" {
		jq, err = gojq.Parse(c.Jq)
		if err != nil {
			return err
		}
	}

	// init assertion workers
	client := newRetriableHTTPClient(c.Retry)
	var wantMatch jsondiff.Difference
	switch c.Match {
	case "superset":
		wantMatch = jsondiff.SupersetMatch
	default:
		wantMatch = jsondiff.FullMatch
	}
	cs := make([]<-chan result, c.Threads)
	for i := 0; i < c.Threads; i++ {
		cs[i] = compare(ctx, client, tChan, c.IgnoreFields, wantMatch, jq)
	}

	collection := make([]test, 0)

	// compute results
	sum := Summary{
		Issues: map[string][]int{},
	}
	results := merge(cs...)
	for r := range results {
		sum.Count++

		if len(r.Diffs) > 0 {
			collection = append(collection, r.e)
			_ = tpl.ExecuteTemplate(os.Stdout, "curl", r.e)
			sum.FailedRows = append(sum.FailedRows, r.e.Row)

			fmt.Println("Diff:")
			for _, v := range r.Diffs {
				sum.Issues[v.Field] = append(sum.Issues[v.Field], r.e.Row)
				fmt.Println(v.Field + ":")
				if log.IsLevelEnabled(log.DebugLevel) {
					fmt.Println(v.Delta)
				} else {
					fmt.Println("Error: Not Equal")
				}
			}
			fmt.Printf("\n\n")
		} else {
			sum.Passed++
		}
	}

	if c.PostmanFilePath != "" {
		postman := PostmanV2{}
		err = postman.GenerateCollection(c.PostmanFilePath, collection)
		if err != nil {
			return fmt.Errorf("postman collection: %w", err)
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

func compare(ctx context.Context, client httpClient, tests <-chan test,
	ignore map[string]struct{}, wantMatch jsondiff.Difference, jq *gojq.Query) <-chan result {
	results := make(chan result)

	go func() {
		for t := range tests {
			r, err := exec(ctx, client, t, ignore, wantMatch, jq)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					log.Infof("row:%d was canceled", t.Row)
				} else {
					_ = tpl.ExecuteTemplate(os.Stdout, "curl", r.e)
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
