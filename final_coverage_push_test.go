package goddgs

import (
	"testing"
	"time"
)

func TestSessionHubCustomTTLAndDynamicCloseNil(t *testing.T) {
	h := NewSessionHub(5 * time.Second)
	if h.ttl != 5*time.Second {
		t.Fatalf("ttl=%v", h.ttl)
	}
	d := &DynamicFetcher{}
	if err := d.Close(); err != nil {
		t.Fatal(err)
	}
}
