package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/user/pos-wms-mvp/services/erp-api/internal/service"
)

type erpResponseEnvelope struct {
	Success bool `json:"success"`
	Data    struct {
		Period             string  `json:"period"`
		Revenue            float64 `json:"revenue"`
		COGS               float64 `json:"cogs"`
		PayrollCost        float64 `json:"payroll_cost"`
		NetProfit          float64 `json:"net_profit"`
		ProfitMargin       float64 `json:"profit_margin"`
		RevenueSourceCount int     `json:"revenue_source_count"`
		CostItemCount      int     `json:"cost_item_count"`
		EmployeeCount      int     `json:"employee_count"`
	} `json:"data"`
}

type errorEnvelope struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

func TestGetFinancialSummary_EndToEndWithMockUpstream_PeriodFilter(t *testing.T) {
	omsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/orders" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":[
			{"total_amount":1000,"created_at":"2026-03-05T10:00:00Z"},
			{"total_amount":2000,"created_at":"2026-03-20T10:00:00Z"},
			{"total_amount":9999,"created_at":"2026-02-28T10:00:00Z"}
		]}`))
	}))
	defer omsServer.Close()

	scmServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/scm/purchase-orders" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":[
			{"quantity":10,"unit_cost":30,"total_cost":300,"created_at":"2026-03-02T10:00:00Z"},
			{"quantity":5,"unit_cost":20,"total_cost":100,"created_at":"2026-03-22T10:00:00Z"},
			{"quantity":8,"unit_cost":50,"total_cost":400,"created_at":"2026-01-15T10:00:00Z"}
		]}`))
	}))
	defer scmServer.Close()

	hrmServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/hrm/payroll" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":{"total_active_salary":1200,"active_employee_count":3}}`))
	}))
	defer hrmServer.Close()

	t.Setenv("OMS_API_URL", omsServer.URL)
	t.Setenv("SCM_API_URL", scmServer.URL)
	t.Setenv("HRM_API_URL", hrmServer.URL)
	t.Setenv("ERP_HTTP_TIMEOUT_MS", "1000")
	t.Setenv("ERP_AGGREGATE_TIMEOUT_MS", "2000")
	t.Setenv("ERP_HTTP_RETRIES", "0")

	svc := service.NewService()
	h := NewHandler(svc)
	app := fiber.New()
	h.RegisterRoutes(app)

	req := httptest.NewRequest(http.MethodGet, "/erp/financial-summary?period=2026-03", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("app test failed: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.StatusCode)
	}

	payload := &erpResponseEnvelope{}
	if err := json.NewDecoder(res.Body).Decode(payload); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}

	if !payload.Success {
		t.Fatalf("expected success=true")
	}
	if payload.Data.Period != "2026-03" {
		t.Fatalf("expected period 2026-03, got %s", payload.Data.Period)
	}
	if payload.Data.Revenue != 3000 {
		t.Fatalf("expected revenue 3000, got %v", payload.Data.Revenue)
	}
	if payload.Data.COGS != 400 {
		t.Fatalf("expected cogs 400, got %v", payload.Data.COGS)
	}
	if payload.Data.PayrollCost != 1200 {
		t.Fatalf("expected payroll 1200, got %v", payload.Data.PayrollCost)
	}
	if payload.Data.NetProfit != 1400 {
		t.Fatalf("expected net profit 1400, got %v", payload.Data.NetProfit)
	}
	if payload.Data.RevenueSourceCount != 2 {
		t.Fatalf("expected revenue source count 2, got %d", payload.Data.RevenueSourceCount)
	}
	if payload.Data.CostItemCount != 2 {
		t.Fatalf("expected cost item count 2, got %d", payload.Data.CostItemCount)
	}
	if payload.Data.EmployeeCount != 3 {
		t.Fatalf("expected employee count 3, got %d", payload.Data.EmployeeCount)
	}
}

func TestGetFinancialSummary_InvalidPeriod_ReturnsBadRequest(t *testing.T) {
	t.Setenv("OMS_API_URL", "http://127.0.0.1:1")
	t.Setenv("SCM_API_URL", "http://127.0.0.1:1")
	t.Setenv("HRM_API_URL", "http://127.0.0.1:1")

	svc := service.NewService()
	h := NewHandler(svc)
	app := fiber.New()
	h.RegisterRoutes(app)

	req := httptest.NewRequest(http.MethodGet, "/erp/financial-summary?period=2026-13", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("app test failed: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, res.StatusCode)
	}

	payload := &errorEnvelope{}
	if err := json.NewDecoder(res.Body).Decode(payload); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}
	if payload.Success {
		t.Fatalf("expected success=false")
	}
}

func TestGetFinancialSummary_PeriodFilter_AcrossPaginationPages(t *testing.T) {
	omsPage2Requested := false
	scmPage2Requested := false

	omsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/orders" {
			http.NotFound(w, r)
			return
		}

		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		w.Header().Set("Content-Type", "application/json")

		switch offset {
		case 0:
			rows := ""
			for i := 0; i < 200; i++ {
				if i > 0 {
					rows += ","
				}
				rows += `{"total_amount":1,"created_at":"2026-02-10T00:00:00Z"}`
			}
			_, _ = w.Write([]byte(`{"success":true,"data":[` + rows + `]}`))
		case 200:
			omsPage2Requested = true
			_, _ = w.Write([]byte(`{"success":true,"data":[
				{"total_amount":100,"created_at":"2026-03-05T10:00:00Z"},
				{"total_amount":300,"created_at":"2026-03-28T10:00:00Z"}
			]}`))
		default:
			_, _ = w.Write([]byte(`{"success":true,"data":[]}`))
		}
	}))
	defer omsServer.Close()

	scmServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/scm/purchase-orders" {
			http.NotFound(w, r)
			return
		}

		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		w.Header().Set("Content-Type", "application/json")

		switch offset {
		case 0:
			rows := ""
			for i := 0; i < 200; i++ {
				if i > 0 {
					rows += ","
				}
				rows += `{"quantity":1,"unit_cost":5,"total_cost":5,"created_at":"2026-04-10T00:00:00Z"}`
			}
			_, _ = w.Write([]byte(`{"success":true,"data":[` + rows + `]}`))
		case 200:
			scmPage2Requested = true
			_, _ = w.Write([]byte(`{"success":true,"data":[
				{"quantity":2,"unit_cost":20,"total_cost":40,"created_at":"2026-03-01T00:00:00Z"},
				{"quantity":3,"unit_cost":20,"total_cost":60,"created_at":"2026-03-31T23:59:59Z"}
			]}`))
		default:
			_, _ = w.Write([]byte(`{"success":true,"data":[]}`))
		}
	}))
	defer scmServer.Close()

	hrmServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/hrm/payroll" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":{"total_active_salary":50,"active_employee_count":2}}`))
	}))
	defer hrmServer.Close()

	t.Setenv("OMS_API_URL", omsServer.URL)
	t.Setenv("SCM_API_URL", scmServer.URL)
	t.Setenv("HRM_API_URL", hrmServer.URL)
	t.Setenv("ERP_HTTP_TIMEOUT_MS", "1000")
	t.Setenv("ERP_AGGREGATE_TIMEOUT_MS", "3000")
	t.Setenv("ERP_HTTP_RETRIES", "0")

	svc := service.NewService()
	h := NewHandler(svc)
	app := fiber.New()
	h.RegisterRoutes(app)

	req := httptest.NewRequest(http.MethodGet, "/erp/financial-summary?period=2026-03", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("app test failed: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.StatusCode)
	}

	payload := &erpResponseEnvelope{}
	if err := json.NewDecoder(res.Body).Decode(payload); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}

	if !omsPage2Requested || !scmPage2Requested {
		t.Fatalf("expected pagination to request second page for both OMS and SCM")
	}
	if payload.Data.Revenue != 400 {
		t.Fatalf("expected revenue 400 from page2 in-period rows, got %v", payload.Data.Revenue)
	}
	if payload.Data.COGS != 100 {
		t.Fatalf("expected cogs 100 from page2 in-period rows, got %v", payload.Data.COGS)
	}
	if payload.Data.PayrollCost != 50 {
		t.Fatalf("expected payroll 50, got %v", payload.Data.PayrollCost)
	}
	if payload.Data.RevenueSourceCount != 2 {
		t.Fatalf("expected revenue source count 2, got %d", payload.Data.RevenueSourceCount)
	}
	if payload.Data.CostItemCount != 2 {
		t.Fatalf("expected cost item count 2, got %d", payload.Data.CostItemCount)
	}
}

