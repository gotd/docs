---
sidebar_position: 7
---

# Editing & chat management

## Editing, forwarding, deleting

```go
bot.EditMessageText(ctx, chat, messageID, "new text")
bot.EditMessageCaption(ctx, chat, messageID, "new caption")
bot.EditMessageReplyMarkup(ctx, chat, messageID, markup)
bot.ForwardMessage(ctx, toChat, fromChat, messageID)
bot.CopyMessage(ctx, toChat, fromChat, messageID)
bot.DeleteMessage(ctx, chat, messageID)
bot.DeleteMessages(ctx, chat, []int{id1, id2})
```

Live locations have their own editing methods: `EditMessageLiveLocation` and
`StopMessageLiveLocation`.

## Chat members

For supergroups and channels:

| Method | Use |
| --- | --- |
| `BanChatMember` / `UnbanChatMember` | Ban or unban a user. |
| `RestrictChatMember` | Apply `ChatPermissions`. |
| `PromoteChatMember` | Grant `ChatAdminRights`. |
| `GetChatMember` | Look up one member. |
| `GetChatAdministrators` | List admins. |
| `GetChatMemberCount` | Member count. |

`ChatPermissions` and `ChatAdminRights` map to the corresponding MTProto rights;
the incoming participant is converted to the `ChatMember` union.

## Chat administration

```go
bot.PinChatMessage(ctx, chat, messageID)
bot.UnpinChatMessage(ctx, chat, messageID)
bot.UnpinAllChatMessages(ctx, chat)
bot.SetChatTitle(ctx, chat, "New title")
bot.SetChatDescription(ctx, chat, "New description")
bot.SetChatPermissions(ctx, chat, perms)
bot.SetChatPhoto(ctx, chat, photo)
bot.DeleteChatPhoto(ctx, chat)
bot.LeaveChat(ctx, chat)
```

### Invite links

```go
bot.ExportChatInviteLink(ctx, chat)
bot.CreateChatInviteLink(ctx, chat, ...)
bot.EditChatInviteLink(ctx, chat, link, ...)
bot.RevokeChatInviteLink(ctx, chat, link)
```

## Stickers

```go
bot.UploadStickerFile(ctx, userID, sticker)
bot.CreateNewStickerSet(ctx, ...)
bot.AddStickerToSet(ctx, ...)
bot.DeleteStickerFromSet(ctx, sticker)
bot.SetStickerPositionInSet(ctx, sticker, position)
```

Stickers are described with `InputSticker` and a `StickerFormat`.

:::note[Deferred methods]
A few methods are not implemented yet: `GetStickerSet` /
`SetStickerSetThumb` (need full sticker-set conversion), `EditMessageMedia`, and
the payment answers `AnswerPreCheckoutQuery` / `AnswerShippingQuery` (waiting on
payment-update plumbing). Track them in the
[roadmap](https://github.com/gotd/botapi/blob/main/docs/roadmap.md). Anything not
yet covered is reachable through the [raw client](./persistence-and-pooling.md#the-escape-hatch).
:::
