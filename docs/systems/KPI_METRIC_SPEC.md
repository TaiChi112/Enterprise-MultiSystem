# KPI Metric Specification (Grafana/BI-Ready)

สเปคนี้ใช้เป็นมาตรฐานสำหรับ implement dashboard ทั้ง Grafana และ BI tools โดยให้ metric naming, owner, และ data logic ชัดเจน

## 1) Metric Contract Standard
| Field | Description |
|---|---|
| metric_key | unique metric name (snake_case) |
| display_name | label บน dashboard |
| domain | business domain/system |
| owner_team | owner สำหรับ monthly review |
| direction | higher_is_better or lower_is_better |
| unit | %, sec, count, currency |
| grain | daily, weekly, monthly |
| baseline_method | median_6m, p95_median_6m, trailing_3m_avg |
| source_table_or_stream | source dataset |
| calculation_sql_pseudocode | สูตรคำนวณ |
| data_freshness_sla | SLA ของ data latency |
| alert_threshold_monthly | threshold เทียบ monthly target |
| alert_threshold_quarterly | threshold เทียบ quarterly target |

## 2) Core Metric Spec
| metric_key | display_name | domain | owner_team | direction | unit | grain | baseline_method | source_table_or_stream | calculation_sql_pseudocode | data_freshness_sla |
|---|---|---|---|---|---|---|---|---|---|---|
| gateway_availability_pct | Gateway Availability | api-gateway | Platform Engineering | higher_is_better | % | monthly | median_6m | gateway_request_log | uptime_minutes / total_minutes * 100 | D+1 08:00 |
| login_success_rate_pct | Login Success Rate | iam | Identity and Security | higher_is_better | % | monthly | median_6m | iam_auth_events | success_logins / total_login_attempts * 100 | D+1 08:00 |
| checkout_success_rate_pct | Checkout Success Rate | pos | Retail Operations Product | higher_is_better | % | monthly | median_6m | pos_sales_events | completed_sales / initiated_checkouts * 100 | D+1 08:00 |
| repeat_purchase_rate_pct | Repeat Purchase Rate | crm | Customer Growth | higher_is_better | % | monthly | trailing_3m_avg | crm_customer_orders | repeat_customers / active_customers * 100 | D+2 10:00 |
| order_lifecycle_completion_pct | Order Lifecycle Completion | oms | Order Orchestration | higher_is_better | % | monthly | median_6m | oms_orders | completed_orders / created_orders * 100 | D+1 08:00 |
| stockout_rate_pct | Stockout Rate | scm | Supply Planning | lower_is_better | % | monthly | median_6m | scm_inventory_daily | stockout_sku_days / total_sku_days * 100 | D+1 08:00 |
| edi_transmission_success_pct | EDI Transmission Success | edi | B2B Integration | higher_is_better | % | monthly | median_6m | edi_transmissions | successful_transmissions / total_transmissions * 100 | D+1 08:00 |
| payroll_summary_accuracy_pct | Payroll Summary Accuracy | hrm | People Ops Technology | higher_is_better | % | monthly | median_6m | hrm_payroll_recon | matched_records / total_records * 100 | D+2 10:00 |
| pl_reconciliation_accuracy_pct | P&L Reconciliation Accuracy | erp | Finance Systems | higher_is_better | % | monthly | median_6m | erp_financial_recon | reconciled_lines / total_lines * 100 | D+2 10:00 |
| master_data_validation_pass_pct | Master Data Validation Pass | mdm | Data Governance | higher_is_better | % | monthly | median_6m | mdm_validation_events | passed_validations / total_validations * 100 | D+1 08:00 |
| insight_adoption_rate_pct | Insight Adoption Rate | dss | Decision Intelligence | higher_is_better | % | monthly | trailing_3m_avg | dss_review_usage | reviews_using_dss / total_monthly_reviews * 100 | D+3 12:00 |
| document_retrieval_time_p95_sec | Document Retrieval Time P95 | ecm | Enterprise Content Ops | lower_is_better | sec | monthly | p95_median_6m | ecm_access_log | percentile_cont(0.95) over retrieval_time_sec | D+1 08:00 |
| idp_extraction_accuracy_pct | IDP Extraction Accuracy | idp | Intelligent Automation | higher_is_better | % | monthly | median_6m | idp_validation_results | correct_fields / total_fields * 100 | D+2 10:00 |

## 3) Dashboard Implementation Notes
- Grafana:
  - Use one folder: Enterprise-KPI
  - Dashboard 1: Executive Summary (6 KPIs)
  - Dashboard 2: Owner Scoreboard (13 KPIs with RAG)
  - Dashboard 3: Risk and Trend (variance and month-over-month)
- BI:
  - Build semantic model with monthly grain as default
  - Add dimensions: owner_team, domain, business_unit, region, month

## 4) Alert Policy (Suggested)
| Alert Key | Condition | Severity | Owner |
|---|---|---|---|
| kpi_miss_monthly | monthly_actual misses monthly target | warning | KPI owner team |
| kpi_miss_quarterly | quarter_to_date misses quarterly target | critical | KPI owner + executive sponsor |
| data_freshness_breach | dataset delayed beyond SLA | warning | Data Platform |

## 5) Example SQL Templates
```sql
-- Example: monthly checkout success rate
SELECT
  DATE_TRUNC('month', event_time) AS month,
  100.0 * SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) / NULLIF(COUNT(*), 0) AS checkout_success_rate_pct
FROM pos_sales_events
GROUP BY 1
ORDER BY 1;
```

```sql
-- Example: monthly P95 document retrieval time
SELECT
  DATE_TRUNC('month', retrieved_at) AS month,
  PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY retrieval_time_sec) AS document_retrieval_time_p95_sec
FROM ecm_access_log
GROUP BY 1
ORDER BY 1;
```
