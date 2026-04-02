package domain

import "time"

// ============================================================================
// FINANCIAL SUMMARY - Aggregated Financial Data for P&L Report
// ============================================================================

// FinancialSummary represents aggregated financial data from across the system.
// It pulls revenue from OMS/POS, COGS from SCM, and Payroll from HRM.
type FinancialSummary struct {
	Period             string    `json:"period"`
	Revenue            float64   `json:"revenue"`              // Total sales revenue from OMS/POS
	COGS               float64   `json:"cogs"`                 // Cost of goods sold from SCM
	PayrollCost        float64   `json:"payroll_cost"`         // Total payroll from HRM
	NetProfit          float64   `json:"net_profit"`           // Revenue - COGS - PayrollCost
	ProfitMargin       float64   `json:"profit_margin"`        // (NetProfit / Revenue) * 100
	RevenueSourceCount int       `json:"revenue_source_count"` // Total completed orders
	CostItemCount      int       `json:"cost_item_count"`      // Total transmitted POs
	EmployeeCount      int       `json:"employee_count"`       // Total active employees
	GeneratedAt        time.Time `json:"generated_at"`         // When report was generated
}

// ============================================================================
// REQUEST / RESPONSE DTOs
// ============================================================================

// FinancialSummaryRequest - Input DTO for requesting financial summary
type FinancialSummaryRequest struct {
	Period string `json:"period" validate:"omitempty"` // Optional: if not provided, use current month
}

// ============================================================================
// RESPONSE TYPES
// ============================================================================

// SuccessResponse - Standard success response
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// ErrorResponse - Standard error response
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
}

// AggregationStats - Statistics for debugging and monitoring aggregation performance
type AggregationStats struct {
	OMS_LatencyMs  int    `json:"oms_latency_ms"`
	SCM_LatencyMs  int    `json:"scm_latency_ms"`
	HRM_LatencyMs  int    `json:"hrm_latency_ms"`
	TotalLatencyMs int    `json:"total_latency_ms"`
	OMS_Error      string `json:"oms_error,omitempty"`
	SCM_Error      string `json:"scm_error,omitempty"`
	HRM_Error      string `json:"hrm_error,omitempty"`
}
