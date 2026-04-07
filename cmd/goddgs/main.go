package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/net/html"

	goddgs "github.com/velariumai/go-ddgs-stealth"
)

func main() {
	os.Exit(run(os.Args))
}

func run(argv []string) int {
	if len(argv) < 2 {
		usage()
		return 2
	}
	switch argv[1] {
	case "search":
		return runSearch(argv[2:])
	case "providers":
		return runProviders()
	case "doctor":
		return runDoctor()
	case "stealth-fetch":
		return runStealthFetch(argv[2:])
	case "stealth-crawl":
		return runStealthCrawl(argv[2:])
	case "repl":
		return runREPL()
	default:
		usage()
		return 2
	}
}

func usage() {
	fmt.Println("goddgs commands: search, providers, doctor, stealth-fetch, stealth-crawl, repl")
}

func runSearch(args []string) int {
	fs := flag.NewFlagSet("search", flag.ContinueOnError)
	q := fs.String("q", "", "query")
	max := fs.Int("max", 10, "max results")
	region := fs.String("region", "us-en", "region")
	asJSON := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		fmt.Println(err)
		return 2
	}
	if strings.TrimSpace(*q) == "" {
		fmt.Println("query is required (--q)")
		return 2
	}
	cfg := goddgs.LoadConfigFromEnv()
	engine, err := goddgs.NewDefaultEngineFromConfig(cfg)
	if err != nil {
		fmt.Println("engine init error:", err)
		return 4
	}
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()
	res, err := engine.Search(ctx, goddgs.SearchRequest{Query: *q, MaxResults: *max, Region: *region})
	if err != nil {
		if goddgs.IsBlocked(err) {
			fmt.Println("blocked by target protection:", err)
			return 2
		}
		var se *goddgs.SearchError
		if errors.As(err, &se) {
			fmt.Println(se.Error())
			switch se.Kind {
			case goddgs.ErrKindInvalidInput:
				return 2
			case goddgs.ErrKindNoResults:
				return 2
			}
		}
		fmt.Println("search error:", err)
		return 4
	}
	if *asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(res)
		return 0
	}
	fmt.Printf("Provider: %s (fallback=%v)\n", res.Provider, res.FallbackUsed)
	for i, r := range res.Results {
		fmt.Printf("%d. %s\n   %s\n", i+1, r.Title, r.URL)
	}
	return 0
}

func runProviders() int {
	cfg := goddgs.LoadConfigFromEnv()
	engine, err := goddgs.NewDefaultEngineFromConfig(cfg)
	if err != nil {
		fmt.Println("engine init error:", err)
		return 4
	}
	enabled := engine.EnabledProviders()
	fmt.Println("Enabled providers:", strings.Join(enabled, ", "))
	return 0
}

func runDoctor() int {
	cfg := goddgs.LoadConfigFromEnv()
	fmt.Println("Timeout:", cfg.Timeout)
	fmt.Println("Max retries:", cfg.MaxRetries)
	fmt.Println("Provider order:", strings.Join(cfg.ProviderOrder, ","))
	fmt.Println("Brave key set:", cfg.BraveAPIKey != "")
	fmt.Println("Tavily key set:", cfg.TavilyAPIKey != "")
	fmt.Println("SerpAPI key set:", cfg.SerpAPIKey != "")
	fmt.Println("Stealth level env:", getenv("GODDGS_STEALTH_LEVEL", "strong"))
	fmt.Println("Stealth headless env:", getenv("GODDGS_STEALTH_HEADLESS", "true"))
	fmt.Println("Stealth profile dir:", getenv("GODDGS_STEALTH_PROFILE_DIR", ""))

	engine, err := goddgs.NewDefaultEngineFromConfig(cfg)
	if err != nil {
		fmt.Println("engine init error:", err)
		return 4
	}
	ctx, cancel := context.WithTimeout(context.Background(), min(cfg.Timeout, 10*time.Second))
	defer cancel()
	_, err = engine.Search(ctx, goddgs.SearchRequest{Query: "golang", MaxResults: 1, Region: "us-en"})
	if err != nil {
		fmt.Println("probe failed:", err)
		return 3
	}

	hf, err := goddgs.NewHTTPFetcher(goddgs.StealthOptions{AntiBotConfig: goddgs.NewAntiBotConfig()})
	if err != nil {
		fmt.Println("stealth http init failed:", err)
		return 3
	}
	_, err = hf.Fetch(ctx, goddgs.FetchRequest{Method: "GET", URL: cfg.DuckDuckGoBase + "/"})
	if err != nil {
		fmt.Println("stealth http probe failed:", err)
	} else {
		fmt.Println("stealth http probe ok")
	}

	fmt.Println("probe ok")
	return 0
}

