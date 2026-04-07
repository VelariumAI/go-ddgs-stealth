package main

import (
	"os"
	"testing"
)

func TestRunWrapperPath(t *testing.T) {
	stop := make(chan os.Signal, 1)
	stop <- os.Interrupt
	t.Setenv("GODDGS_DDG_BASE", "http://127.0.0.1:1")
	t.Setenv("GODDGS_LINKS_BASE", "http://127.0.0.1:1")
	t.Setenv("GODDGS_HTML_BASE", "http://127.0.0.1:1")
	t.Setenv("GODDGS_ADDR", ":0")
	_ = run(stop)
}
