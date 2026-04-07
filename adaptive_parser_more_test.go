package goddgs

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestCSSCandidateAndRegisterErrors(t *testing.T) {
	p := NewAdaptiveParser()
	if err := p.Register("", "a.link"); err == nil {
		t.Fatal("expected register error")
	}
	if err := p.Register("x", "???"); err == nil {
		t.Fatal("expected invalid css error")
	}

	doc, _ := html.Parse(strings.NewReader(`<div id="n1" class="c1 c2">x</div>`))
	var n *html.Node
	var find func(*html.Node)
	find = func(cur *html.Node) {
		if cur.Type == html.ElementNode && cur.Data == "div" {
			n = cur
			return
		}
		for c := cur.FirstChild; c != nil; c = c.NextSibling {
			find(c)
		}
	}
	find(doc)
	if got := cssCandidate(n); got != "div#n1" {
		t.Fatalf("cssCandidate=%q want div#n1", got)
	}
}
