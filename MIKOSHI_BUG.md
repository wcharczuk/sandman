# Mikoshi bug: vectorized engine falls back to row-by-row on UUID scan+sort

## Summary

A plain `TableReader → Sorter` plan over a table whose primary key is `UUID`
drops out of the vectorized engine with:

```
engine: row-by-row (fallback: unhandled type semantic_type:UUID)
```

The query doesn't touch the UUID column in any non-trivial way — the projection
is `id` and a computed integer expression, and the sort is on the computed
integer. UUID appears only as the PK being read off the primary index. There's
no obvious reason vectorized execution should bail here.

Fix impact is meaningful in practice: with ~55K rows passing the filter on a
single range, row-by-row execution takes hundreds of ms of stall time where
vectorized should be in the tens.

## Environment

- Binary: `v0.0.0-20260411202707-2995001d82d6` (rev `2995001d82d6`)
- Platform: `linux/amd64`, Go `1.25.7`
- Cluster: single-node `mikoshi start --insecure --listen-addr=localhost`
- Session defaults (unchanged): `sql.defaults.vectorize = auto`

## Repro

Schema (PK is `UUID`; only other index is a unique on `name`):

```sql
CREATE DATABASE repro;
USE repro;

CREATE TABLE timers (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    name STRING NOT NULL,
    labels JSONB NULL,
    priority INT8 NOT NULL,
    shard_key STRING NOT NULL,
    shard INT8 NOT NULL,
    created_utc TIMESTAMP NOT NULL,
    due_utc TIMESTAMP NOT NULL,
    assigned_until_utc TIMESTAMP NULL,
    retry_utc TIMESTAMP NULL,
    attempt INT8 NOT NULL,
    assigned_worker STRING NULL,
    hook_url STRING NOT NULL,
    hook_method STRING NOT NULL,
    hook_headers JSONB NULL,
    hook_body BYTES NULL,
    delivered_utc TIMESTAMP NULL,
    delivered_status_code INT8 NOT NULL,
    delivered_err STRING NOT NULL,
    CONSTRAINT pk_timers_id PRIMARY KEY (id ASC),
    UNIQUE INDEX uk_timers_name (name ASC)
);

-- Populate ~75K rows, all "due" and undelivered.
INSERT INTO timers (
    name, labels, priority, shard_key, shard, created_utc, due_utc,
    attempt, hook_url, hook_method, delivered_status_code, delivered_err
)
SELECT
    gen_random_uuid()::STRING,
    NULL,
    (random()*1000)::INT8,
    'k',
    (random()*10000)::INT8,
    now()::TIMESTAMP,
    (now() - '1 hour'::INTERVAL)::TIMESTAMP,
    0,
    'http://example/hook',
    'GET',
    0,
    ''
FROM generate_series(1, 75000);
```

Run the worker's poll predicate under `EXPLAIN ANALYZE`:

```sql
EXPLAIN ANALYZE
SELECT id
FROM timers
WHERE
    due_utc < now()::TIMESTAMP
    AND (assigned_until_utc IS NULL OR assigned_until_utc < now()::TIMESTAMP)
    AND (retry_utc IS NULL OR retry_utc < now()::TIMESTAMP)
    AND attempt < 5
    AND delivered_utc IS NULL
ORDER BY
    priority
    + ((mod(shard, 3600)
        + CAST(extract('minute', now()::TIMESTAMP)
             * extract('second', now()::TIMESTAMP) AS INT8)
      ) * 100)
    DESC
LIMIT 255;
```

## Expected

Vectorized execution: `TableReader` feeds columnar batches into an `ordered
aggregator`-style sorter with a `topK` cutoff. `engine` should report the
vectorized row flow, not `row-by-row`.

## Actual

```
tree         field         description
              engine        row-by-row (fallback: unhandled type semantic_type:UUID)
Response
 └── Sorter/1
                            @1-
                 Out        @2
                            Limit 255
                 rows read  55175
                 stall time 148.754ms
                 max memory used 20 KiB
                 max disk used   0 B
      └── TableReader/0
                            timers@primary
                 Spans      -
                 Filter     ((((@8 < now()::TIMESTAMP) AND ((@9 IS NULL) OR (@9 < now()::TIMESTAMP))) AND ((@10 IS NULL) OR (@10 < now()::TIMESTAMP))) AND (@11 < 5)) AND (@17 IS NULL)
                 Render     expr:"@4+((mod(@6,3600)+(extract('minute',now()::TIMESTAMP)*extract('second',now()::TIMESTAMP)))*100)", expr:"@1"
                 rows read  74288
                 stall time 62.336ms
```

Note the projection is `@1` (the UUID `id`) + the computed INT8 sort key. The
sort is entirely on the INT8 — UUID is passed through, not compared.

## Why it matters

This is the hot-path query for sandman's worker (`pkg/model/manager.go:146`),
which polls every 5s per worker. With two workers and a 52K-row backlog,
observed service latencies from `mikoshi_internal.node_statement_statistics`:

| Statement | Count | avg | max |
|---|---|---|---|
| Worker poll UPDATE (with this SELECT as subquery) | 134 | 2.57s | 5.33s |
| Same UPDATE, other fingerprint | 162 | 196ms | 2.50s |

Polls take longer than the 5s interval, so workers effectively never sleep and
the lock-table grows unbounded (10K held locks on `timers` during the incident,
2 holder txns, 0 waiters — all from `FOR UPDATE SKIP LOCKED`). Client queries
from outside the workers (e.g. a `SELECT count(*)`) time out.

Sandman can and should add a partial index to reduce the scan — but even
without an index, a 74K-row columnar scan+sort on INT8 should be fast. The
row-by-row fallback is the part that makes this much worse than it needs to
be.

## Suspected area

Vectorized column-type gating. Somewhere in the planner decides the projection
set contains `UUID` and gives up instead of emitting a pass-through column.
Likely candidates:

- `pkg/sql/colexec/` — type switch over projected column types
- `pkg/sql/colflow/` — plan-to-vector translation that decides the fallback

Look for the emitter of the exact string
`unhandled type semantic_type:%s` (or the `semantic_type:UUID` render path)
— that's where the short-circuit is.

## Minimal hypothesis to test first

If you replace the PK type with `INT8` / `BYTES` and rerun the same
`EXPLAIN ANALYZE`, does the engine stay vectorized? If yes, the bug is purely
in the projected-column type switch — UUID pass-through isn't wired up, even
though nothing in the plan actually operates on the UUID values.
