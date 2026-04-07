package goddgs

import (
	"path/filepath"
	"testing"
)

func TestAdaptiveParserSaveLoad(t *testing.T) {
	p := NewAdaptiveParser()
	if err := p.Register("item", "a.item"); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "selectors.json")
	if err := p.Save(path); err != nil {
		t.Fatal(err)
	}
	q := NewAdaptiveParser()
	if err := q.Load(path); err != nil {
		t.Fatal(err)
	}
	q.mu.RLock()
	defer q.mu.RUnlock()
	if _, ok := q.selectors["item"]; !ok {
		t.Fatal("expected loaded selector")
	}
}
