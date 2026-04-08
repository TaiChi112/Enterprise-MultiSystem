-- KPI dashboard tables for Grafana/BI integration
-- Applies to development/demo environments.

CREATE TABLE IF NOT EXISTS kpi_monthly_actuals (
    id BIGSERIAL PRIMARY KEY,
    month DATE NOT NULL,
    metric_key VARCHAR(128) NOT NULL,
    owner_team VARCHAR(128) NOT NULL,
    direction VARCHAR(32) NOT NULL CHECK (direction IN ('higher_is_better', 'lower_is_better')),
    actual_value NUMERIC(12,4) NOT NULL,
    unit VARCHAR(16) NOT NULL,
    source_system VARCHAR(64),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE (month, metric_key)
);

CREATE INDEX IF NOT EXISTS idx_kpi_monthly_actuals_metric_month
    ON kpi_monthly_actuals (metric_key, month DESC);

CREATE INDEX IF NOT EXISTS idx_kpi_monthly_actuals_owner_month
    ON kpi_monthly_actuals (owner_team, month DESC);

CREATE TABLE IF NOT EXISTS kpi_owner_scoreboard_current (
    id BIGSERIAL PRIMARY KEY,
    cycle_month DATE NOT NULL,
    owner_team VARCHAR(128) NOT NULL,
    metric_key VARCHAR(128) NOT NULL,
    primary_kpi VARCHAR(255) NOT NULL,
    direction VARCHAR(32) NOT NULL CHECK (direction IN ('higher_is_better', 'lower_is_better')),
    baseline_value NUMERIC(12,4) NOT NULL,
    current_month_value NUMERIC(12,4) NOT NULL,
    monthly_target NUMERIC(12,4) NOT NULL,
    quarterly_target NUMERIC(12,4) NOT NULL,
    status VARCHAR(16) NOT NULL CHECK (status IN ('Green', 'Amber', 'Red')),
    priority_action TEXT,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE (cycle_month, metric_key)
);

CREATE INDEX IF NOT EXISTS idx_kpi_owner_scoreboard_current_status
    ON kpi_owner_scoreboard_current (status);

CREATE INDEX IF NOT EXISTS idx_kpi_owner_scoreboard_current_owner
    ON kpi_owner_scoreboard_current (owner_team);

-- Helper view for latest monthly snapshot per metric.
CREATE OR REPLACE VIEW v_kpi_latest_actuals AS
SELECT DISTINCT ON (metric_key)
    metric_key,
    owner_team,
    month,
    actual_value,
    unit,
    direction
FROM kpi_monthly_actuals
ORDER BY metric_key, month DESC;

-- DB-side RAG computation to avoid BI-side status logic.
CREATE OR REPLACE VIEW v_kpi_owner_scoreboard_rag AS
SELECT
    s.cycle_month,
    s.owner_team,
    s.metric_key,
    s.primary_kpi,
    s.direction,
    s.baseline_value,
    s.current_month_value,
    s.monthly_target,
    s.quarterly_target,
    CASE
        WHEN s.direction = 'higher_is_better' AND s.current_month_value >= s.monthly_target THEN 'Green'
        WHEN s.direction = 'lower_is_better' AND s.current_month_value <= s.monthly_target THEN 'Green'
        WHEN s.direction = 'higher_is_better' AND ((s.monthly_target - s.current_month_value) / NULLIF(s.monthly_target, 0)) <= 0.10 THEN 'Amber'
        WHEN s.direction = 'lower_is_better' AND ((s.current_month_value - s.monthly_target) / NULLIF(s.monthly_target, 0)) <= 0.10 THEN 'Amber'
        ELSE 'Red'
    END AS computed_status,
    CASE
        WHEN s.direction = 'higher_is_better' THEN ROUND(100.0 * (s.current_month_value - s.monthly_target) / NULLIF(s.monthly_target, 0), 4)
        ELSE ROUND(100.0 * (s.monthly_target - s.current_month_value) / NULLIF(s.monthly_target, 0), 4)
    END AS monthly_gap_pct,
    CASE
        WHEN s.direction = 'higher_is_better' THEN ROUND(100.0 * (s.current_month_value - s.quarterly_target) / NULLIF(s.quarterly_target, 0), 4)
        ELSE ROUND(100.0 * (s.quarterly_target - s.current_month_value) / NULLIF(s.quarterly_target, 0), 4)
    END AS quarterly_gap_pct,
    s.priority_action,
    s.updated_at
FROM kpi_owner_scoreboard_current s;
