package goddgs

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

type solverRecoveryMock struct{}

func (solverRecoveryMock) Supports(signal BlockSignal) bool { return signal != BlockSignalNone }
func (solverRecoveryMock) Solve(ctx context.Context, pageURL string, info BlockInfo, body []byte) (*ChallengeSolution, error) {
	return &ChallengeSolution{}, nil
}

func TestHTTPFetcherRecoversViaSolver(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&hits, 1) == 1 {
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte("access denied"))
			return
		}
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	cfg := NewAntiBotConfig()
	cfg.ChallengeSolvers = []ChallengeSolver{solverRecoveryMock{}}
	f, err := NewHTTPFetcher(StealthOptions{AntiBotConfig: cfg})
	if err != nil {
		t.Fatal(err)
	}
	res, err := f.Fetch(context.Background(), FetchRequest{Method: "GET", URL: srv.URL})
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status=%d", res.StatusCode)
	}
}
