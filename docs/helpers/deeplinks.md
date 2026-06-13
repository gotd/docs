---
sidebar_position: 9
---

# Deeplinks

Telegram links like `https://t.me/durov`, `tg://resolve?domain=durov` and
`t.me/+AbCdEf` encode an action and its parameters. The
[`deeplink`](https://ref.gotd.dev/pkg/github.com/gotd/td/telegram/deeplink.html) package
parses them **offline** — no client, no login, no network.

```go
import "github.com/gotd/td/telegram/deeplink"

if !deeplink.IsDeeplinkLike(link) {
	return fmt.Errorf("%q is not a deeplink", link)
}

d, err := deeplink.Parse(link)
if err != nil {
	return err
}

switch d.Type {
case deeplink.Resolve:
	fmt.Println("resolve username:", d.Args.Get("domain"))
case deeplink.Join:
	fmt.Println("join via invite:", d.Args.Get("invite"))
case deeplink.BusinessChat:
	fmt.Println("business chat slug:", d.Args.Get("slug"))
}
```

`Parse` returns a `DeepLink` with a `Type` and a `url.Values` of `Args`. The recognised
types are:

| Type                   | Example links                                            | Key argument |
|------------------------|----------------------------------------------------------|--------------|
| `deeplink.Resolve`     | `t.me/durov`, `tg://resolve?domain=durov`               | `domain`     |
| `deeplink.Join`        | `t.me/+AbCd`, `t.me/joinchat/AbCd`, `tg://join?invite=…`| `invite`     |
| `deeplink.BusinessChat`| `t.me/m/slug`, `tg://message?slug=…`                    | `slug`       |

`IsDeeplinkLike` is a cheap pre-check so you can route arbitrary user input before
attempting a full parse. This is the
[`deeplink`](https://github.com/gotd/td/tree/main/examples/deeplink) example.

## Resolving the link

Parsing gives you the *intent*; to act on it you still resolve the target online. The
[message sender](./message-sender.md) and [peers manager](./peers.md) accept links
directly:

```go
sender.ResolveDeeplink(link).Text(ctx, "hi")
```
