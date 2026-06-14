---
sidebar_position: 7
---

# Support

This project is fully open-source and support is done voluntarily by the
community, so **no SLA is provided**.

## Community channels

Important news, major updates and security issues are posted in the news channel:

* [**@gotd_news**](https://t.me/gotd_news) — the gotd news channel.

Development and user support are provided in our chat groups:

[![Telegram: English chat](https://api.go-faster.org/badge/telegram/gotd_en?title=EN&v=1)](https://t.me/gotd_en)
[![Telegram: Russian chat](https://api.go-faster.org/badge/telegram/gotd_ru?title=RU&v=1)](https://t.me/gotd_ru)
[![Telegram: Chinese chat](https://api.go-faster.org/badge/telegram/gotd_zhcn?title=CN&v=1)](https://t.me/gotd_zhcn)
[![Telegram: Online count](https://api.go-faster.org/badge/telegram/online?groups=gotd_ru&groups=gotd_en&groups=gotd_zhcn)](https://t.me/gotd_en)

* [**@gotd_en**](https://t.me/gotd_en) — English.
* [**@gotd_ru**](https://t.me/gotd_ru) — Russian.
* [**@gotd_zhcn**](https://t.me/gotd_zhcn) — Chinese.

While we recommend using [test servers](https://core.telegram.org/api/datacenter#test-datacenters),
you can join [**@gotd_test**](https://t.me/gotd_test) for testing in production.

## How to not get banned?

**Do not share your app's ID and hash!** They cannot be regenerated and are
bound to your Telegram account.

:::warning
All clients are strictly monitored to prevent abuse.

If you try to use a Telegram client for flooding, spamming, faking subscribers
and view counts you will be banned permanently.
:::

> Due to excessive abuse of the Telegram API, **all accounts that sign up or
> log in using unofficial Telegram clients are automatically put under
> observation** to avoid violations of the
> [Terms of Service](https://core.telegram.org/api/terms).
>
> — [Official documentation](https://core.telegram.org/api/obtaining_api_id)

### A note from MadelineProto

There is a useful quote from the [MadelineProto](https://docs.madelineproto.xyz/docs/LOGIN.html)
docs on getting permission to use the Telegram API.

Before you start using the Telegram API, you have to understand that Telegram
strictly controls userbots created on their platform. If you use any Telegram
client, **including official clients**, for flooding, spamming or botting
channels, you **will be banned forever**.

Due to excessive abuse of the Telegram API, **all phone numbers** that sign up
or log in — **even using official or Telegram-approved API clients** — are
automatically put **under observation**, and **may** get banned **even if you
did nothing wrong**, simply because some internal flags are triggered on the
Telegram servers.

To avoid this, you must let Telegram know that you will use your account with a
userbot. When or before you first sign up or log in, send an email with the
phone number to [recover@telegram.org](mailto:recover@telegram.org) explaining
**what your userbot will do**. **Do not lie** — just tell them what you intend
to do, asking them not to ban your account. If your account does get banned,
write to the same address asking to unban it.

**Do not abuse this or any other API for flooding, spamming or botting** — the
consequences fall not only on you, but on all other users of this and other
libraries, and even normal users. There were cases when several **normal user
accounts that did nothing wrong** were banned when Telegram deployed a new
spambot detection system: this is bad for the community and bad for Telegram, so
please do not abuse.

### Summary

1. This client is unofficial; Telegram treats such clients suspiciously,
   especially fresh ones.
2. Use bots whenever possible.
3. If you still want to automate things with a user, use it passively (i.e.
   receive more than you send).
4. When using it with a user:
   * Do it with extreme care.
   * Do not use VoIP numbers.
   * Do not abuse, spam or use it for other suspicious activities.
   * Implement a rate-limiting system.
   * _Generally_, this is a bad idea if you're not 100% sure what you're doing.

Bad usage of the API can trigger Telegram's anti-abuse system and ban all your
accounts forever.

## What to do if I got banned?

First of all, there's no reason to panic. The automated anti-abuse system makes
incorrect bans often. See
[discussions](https://github.com/lonamiwebs/telethon/issues/824#issuecomment-432182634)
in other Telegram API libraries for more context.

Second, write to [recover@telegram.org](mailto:recover@telegram.org) explaining
what you intend to do with the API, asking to unban your account.

Third, if you follow the "How to not get banned?" recommendations and suspect
that something in this library can trigger the anti-abuse system,
[create an issue](https://github.com/gotd/td/issues/new) with a detailed
description of what you were doing.
