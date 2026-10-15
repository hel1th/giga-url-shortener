CREATE DATABASE IF NOT EXISTS analytics;

CREATE TABLE IF NOT EXISTS analytics.link_clicks (
    short_code FixedString (7),
    clicked_at DateTime64 (3, 'UTC'),
    ip String,
    user_agent String,
    browser LowCardinality (String),
    os LowCardinality (String),
    device LowCardinality (String),
    referer String
) ENGINE = MergeTree ()
PARTITION BY
    toYYYYMM (clicked_at)
ORDER BY (short_code, clicked_at);