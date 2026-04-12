# Mikoshi bug: concurrent `FOR UPDATE SKIP LOCKED` polls trip "transaction is too large to complete"

## Summary

When two workers run sandman's poll query against `timers` on the same 5s
tick, ~30% of their `UPDATE … WHERE id IN (SELECT … FOR UPDATE SKIP LOCKED)`
calls fail with:

```
ERROR: transaction is too large to complete; try splitting into pieces (SQLSTATE XX000)
```

The error is non-retryable (`first_attempt_count == count`, `max_retries == 0`
in `node_statement_statistics`). The losing worker drops its entire batch and
sleeps until the next tick. With one worker the error never fires.

The error originates at `pkg/kv/txn_interceptor_span_refresher.go:231` and is
gated on `pushed && refreshInvalid`:

- `refreshInvalid` flips when `refreshSpansBytes` exceeds
  `kv.transaction.max_refresh_spans_bytes` (default `256000`, registered at
  `txn_interceptor_span_refresher.go:55-59`).
- `pushed` happens when a concurrent writer bumps this txn's read timestamp.

So the error is *only* reachable when (a) the refresh-span buffer overflows
**and** (b) someone else writes into the keys this txn already read. Either
condition alone is harmless. With two workers polling the same table, both
fire on every tick.

## Environment

- Binary: `v0.0.0-20260412004146-d12d50445bc7` (rev `d12d50445bc7`)
- Platform: `linux/amd64`, Go `1.25.7`
- Cluster: single-node `mikoshi start --insecure --listen-addr=localhost`
- Cluster setting: `kv.transaction.max_refresh_spans_bytes = 256000` (default)
- Session defaults: `sql.defaults.vectorize = auto` (default)

## Repro

Same schema and seed as `MIKOSHI_BUG.md` (UUID PK `timers` table populated
with ~75K due/undelivered rows). Then run two clients, each in a tight loop,
issuing the worker poll UPDATE:

```sql
-- save as poll.sql
UPDATE timers
SET assigned_worker = 'probe-' || gen_random_uuid()::string,
    attempt = attempt + 1,
    assigned_until_utc = now()::timestamp + interval '1 minute',
    retry_utc = now()::timestamp + interval '5 minutes'
WHERE id IN (
    SELECT id FROM timers
    WHERE due_utc < now()::timestamp
      AND (assigned_until_utc IS NULL OR assigned_until_utc < now()::timestamp)
      AND (retry_utc IS NULL OR retry_utc < now()::timestamp)
      AND attempt < 5
      AND delivered_utc IS NULL
    ORDER BY priority + ((mod(shard, 3600)
                          + CAST(extract('minute', now()::timestamp)
                               * extract('second', now()::timestamp) AS INT8))
                         * 100) DESC
    LIMIT 255
    FOR UPDATE SKIP LOCKED
)
RETURNING id;
```

```bash
( while :; do mikoshi sql --insecure --database=sandman < poll.sql; done ) &
( while :; do mikoshi sql --insecure --database=sandman < poll.sql; done ) &
```

After a minute or so, `mikoshi_internal.node_statement_statistics` shows:

```
flags  count  first_attempt_count  max_retries  last_error
       304    302                  1            <nil>
!      82     82                   0            transaction is too large to complete; try splitting into pieces
```

The `!` bucket grows monotonically. With only one client running, it never
appears.

## Expected

`FOR UPDATE SKIP LOCKED` is the textbook pattern for cooperative work-stealing
queues, and serializable refresh is supposed to be transparent for
short-lived single-statement transactions like this. Either:

- Refresh spans should stay bounded for a Scan that reads N rows (one or a
  few spans per range, not per row), so 256 KB is plenty, OR
- A refresh-span overflow on a single-statement implicit txn should fall
  through to a client-visible serializable retry rather than a
  non-retryable "too large" error — at minimum it should not be flagged
  with the !-fingerprint that prevents pgx/database/sql from auto-retrying.

## Actual

Tracing one poll (`SET tracing = on`):

```
operation                                  count
/mikoshi.mikoshipb.Internal/Batch          407
dist sender send                           161
table reader                                40
waiting for read lock                       40
Scan /Table/57/{...}                        37
per-resume Scan /Table/57/{2/"<uuid>"/0-3}  34
```

For ~100 returned rows, that's **161 dist-sender batches** and **34 per-row
resume Scan requests**, each contributing its own refresh-span entry. The
storage-level reason: under row-by-row execution the SELECT FOR UPDATE
acquires locks one row at a time, splitting what would be a single
columnar Scan into a long sequence of per-key resume Scans. EXPLAIN ANALYZE
of the same query reports the fallback explicitly:

