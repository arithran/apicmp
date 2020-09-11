package diff

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"

	"github.com/hashicorp/go-retryablehttp"
	log "github.com/sirupsen/logrus"
)

type httpClient interface {
	Do(req *retryablehttp.Request) (*http.Response, error)
}

func newRetriableHTTPClient(retry map[int]struct{}) httpClient {
	c := retryablehttp.NewClient()
	c.Logger = log.New()
	// c.CheckRetry = newRetryPolicy(retry)
	c.RetryMax = 0
	return c
}

func newRetryPolicy(retry map[int]struct{}) retryablehttp.CheckRetry {
	return func(ctx context.Context, resp *http.Response, err error) (bool, error) {
		fmt.Println("--- DEBUG --- in newRetrypolicy")
		// do not retry on context.Canceled or context.DeadlineExceeded
		if ctx.Err() != nil {
			return false, ctx.Err()
		}

		if _, ok := retry[resp.StatusCode]; ok {
			log.Info("Retrying")
			return true, nil
		}

		return false, nil
	}
}

func httpTrace(req *http.Request, resp *http.Response) {
	reqStr, _ := httputil.DumpRequestOut(req, true)
	log.Tracef("---TRACE REQUEST---\n%s\n--- END ---\n\n", reqStr)

	respStr, _ := httputil.DumpResponse(resp, true)
	log.Tracef("---TRACE RESPONSE---\n%s\n--- END ---\n\n", respStr)
}
