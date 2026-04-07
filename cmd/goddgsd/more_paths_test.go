package main

import (
	"errors"
	"net/http"
	"os"
	"testing"
)

func TestRunWithServeAdditionalPaths(t *testing.T) {
	stop := make(chan os.Signal, 1)
	stop <- os.Interrupt
	t.Setenv("GODDGS_DDG_BASE", "http://127.0.0.1:1")
	t.Setenv("GODDGS_LINKS_BASE", "http://127.0.0.1:1")
	t.Setenv("GODDGS_HTML_BASE", "http://127.0.0.1:1")
	t.Setenv("GODDGS_ADDR", ":0")
	if err := runWithServe(stop, nil); err != nil {
		t.Fatalf("runWithServe nil error: %v", err)
	}

	stop2 := make(chan os.Signal, 1)
	stop2 <- os.Interrupt
	if err := runWithServe(stop2, func(_ *http.Server) error { return errors.New("serve failed") }); err != nil {
		t.Fatalf("runWithServe custom error path returned: %v", err)
	}
}
