package diff

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"

	"github.com/arithran/jsondiff"
	"github.com/hashicorp/go-retryablehttp"
)

var opts jsondiff.Options

type (
	result struct {
		e      test
		Before output
		After  output
		Diffs  []diff
	}
	diff struct {
		Field string
		Delta string
	}
)

func exec(ctx context.Context, c httpClient, t test, ignore map[string]struct{}, wantMatch jsondiff.Difference) (result, error) {
	var err error
	res := result{
		e: t,
	}

	res.Before, err = newOutput(ctx, c, t.Before)
	if err != nil {
		return res, err
	}
	res.After, err = newOutput(ctx, c, t.After)
	if err != nil {
		return res, err
	}

	// calculate diff only if status codes are equal
	if res.Before.Code == res.After.Code {
		for k, v := range res.Before.Body {
			if _, ok := ignore[k]; ok {
				continue
			}

			match, delta := jsondiff.Compare(res.After.Body[k], v, &opts)
			if match > wantMatch {
				res.Diffs = append(res.Diffs, diff{
					Field: k,
					Delta: cleanDiff(delta),
				})
			}
		}
	} else {
		res.Diffs = append(res.Diffs, diff{
			Field: "_http.StatusCode",
			Delta: fmt.Sprintf("StatusCodes didn't match,\n before: %s\n after : %s", res.Before.Code, res.After.Code),
		})
	}

	sort.Slice(res.Diffs, func(i, j int) bool {
		return res.Diffs[i].Field < res.Diffs[j].Field
	})
	return res, nil
}

type output struct {
	Code string
	Body map[string]json.RawMessage
}

func newOutput(ctx context.Context, c httpClient, i input) (output, error) {
	o := output{}

	// request
	var err error
	var req *http.Request
	if i.Body != "" {
		req, err = http.NewRequestWithContext(ctx, i.Method, i.Path, bytes.NewReader([]byte(i.Body)))
	} else {
		req, err = http.NewRequestWithContext(ctx, i.Method, i.Path, nil)
	}
	if err != nil {
		fmt.Println("HI NewRequestWithContext")
		return o, err
	}
	for k, v := range i.Headers {
		req.Header.Add(k, v)
	}
	retryableReq, err := retryablehttp.FromRequest(req)
	if err != nil {
		fmt.Println("HI FromRequest")
		return o, err
	}

	// response
	fmt.Println("HI Do")
	resp, err := c.Do(retryableReq)
	if err != nil {
		fmt.Println("err Do")
		return o, err
	}
	httpTrace(req, resp)

	// decode
	o.Code = resp.Status
	fmt.Println("HI Decode")
	err = json.NewDecoder(resp.Body).Decode(&o.Body)
	if err != nil {
		fmt.Println("err Decode")
		return o, err
	}
	resp.Body.Close()

	fmt.Println("HI Closed")
	return o, nil
}

func init() {
	opts = jsondiff.DefaultConsoleOptions()
	opts.PrintTypes = false
}
