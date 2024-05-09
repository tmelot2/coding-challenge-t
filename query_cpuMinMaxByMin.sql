-- Queries per-minute CPU min & max, for given hostname, within given time range.
-- NOTE: Time range is inclusive on both start & end!
--
-- $1 start timestampz
-- $2 end timestampz
-- $3 hostname

WITH minutes AS (
    SELECT generate_series(
        $1::TIMESTAMPTZ,
        $2::TIMESTAMPTZ,
        '1 minute'::interval
    ) AS minute
)
SELECT
    cpu.host,
    m.minute,
    MIN(usage) as cpuMin,
    MAX(usage) as cpuMax
FROM minutes m
LEFT JOIN cpu_usage cpu
    ON cpu.ts >= m.minute
    AND cpu.ts <= m.minute + INTERVAL '1 minute'
WHERE
    cpu.host = $3
    and cpu.ts <= $2
GROUP BY cpu.host, m.minute
ORDER BY m.minute;
