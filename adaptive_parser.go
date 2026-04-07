package goddgs

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/andybalholm/cascadia"
	"golang.org/x/net/html"
)

// AdaptiveSelector stores a stable selector plus learned fingerprints.
type AdaptiveSelector struct {
	Name         string
	PrimaryCSS   string
	Fingerprints []ElementFingerprint
}

// ElementFingerprint describes a previously matched element for similarity fallback.
type ElementFingerprint struct {
	Tag      string
	ID       string
	Class    string
	TextHash string
}

// AdaptiveParser implements selector self-healing with lightweight similarity matching.
type AdaptiveParser struct {
	mu        sync.RWMutex
	selectors map[string]*AdaptiveSelector
}

func NewAdaptiveParser() *AdaptiveParser {
	return &AdaptiveParser{selectors: map[string]*AdaptiveSelector{}}
}

// Save persists adaptive selector state to disk as JSON.
func (p *AdaptiveParser) Save(path string) error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(p.selectors); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// Load restores adaptive selector state from disk.
func (p *AdaptiveParser) Load(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	var m map[string]*AdaptiveSelector
	if err := json.NewDecoder(f).Decode(&m); err != nil {
		return err
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if m == nil {
		m = map[string]*AdaptiveSelector{}
	}
	p.selectors = m
	return nil
}

func (p *AdaptiveParser) Register(name, css string) error {
	name = strings.TrimSpace(name)
	css = strings.TrimSpace(css)
	if name == "" || css == "" {
		return fmt.Errorf("name and css are required")
	}
	if _, err := cascadia.Compile(css); err != nil {
		return fmt.Errorf("invalid css selector %q: %w", css, err)
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.selectors[name] = &AdaptiveSelector{Name: name, PrimaryCSS: css}
	return nil
}

func (p *AdaptiveParser) Select(doc *html.Node, name string) (*html.Node, bool, error) {
	if doc == nil {
		return nil, false, fmt.Errorf("document is nil")
	}
	p.mu.RLock()
	sel, ok := p.selectors[name]
	p.mu.RUnlock()
	if !ok {
		return nil, false, fmt.Errorf("selector %q not registered", name)
	}

	compiled, err := cascadia.Compile(sel.PrimaryCSS)
	if err != nil {
		return nil, false, err
	}
	if node := cascadia.Query(doc, compiled); node != nil {
		p.learn(name, node)
		return node, false, nil
	}

	// Fallback: similarity search over DOM.
	node, score := p.findSimilar(doc, sel.Fingerprints)
	if node == nil || score < 0.55 {
		return nil, false, nil
	}

	if updated := cssCandidate(node); updated != "" {
		p.mu.Lock()
		if current, ok := p.selectors[name]; ok {
			current.PrimaryCSS = updated
		}
		p.mu.Unlock()
	}
	p.learn(name, node)
	return node, true, nil
}

func (p *AdaptiveParser) learn(name string, n *html.Node) {
	fp := fingerprintForNode(n)
	p.mu.Lock()
	defer p.mu.Unlock()
	sel, ok := p.selectors[name]
	if !ok {
		return
	}
	if len(sel.Fingerprints) >= 8 {
		sel.Fingerprints = sel.Fingerprints[1:]
	}
	sel.Fingerprints = append(sel.Fingerprints, fp)
}

func (p *AdaptiveParser) findSimilar(doc *html.Node, refs []ElementFingerprint) (*html.Node, float64) {
	if len(refs) == 0 {
		return nil, 0
	}
	var (
		bestNode  *html.Node
		bestScore float64
	)
	var visit func(*html.Node)
	visit = func(n *html.Node) {
		if n == nil {
			return
		}
		if n.Type == html.ElementNode {
			cand := fingerprintForNode(n)
			s := scoreFingerprint(cand, refs)
			if s > bestScore {
				bestScore = s
				bestNode = n
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			visit(c)
		}
	}
	visit(doc)
	return bestNode, bestScore
}

func scoreFingerprint(cand ElementFingerprint, refs []ElementFingerprint) float64 {
	best := 0.0
	for _, ref := range refs {
		s := 0.0
		if cand.Tag != "" && cand.Tag == ref.Tag {
			s += 0.35
		}
		if cand.ID != "" && cand.ID == ref.ID {
			s += 0.35
		}
		if cand.Class != "" && cand.Class == ref.Class {
			s += 0.2
		}
		if cand.TextHash != "" && cand.TextHash == ref.TextHash {
			s += 0.1
		}
		if s > best {
			best = s
		}
	}
	return best
}

func fingerprintForNode(n *html.Node) ElementFingerprint {
	fp := ElementFingerprint{Tag: strings.ToLower(strings.TrimSpace(n.Data))}
	for _, a := range n.Attr {
		switch strings.ToLower(a.Key) {
		case "id":
			fp.ID = strings.TrimSpace(a.Val)
		case "class":
			fp.Class = strings.Join(strings.Fields(a.Val), " ")
		}
	}
	text := strings.TrimSpace(nodeText(n))
	if text != "" {
		sum := sha1.Sum([]byte(text))
		fp.TextHash = hex.EncodeToString(sum[:])
	}
	return fp
}

func nodeText(n *html.Node) string {
	if n == nil {
		return ""
	}
	var b strings.Builder
	var visit func(*html.Node)
	visit = func(cur *html.Node) {
		if cur.Type == html.TextNode {
			b.WriteString(cur.Data)
			b.WriteByte(' ')
		}
		for c := cur.FirstChild; c != nil; c = c.NextSibling {
			visit(c)
		}
	}
	visit(n)
	return strings.Join(strings.Fields(b.String()), " ")
}

func cssCandidate(n *html.Node) string {
	if n == nil || n.Type != html.ElementNode {
		return ""
	}
	tag := strings.ToLower(n.Data)
	if tag == "" {
		return ""
	}
	for _, a := range n.Attr {
		if strings.EqualFold(a.Key, "id") && strings.TrimSpace(a.Val) != "" {
			return fmt.Sprintf("%s#%s", tag, strings.TrimSpace(a.Val))
		}
	}
	for _, a := range n.Attr {
		if strings.EqualFold(a.Key, "class") {
			classes := strings.Fields(a.Val)
			if len(classes) > 0 {
				return fmt.Sprintf("%s.%s", tag, classes[0])
			}
		}
	}
	return tag
}
