package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/user/pos-wms-mvp/services/erp-api/internal/domain"
)

// Service holds configuration and clients for ERP business logic.
// Note: ERP does NOT have a database repository. It is a read-only aggregator.
type Service struct {
	omsAPIURL        string
	scmAPIURL        string
	hrmAPIURL        string
	httpClient       *http.Client
	httpTimeout      time.Duration
	aggregateTimeout time.Duration
	httpRetries      int
}

var ErrInvalidPeriod = errors.New("invalid period format, expected YYYY-MM")

// NewService creates a new ERP service instance.
func NewService() *Service {
	httpTimeoutMs := getEnvInt("ERP_HTTP_TIMEOUT_MS", 5000)
	aggregateTimeoutMs := getEnvInt("ERP_AGGREGATE_TIMEOUT_MS", 7000)
	httpRetries := getEnvInt("ERP_HTTP_RETRIES", 1)
	if httpRetries < 0 {
		httpRetries = 0
	}

	return &Service{
		omsAPIURL:        getEnv("OMS_API_URL", "http://localhost:4003"),
		scmAPIURL:        getEnv("SCM_API_URL", "http://localhost:4004"),
		hrmAPIURL:        getEnv("HRM_API_URL", "http://localhost:4006"),
		httpClient:       &http.Client{Timeout: time.Duration(httpTimeoutMs) * time.Millisecond},
		httpTimeout:      time.Duration(httpTimeoutMs) * time.Millisecond,
		aggregateTimeout: time.Duration(aggregateTimeoutMs) * time.Millisecond,
		httpRetries:      httpRetries,
	}
}

// ============================================================================
// FINANCIAL AGGREGATION (SKELETON FOR STEP 4)
// ============================================================================

// GetFinancialSummary aggregates data from OMS, SCM, and HRM concurrently.
func (s *Service) GetFinancialSummary(ctx context.Context, req *domain.FinancialSummaryRequest) (*domain.FinancialSummary, error) {
	startTime := time.Now()
	ctx, cancel := context.WithTimeout(ctx, s.aggregateTimeout)
	defer cancel()

	if req == nil {
		req = &domain.FinancialSummaryRequest{}
	}

	period := req.Period
	if period == "" {
		period = time.Now().Format("2006-01")
	}

	periodStart, periodEnd, err := parsePeriodRange(period)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidPeriod, err)
	}

	type omsResult struct {
		revenue float64
		count   int
		latency int
		err     error
	}
	type scmResult struct {
		cogs    float64
		count   int
		latency int
		err     error
	}
	type hrmResult struct {
		payroll float64
		count   int
		latency int
		err     error
	}

	omsCh := make(chan omsResult, 1)
	scmCh := make(chan scmResult, 1)
	hrmCh := make(chan hrmResult, 1)

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		revenue, count, latency, err := s.fetchOMSSummary(ctx, periodStart, periodEnd)
		omsCh <- omsResult{revenue: revenue, count: count, latency: latency, err: err}
	}()

	go func() {
		defer wg.Done()
		cogs, count, latency, err := s.fetchSCMSummary(ctx, periodStart, periodEnd)
		scmCh <- scmResult{cogs: cogs, count: count, latency: latency, err: err}
	}()

	go func() {
		defer wg.Done()
		payroll, count, latency, err := s.fetchHRMSummary(ctx)
		hrmCh <- hrmResult{payroll: payroll, count: count, latency: latency, err: err}
	}()

	wg.Wait()

	oms := <-omsCh
	scm := <-scmCh
	hrm := <-hrmCh

	if oms.err != nil || scm.err != nil || hrm.err != nil {
		errParts := make([]string, 0, 3)
		if oms.err != nil {
			errParts = append(errParts, fmt.Sprintf("oms: %v", oms.err))
		}
		if scm.err != nil {
			errParts = append(errParts, fmt.Sprintf("scm: %v", scm.err))
		}
		if hrm.err != nil {
			errParts = append(errParts, fmt.Sprintf("hrm: %v", hrm.err))
		}
		return nil, fmt.Errorf("aggregation failed: %s", strings.Join(errParts, " | "))
	}

	netProfit := oms.revenue - scm.cogs - hrm.payroll
	profitMargin := 0.0
	if oms.revenue > 0 {
		profitMargin = (netProfit / oms.revenue) * 100
	}

	summary := &domain.FinancialSummary{
		Period:             period,
		Revenue:            oms.revenue,
		COGS:               scm.cogs,
		PayrollCost:        hrm.payroll,
		NetProfit:          netProfit,
		ProfitMargin:       profitMargin,
		RevenueSourceCount: oms.count,
		CostItemCount:      scm.count,
		EmployeeCount:      hrm.count,
		GeneratedAt:        time.Now(),
	}

	_ = time.Since(startTime)

	return summary, nil
}

