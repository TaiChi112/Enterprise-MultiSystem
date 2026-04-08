# KPI Baseline Alignment Workbook (3-6 Month Historical Data)

เอกสารนี้ใช้ปรับ KPI baseline ให้สอดคล้องข้อมูลธุรกิจจริง โดยอ้างอิงข้อมูลย้อนหลัง 3-6 เดือนก่อน monthly review

## 1) Current Data Status in Repository
- Current status: repository ยังไม่มีไฟล์ historical business export (monthly actuals) แบบพร้อมใช้งาน
- Interpretation: baseline ในเอกสารระบบปัจจุบันเป็น provisional baseline เพื่อใช้ตั้ง governance รอบแรก
- Next action: เติมข้อมูลย้อนหลังจริงผ่าน template ด้านล่าง แล้วคำนวณ baseline ใหม่ตามสูตรมาตรฐาน

## 2) Baseline Alignment Rules
1. Time window: ใช้ย้อนหลัง 6 เดือน (ถ้ามีไม่ครบให้ใช้ขั้นต่ำ 3 เดือน)
2. Baseline method:
   - Stability KPI (availability, accuracy): ใช้ median ของรายเดือนย้อนหลัง
   - Latency KPI: ใช้ P95 median ของรายเดือนย้อนหลัง
   - Growth KPI: ใช้ trailing 3-month average growth
3. Seasonality handling:
   - หากธุรกิจมี seasonality สูง ให้เก็บทั้ง baseline รวมและ baseline ตาม season
4. Outlier policy:
   - ตัด outlier ที่เกิน 3 sigma หรือเหตุการณ์ one-off ที่มี incident report รองรับ
5. Target setting:
   - Monthly Target = baseline ปรับดีขึ้น 5-12% ตามความยาก
   - Quarterly Target = monthly target x 1.5 improvement step

## 3) Historical Data Intake Template
ใช้ตารางนี้เติมค่าจริงต่อเดือน (ตัวอย่างเดือน Jan-Jun)

| Metric Key | Jan | Feb | Mar | Apr | May | Jun | Unit | Notes |
|---|---|---|---|---|---|---|---|---|
| gateway_availability_pct |  |  |  |  |  |  | % | uptime monthly |
| auth_failure_rate_pct |  |  |  |  |  |  | % | protected requests |
| checkout_success_rate_pct |  |  |  |  |  |  | % | POS completed checkouts |
| stock_accuracy_post_sale_pct |  |  |  |  |  |  | % | inventory reconciliation |
| repeat_purchase_rate_pct |  |  |  |  |  |  | % | CRM monthly cohort |
| order_lifecycle_completion_pct |  |  |  |  |  |  | % | OMS completed lifecycle |
| stockout_rate_pct |  |  |  |  |  |  | % | SCM stockout incidence |
| edi_transmission_success_pct |  |  |  |  |  |  | % | EDI outbound success |
| payroll_summary_accuracy_pct |  |  |  |  |  |  | % | HRM vs finance reconciliation |
| financial_reconciliation_accuracy_pct |  |  |  |  |  |  | % | ERP reconciled summaries |
| master_data_validation_pass_pct |  |  |  |  |  |  | % | MDM validation success |
| insight_adoption_rate_pct |  |  |  |  |  |  | % | DSS usage in review meetings |
| document_retrieval_time_sec |  |  |  |  |  |  | sec | ECM retrieval latency |
| idp_extraction_accuracy_pct |  |  |  |  |  |  | % | IDP extracted fields accuracy |

## 4) Baseline Output Table (Fill After Data Intake)
| KPI | Baseline Type | 3-6M Baseline | Monthly Target | Quarterly Target | Owner |
|---|---|---|---|---|---|
| Gateway Availability | Median |  |  |  | Platform Engineering |
| Login Success Rate | Median |  |  |  | Identity and Security |
| Checkout Success Rate | Median |  |  |  | Retail Operations Product |
| Repeat Purchase Rate | T3M average |  |  |  | Customer Growth |
| Order Lifecycle Completion Rate | Median |  |  |  | Order Orchestration |
| Stockout Rate | Median |  |  |  | Supply Planning |
| EDI Transmission Success Rate | Median |  |  |  | B2B Integration |
| Payroll Summary Accuracy | Median |  |  |  | People Ops Technology |
| P&L Reconciliation Accuracy | Median |  |  |  | Finance Systems |
| Master Data Validation Pass Rate | Median |  |  |  | Data Governance |
| Insight Adoption Rate | T3M average |  |  |  | Decision Intelligence |
| Document Retrieval Time | P95 median |  |  |  | Enterprise Content Ops |
| IDP Extraction Accuracy | Median |  |  |  | Intelligent Automation |

## 5) Governance Cadence
- Weekly: data quality and pipeline completeness check
- Monthly: baseline drift check and owner review
- Quarterly: target reset and threshold governance sign-off

## 6) Decision Log (Current Round)
- Baseline in system pages remains provisional due to missing 3-6 month exported business actuals in repository
- This workbook is now the single template for replacing provisional baselines with real values
