package goddgs

import "testing"

func TestSessionHubDeleteAndLen(t *testing.T) {
	h := NewSessionHub(0)
	_ = h.Put("a", mockFetcher("m"))
	if got := h.Len(); got != 1 {
		t.Fatalf("len=%d want 1", got)
	}
	h.Delete("a")
	if got := h.Len(); got != 0 {
		t.Fatalf("len=%d want 0", got)
	}
}
