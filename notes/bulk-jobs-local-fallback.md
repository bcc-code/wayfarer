# Bulk Jobs: running async operations locally

## The problem

Every `*Async` bulk mutation (`grantQuizSessionAccessAsync`, the bulk challenge
and achievement operations, the fix-missing-progress jobs) publishes a message to
GCP Pub/Sub and returns a `BulkJob` the UI polls. A worker receives the message
on `POST /pubsub/bulk-operations` and processes it.

`PUBSUB_ENABLED` defaults to `false` and is set per environment outside this repo
(Dokploy env vars in production; nothing sets it locally). With it off, the
service used to mark the freshly created job `failed` and return:

```
pub/sub is disabled, use synchronous mutation instead
```

So every async bulk operation was unusable locally, and each attempt left a
`failed` row in `bulk_jobs`. Its own code comment said "If Pub/Sub is disabled,
process synchronously" — the implementation did the opposite.

## The fix

`bulk.Service` now falls back to running the job **in this process** when Pub/Sub
is disabled, instead of failing it.

- `LocalProcessor` (in `internal/services/bulk/service.go`) is a one-method
  interface satisfied by `*pubsub.Processor`. It is an interface because the
  processor is built *from* the bulk service — it dispatches back into the
  service's `ProcessBulk*` methods — so the two can only be joined after both
  exist. `cmd/server/main.go` calls `bulkService.SetLocalProcessor(pubsubProcessor)`
  right after constructing the processor.
- `Service.runLocally` sends the message through that same processor in a
  goroutine, so the job moves through the same states and progress counters as it
  would via Pub/Sub. Only the delivery mechanism differs.
- Both branches that previously failed — `CreateBulkJobAndPublish` and
  `RetryBulkJob` — now use it.

The job context is `context.WithoutCancel(ctx)`: the mutation returns the pending
job immediately, exactly as it does with Pub/Sub, so the request that created the
job is long gone before a large grant finishes.

## Why this is a fallback and not the default

**Production keeps Pub/Sub.** A quiz session is granted to ~10k users at once,
which is far past what belongs in a request, and the in-process path gives up what
Pub/Sub provides:

- no durability — a restart mid-job leaves the row in `processing` with no
  redelivery
- no retries
- no back-pressure; the work runs on the API process

Production sets `PUBSUB_ENABLED=true` and never takes this path. The fallback logs
at WARN (`"Pub/Sub disabled, processing bulk operation in-process"`) so a
misconfigured environment is visible in the logs rather than silent.

## Using it

Nothing to configure — with `PUBSUB_ENABLED` unset locally, the admin UI's
"grant access" buttons now work and the job completes in the background. Poll the
`bulkJob` query (or reload the sessions page) to see it finish.

To exercise the real Pub/Sub path locally instead, set `PUBSUB_ENABLED=true`,
`PUBSUB_PROJECT_ID` and `PUBSUB_TOPIC_ID` against the Pub/Sub emulator, and point
a push subscription at `POST /pubsub/bulk-operations`.

## Note on the synchronous mutations

`grantQuizSessionAccess` (no `Async`) still exists and calls the same
`BulkService.GrantQuizSessionAccess` inline, returning the number of grants. It is
fine for small, known-small grants, but the admin UI deliberately uses the async
variant because the all-project case is not small.
