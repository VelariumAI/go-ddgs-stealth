package goddgs

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestAdaptiveParser_SelectPrimaryAndHeal(t *testing.T) {
	p := NewAdaptiveParser()
	if err := p.Register("title", "h1.title"); err != nil {
		t.Fatal(err)
	}

	doc1, _ := html.Parse(strings.NewReader(`<html><body><h1 id="main-title" class="title">Hello</h1></body></html>`))
	n, healed, err := p.Select(doc1, "title")
	if err != nil || n == nil || healed {
		t.Fatalf("first select err=%v healed=%v node=nil?%v", err, healed, n == nil)
	}

	doc2, _ := html.Parse(strings.NewReader(`<html><body><h1 id="main-title" class="headline">Hello</h1></body></html>`))
	n, healed, err = p.Select(doc2, "title")
	if err != nil || n == nil || !healed {
		t.Fatalf("heal select err=%v healed=%v node=nil?%v", err, healed, n == nil)
	}
}
