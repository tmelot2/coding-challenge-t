--CREATE DATABASE homework;
-- \c homework

CREATE EXTENSION IF NOT EXISTS timescaledb;

CREATE TABLE IF NOT EXISTS cpu_usage(
  ts    TIMESTAMPTZ,
  host  TEXT,
  usage DOUBLE PRECISION
);

SELECT create_hypertable('cpu_usage', 'ts');
