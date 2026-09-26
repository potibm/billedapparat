# ADR 007: Health and Readiness Probe Scope

**Status:** Accepted

## Context

Containers and reverse proxies need to know whether the Hub is worth sending traffic to. Docker restarts a container whose `HEALTHCHECK` keeps failing, and orchestrators remove an unready replica from the load balancer. Both decisions act on the answer, so a probe that reports a failure the operator cannot act on is worse than no probe at all: it removes healthy capacity, and it converts a transient external hiccup into a restart storm across every replica.

Billedapparat's `serve` command has one runtime dependency — the SQLite database at `data/db/`. It additionally may be configured to authenticate `/api/admin` against an external OIDC provider, and the wider Billedapparat ecosystem includes Redis-backed Collectors.

## Decision

The Hub exposes two unauthenticated probes on the root of the application, not under `/api`:

- `GET /health` — liveness. Always answers `200` while the process runs and performs no dependency checks. The container `HEALTHCHECK` probes this endpoint.
- `GET /ready` — readiness. Answers `200` once the SQLite database is reachable and `503` when it is not.

| Situation | HTTP | `status`      | Body                                               |
| --------- | ---- | ------------- | -------------------------------------------------- |
| `/health` | 200  | `ok`          | `{"status":"ok"}`                                  |
| DB up     | 200  | `ready`       | `{"status":"ready"}`                               |
| DB down   | 503  | `unavailable` | `{"status":"unavailable","error":"database down"}` |

`/ready` checks the database and nothing else. The underlying error is never returned, so the response contract cannot leak driver internals to an unauthenticated caller. It is logged on every failure and reported to Sentry at most once a minute; see _Observability_ below.

`Store.Ping` (`internal/app/store/gorm/store.go`) performs one step: a `PingContext` on the underlying `*sql.DB`, bounded by `readinessCheckTimeout`. That is the entire check, and the reasoning is worth recording so the check is not silently "strengthened" into something misleading.

### Why the database check is only connection-level

The driver does implement `driver.Pinger`, but its implementation runs `select 1` — a pure expression that reads neither the schema nor the file — and connections are opened with `SQLITE_OPEN_CREATE`. Two consequences:

- A ping cannot tell a healthy database from an empty one.
- A ping still succeeds after the database file has been deleted, because opening a missing path simply creates it.

An earlier version added a second step that looked up a migrated table in `sqlite_master`. It was removed. It looked stronger, but the two states it was meant to catch cannot reach a running Hub:

- **A corrupt file** never gets that far. `gorm.Open` survives it (its version probe is another pure expression), but `newStore` runs `AutoMigrate` immediately afterwards, and the DDL must read the schema. A corrupt file therefore aborts startup and the process exits before a router exists.
- **A lost volume** is unreliable in both directions. A pooled connection keeps reading through a file handle that stays valid after the file is unlinked, so `/ready` keeps answering `200`; a freshly opened connection instead creates an empty file and reports a failure. The answer depends on pool state rather than on the fault. Worse, opening that fresh connection _recreates_ the database file, so the restart that follows AutoMigrates an empty database into a valid schema and the Hub reports itself ready with no data.

### What the probe deliberately does not report

- **Write-lock contention.** WAL lets readers proceed while a Collector holds the write lock, and Beamers keep streaming. A busy database is not a reason to pull the Hub out of rotation, so contention is not a failure.
- **A lost or corrupt volume**, for the reasons above. Alert on the freshness of ingested slides instead. Treat `/ready` as a process-and-connection signal, not a storage-integrity one.

### Observability

`/ready` is unauthenticated and polled on a timer, so a naive error report would turn one outage into thousands of duplicate events. Failures are therefore split:

- **Logged every time**, at `WARN`, with the real driver error. This is the audit trail and the intended alerting source.
- **Reported to Sentry at most once per minute**, at `LevelWarning` with a `probe=readiness` tag. The first failure of each window is always sent.

Sentry is the alerting path here because `sentrygin` only recovers panics, so the `503` response is otherwise invisible to error tracking.

### The scope rule

