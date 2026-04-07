package goddgs

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDynamicFetcherBasicFlow(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	df, err := NewDynamicFetcher(StealthOptions{AntiBotConfig: NewAntiBotConfig(), BrowserBinary: "/no/such/browser"})
	if err != nil {
		t.Fatal(err)
	}
	defer df.Close()

	res, err := df.FetchDynamic(context.Background(), DynamicFetchRequest{FetchRequest: FetchRequest{Method: "GET", URL: srv.URL}})
	if err != nil {
		t.Fatal(err)
	}
	if res == nil {
		t.Fatal("expected response")
	}
}
