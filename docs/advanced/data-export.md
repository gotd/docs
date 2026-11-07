---
sidebar_position: 4
---

# Exporting account data (takeout)

Telegram's "Export Telegram data" feature runs inside a special *takeout session* that
relaxes some rate limits for bulk export but is tracked separately by the server. The
[`takeout`](https://ref.gotd.dev/pkg/github.com/gotd/td/telegram/takeout.html) package
wraps `account.initTakeoutSession` and the session lifecycle.

## Running a takeout

Declare what you want to export with `takeout.Config`, then do your work inside
`takeout.Run`. The session is finished automatically when your function returns —
successfully if it returns `nil`:

```go
import "github.com/gotd/td/telegram/takeout"

cfg := takeout.Config{Contacts: true}

return takeout.Run(ctx, client, cfg, func(ctx context.Context, t *takeout.Client) error {
	// t is a tg.Invoker that wraps calls in the takeout session.
	api := tg.NewClient(t)

	res, err := api.ContactsGetContacts(ctx, 0)
	if err != nil {
		return err
	}
	contacts, ok := res.(*tg.ContactsContacts)
	if !ok {
		return nil
	}
	for _, u := range contacts.Users {
		if user, ok := u.AsNotEmpty(); ok {
			fmt.Printf("%d %s %s\n", user.ID, user.FirstName, user.LastName)
		}
	}
	return nil
})
```

The key idea: build a `tg.NewClient(t)` from the takeout client `t`, and every call you
make through it is wrapped in the takeout session.

## Config options

`takeout.Config` mirrors the official export dialog:

| Field                | Exports                          |
|----------------------|----------------------------------|
| `Contacts`           | Contact list                     |
| `MessageUsers`       | Messages from private chats      |
| `MessageChats`       | Messages from basic groups       |
| `MessageMegagroups`  | Messages from supergroups        |
| `MessageChannels`    | Messages from channels           |
| `Files`              | Media files                      |
| `FileMaxSize`        | Cap on exported file size (bytes)|

This is the [`takeout`](https://github.com/gotd/td/tree/main/examples/takeout) example.

## Long exports

A takeout session lives until it is finished, so a long export can save its ID and
resume it after a restart instead of starting a new session (which may require the
user to confirm again):

```go
t, err := takeout.Init(ctx, client, cfg)
if err != nil {
	return err
}
saveTakeoutID(t.ID())

// After restart:
t = takeout.NewClient(loadTakeoutID(), client)
// ... export ...
return t.Finish(ctx, true)
```

New accounts must confirm the export on another device, or wait: init fails with
`TAKEOUT_INIT_DELAY_X`. Set `OnInitDelay` to keep polling until the user confirms:

```go
cfg := takeout.Config{
	Contacts: true,
	OnInitDelay: func(ctx context.Context, d time.Duration) error {
		fmt.Printf("Confirm the export in Telegram (or wait %s)\n", d)
		return nil // keep polling every InitPollInterval (30s)
	},
}
```

`takeout.Run` finishes the session even if its context is canceled.
