# Bulk Jobs

Long-running admin operations (`*Async` mutations, `retryBulkJob`) are tracked as
rows in `bulk_jobs` and executed on the instance that received the request.
There is no queue — the former GCP Pub/Sub round-trip was removed.

## Code

| File | Role |
| --- | --- |
| `backend/internal/services/bulk/jobs.go` | `OperationType` constants, typed `*Params` structs, `JobStatus`, `BatchSize` |
| `backend/internal/services/bulk/runner.go` | `Runner`: marks job PROCESSING, dispatches by operation type, marks FAILED on error/panic |
| `backend/internal/services/bulk/processor.go` | `Processor`: per-operation logic, marks COMPLETED with counts |
| `backend/internal/services/bulk/service.go` | `CreateAndRunBulkJob`, `RetryBulkJob`, listing/queries |

Adding an operation: add a constant + params type in `jobs.go`, a method on
`OperationHandler`/`Processor`, a case in `Runner.dispatch`, then `make generate`
(regenerates mocks).

## Lifecycle

`PENDING` (row inserted) → `PROCESSING` → `COMPLETED` | `FAILED`.
Failures are recorded on the row (`error_message`), not returned as GraphQL errors.

## Sync vs async

Controlled by `BULK_JOBS_SYNC` (config `Jobs.Sync`):

- **Default off-GCP (Dokploy, local)** — async. The job runs in a goroutine
  detached from request cancellation (`context.WithoutCancel`); the mutation returns
  the PENDING job immediately. Poll `bulkJob(id)` for status. On shutdown the
  server waits up to 30s for in-flight jobs.
- **Default on Cloud Run** (`K_SERVICE` or `CLOUD_RUN_JOB` is set) — sync. Cloud
  Run services throttle CPU once the response is sent, and Cloud Run jobs kill
  leftover goroutines when the process exits, so the job runs inside the request
  and the mutation returns the job with its final status.

## Caveats

- Async jobs interrupted by a crash/redeploy stay `PROCESSING` forever; use
  `retryBulkJob` to rerun.
- Sync jobs are bounded by `SERVER_WRITE_TIMEOUT` (default 5m) and the Cloud Run
  request timeout. Very large jobs should be split (see the batching in
  `fixMissing*Async` resolvers).
- `bulk_jobs.message_id` is unused since Pub/Sub was removed; it can be dropped
  in a future migration.