A dependency may gate readiness **only if restarting the Hub would plausibly repair its failure**. If a restart cannot fix it, a failing probe is not actionable and must not influence routing or restarts.

The database qualifies: it holds all state, the Hub cannot serve any request without it, and a restart is a reasonable remedy.

### Considered and rejected: probing the OIDC provider

An earlier proposal was to add a cached, timeout-bounded probe of the provider's discovery document to `/ready`. It was dropped, so it is not re-proposed without context:

1. **A restart does not fix it.** If the identity provider is unavailable, restarting every Hub replica would not restore logins — it would only replace working processes with identical ones. Gating readiness on the provider converts a login blip into a full outage.
2. **Issued tokens are verified offline.** The OIDC middleware (`internal/app/middleware/auth.go`) validates a bearer JWT against a `keyfunc` JWKS cache (`defaultJWKSRefreshInterval` = 1 hour, `defaultJWKSRefreshRateLimit` = 5 minutes). It performs no per-request introspection. An admin client that already holds a token keeps working while the provider is down, so reporting the Hub as not-ready would misdescribe what authenticated users actually experience.
3. **Operational complexity outweighs the benefit.** A correct implementation needs a cached probe with a TTL, a short timeout, a decision on how to surface the result in the response contract, and monitoring guidance. None of that surface area is worth paying for a signal that is available more cheaply elsewhere.

### Considered and rejected: probing Redis

Not even a candidate. Per ADR 001 the Hub does not manage Collectors, and the `serve` process opens no Redis connection at all — Redis is consumed only by the separate `collect` process (`cmd/collect.go`). A failure there affects ingestion, not the Hub's ability to serve the content it already holds, so a Hub restart cannot repair it.

## Consequences

- **Positive:** Readiness is cheap and deterministic. The endpoint performs one local operation bounded by `readinessCheckTimeout` (2 seconds), with no external network dependency that could make it slow or flaky.
- **Positive:** A dependency outage can never make Hub instances appear unhealthy, so no restart storm is possible and Beamers keep playing whatever the Hub already has.
- **Positive:** The probe cannot report a fault that a restart would not fix, which keeps it from being a source of false negatives under normal operation.
- **Positive:** The probes are registered before the embedded asset middleware, so probing them never pays for a static file lookup, and they are traced automatically by the existing `otelgin` middleware.
- **Positive:** Readiness failures reach error tracking, so an unready Hub is alertable without scraping logs.
- **Positive:** Because the endpoint is unauthenticated, the store caps its connection pool (`maxOpenConns`) so a burst of probes cannot open unbounded connections and file handles. The cap is greater than one on purpose: SQLite serialises writers anyway, while a single connection would let one slow query occupy the only connection and make the probe itself time out, turning ordinary contention into needless restarts.
- **Positive:** The rule is general. New dependencies are evaluated against one question instead of re-litigating the trade-off.
- **Negative:** **An identity provider outage is not observable through `/ready`.** This is the accepted cost. Monitor the provider directly, or alert on the rate of `401` responses from `/api/admin/*`, which the OIDC middleware already logs.
- **Negative:** **Storage loss is not observable through `/ready`.** A deleted volume keeps answering `200` while a pooled connection holds the old file handle, and a corrupt file stops the process from booting, so neither surfaces as a failed probe. Alert on the freshness of ingested slides. This is the main operational cost of the decision, and it is the reason the check is not quietly strengthened.
- **Negative:** The container image now pins the port. `ENV APP_PORT=8080` feeds viper's `AutomaticEnv`, and environment variables outrank `app.port` in a mounted `config.yaml` (only an explicit `--port` flag wins). Docker deployments that change the port must now set `APP_PORT` rather than editing the config file.
- **Neutral:** Startup is the one point where the identity provider _is_ contacted. `middleware.AuthMiddleware` performs a live OIDC discovery request before the router is built, so an unreachable issuer prevents the process from booting. That request is bounded by `defaultTimeout` (10 seconds) rather than `http.DefaultClient`, which has no timeout and would block startup indefinitely. The timeout is generous because failing to start is a harder outcome than a slow boot.
