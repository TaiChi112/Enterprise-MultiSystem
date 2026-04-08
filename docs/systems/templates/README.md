# KPI Actual Data Templates

ใช้ไฟล์ในโฟลเดอร์นี้เพื่อส่ง actual data ย้อนหลัง 3-6 เดือน และคำนวณ baseline + RAG อัตโนมัติ

## Files
- kpi_actuals_6m_template.csv: ใส่ค่าจริงย้อนหลังรายเดือน 3-6 เดือน
- kpi_current_month_input.csv: ใส่ baseline + current month + targets สำหรับ monthly review
- kpi_owner_scoreboard_current.csv: scoreboard ปัจจุบันในรูป CSV สำหรับ import เข้า BI
- kpi_monthly_actuals_seed.csv: sample actuals สำหรับ seed เข้า kpi_monthly_actuals

Related setup files:
- SQL DDL: [migrations/20260408_create_kpi_dashboard_tables.sql](../../../migrations/20260408_create_kpi_dashboard_tables.sql)
- SQL Seed (one-shot COPY): [migrations/20260408_seed_kpi_dashboard_tables.sql](../../../migrations/20260408_seed_kpi_dashboard_tables.sql)
- Grafana provider: [observability/grafana/provisioning/dashboards/enterprise-kpi-dashboards.yml](../../../observability/grafana/provisioning/dashboards/enterprise-kpi-dashboards.yml)
- Grafana datasource: [observability/grafana/provisioning/datasources/postgres-kpi.yml](../../../observability/grafana/provisioning/datasources/postgres-kpi.yml)
- Seed refresh script: [scripts/refresh-kpi-seed.sh](../../../scripts/refresh-kpi-seed.sh)
- VS Code task: [.vscode/tasks.json](../../../.vscode/tasks.json)

## Datasource Default for Enterprise-KPI Folder
Grafana ไม่มีฟีเจอร์ default datasource รายโฟลเดอร์โดยตรง
แนวทางที่ใช้อยู่ในโปรเจกต์นี้คือ:
1. provision datasource `postgres-kpi`
2. provision dashboard provider แยกโฟลเดอร์ `Enterprise-KPI`
3. lock datasource ของ dashboard ในโฟลเดอร์นี้ให้ใช้ `postgres-kpi` ทุก panel/query

ผลลัพธ์: dashboard ในโฟลเดอร์ Enterprise-KPI ใช้ Postgres โดยอัตโนมัติ โดยไม่กระทบ dashboard อื่น

## One-shot Initialization on Startup
เมื่อ `postgres` เริ่มครั้งแรก (data volume ยังใหม่) ระบบจะรันไฟล์ init ตามลำดับ:
1. `01-schema.sql`
2. `02-kpi-ddl.sql`
3. `03-kpi-seed.sql`

ข้อสำคัญ: ถ้ามี volume เดิมอยู่แล้ว script init จะไม่รันซ้ำอัตโนมัติ

หากต้องการให้ init ใหม่ทั้งหมด:

```bash
docker compose down -v
docker compose up -d
```

## One-Command Refresh Seed (No Volume Drop)

```bash
bash scripts/refresh-kpi-seed.sh
```

หรือผ่าน VS Code Task: `refresh-kpi-seed`

## Recommended Flow
1. เติมข้อมูลย้อนหลังใน kpi_actuals_6m_template.csv
2. คำนวณ baseline ต่อ metric
3. เติมไฟล์ kpi_current_month_input.csv ด้วย baseline + current + targets
4. คำนวณ RAG status อัตโนมัติด้วยสูตรด้านล่าง

## Excel Formula for Auto-RAG
สมมติคอลัมน์:
- D: direction
- F: current_month_value
- G: monthly_target

สูตร:

```excel
=IF(D2="higher_is_better",IF(F2>=G2,"Green",IF((G2-F2)/G2<=0.1,"Amber","Red")),IF(F2<=G2,"Green",IF((F2-G2)/G2<=0.1,"Amber","Red")))
```

## Baseline Calculation Rules
- median_6m: ใช้ MEDIAN ของ 6 เดือนล่าสุด
- trailing_3m_avg: ใช้ AVERAGE 3 เดือนล่าสุด
- p95_median_6m: ใช้ median ของค่า P95 รายเดือน
