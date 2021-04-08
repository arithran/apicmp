package diff

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sort"

	"github.com/hashicorp/go-retryablehttp"
)

type outputType uint8

const (
	bodyOutput outputType = 1 << iota
	rawOutput
)

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

func exec(ctx context.Context, c httpClient, t test, outCmp outputComparer) (result, error) {
	var err error
	res := result{
		e: t,
	}

	res.Before, err = newOutput(ctx, c, t.Before, outCmp.outputType())
	if err != nil {
		return res, err
	}
	res.After, err = newOutput(ctx, c, t.After, outCmp.outputType())
	if err != nil {
		return res, err
	}

	// calculate diff only if status codes are equal
	if res.Before.Code == res.After.Code {
		if res.Before.StatusCode > 199 && res.Before.StatusCode < 300 {
			d, ok := outCmp.cmp(res.Before, res.After)
			if !ok {
				res.Diffs = append(res.Diffs, d)
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
	StatusCode int
	Code       string
	Body       map[string]json.RawMessage
	Raw        []byte
}

func newOutput(ctx context.Context, c httpClient, i input, outType outputType) (output, error) {
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
	o.StatusCode = resp.StatusCode
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return o, err
	}
	resp.Body.Close()
	if outType&bodyOutput != 0 {
		err = json.Unmarshal(data, &o.Body)
		if err != nil {
			return o, err
		}
	}
	if outType&rawOutput != 0 {
		o.Raw = data
	}
	return o, nil
}