type successEnvelope struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
	Error   string          `json:"error"`
}

type httpStatusError struct {
	StatusCode int
	Body       string
}

func (e *httpStatusError) Error() string {
	if e.Body == "" {
		return fmt.Sprintf("unexpected status code: %d", e.StatusCode)
	}
	return fmt.Sprintf("unexpected status code: %d, body: %s", e.StatusCode, e.Body)
}

type omsOrder struct {
	TotalAmount float64   `json:"total_amount"`
	CreatedAt   time.Time `json:"created_at"`
}

type scmPurchaseOrder struct {
	Quantity  int       `json:"quantity"`
	UnitCost  float64   `json:"unit_cost"`
	Cost      float64   `json:"cost"`
	LineTotal float64   `json:"line_total"`
	TotalCost float64   `json:"total_cost"`
	CreatedAt time.Time `json:"created_at"`
}

type hrmPayroll struct {
	TotalActiveSalary   float64 `json:"total_active_salary"`
	ActiveEmployeeCount int     `json:"active_employee_count"`
}

type scmSupplier struct {
	ID int `json:"id"`
}

func (s *Service) fetchOMSSummary(ctx context.Context, periodStart, periodEnd time.Time) (float64, int, int, error) {
	start := time.Now()
	limit := 200
	offset := 0
	totalRevenue := 0.0
	totalCount := 0

	for page := 0; page < 50; page++ {
		path := fmt.Sprintf("/api/orders?status=completed&limit=%d&offset=%d", limit, offset)
		data, err := s.doJSONGet(ctx, s.omsAPIURL, path)
		if err != nil {
			return 0, 0, int(time.Since(start).Milliseconds()), err
		}

		var orders []omsOrder
		if err := json.Unmarshal(data, &orders); err != nil {
			return 0, 0, int(time.Since(start).Milliseconds()), fmt.Errorf("decode oms orders: %w", err)
		}

		if len(orders) == 0 {
			break
		}

		for _, order := range orders {
			if !isWithinPeriod(order.CreatedAt, periodStart, periodEnd) {
				continue
			}
			totalRevenue += order.TotalAmount
			totalCount++
		}

		if len(orders) < limit {
			break
		}
		offset += limit
	}

	return totalRevenue, totalCount, int(time.Since(start).Milliseconds()), nil
}

func (s *Service) fetchSCMSummary(ctx context.Context, periodStart, periodEnd time.Time) (float64, int, int, error) {
	start := time.Now()
	limit := 200
	offset := 0
	totalCOGS := 0.0
	totalCount := 0

	for page := 0; page < 50; page++ {
		path := fmt.Sprintf("/scm/purchase-orders?limit=%d&offset=%d", limit, offset)
		data, err := s.doJSONGet(ctx, s.scmAPIURL, path)
		if err != nil {
			var statusErr *httpStatusError
			if errors.As(err, &statusErr) && statusErr.StatusCode == http.StatusNotFound {
				count, fallbackErr := s.fetchSCMSupplierCount(ctx)
				if fallbackErr != nil {
					return 0, 0, int(time.Since(start).Milliseconds()), fallbackErr
				}
				return 0, count, int(time.Since(start).Milliseconds()), nil
			}
			return 0, 0, int(time.Since(start).Milliseconds()), err
		}

		var orders []scmPurchaseOrder
		if err := json.Unmarshal(data, &orders); err != nil {
			return 0, 0, int(time.Since(start).Milliseconds()), fmt.Errorf("decode scm purchase orders: %w", err)
		}

		if len(orders) == 0 {
			break
		}

		for _, po := range orders {
			if !isWithinPeriod(po.CreatedAt, periodStart, periodEnd) {
				continue
			}
			totalCount++
			switch {
			case po.TotalCost > 0:
				totalCOGS += po.TotalCost
			case po.LineTotal > 0:
				totalCOGS += po.LineTotal
			case po.Cost > 0:
				totalCOGS += po.Cost
			case po.UnitCost > 0 && po.Quantity > 0:
				totalCOGS += po.UnitCost * float64(po.Quantity)
			}
		}

		if len(orders) < limit {
			break
		}
		offset += limit
	}

	return totalCOGS, totalCount, int(time.Since(start).Milliseconds()), nil
}

