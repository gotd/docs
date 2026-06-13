---
sidebar_position: 10
---

# Architecture

botapi exposes the **Telegram Bot API** surface (types, methods, updates) but
implements it **directly over MTProto** via [`gotd/td`](https://github.com/gotd/td)
— not over HTTP to `api.telegram.org`. This page sketches how the Bot API
surface maps onto the MTProto building blocks underneath. The Bot API docs
([core.telegram.org/bots/api](https://core.telegram.org/bots/api)) are the spec.

## The translation layer

A user holds a `botapi.Message`, never a `tg.Message`. Between the two sits a
translation layer that botapi owns; everything below it is `gotd/td` doing the
protocol heavy lifting. botapi never re-implements MTProto, peer resolution, gap
recovery or file transfer — it translates.

```
                      ┌─────────────────────────────────────────┐
   BotFather token →  │ telegram.Client (MTProto, one per bot)   │
                      │   Auth().Bot(token)                      │
                      └───────────────┬──────────────────────────┘
                                      │ API() *tg.Client
        ┌─────────────────────────────┼───────────────────────────────┐
        ▼ outgoing                     ▼ peers                          ▼ incoming
  message.Sender            peers.Manager (Storage/Cache)      UpdateHandler chain:
  uploader/downloader       ResolveTDLibID / ResolveDomain      peers.UpdateHook
  fileid codec              InputPeer + access hashes            → updates.Manager (gaps)
                                                                 → tg.UpdateDispatcher
                                                                    (OnNewMessage, …)
        └──────────────── translation layer (botapi) ─────────────────┘
                                      │
                        Bot API types & methods (hand-written)
```

## The building blocks

botapi sits on a handful of `gotd/td` primitives, each documented in this site:

| Block | Role | Docs |
| --- | --- | --- |
| `telegram.Client` | The MTProto connection; `Auth().Bot(token)` logs the bot in. | [First client](../getting-started/first-client.md), [Bot auth](../authentication/bot.md) |
| `message.Sender` | Fluent builder for everything outgoing. | [Message sender](../helpers/message-sender.md) |
| `telegram/peers` | Resolves bare Bot-API chat IDs into `InputPeer`s with access hashes. | [Peers](../helpers/peers.md) |
| `telegram/updates` | Gap-aware update manager (`getDifference` recovery). | [Updates recovery](../helpers/updates-recovery.md) |
| `tg.UpdateDispatcher` | Typed update fan-out — the terminal handler. | [Handling updates](../basics/handling-updates.md) |
| `uploader` / `downloader` | File transfer for `GetFile`-style operations. | [Uploads](../helpers/uploading-files.md), [Downloads](../helpers/downloading-files.md) |
| `fileid` | Codec between Bot API `file_id` strings and MTProto locations. | [Downloading files](../helpers/downloading-files.md) |
| `tgerr` | RPC error matching, mapped onto Bot API `{error_code, description}`. | [Errors & resilience](./errors-and-resilience.md) |

## Update flow — no long-poll, no webhook

Because botapi is on MTProto, there is **no `getUpdates` and no webhook**.
Updates arrive on the persistent connection and flow through a verified chain:

```
tg updates → peers.Manager.UpdateHook (harvest access hashes)
           → updates.Manager.Handle    (gap recovery)
           → tg.UpdateDispatcher        (typed fan-out)
           → botapi mapping             (tg.Update* → botapi.Update)
           → handler router             (predicates, middleware, handlers)
```

The `peers.UpdateHook` is what keeps access hashes fresh without explicit calls:
every update's users and chats are harvested into storage before your handler
sees it. That is why a bot can later address any peer it has "seen."

## Type-safe by construction

* **Sealed-interface unions** — where the Bot API uses "one of" objects
  (`ChatID`, `InputFile`, `ReplyMarkup`, `InputMedia`, `ChatMember`,
  `InlineQueryResult`, `InputMessageContent`, …), botapi uses an interface with
  an unexported marker method and a fixed set of implementations. Illegal states
  are unrepresentable, and switches over them are checked for exhaustiveness by a
  linter.
* **Typed enums** — `ParseMode`, `ChatType`, `ChatAction`, `MessageEntityType`,
  … are typed constants, not bare strings the caller can mistype.
* **Zero-reflection request building** — there is no JSON marshaling on the wire;
  methods translate typed params straight into `gotd/td` builder calls. Hot paths
  get allocation-test coverage like `gotd/td` itself.

## Conformance

botapi keeps a copy of the published Bot API docs (`internal/botdoc`) as a
reference oracle. A conformance test asserts that every published method is
either implemented on `*Bot`, covered by other means, or explicitly categorized
as deferred — so when Telegram ships a new Bot API version, the drift surfaces as
a failing test rather than a silent gap.

For the full design rationale and rebuild history, see the upstream
[`architecture.md`](https://github.com/gotd/botapi/blob/main/docs/architecture.md),
[`building-blocks.md`](https://github.com/gotd/botapi/blob/main/docs/building-blocks.md)
and [`roadmap.md`](https://github.com/gotd/botapi/blob/main/docs/roadmap.md).