```
engine: row-by-row (fallback: unhandled type semantic_type:JSONB)
... rows read 134605 ... stall time 270ms (Sorter) + 151ms (TableReader)
```

Note this is the same class of bug `MIKOSHI_BUG.md` documents for UUID, but
the trigger here is the JSONB columns (`labels`, `hook_headers`) carried
through the outer UPDATE's RETURNING clause.

With both workers running, the per-resume scans interleave and the resulting
fragmentation × row count crosses the 256 KB refresh-span budget. The
moment the *other* worker's UPDATE Puts bump this txn's timestamp, the
refresher discovers `refreshInvalid == true` and emits the non-retryable
error at `txn_interceptor_span_refresher.go:231`.

## Confirmation

Bumping the budget eliminates the failure mode without any other change:

```sql
SET CLUSTER SETTING kv.transaction.max_refresh_spans_bytes = 16000000;  -- 16 MiB
```

| Window                                  | Successful polls | "Too large" errors | Deliveries / min |
|-----------------------------------------|------------------|--------------------|------------------|
| ~12 min before, default 256 KB budget   | 304              | 180                | ~37              |
| ~2 min after, 16 MB budget              | +149             | +1 (in-flight)     | ~1,500           |

Throughput jumps ~40×, and worker-01 (which had been all but starved by the
errors — `timers_processed` 1,144 vs worker-00's 8,455) catches up to
6,345 within two minutes. The errors stop without any change to the SQL,
the schema, or the workers.

## Why it matters

This is the same hot path as `MIKOSHI_BUG.md`, and it cascades. The
row-by-row fallback alone is a 5–10× performance hit; combine it with the
refresh-span overflow and one of two concurrent workers becomes a no-op.
Sandman's whole horizontal-scaling story collapses: adding workers
*reduces* throughput because each new worker steals refresh-span budget
from its peers without adding successful polls.

The error message ("try splitting into pieces") is also misleading for
this workload — there is nothing to split. The query is a single
statement; the implicit txn is one statement long; the user can't shorten
it. The actionable thing is "raise `max_refresh_spans_bytes`," which is
not what the message suggests.

Three things compound here, and I think all three are worth fixing
independently:

1. **Vectorized fallback for JSONB pass-through columns.** Same root as
   `MIKOSHI_BUG.md` — the planner bails on the columnar engine because
   the projected set contains a type it doesn't have a vector for, even
   though nothing in the plan operates on the JSONB values. Fixing this
   alone would shrink refresh-span fragmentation enough that the 256 KB
   budget is fine.
2. **Per-row refresh-span emission under `FOR UPDATE SKIP LOCKED`.** Each
   per-resume Scan request emits its own refresh span. For a SELECT that
   touches N rows under row-by-row locking, refresh-span growth is O(N)
   in span *count*, not just in covered key bytes. Coalescing adjacent
   refresh spans on append (or recording one bounding span per
   range-level batch and only narrowing on conflict) would fix this.
3. **Non-retryable surface for refresh-span overflow on a single-statement
   txn.** The current behavior — fail the implicit txn with a
   non-retryable error — is the worst possible outcome for the database/sql
   driver, which has no way to recover. A `RETRY_SERIALIZABLE` would at
   least let the SQL layer auto-retry (and probably succeed once the
   conflicting txn cleared).

## Suspected area

- `pkg/kv/txn_interceptor_span_refresher.go:55` — `MaxTxnRefreshSpansBytes`
  default and `:204-217` overflow path.
- `pkg/kv/txn_interceptor_span_refresher.go:225-233` — the `pushed &&
  refreshInvalid` short-circuit that turns the overflow into a non-retryable
  error rather than `RETRY_SERIALIZABLE`.
- `pkg/kv/txn_interceptor_span_refresher.go:460-499` — `appendRefreshSpans`
  appends one entry per request span without coalescing.
- `pkg/sql/colflow/` (or wherever the planner decides JSONB → row-by-row
  fallback) — same hunting ground as `MIKOSHI_BUG.md`'s UUID case.
- `pkg/sql/row/cfetcher.go` / `pkg/sql/scan.go` — how `FOR UPDATE SKIP
  LOCKED` is lowered to per-row resume Scans under row-by-row execution.

## Minimal hypothesis to test first

In `appendRefreshSpans`, before appending, check if the new span is
contiguous with the last span recorded for the same direction (read/write)
and merge them in place. If yes, the fragmentation goes away and the
default 256 KB budget should hold even under row-by-row execution.

Independently: if you replace `labels JSONB` and `hook_headers JSONB` with
`STRING`, does the engine stay vectorized? If yes, the JSONB pass-through
fallback is the upstream cause and fixing it makes the refresh-span issue
unobservable in this workload.