func (s *Service) fetchSCMSupplierCount(ctx context.Context) (int, error) {
	limit := 200
	offset := 0
	totalCount := 0

	for page := 0; page < 50; page++ {
		path := fmt.Sprintf("/scm/suppliers?limit=%d&offset=%d", limit, offset)
		data, err := s.doJSONGet(ctx, s.scmAPIURL, path)
		if err != nil {
			return 0, err
		}

		var suppliers []scmSupplier
		if err := json.Unmarshal(data, &suppliers); err != nil {
			return 0, fmt.Errorf("decode scm suppliers: %w", err)
		}

		if len(suppliers) == 0 {
			break
		}

		totalCount += len(suppliers)
		if len(suppliers) < limit {
			break
		}
		offset += limit
	}

	return totalCount, nil
}

func (s *Service) fetchHRMSummary(ctx context.Context) (float64, int, int, error) {
	start := time.Now()
	data, err := s.doJSONGet(ctx, s.hrmAPIURL, "/hrm/payroll")
	if err != nil {
		return 0, 0, int(time.Since(start).Milliseconds()), err
	}

	var payroll hrmPayroll
	if err := json.Unmarshal(data, &payroll); err != nil {
		return 0, 0, int(time.Since(start).Milliseconds()), fmt.Errorf("decode hrm payroll: %w", err)
	}

	return payroll.TotalActiveSalary, payroll.ActiveEmployeeCount, int(time.Since(start).Milliseconds()), nil
}

func (s *Service) doJSONGet(ctx context.Context, baseURL, path string) (json.RawMessage, error) {
	baseURL = strings.TrimRight(baseURL, "/")
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	if _, err := url.Parse(baseURL); err != nil {
		return nil, fmt.Errorf("invalid base url: %w", err)
	}
	fullURL := baseURL + path

	var lastErr error
	for attempt := 0; attempt <= s.httpRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(150 * time.Millisecond):
			}
		}

		reqCtx, cancel := context.WithTimeout(ctx, s.httpTimeout)
		req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, fullURL, nil)
		if err != nil {
			cancel()
			return nil, err
		}

		resp, err := s.httpClient.Do(req)
		if err != nil {
			cancel()
			lastErr = err
			continue
		}

		body, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		cancel()
		if readErr != nil {
			lastErr = readErr
			continue
		}

		if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
			statusErr := &httpStatusError{StatusCode: resp.StatusCode, Body: strings.TrimSpace(string(body))}
			if resp.StatusCode >= http.StatusInternalServerError && attempt < s.httpRetries {
				lastErr = statusErr
				continue
			}
			return nil, statusErr
		}

		env := &successEnvelope{}
		if err := json.Unmarshal(body, env); err != nil {
			return nil, fmt.Errorf("decode envelope: %w", err)
		}

		if !env.Success {
			if env.Error == "" {
				env.Error = "upstream returned unsuccessful response"
			}
			return nil, errors.New(env.Error)
		}

		return env.Data, nil
	}

	if lastErr == nil {
		lastErr = errors.New("request failed")
	}
	return nil, lastErr
}

func parsePeriodRange(period string) (time.Time, time.Time, error) {
	start, err := time.Parse("2006-01", period)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	start = start.UTC()
	end := start.AddDate(0, 1, 0)
	return start, end, nil
}

func isWithinPeriod(createdAt, periodStart, periodEnd time.Time) bool {
	if createdAt.IsZero() {
		return true
	}
	createdAt = createdAt.UTC()
	return (createdAt.Equal(periodStart) || createdAt.After(periodStart)) && createdAt.Before(periodEnd)
}

// ============================================================================
// SERVICE URLs (For STEP 4 integration)
// ============================================================================

// GetOMSAPIURL returns the OMS API URL for revenue aggregation
func (s *Service) GetOMSAPIURL() string {
	return s.omsAPIURL
}

// GetSCMAPIURL returns the SCM API URL for COGS aggregation
func (s *Service) GetSCMAPIURL() string {
	return s.scmAPIURL
}

// GetHRMAPIURL returns the HRM API URL for payroll aggregation
func (s *Service) GetHRMAPIURL() string {
	return s.hrmAPIURL
}

// ============================================================================
// UTILITY FUNCTIONS
// ============================================================================

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
