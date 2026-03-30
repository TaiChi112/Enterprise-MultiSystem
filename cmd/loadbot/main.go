package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

type config struct {
	TargetURL      string
	Profile        string
	Workers        int
	TotalRequests  int
	BranchID       int
	ProductIDs     []int
	Quantity       int
	Discount       float64
	RequestTimeout time.Duration
}

type saleItem struct {
	ProductID int     `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Discount  float64 `json:"discount"`
}

type saleRequest struct {
	BranchID     int        `json:"branch_id"`
	CustomerName string     `json:"customer_name"`
	Items        []saleItem `json:"items"`
}

type result struct {
	Duration time.Duration
	Status   int
	Err      error
}

type summary struct {
	TotalRequests  int
	TotalCompleted int
	TotalErrors    int
	StatusCounts   map[int]int
	RPS            float64
	MinLatency     time.Duration
	AvgLatency     time.Duration
	P50Latency     time.Duration
	P95Latency     time.Duration
	P99Latency     time.Duration
	MaxLatency     time.Duration
	Elapsed        time.Duration
}

type profileConfig struct {
	Workers       int
	TotalRequests int
	Timeout       time.Duration
}

var profiles = map[string]profileConfig{
	"smoke": {
		Workers:       5,
		TotalRequests: 100,
		Timeout:       5 * time.Second,
	},
	"stress": {
		Workers:       100,
		TotalRequests: 5000,
		Timeout:       10 * time.Second,
	},
	"spike": {
		Workers:       300,
		TotalRequests: 10000,
		Timeout:       12 * time.Second,
	},
}

func main() {
	cfg, err := parseConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "configuration error: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	fmt.Println("POS & WMS Load Bot")
	fmt.Printf("target=%s profile=%s workers=%d requests=%d branch=%d product_ids=%v timeout=%s\n",
		cfg.TargetURL, cfg.Profile, cfg.Workers, cfg.TotalRequests, cfg.BranchID, cfg.ProductIDs, cfg.RequestTimeout)

	client := &http.Client{
		Timeout: cfg.RequestTimeout,
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			DialContext:           (&net.Dialer{Timeout: 5 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          cfg.Workers * 4,
			MaxIdleConnsPerHost:   cfg.Workers * 2,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   5 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	s, err := runLoad(ctx, client, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load execution failed: %v\n", err)
		os.Exit(1)
	}

	printSummary(s)
}

func parseConfig() (config, error) {
	var cfg config
	var productIDsRaw string

	flag.StringVar(&cfg.TargetURL, "url", "http://localhost:3000/api/sales", "Target endpoint URL")
	flag.StringVar(&cfg.Profile, "profile", "custom", "Load profile: custom|smoke|stress|spike")
	flag.IntVar(&cfg.Workers, "workers", 50, "Number of concurrent workers")
	flag.IntVar(&cfg.TotalRequests, "requests", 1000, "Total number of requests to send")
	flag.IntVar(&cfg.BranchID, "branch-id", 1, "Branch ID for sale request")
	flag.StringVar(&productIDsRaw, "product-ids", "1", "Comma separated product IDs, e.g. 1,2,3")
	flag.IntVar(&cfg.Quantity, "quantity", 1, "Quantity per product item")
	flag.Float64Var(&cfg.Discount, "discount", 0, "Discount per product item")
	flag.DurationVar(&cfg.RequestTimeout, "timeout", 10*time.Second, "HTTP request timeout")
	flag.Parse()

	if err := applyProfile(&cfg); err != nil {
		return cfg, err
	}

	if cfg.Workers <= 0 {
		return cfg, fmt.Errorf("workers must be > 0")
	}
	if cfg.TotalRequests <= 0 {
		return cfg, fmt.Errorf("requests must be > 0")
	}
	if cfg.BranchID <= 0 {
		return cfg, fmt.Errorf("branch-id must be > 0")
	}
	if cfg.Quantity <= 0 {
		return cfg, fmt.Errorf("quantity must be > 0")
	}
	if cfg.Discount < 0 {
		return cfg, fmt.Errorf("discount must be >= 0")
	}

	productIDs, err := parseProductIDs(productIDsRaw)
	if err != nil {
		return cfg, err
	}
	cfg.ProductIDs = productIDs

	return cfg, nil
}

func applyProfile(cfg *config) error {
	profile := strings.ToLower(strings.TrimSpace(cfg.Profile))
	if profile == "" {
		profile = "custom"
	}
	if profile == "custom" {
		cfg.Profile = "custom"
		return nil
	}

	p, ok := profiles[profile]
	if !ok {
		return fmt.Errorf("invalid profile: %s (allowed: custom, smoke, stress, spike)", cfg.Profile)
	}

	cfg.Profile = profile
	cfg.Workers = p.Workers
	cfg.TotalRequests = p.TotalRequests
	cfg.RequestTimeout = p.Timeout
	return nil
}

func parseProductIDs(raw string) ([]int, error) {
	parts := strings.Split(raw, ",")
	ids := make([]int, 0, len(parts))

	for _, p := range parts {
		v := strings.TrimSpace(p)
		if v == "" {
			continue
		}
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			return nil, fmt.Errorf("invalid product id: %q", v)
		}
		ids = append(ids, n)
	}

	if len(ids) == 0 {
		return nil, fmt.Errorf("product-ids must contain at least one id")
	}

	return ids, nil
}

func buildPayload(cfg config, requestID int) ([]byte, error) {
	items := make([]saleItem, 0, len(cfg.ProductIDs))
	for _, pid := range cfg.ProductIDs {
		items = append(items, saleItem{
			ProductID: pid,
			Quantity:  cfg.Quantity,
			Discount:  cfg.Discount,
		})
	}

	body := saleRequest{
		BranchID:     cfg.BranchID,
		CustomerName: fmt.Sprintf("LoadBot-%d", requestID),
		Items:        items,
	}

	return json.Marshal(body)
}

func runLoad(ctx context.Context, client *http.Client, cfg config) (summary, error) {
	jobs := make(chan int, cfg.Workers*2)
	results := make(chan result, cfg.Workers*4)

	var wg sync.WaitGroup
	var completed int64
	start := time.Now()

	for i := 0; i < cfg.Workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for reqID := range jobs {
				res := executeOnce(ctx, client, cfg, reqID)
				results <- res
				atomic.AddInt64(&completed, 1)
			}
		}(i + 1)
	}

	go func() {
		for i := 1; i <= cfg.TotalRequests; i++ {
			select {
			case <-ctx.Done():
				close(jobs)
				return
			case jobs <- i:
			}
		}
		close(jobs)
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	statusCounts := make(map[int]int)
	durations := make([]time.Duration, 0, cfg.TotalRequests)
	totalErrors := 0

	for r := range results {
		durations = append(durations, r.Duration)
		if r.Err != nil {
			totalErrors++
			statusCounts[0]++
			continue
		}
		statusCounts[r.Status]++
	}

	elapsed := time.Since(start)
	if elapsed <= 0 {
		elapsed = time.Millisecond
	}

	s := summarize(cfg.TotalRequests, int(completed), totalErrors, statusCounts, durations, elapsed)
	if ctx.Err() != nil {
		return s, fmt.Errorf("execution interrupted: %w", ctx.Err())
	}

	return s, nil
}

func executeOnce(ctx context.Context, client *http.Client, cfg config, requestID int) result {
	payload, err := buildPayload(cfg, requestID)
	if err != nil {
		return result{Err: err}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.TargetURL, bytes.NewReader(payload))
	if err != nil {
		return result{Err: err}
	}
	req.Header.Set("Content-Type", "application/json")

	start := time.Now()
	resp, err := client.Do(req)
	duration := time.Since(start)

	if err != nil {
		return result{Duration: duration, Err: err}
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)

	return result{Duration: duration, Status: resp.StatusCode}
}

func summarize(totalRequests, completed, totalErrors int, statusCounts map[int]int, durations []time.Duration, elapsed time.Duration) summary {
	s := summary{
		TotalRequests:  totalRequests,
		TotalCompleted: completed,
		TotalErrors:    totalErrors,
		StatusCounts:   statusCounts,
		Elapsed:        elapsed,
	}

	if elapsed > 0 {
		s.RPS = float64(completed) / elapsed.Seconds()
	}

	if len(durations) == 0 {
		return s
	}

	sorted := make([]time.Duration, len(durations))
	copy(sorted, durations)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })

	s.MinLatency = sorted[0]
	s.MaxLatency = sorted[len(sorted)-1]
	s.P50Latency = percentile(sorted, 0.50)
	s.P95Latency = percentile(sorted, 0.95)
	s.P99Latency = percentile(sorted, 0.99)

	var total time.Duration
	for _, d := range sorted {
		total += d
	}
	s.AvgLatency = total / time.Duration(len(sorted))

	return s
}

func percentile(sorted []time.Duration, p float64) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	if p <= 0 {
		return sorted[0]
	}
	if p >= 1 {
		return sorted[len(sorted)-1]
	}

	idx := int(math.Ceil((float64(len(sorted)) * p) - 1.0))
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

func printSummary(s summary) {
	fmt.Println("\nLoad test summary")
	fmt.Println("-----------------")
	fmt.Printf("Total requested  : %d\n", s.TotalRequests)
	fmt.Printf("Total completed  : %d\n", s.TotalCompleted)
	fmt.Printf("Total errors     : %d\n", s.TotalErrors)
	fmt.Printf("Elapsed          : %s\n", s.Elapsed)
	fmt.Printf("Throughput (RPS) : %.2f\n", s.RPS)

	if len(s.StatusCounts) > 0 {
		fmt.Println("Status counts:")
		codes := make([]int, 0, len(s.StatusCounts))
		for c := range s.StatusCounts {
			codes = append(codes, c)
		}
		sort.Ints(codes)
		for _, c := range codes {
			if c == 0 {
				fmt.Printf("  error: %d\n", s.StatusCounts[c])
				continue
			}
			fmt.Printf("  %d: %d\n", c, s.StatusCounts[c])
		}
	}

	fmt.Println("Latency:")
	fmt.Printf("  min : %s\n", s.MinLatency)
	fmt.Printf("  avg : %s\n", s.AvgLatency)
	fmt.Printf("  p50 : %s\n", s.P50Latency)
	fmt.Printf("  p95 : %s\n", s.P95Latency)
	fmt.Printf("  p99 : %s\n", s.P99Latency)
	fmt.Printf("  max : %s\n", s.MaxLatency)
}
