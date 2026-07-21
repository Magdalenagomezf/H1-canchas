package mercadopago

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// rewriteRequester redirects every outgoing request to a local httptest.Server
// instead of the SDK's hardcoded production host, so CreatePreference/GetPayment
// can be exercised end-to-end without hitting the real Mercado Pago API.
type rewriteRequester struct {
	client *http.Client
	target *url.URL
}

func (r rewriteRequester) Do(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = r.target.Scheme
	req.URL.Host = r.target.Host

	return r.client.Do(req)
}

func newTestRequester(t *testing.T, server *httptest.Server) rewriteRequester {
	t.Helper()

	target, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("url.Parse(%q) error = %v", server.URL, err)
	}

	return rewriteRequester{client: server.Client(), target: target}
}
