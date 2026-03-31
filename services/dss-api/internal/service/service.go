package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/user/pos-wms-mvp/services/dss-api/internal/domain"
)

// Service contains DSS analytics logic over ERP aggregated data.
type Service struct {
	erpAPIURL   string
	httpClient  *http.Client
	timeout     time.Duration
	httpRetries int
}

type erpEnvelope struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
	Error   string          `json:"error"`
}

type erpFinancialSummary struct {
	Period    string  `json:"period"`
	Revenue   float64 `json:"revenue"`
	COGS      float64 `json:"cogs"`
	Payroll   float64 `json:"payroll_cost"`
	NetProfit float64 `json:"net_profit"`
}

// NewService creates a DSS service instance.
func NewService() *Service {
	timeoutMs := getEnvInt("DSS_HTTP_TIMEOUT_MS", 4000)
	retries := getEnvInt("DSS_HTTP_RETRIES", 1)
	if retries < 0 {
		retries = 0
	}

	return &Service{
		erpAPIURL:   getEnv("ERP_API_URL", "http://localhost:4007"),
		httpClient:  &http.Client{Timeout: time.Duration(timeoutMs) * time.Millisecond},
		timeout:     time.Duration(timeoutMs) * time.Millisecond,
		httpRetries: retries,
	}
}

// GetSalesTrendInsights pulls monthly ERP data concurrently and calculates MoM growth.
func (s *Service) GetSalesTrendInsights(ctx context.Context, req *domain.SalesTrendRequest) (*domain.SalesTrendInsight, error) {
	period := time.Now().UTC().Format("2006-01")
	months := 3
	if req != nil {
		if req.Period != "" {
			period = req.Period
		}
		if req.Months > 0 {
			months = req.Months
		}
	}

	if months < 2 || months > 12 {
		return nil, fmt.Errorf("months must be between 2 and 12")
	}

	anchor, err := time.Parse("2006-01", period)
	if err != nil {
		return nil, fmt.Errorf("invalid period format, expected YYYY-MM")
	}
	anchor = anchor.UTC()

	periods := buildPeriods(anchor, months)
	type result struct {
		idx   int
		point domain.MonthlyFinancialPoint
		err   error
	}

	results := make(chan result, len(periods))
	var wg sync.WaitGroup
	wg.Add(len(periods))

	for idx, p := range periods {
		idx := idx
		p := p
		go func() {
			defer wg.Done()
			summary, fetchErr := s.fetchERPSummary(ctx, p)
			if fetchErr != nil {
				results <- result{idx: idx, err: fetchErr}
				return
			}
			results <- result{
				idx: idx,
				point: domain.MonthlyFinancialPoint{
					Period:    summary.Period,
					Revenue:   summary.Revenue,
					COGS:      summary.COGS,
					Payroll:   summary.Payroll,
					NetProfit: summary.NetProfit,
				},
			}
		}()
	}

	wg.Wait()
	close(results)

	points := make([]domain.MonthlyFinancialPoint, len(periods))
	for res := range results {
		if res.err != nil {
			return nil, res.err
		}
		points[res.idx] = res.point
	}

	if len(points) < 2 {
		return nil, fmt.Errorf("insufficient data points for trend analysis")
	}

	prev := points[len(points)-2]
	curr := points[len(points)-1]
	revenueGrowth := calcGrowthPct(prev.Revenue, curr.Revenue)
	netProfitGrowth := calcGrowthPct(prev.NetProfit, curr.NetProfit)

	trendDirection := "flat"
	if curr.Revenue > points[0].Revenue {
		trendDirection = "up"
	} else if curr.Revenue < points[0].Revenue {
		trendDirection = "down"
	}

	return &domain.SalesTrendInsight{
		PeriodAnchor:          anchor.Format("2006-01"),
		MonthsAnalyzed:        months,
		RevenueMoMGrowthPct:   revenueGrowth,
		NetProfitMoMGrowthPct: netProfitGrowth,
		TrendDirection:        trendDirection,
		Points:                points,
		GeneratedAt:           time.Now().UTC(),
	}, nil
}

func (s *Service) fetchERPSummary(ctx context.Context, period string) (*erpFinancialSummary, error) {
	path := "/erp/financial-summary?period=" + url.QueryEscape(period)
	data, err := s.doJSONGet(ctx, s.erpAPIURL, path)
	if err != nil {
		return nil, err
	}

	summary := &erpFinancialSummary{}
	if err := json.Unmarshal(data, summary); err != nil {
		return nil, fmt.Errorf("decode erp summary failed: %w", err)
	}
	if summary.Period == "" {
		summary.Period = period
	}
	return summary, nil
}

func (s *Service) doJSONGet(ctx context.Context, baseURL, path string) ([]byte, error) {
	targetURL := stringsTrimRightSlash(baseURL) + path
	var lastErr error

	for attempt := 0; attempt <= s.httpRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
		if err != nil {
			return nil, err
		}

		res, err := s.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		body, readErr := io.ReadAll(res.Body)
		res.Body.Close()
		if readErr != nil {
			lastErr = readErr
			continue
		}

		if res.StatusCode < 200 || res.StatusCode >= 300 {
			lastErr = fmt.Errorf("unexpected ERP status: %d", res.StatusCode)
			continue
		}

		envelope := &erpEnvelope{}
		if err := json.Unmarshal(body, envelope); err != nil {
			return nil, fmt.Errorf("decode ERP envelope failed: %w", err)
		}
		if !envelope.Success {
			if envelope.Error != "" {
				return nil, fmt.Errorf("erp request failed: %s", envelope.Error)
			}
			return nil, fmt.Errorf("erp request failed")
		}

		return envelope.Data, nil
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("erp request failed")
	}
	return nil, lastErr
}

func buildPeriods(anchor time.Time, months int) []string {
	points := make([]string, 0, months)
	for i := months - 1; i >= 0; i-- {
		points = append(points, anchor.AddDate(0, -i, 0).Format("2006-01"))
	}
	return points
}

func calcGrowthPct(prev, curr float64) float64 {
	if prev == 0 {
		if curr == 0 {
			return 0
		}
		return 100
	}
	return ((curr - prev) / prev) * 100
}

func stringsTrimRightSlash(in string) string {
	if len(in) > 0 && in[len(in)-1] == '/' {
		return in[:len(in)-1]
	}
	return in
}

func getEnv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
