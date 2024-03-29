package diff

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/arithran/jsondiff"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/itchyny/gojq"
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

func exec(ctx context.Context, c httpClient, t test,
	ignore map[string]struct{}, wantMatch jsondiff.Difference, jq *gojq.Query) (result, error) {
	var err error
	res := result{
		e: t,
	}

	res.Before, err = newOutput(ctx, c, t.Before, jq)
	if err != nil {
		return res, err
	}
	res.After, err = newOutput(ctx, c, t.After, jq)
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

func newOutput(ctx context.Context, c httpClient, i input, jq *gojq.Query) (output, error) {
	o := output{}

	// request
	var err error
	var req *retryablehttp.Request
	if i.Body != "" {
		req, err = retryablehttp.NewRequest(i.Method, i.Path, []byte(i.Body))
	} else {
		req, err = retryablehttp.NewRequest(i.Method, i.Path, nil)
	}
	if err != nil {
		return o, err
	}
	for k, v := range i.Headers {
		req.Header.Add(k, v)
	}

	// response
	httpTraceReq(req)
	resp, err := c.Do(req.WithContext(ctx))
	if err != nil {
		return o, err
	}
	httpTraceResp(resp)

	// decode
	o.Code = resp.Status
	defer resp.Body.Close()
	if jq != nil {
		var body interface{}
		err = json.NewDecoder(resp.Body).Decode(&body)
		if err != nil {
			return o, err
		}
		o.Body, err = applyJqQueryToBody(jq, body)
		if err != nil {
			return o, err
		}
	} else {
		err = json.NewDecoder(resp.Body).Decode(&o.Body)
		if err != nil {
			return o, err
		}
	}

	return o, nil
}

func init() {
	opts = jsondiff.DefaultConsoleOptions()
	opts.PrintTypes = false
}
