package diff

import (
	"net/http"
	"net/http/httputil"

	log "github.com/sirupsen/logrus"
)

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type retriableHttpClient struct {
	client httpClient
	retry  map[int]struct{}
}

func newRetriableHTTPClient(client httpClient, retry map[int]struct{}) httpClient {
	return &retriableHttpClient{
		client: client,
		retry:  retry,
	}
}

func (c *retriableHttpClient) Do(req *http.Request) (*http.Response, error) {
	resp, err := c.client.Do(req)
	if err == nil {
		if _, ok := c.retry[resp.StatusCode]; ok {
			log.Info("Retrying")
			return c.client.Do(req)
		}
	}

	return resp, err
}

func httpTrace(req *http.Request, resp *http.Response) {
	reqStr, _ := httputil.DumpRequestOut(req, true)
	log.Tracef("---TRACE REQUEST---\n%s\n--- END ---\n\n", reqStr)

	respStr, _ := httputil.DumpResponse(resp, true)
	log.Tracef("---TRACE RESPONSE---\n%s\n--- END ---\n\n", respStr)
}
