-- One-shot CSV seed for KPI dashboard tables
-- Requires CSV files mounted under /docker-entrypoint-initdb.d/

BEGIN;

TRUNCATE TABLE kpi_owner_scoreboard_current RESTART IDENTITY;
TRUNCATE TABLE kpi_monthly_actuals RESTART IDENTITY;

COPY kpi_monthly_actuals (month, metric_key, owner_team, direction, actual_value, unit, source_system)
FROM '/docker-entrypoint-initdb.d/kpi_monthly_actuals_seed.csv'
WITH (FORMAT csv, HEADER true);

COPY kpi_owner_scoreboard_current (
    cycle_month,
    owner_team,
    metric_key,
    primary_kpi,
    direction,
    baseline_value,
    current_month_value,
    monthly_target,
    quarterly_target,
    status,
    priority_action
)
FROM '/docker-entrypoint-initdb.d/kpi_owner_scoreboard_current.csv'
WITH (FORMAT csv, HEADER true);

COMMIT;