func TestGetFinancialSummary_PeriodFilter_UTCMonthBoundaries(t *testing.T) {
	omsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/orders" {
			http.NotFound(w, r)
			return
		}
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		if offset > 0 {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"success":true,"data":[]}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":[
			{"total_amount":900,"created_at":"2026-02-28T23:59:59Z"},
			{"total_amount":100,"created_at":"2026-03-01T00:00:00Z"},
			{"total_amount":200,"created_at":"2026-03-31T23:59:59Z"},
			{"total_amount":800,"created_at":"2026-04-01T00:00:00Z"}
		]}`))
	}))
	defer omsServer.Close()

	scmServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/scm/purchase-orders" {
			http.NotFound(w, r)
			return
		}
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		if offset > 0 {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"success":true,"data":[]}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":[
			{"quantity":1,"unit_cost":70,"total_cost":70,"created_at":"2026-02-28T23:59:59Z"},
			{"quantity":1,"unit_cost":30,"total_cost":30,"created_at":"2026-03-01T00:00:00Z"},
			{"quantity":1,"unit_cost":40,"total_cost":40,"created_at":"2026-03-31T23:59:59Z"},
			{"quantity":1,"unit_cost":80,"total_cost":80,"created_at":"2026-04-01T00:00:00Z"}
		]}`))
	}))
	defer scmServer.Close()

	hrmServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/hrm/payroll" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":{"total_active_salary":10,"active_employee_count":1}}`))
	}))
	defer hrmServer.Close()

	t.Setenv("OMS_API_URL", omsServer.URL)
	t.Setenv("SCM_API_URL", scmServer.URL)
	t.Setenv("HRM_API_URL", hrmServer.URL)
	t.Setenv("ERP_HTTP_TIMEOUT_MS", "1000")
	t.Setenv("ERP_AGGREGATE_TIMEOUT_MS", "3000")
	t.Setenv("ERP_HTTP_RETRIES", "0")

	svc := service.NewService()
	h := NewHandler(svc)
	app := fiber.New()
	h.RegisterRoutes(app)

	req := httptest.NewRequest(http.MethodGet, "/erp/financial-summary?period=2026-03", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("app test failed: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.StatusCode)
	}

	payload := &erpResponseEnvelope{}
	if err := json.NewDecoder(res.Body).Decode(payload); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}

	if payload.Data.Revenue != 300 {
		t.Fatalf("expected revenue 300 from UTC month boundaries, got %v", payload.Data.Revenue)
	}
	if payload.Data.COGS != 70 {
		t.Fatalf("expected cogs 70 from UTC month boundaries, got %v", payload.Data.COGS)
	}
	if payload.Data.RevenueSourceCount != 2 {
		t.Fatalf("expected revenue source count 2, got %d", payload.Data.RevenueSourceCount)
	}
	if payload.Data.CostItemCount != 2 {
		t.Fatalf("expected cost item count 2, got %d", payload.Data.CostItemCount)
	}
}

func TestGetFinancialSummary_ConcurrentUpstreamCalls_BasicPerformanceAssertion(t *testing.T) {
	var mu sync.Mutex
	arrivals := make(map[string]int)
	arrived := make(chan string, 3)
	release := make(chan struct{})

	makeBlockingServer := func(route string, body string) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != route {
				http.NotFound(w, r)
				return
			}

			mu.Lock()
			arrivals[route]++
			mu.Unlock()

			arrived <- route
			<-release

			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(body))
		}))
	}

	omsServer := makeBlockingServer("/api/orders", `{"success":true,"data":[{"total_amount":100,"created_at":"2026-03-10T10:00:00Z"}]}`)
	defer omsServer.Close()
	scmServer := makeBlockingServer("/scm/purchase-orders", `{"success":true,"data":[{"quantity":1,"unit_cost":40,"total_cost":40,"created_at":"2026-03-10T10:00:00Z"}]}`)
	defer scmServer.Close()
	hrmServer := makeBlockingServer("/hrm/payroll", `{"success":true,"data":{"total_active_salary":10,"active_employee_count":1}}`)
	defer hrmServer.Close()

	t.Setenv("OMS_API_URL", omsServer.URL)
	t.Setenv("SCM_API_URL", scmServer.URL)
	t.Setenv("HRM_API_URL", hrmServer.URL)
	t.Setenv("ERP_HTTP_TIMEOUT_MS", "2000")
	t.Setenv("ERP_AGGREGATE_TIMEOUT_MS", "3000")
	t.Setenv("ERP_HTTP_RETRIES", "0")

	svc := service.NewService()
	h := NewHandler(svc)
	app := fiber.New()
	h.RegisterRoutes(app)

	resCh := make(chan *http.Response, 1)
	errCh := make(chan error, 1)
	start := time.Now()
	go func() {
		req := httptest.NewRequest(http.MethodGet, "/erp/financial-summary?period=2026-03", nil)
		res, err := app.Test(req)
		if err != nil {
			errCh <- err
			return
		}
		resCh <- res
	}()

	seen := map[string]bool{}
	timer := time.NewTimer(400 * time.Millisecond)
	defer timer.Stop()
	for len(seen) < 3 {
		select {
		case route := <-arrived:
			seen[route] = true
		case <-timer.C:
			close(release)
			t.Fatalf("expected all 3 upstreams to be called concurrently within threshold, seen=%v", seen)
		}
	}

	close(release)

	select {
	case err := <-errCh:
		t.Fatalf("app test failed: %v", err)
	case res := <-resCh:
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, res.StatusCode)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("request did not complete in time")
	}

	elapsed := time.Since(start)
	if elapsed > 800*time.Millisecond {
		t.Fatalf("expected concurrent request completion under threshold, elapsed=%s", elapsed)
	}

	mu.Lock()
	defer mu.Unlock()
	if arrivals["/api/orders"] != 1 || arrivals["/scm/purchase-orders"] != 1 || arrivals["/hrm/payroll"] != 1 {
		t.Fatalf("expected each upstream called exactly once, arrivals=%v", arrivals)
	}
}

func TestGetFinancialSummary_TableDriven_PeriodCoverage(t *testing.T) {
	omsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/orders" {
			http.NotFound(w, r)
			return
		}
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		w.Header().Set("Content-Type", "application/json")
		if offset > 0 {
			_, _ = w.Write([]byte(`{"success":true,"data":[]}`))
			return
		}
		_, _ = w.Write([]byte(`{"success":true,"data":[
			{"total_amount":100,"created_at":"2026-02-10T00:00:00Z"},
			{"total_amount":200,"created_at":"2026-03-10T00:00:00Z"},
			{"total_amount":300,"created_at":"2026-04-10T00:00:00Z"}
		]}`))
	}))
	defer omsServer.Close()

	scmServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/scm/purchase-orders" {
			http.NotFound(w, r)
			return
		}
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		w.Header().Set("Content-Type", "application/json")
		if offset > 0 {
			_, _ = w.Write([]byte(`{"success":true,"data":[]}`))
			return
		}
		_, _ = w.Write([]byte(`{"success":true,"data":[
			{"quantity":1,"unit_cost":10,"total_cost":10,"created_at":"2026-02-15T00:00:00Z"},
			{"quantity":1,"unit_cost":20,"total_cost":20,"created_at":"2026-03-15T00:00:00Z"},
			{"quantity":1,"unit_cost":30,"total_cost":30,"created_at":"2026-04-15T00:00:00Z"}
		]}`))
	}))
	defer scmServer.Close()

	hrmServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/hrm/payroll" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":{"total_active_salary":50,"active_employee_count":5}}`))
	}))
	defer hrmServer.Close()

	t.Setenv("OMS_API_URL", omsServer.URL)
	t.Setenv("SCM_API_URL", scmServer.URL)
	t.Setenv("HRM_API_URL", hrmServer.URL)
	t.Setenv("ERP_HTTP_TIMEOUT_MS", "1000")
	t.Setenv("ERP_AGGREGATE_TIMEOUT_MS", "3000")
	t.Setenv("ERP_HTTP_RETRIES", "0")

	svc := service.NewService()
	h := NewHandler(svc)
	app := fiber.New()
	h.RegisterRoutes(app)

	tests := []struct {
		name                 string
		period               string
		expectedRevenue      float64
		expectedCOGS         float64
		expectedRevenueCount int
		expectedCostCount    int
	}{
		{name: "period_feb", period: "2026-02", expectedRevenue: 100, expectedCOGS: 10, expectedRevenueCount: 1, expectedCostCount: 1},
		{name: "period_mar", period: "2026-03", expectedRevenue: 200, expectedCOGS: 20, expectedRevenueCount: 1, expectedCostCount: 1},
		{name: "period_apr", period: "2026-04", expectedRevenue: 300, expectedCOGS: 30, expectedRevenueCount: 1, expectedCostCount: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/erp/financial-summary?period="+tt.period, nil)
			res, err := app.Test(req)
			if err != nil {
				t.Fatalf("app test failed: %v", err)
			}
			defer res.Body.Close()

			if res.StatusCode != http.StatusOK {
				t.Fatalf("expected status %d, got %d", http.StatusOK, res.StatusCode)
			}

			payload := &erpResponseEnvelope{}
			if err := json.NewDecoder(res.Body).Decode(payload); err != nil {
				t.Fatalf("decode response failed: %v", err)
			}

			if payload.Data.Period != tt.period {
				t.Fatalf("expected period %s, got %s", tt.period, payload.Data.Period)
			}
			if payload.Data.Revenue != tt.expectedRevenue {
				t.Fatalf("expected revenue %v, got %v", tt.expectedRevenue, payload.Data.Revenue)
			}
			if payload.Data.COGS != tt.expectedCOGS {
				t.Fatalf("expected cogs %v, got %v", tt.expectedCOGS, payload.Data.COGS)
			}
			if payload.Data.RevenueSourceCount != tt.expectedRevenueCount {
				t.Fatalf("expected revenue source count %d, got %d", tt.expectedRevenueCount, payload.Data.RevenueSourceCount)
			}
			if payload.Data.CostItemCount != tt.expectedCostCount {
				t.Fatalf("expected cost item count %d, got %d", tt.expectedCostCount, payload.Data.CostItemCount)
			}
			if payload.Data.PayrollCost != 50 {
				t.Fatalf("expected payroll cost 50, got %v", payload.Data.PayrollCost)
			}
		})
	}
}