func runStealthFetch(args []string) int {
	fs := flag.NewFlagSet("stealth-fetch", flag.ContinueOnError)
	rawURL := fs.String("url", "", "target url")
	mode := fs.String("mode", "http", "http|stealth")
	asJSON := fs.Bool("json", false, "json output")
	level := fs.String("level", "strong", "basic|strong|aggressive")
	human := fs.Bool("human", false, "enable human-like pacing")
	if err := fs.Parse(args); err != nil {
		fmt.Println(err)
		return 2
	}
	if strings.TrimSpace(*rawURL) == "" {
		fmt.Println("url is required (--url)")
		return 2
	}
	stealthOpts := goddgs.StealthOptions{
		AntiBotConfig:     goddgs.NewAntiBotConfig(),
		HumanLikeBehavior: *human,
		StealthLevel:      goddgs.StealthLevel(strings.ToLower(strings.TrimSpace(*level))),
	}

	var fetcher goddgs.Fetcher
	if strings.EqualFold(*mode, "stealth") {
		sf, err := goddgs.NewStealthyFetcher(stealthOpts)
		if err != nil {
			fmt.Println("stealth fetcher init failed:", err)
			return 4
		}
		defer sf.Close()
		fetcher = sf
	} else {
		hf, err := goddgs.NewHTTPFetcher(stealthOpts)
		if err != nil {
			fmt.Println("http fetcher init failed:", err)
			return 4
		}
		fetcher = hf
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	res, err := fetcher.Fetch(ctx, goddgs.FetchRequest{Method: "GET", URL: *rawURL})
	if err != nil {
		fmt.Println("fetch failed:", err)
		return 3
	}
	if *asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(res)
		return 0
	}
	fmt.Printf("Fetcher: %s\nStatus: %d\nFinal URL: %s\nBytes: %d\n", res.Fetcher, res.StatusCode, res.FinalURL, len(res.Body))
	return 0
}

func runStealthCrawl(args []string) int {
	fs := flag.NewFlagSet("stealth-crawl", flag.ContinueOnError)
	startURL := fs.String("url", "", "start url")
	maxPages := fs.Int("max", 50, "max pages")
	concurrency := fs.Int("concurrency", 4, "worker count")
	out := fs.String("out", "", "optional jsonl output path")
	if err := fs.Parse(args); err != nil {
		fmt.Println(err)
		return 2
	}
	if strings.TrimSpace(*startURL) == "" {
		fmt.Println("url is required (--url)")
		return 2
	}

	hf, err := goddgs.NewHTTPFetcher(goddgs.StealthOptions{AntiBotConfig: goddgs.NewAntiBotConfig()})
	if err != nil {
		fmt.Println("http fetcher init failed:", err)
		return 4
	}

	spider, err := goddgs.NewSpider(hf, goddgs.SpiderConfig{
		StartURLs:          []string{*startURL},
		Concurrency:        *concurrency,
		MaxPages:           *maxPages,
		StreamingJSONL:     strings.TrimSpace(*out),
		CheckpointFileJSON: strings.TrimSpace(*out) + ".checkpoint.json",
		Parse:              parseHTMLAnchors,
	})
	if err != nil {
		fmt.Println("spider init failed:", err)
		return 4
	}
	defer spider.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	if err := spider.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		fmt.Println("crawl failed:", err)
		return 3
	}
	fmt.Println("crawl completed")
	return 0
}

func runREPL() int {
	fmt.Println("goddgs REPL started. Commands: search <query> | fetch <url> | quit")
	s := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !s.Scan() {
			break
		}
		line := strings.TrimSpace(s.Text())
		if line == "" {
			continue
		}
		if line == "quit" || line == "exit" {
			return 0
		}
		parts := strings.SplitN(line, " ", 2)
		cmd := parts[0]
		arg := ""
		if len(parts) > 1 {
			arg = strings.TrimSpace(parts[1])
		}
		switch cmd {
		case "search":
			if runSearch([]string{"--q", arg}) != 0 {
				fmt.Println("search command failed")
			}
		case "fetch":
			if runStealthFetch([]string{"--url", arg}) != 0 {
				fmt.Println("fetch command failed")
			}
		default:
			fmt.Println("unknown command")
		}
	}
	if err := s.Err(); err != nil {
		fmt.Println("repl error:", err)
		return 1
	}
	return 0
}

func parseHTMLAnchors(_ context.Context, _ goddgs.CrawlResult, body []byte) ([]string, error) {
	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	links := make([]string, 0, 32)
	var visit func(*html.Node)
	visit = func(n *html.Node) {
		if n.Type == html.ElementNode && strings.EqualFold(n.Data, "a") {
			for _, a := range n.Attr {
				if strings.EqualFold(a.Key, "href") {
					links = append(links, a.Val)
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			visit(c)
		}
	}
	visit(doc)
	return links, nil
}

func getenv(k, def string) string {
	v := strings.TrimSpace(os.Getenv(k))
	if v == "" {
		return def
	}
	return v
}

func min(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
