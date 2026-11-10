---
sidebar_position: 8
---

# Observability

gotd can report what it does through three channels:

- [**Metrics**](#metrics) via OpenTelemetry. Off unless you pass a `MeterProvider`.
- [**Logs**](#logs) via [`gotd/log`](https://github.com/gotd/log). Off unless you pass a
  `Logger`.
- [**Hooks**](#hooks), which are callbacks for events your application should react to.

Nothing is recorded by default. gotd never falls back to the global OpenTelemetry
providers, and with no provider set, instrumentation costs no allocations on the request
path.

## Setup

Pass the same providers to the client, the updates manager and the downloader:

```go
import (
	"github.com/gotd/log"
	"github.com/gotd/log/logzap"
	"go.opentelemetry.io/otel"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/telegram/updates"
)

mp := otel.GetMeterProvider() // or your SDK MeterProvider
tp := otel.GetTracerProvider()
lg := logzap.New(zapLogger)

gaps := updates.New(updates.Config{
	Handler:        dispatcher,
	Logger:         log.Named(lg, "updates"),
	MeterProvider:  mp,
	TracerProvider: tp,
})

client := telegram.NewClient(appID, appHash, telegram.Options{
	UpdateHandler:  gaps,
	Logger:         lg,
	MeterProvider:  mp,
	TracerProvider: tp,
})

// client.Downloader() and client.Download* use the client's MeterProvider and Logger.
// The updates manager needs its own MeterProvider (above). For your own Downloader:
d := downloader.NewDownloader().
	WithMeterProvider(mp).
	WithLogger(log.Named(lg, "downloader"))
```

Most events are logged at debug level, so with an info-level logger you only see the
abnormal ones (see [Logs](#logs)). Raise the level when you need to see more.

Metric attributes never contain user, chat or channel IDs, file locations or message
content, so the number of series stays bounded. IDs only appear in logs.

### Several clients in one process

gotd doesn't label metrics by client. Two clients (e.g. two accounts) sharing one
`MeterProvider` add up into the same counters, and their gauges such as
`gotd.client.pools` overwrite each other. Give each client its own provider that adds an
instrumentation-scope attribute to every meter gotd creates:

```go
import (
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// scoped adds attrs to the instrumentation scope of every meter.
type scoped struct {
	metric.MeterProvider
	attrs []attribute.KeyValue
}

func (p scoped) Meter(name string, opts ...metric.MeterOption) metric.Meter {
	return p.MeterProvider.Meter(name, append(opts, metric.WithInstrumentationAttributes(p.attrs...))...)
}

mp := scoped{MeterProvider: shared, attrs: []attribute.KeyValue{attribute.String("account", name)}}
// Use mp in telegram.Options and updates.Config of that account.
```

`client.Downloader()` uses the client's provider, so its metrics are separated too. A
`Downloader` you share between clients is not.

With Prometheus' OTLP receiver, set `otlp.promote_scope_metadata: true` to turn the
scope attribute into a label (`otel_scope_account` above).

## Metrics

### Client (`telegram.Options.MeterProvider`)

| Metric | Type | Attributes | Meaning |
| --- | --- | --- | --- |
| `gotd.client.connection.state` | counter | `dc`, `state` (`connecting`, `ready`, `disconnected`) | Primary connection state changes. |
| `gotd.client.connection.dead` | counter | `dc`, `kind` (`primary`, `data`, `cdn`) | Connections found dead (pings, read errors). |
| `gotd.client.connection.reconnects` | counter | `dc` | Reconnects of the primary connection. |
| `gotd.client.connection.restarts` | counter | `reason` (`timeout`, `migration`) | Restarts requested by the client itself; each is also counted in `connection.reconnects`. |
| `gotd.client.migrations` | counter | `dc` (target) | Home DC changes (`*_MIGRATE`). |
| `gotd.client.auth.lost` | counter | `reason`, `kind` | Authorized → lost transitions (see `OnAuthLost`). `reason` is the error type, `kind` is `banned`, `deleted`, `revoked`, `unregistered`, `expired` or `key_invalid`. |
| `gotd.client.account.restricted` | counter | `reason` | Errors telling that the account is restricted: `FROZEN_METHOD_INVALID`, `FROZEN_PARTICIPANT_MISSING`, `PHONE_NUMBER_BANNED`. |
| `gotd.client.request.timeouts` | counter | `result` (`resent`, `failed`), `method` | Requests that missed their `RequestTimeout` deadline. |
| `gotd.client.pool.created` | counter | `dc`, `kind` (`file`, `migrate`), `result` (`ok`, `error`, `canceled`) | Connection pools created for other DCs. |
| `gotd.client.pool.create.duration` | histogram, s | `kind`, `result` | Time to create a pool (connect and authorization transfer). |
| `gotd.client.pool.evicted` | counter | `dc`, `kind`, `reason` | Cached pools dropped, e.g. after `AUTH_KEY_UNREGISTERED`. |
| `gotd.client.file.migrations` | counter | `dc` (target), `source` | `FILE_MIGRATE` hops. |
| `gotd.client.pools` | gauge | `kind` (`file`, `migrate`, `cdn`) | Open pools while `Run` runs. |
| `gotd.client.invoke.duration` | histogram, s | `result` (`ok`, `rpc_error` with `code`, `canceled`, `error`) | Duration of every request, including `client.API()` calls. |

`invoke.duration` has no method attribute: one series per TL method would be too many.
For per-method metrics, add a [middleware](../helpers/middleware).

### Updates (`updates.Config.MeterProvider`)

| Metric | Type | Attributes | Meaning |
| --- | --- | --- | --- |
| `gotd.updates.difference.requests` | counter | `kind` (`common`, `channel`), `result` (`empty`, `slice`, `full`, `too_long`, `error`, `inaccessible`) | `getDifference` / `getChannelDifference` calls. |
| `gotd.updates.difference.duration` | histogram, s | `kind`, `result` | Their duration. |
| `gotd.updates.difference.triggers` | counter | `kind`, `reason` (`startup`, `pts-gap-timeout`, `idle-timeout`, `channel-too-long`, …) | Why a difference was requested. |
| `gotd.updates.gaps` | counter | `kind`, `seq` (`pts`, `qts`, `seq`), `action` (`detected`, `resolved_by_waiting`, `resolved_by_pending`) | Sequence gaps and how they were closed. |
| `gotd.updates.channel.too_long` | counter | `dialog` (`ok`, `stale`, `invalid`) | `channelDifferenceTooLong` gaps, one per `OnChannelGap` call. |
| `gotd.updates.channel.gap.size` | histogram, pts | | `RemotePts - LocalPts` of those gaps. |
| `gotd.updates.channel.gap.messages` | histogram | | Number of messages returned with a gap. |
| `gotd.updates.before_run` | counter, batches | `action` (`passed`, `buffered`, `dropped`, `replayed`, `passed_raw`) | Updates received before `Run`, see `BeforeRun`. |
| `gotd.updates.channel.track` | counter | `source` (`api`, `stored`, `update`, `too_long`), `result` | Channels starting to be tracked. |
| `gotd.updates.channels.tracked` | up-down counter | | Channels tracked now. |
| `gotd.updates.channel.inaccessible` | counter | | Channels dropped after `CHANNEL_PRIVATE`. |
| `gotd.updates.stopped` | counter | `op` | Calls rejected with `updates.ErrStopped`. |
| `gotd.updates.dispatched` | counter | `source` (`live`, `difference`) | Updates passed to the handler. |

### Downloader (`Downloader.WithMeterProvider`)

| Metric | Type | Attributes | Meaning |
| --- | --- | --- | --- |
| `gotd.downloader.downloads` | counter | `result` (`ok`, `error`, `canceled`) | Finished downloads. |
| `gotd.downloader.requests` | counter | `kind` (`master`, `cdn`, `web`), `result` (`ok`, `flood_wait`, `timeout`, `canceled`, `error`) | Part requests. |
| `gotd.downloader.request.duration` | histogram, s | `kind` | Part request duration. |
| `gotd.downloader.bytes` | counter, By | `kind` | Bytes downloaded. |
| `gotd.downloader.resumes` | counter | | Downloads started with `WithOffset`. |
| `gotd.downloader.cdn.redirects` | counter | | Files served from a CDN DC. |
| `gotd.downloader.cdn.reuploads` | counter | | `upload.reuploadCdnFile` requests. |
| `gotd.downloader.hash.mismatches` | counter | `kind` (`cdn`, `verify`) | Parts that failed hash verification. |
| `gotd.downloader.file_reference.refreshes` | counter | `result` (`ok`, `not_found`, `error`, `limit`) | `WithRefresh` attempts. |
| `gotd.downloader.verify` | counter | `result` (`complete`, `partial`, `none`, `error`) | `Builder.Verify` results. |
| `gotd.downloader.retries` | counter | `operation` | Retried part requests. |
| `gotd.downloader.retries.exhausted` | counter | `operation` | Requests that ran out of retries. |
| `gotd.downloader.scheduler.backoffs` | counter | `reason` (`flood_wait`, `timeout`) | Backoffs reported to a `Scheduler`. |
| `gotd.downloader.scheduler.window` | gauge | your attributes | Current window of a `Scheduler`. |
| `gotd.downloader.scheduler.in_flight` | gauge | your attributes | Part requests in flight. |
| `gotd.downloader.scheduler.waiting` | gauge | your attributes | Part requests waiting for a slot. |
| `gotd.downloader.scheduler.decreases` | counter | your attributes | Window decreases. |

Scheduler gauges are registered by `Scheduler.Instrument`, with your own attributes, for
example the DC. Schedulers the client creates from `Options.DownloadScheduler` are
registered with a `dc` attribute while `Run` runs. For your own schedulers:

```go
reg, err := sched.Instrument(meter, attribute.Int("dc", 2))
defer reg.Unregister()
```

### What to watch

- **Counters versus state.** `connection.state` counts transitions. The first `ready`
  after start is a counter's first value, which PromQL's `increase()` doesn't show, and
  an idle client looks the same as a dead one. To show disconnects, chart increases of
  `state="disconnected"` and `connection.dead`. "Connected now" can't be read from these
  counters.
- **The client is healthy:**
  - `connection.state{state="ready"}` grows only on reconnects;
  - `connection.dead` and `request.timeouts` stay near zero;
  - `invoke.duration` p95 is stable.
- **Updates are complete:**
  - `difference.requests{result="too_long"}` and `channel.too_long` are rare;
  - `gaps{action="detected"}` is followed by a resolution;
  - `before_run{action="dropped"}` is zero, unless you chose `BeforeRunDrop`.
- **Authorization:** any increase of `auth.lost` needs attention; `kind="banned"` or
  `"deleted"` means the account is gone. Any `account.restricted` means a frozen
  account or a banned number.
- **Downloads keep up:**
  - `downloads{result="error"}`, `hash.mismatches` and `retries.exhausted` stay near zero;
  - `scheduler.window` recovers after `scheduler.backoffs`.

## Logs

Events are logged through the `Logger` you pass. Attributes are structured fields, and
IDs (channel, DC, pts) appear here, never in metrics.

| Level | Component | Events |
| --- | --- | --- |
| Debug | client | Connection state, dead data and CDN connections, pool created, `FILE_MIGRATE` hop, restart requested |
| Info | client | Home DC migration (from, to), popup service notifications (type, text) |
| Warn | client | Dead primary connection, pool creation failed, pool evicted after `AUTH_KEY_UNREGISTERED`, lost authorization, request timeout, `AUTH_KEY_DROP_*` notification, frozen account (since, until, appeal URL) |
| Error | client | Lost authorization of a banned or deleted account, `PHONE_NUMBER_BANNED` |
| Debug | updates | Each difference request with result and duration, channel tracked, update received before `Run` |
| Info | updates | Channel gap (channel, local and remote pts, top message, number of messages), replay of updates buffered before `Run` |
| Warn | updates | `updates.differenceTooLong`, channel pts that did not advance, failed `TrackChannel` |
| Debug | downloader | Part requests, CDN redirect, reupload, file reference refreshed |
| Warn | downloader | Hash mismatch, failed refresh, retries exhausted |

## Hooks

Use hooks to act on events. Use metrics and logs to watch them.

### Client (`telegram.Options`)

| Hook | Called when |
| --- | --- |
| `OnConnectionStateEvent func(ConnectionStateEvent)` | The primary connection state changes; `DC`, `State`. Prefer it to `OnConnectionState`. |
| `OnConnDead func(ConnDeadEvent)` | Any connection is dead; `DC`, `Kind`, `Err`. Prefer it to `OnDead`. |
| `OnAuthLost func(ctx, err) error` | An authorized session is lost. Runs on its own goroutine with `Run`'s context. Return an error to stop `Run` with `*AuthLostError`. See [Long-running clients](long-running). |
| `OnFrozen func(ctx, FreezeState)` | A request failed because the account is frozen, and app config confirms it. Checked at most once an hour. See [Frozen accounts](long-running#frozen-accounts). |
| `OnSelfError` / `OnSelfSuccess` | The `Self` call after connecting fails or succeeds. |
| `OnTransfer` | Authorization is transferred to another DC. |
| `RequestTimeout func(bin.Encoder) time.Duration` | Per-request response deadline; one resend after a reconnect. Return 0 for requests that must not be sent twice. |
| `PingInterval`, `PingTimeout` | Liveness checks of connections. |
| `Middlewares` | Wrap every request, e.g. for per-method metrics or rate limits. |

### Updates (`updates.Config`)

| Hook | Called when |
| --- | --- |
| `OnChannelGap func(ChannelGap)` | A channel gap can't be recovered. Has the pts range, the top message, and the latest `Messages` with `Users` and `Chats`; backfill with `messages.getHistory`. |
| `OnChannelTooLong func(channelID)` | Same event, ID only. |
| `OnChannelInaccessible func(channelID)` | Access to a channel is lost (`CHANNEL_PRIVATE`); it stops being tracked. |
| `OnTooLong func()` | The common state gap can't be recovered (`updates.differenceTooLong`). |
| `OnLoadUserStateFailed func()` | No stored state on start, so a full resync follows. |
| `OnLoadChannelStateFailed func(channelID)` | A stored channel can't be loaded; track it again with `Manager.TrackChannel`. |
| `BeforeRun`, `BeforeRunBufferSize` | What happens to updates received before `Run`: pass, buffer or drop. |

### Dispatcher (`tg.UpdateDispatcher`)

| Hook | Called when |
| --- | --- |
| `OnBefore func(ctx, e, update) error` | Every update, before its handler, including updates without one. Return an error to skip the handler. Useful for per-type counters or tracing. |
| `OnFallback` | An update has no handler. |
| `OnAny*` | One handler for private chats, groups and channels; see [Handling updates](../basics/handling-updates). |

### Downloads

| Hook | Called when |
| --- | --- |
| `Builder.WithProgress(func(done, total int64))` | Bytes are written. |
| `Builder.WithRefresh(RefreshFunc)` | A file reference expired; return a fresh location (see `downloader/fileref`). |
| `Scheduler.Snapshot()` | Read the window, in-flight, waiting and decreases at any time. |

### Other packages

| Hook | Package |
| --- | --- |
| `Config.OnInitDelay` | `telegram/takeout`: `TAKEOUT_INIT_DELAY` while waiting for the export to be allowed. |
| `Waiter.WithCallback(func(ctx, FloodWait))`, `Waiter.Snapshot()` | `gotd/contrib/middleware/floodwait`: every flood wait (type, wait, error) and per-type state. |

## Tracing

`TracerProvider` in `telegram.Options` and `updates.Config` adds spans for requests and
update processing; see [Debugging](debugging).
