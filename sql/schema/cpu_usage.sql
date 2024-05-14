-- Sets up db schema for CPU usage tracking data.

DROP TABLE IF EXISTS cpu_usage;

-- Setup TimescaleDB extension
CREATE EXTENSION IF NOT EXISTS timescaledb;

-- Create the table & convert to a Timescale hypertable
CREATE TABLE IF NOT EXISTS cpu_usage(
  ts    TIMESTAMPTZ,
  host  TEXT,
  usage DOUBLE PRECISION
);
SELECT create_hypertable('cpu_usage', 'ts');
