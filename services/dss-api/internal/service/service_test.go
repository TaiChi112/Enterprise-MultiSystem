package service

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/user/pos-wms-mvp/services/dss-api/internal/domain"
)

func TestGetSalesTrendInsights_CalculatesMoMGrowth(t *testing.T) {
	erpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		period := r.URL.Query().Get("period")
		revenueByPeriod := map[string]float64{
			"2026-01": 1000,
			"2026-02": 1200,
			"2026-03": 1500,
		}
		netProfitByPeriod := map[string]float64{
			"2026-01": 200,
			"2026-02": 300,
			"2026-03": 450,
		}
		revenue := revenueByPeriod[period]
		netProfit := netProfitByPeriod[period]
		cogs := revenue * 0.5
		payroll := revenue - cogs - netProfit

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(fmt.Sprintf(`{"success":true,"data":{"period":"%s","revenue":%.0f,"cogs":%.0f,"payroll_cost":%.0f,"net_profit":%.0f}}`, period, revenue, cogs, payroll, netProfit)))
	}))
	defer erpServer.Close()

	t.Setenv("ERP_API_URL", erpServer.URL)
	t.Setenv("DSS_HTTP_RETRIES", "0")
	t.Setenv("DSS_HTTP_TIMEOUT_MS", "1000")

	svc := NewService()
	res, err := svc.GetSalesTrendInsights(context.Background(), &domain.SalesTrendRequest{
		Period: "2026-03",
		Months: 3,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if res.PeriodAnchor != "2026-03" {
		t.Fatalf("expected anchor 2026-03, got %s", res.PeriodAnchor)
	}
	if len(res.Points) != 3 {
		t.Fatalf("expected 3 points, got %d", len(res.Points))
	}
	if res.Points[0].Period != "2026-01" || res.Points[1].Period != "2026-02" || res.Points[2].Period != "2026-03" {
		t.Fatalf("unexpected periods: %+v", res.Points)
	}

	if res.RevenueMoMGrowthPct != 25 {
		t.Fatalf("expected revenue growth 25, got %v", res.RevenueMoMGrowthPct)
	}
	if res.NetProfitMoMGrowthPct != 50 {
		t.Fatalf("expected net profit growth 50, got %v", res.NetProfitMoMGrowthPct)
	}
	if res.TrendDirection != "up" {
		t.Fatalf("expected trend direction up, got %s", res.TrendDirection)
	}
}

func TestGetSalesTrendInsights_InvalidPeriod(t *testing.T) {
	svc := NewService()
	_, err := svc.GetSalesTrendInsights(context.Background(), &domain.SalesTrendRequest{Period: "03-2026", Months: 3})
	if err == nil {
		t.Fatal("expected invalid period error")
	}
	if !strings.Contains(err.Error(), "invalid period format") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetSalesTrendInsights_InvalidMonths(t *testing.T) {
	svc := NewService()
	_, err := svc.GetSalesTrendInsights(context.Background(), &domain.SalesTrendRequest{Period: "2026-03", Months: 1})
	if err == nil {
		t.Fatal("expected invalid months error")
	}
	if !strings.Contains(err.Error(), "months must be between 2 and 12") {
		t.Fatalf("unexpected error: %v", err)
	}
}
