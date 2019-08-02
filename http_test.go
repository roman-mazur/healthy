package healthy

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHttpCheck_Run(t *testing.T) {
	var endpointHit bool

	var stubHttp http.HandlerFunc = func (rw http.ResponseWriter, r *http.Request) {
		endpointHit = true
		rw.WriteHeader(207)
	}
	server := httptest.NewServer(stubHttp)

	check := &HttpCheck{Url: server.URL, ExpectedStatusCode: 200, Timeout: 3 * time.Second}
	err := check.Run(context.Background())
	if err != nil {
		if !strings.Contains(err.Error(), "expected 200, got 207") {
			t.Errorf("Unexpected error message %s", err)
		}
	} else {
		t.Errorf("Error was expected, but check succeded")
	}

	if !endpointHit {
		t.Errorf("Endpoint is not hit!")
	}
}
