---
sidebar_position: 5
---

# Voice and video calls

gotd implements Telegram's [tgcalls](https://core.telegram.org/api/calls) over WebRTC in
the [`calls`](https://ref.gotd.dev/pkg/github.com/gotd/td/telegram/calls.html) package,
supporting both **1:1 calls** and **group voice chats**. This is an advanced feature
that builds on the rest of the library — authenticate and handle updates first.

:::note Audio pipeline
The examples transcode MP3 to Opus with **ffmpeg** and feed the resulting RTP packets
into the call. You need ffmpeg available, and you bring your own audio source. The
helper `examples.StreamMP3(ctx, writePacket, path)` encapsulates the transcoding.
:::

## 1:1 calls

Create a call client, register it on your dispatcher so it sees signaling updates, then
place a call to a user:

```go
import "github.com/gotd/td/telegram/calls"

callClient := calls.NewClient(api, calls.Options{Logger: logzap.New(log)})
callClient.Register(dispatcher)

conn, err := callClient.Request(ctx, user) // user is a tg.InputUserClass
if err != nil {
	return err
}
defer callClient.Discard(ctx, calls.DiscardHangup)

connected := make(chan struct{})
conn.OnConnected(func() { close(connected) })
<-connected

// Stream audio into the call's RTP track.
if err := examples.StreamMP3(ctx, conn.AudioTrack().WriteRTP, audioPath); err != nil {
	return err
}
```

`conn.AudioTrack()` is a `webrtc.TrackLocalStaticRTP` you write packets to. This is the
[`call`](https://github.com/gotd/td/tree/main/examples/call) example.

## Group voice chats

Join a supergroup or channel's voice chat with a `GroupCall`:

```go
gc := calls.NewGroupCall(client.API(), calls.Options{Logger: logzap.New(log)})
gc.Register(dispatcher)
gc.OnParticipants(func(p []tg.GroupCallParticipant) {
	log.Info("participants", zap.Int("count", len(p)))
})

joinAs := &tg.InputPeerUser{UserID: self.ID, AccessHash: self.AccessHash}
if err := gc.Join(ctx, call, joinAs); err != nil { // call is a *tg.InputGroupCall
	return err
}
defer gc.Leave(ctx)

if err := examples.StreamMP3(ctx, gc.WriteAudio, audioPath); err != nil {
	return err
}
```

`gc.WriteAudio(pkt *rtp.Packet)` feeds audio into the group call. This is the
[`groupcall`](https://github.com/gotd/td/tree/main/examples/groupcall) example.

## Caveats

* Calls require a **user** session (QR or code login), not a bot.
* You must obtain the `*tg.InputGroupCall` for a chat before joining (resolve the chat,
  then read its full info).
* Video is supported by the underlying protocol, but the examples stream audio only.
* This is one of the more involved parts of the library — read the example sources in
  full before building on them.
