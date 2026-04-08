# Monthly Review Pack (Executive) - 1 Page + 5 Slide Narrative

รอบรายงาน: Current Cycle (อ้างอิงจาก Owner-Level KPI Scoreboard ล่าสุด)

## 1-Page Executive Snapshot

### Headline
- Overall status: 0 Green, 13 Amber, 0 Red
- Interpretation: ระบบส่วนใหญ่ใกล้เป้ารายเดือน แต่ยังต้องปิด gap เชิงปฏิบัติการเพื่อให้ถึง quarterly targets
- Priority themes this month: Reliability uplift, cost control, data trust, and adoption of decision insights

### KPI Snapshot by Theme
| Theme | KPI | Baseline | Current | Monthly Target | Gap | Status |
|---|---|---:|---:|---:|---:|---|
| Reliability | Checkout Success Rate (%) | 97.6 | 98.3 | 98.5 | -0.2 | Amber |
| Growth | Repeat Purchase Rate (%) | 26.0 | 27.4 | 28.0 | -0.6 | Amber |
| Cost Control | Stockout Rate (%) | 4.7 | 4.0 | 3.8 | +0.2 | Amber |
| Workforce Efficiency | Payroll Summary Accuracy (%) | 98.4 | 98.9 | 99.0 | -0.1 | Amber |
| Data Trust | Master Data Validation Pass (%) | 89.0 | 92.5 | 93.0 | -0.5 | Amber |
| Decision Effectiveness | Insight Adoption Rate (%) | 58.0 | 66.0 | 68.0 | -2.0 | Amber |

### Top 5 Actions Before Next Review
1. Platform Engineering + Identity and Security: ลด auth and routing failure เพื่อเพิ่ม gateway/login reliability
2. Retail Operations + Order Orchestration: ลด checkout and order transition defects ตามสาขาที่กระทบสูง
3. Supply Planning + B2B Integration: ปิด gap stockout และ partner transmission latency
4. Data Governance + Finance Systems: เร่ง data reconciliation สำหรับ monthly close
5. Decision Intelligence + Executive Office: เพิ่มการใช้ DSS ใน monthly business review

### Decisions Needed from Leadership
1. Confirm owner priority stack for the next 30 days
2. Approve cross-team capacity for data quality and integration hardening
3. Confirm risk appetite on amber-only status for another cycle

---

## 5 Slide Narrative

### Slide 1 - Executive Performance Overview
- Message: ธุรกิจเดินหน้าได้ แต่ยังอยู่ช่วง close-the-gap ก่อนเข้าสถานะ green
- Evidence: 13 amber / 0 red / 0 green
- Ask: ยืนยันเป้าหมายเดือนถัดไปไม่ลดมาตรฐาน

### Slide 2 - Reliability and Revenue Path
- Focus KPIs: Gateway Availability, Login Success Rate, Checkout Success Rate
- Story: ช่องทางเข้าระบบดีขึ้น แต่ conversion จุด checkout ยังต่ำกว่าเป้ารายเดือนเล็กน้อย
- Executive action: อนุมัติ reliability sprint ข้าม Platform + Retail

### Slide 3 - Cost and Operational Efficiency
- Focus KPIs: Stockout Rate, Payroll Summary Accuracy, P&L Reconciliation Accuracy
- Story: ต้นทุนและคุณภาพข้อมูลการเงินดีขึ้น แต่ต้องเร่งเพื่อลด monthly closing risk
- Executive action: lock dependency resolution ระหว่าง SCM-HRM-ERP

### Slide 4 - Data Trust and Automation
- Focus KPIs: Master Data Validation Pass, IDP Extraction Accuracy, Document Retrieval Time
- Story: data quality และ document automation ดีขึ้นต่อเนื่อง แต่ยังไม่ถึง threshold ที่นิ่งพอ
- Executive action: prioritize data governance backlog and automation QA cycle

### Slide 5 - Commitments and Next-Month Outcomes
- 30-day commitments:
  - ลด KPI gaps ของ amber metrics ให้เหลือไม่เกิน 5% จาก target
  - เพิ่ม green metrics อย่างน้อย 4 ตัว
  - ไม่มี red KPI ใหม่
- Review checkpoint: owner readout รายสัปดาห์ + executive escalation เฉพาะ blockers สำคัญ

```mermaid
flowchart LR
    A[Collect Monthly KPI Actuals] --> B[Owner Scoreboard Refresh]
    B --> C[Executive Review Meeting]
    C --> D[Commitments and Actions]
    D --> E[Weekly Follow-up]
    E --> A
```
