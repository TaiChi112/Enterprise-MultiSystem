package domain

import "time"

// SalesTrendRequest controls the period window to analyze.
type SalesTrendRequest struct {
	Period string `json:"period"`
	Months int    `json:"months"`
}

// MonthlyFinancialPoint holds one month summary pulled from ERP.
type MonthlyFinancialPoint struct {
	Period    string  `json:"period"`
	Revenue   float64 `json:"revenue"`
	COGS      float64 `json:"cogs"`
	Payroll   float64 `json:"payroll"`
	NetProfit float64 `json:"net_profit"`
}

// SalesTrendInsight is the DSS analytical response for sales trend.
type SalesTrendInsight struct {
	PeriodAnchor          string                  `json:"period_anchor"`
	MonthsAnalyzed        int                     `json:"months_analyzed"`
	RevenueMoMGrowthPct   float64                 `json:"revenue_mom_growth_pct"`
	NetProfitMoMGrowthPct float64                 `json:"net_profit_mom_growth_pct"`
	TrendDirection        string                  `json:"trend_direction"`
	Points                []MonthlyFinancialPoint `json:"points"`
	GeneratedAt           time.Time               `json:"generated_at"`
}

// SuccessResponse is the standard success response envelope.
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// ErrorResponse is the standard error response envelope.
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
}
