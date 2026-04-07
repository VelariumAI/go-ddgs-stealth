package goddgs

import (
	"testing"
	"time"
)

func TestSessionHubLifecycle(t *testing.T) {
	h := NewSessionHub(20 * time.Millisecond)
	if err := h.Put("s1", mockFetcher("http")); err != nil {
		t.Fatal(err)
	}
	if _, ok := h.Get("s1"); !ok {
		t.Fatal("expected session")
	}
	time.Sleep(30 * time.Millisecond)
	if _, ok := h.Get("s1"); ok {
		t.Fatal("expected session eviction")
	}
}
