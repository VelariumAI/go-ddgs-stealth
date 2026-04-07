package bench

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	goddgs "github.com/velariumai/go-ddgs-stealth"
)

func BenchmarkHTTPFetcher(b *testing.B) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	f, err := goddgs.NewHTTPFetcher(goddgs.StealthOptions{AntiBotConfig: goddgs.NewAntiBotConfig()})
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := f.Fetch(context.Background(), goddgs.FetchRequest{Method: "GET", URL: srv.URL}); err != nil {
			b.Fatal(err)
		}
	}
}
