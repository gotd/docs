package main

// namespaceDescriptions maps a Telegram API namespace to a short, human
// description for the reference overview. Namespaces not listed fall back to an
// empty description.
var namespaceDescriptions = map[string]string{
	"account":   "Account settings: privacy, sessions, two-factor auth, profile and notifications.",
	"aicompose": "AI-assisted message composition.",
	"auth":      "Authorization: sign in and out, login codes, passwords and bot login.",
	"bots":      "Bot-side methods: commands, inline results, menus and web apps.",
	"channels":  "Channels and supergroups: administration, members and settings.",
	"chatlists": "Shareable chat-folder (chat list) invite links.",
	"contacts":  "Contacts: add, block, search and import/export.",
	"folders":   "Peer folders, including the archive.",
	"fragment":  "Fragment.com collectible usernames and phone numbers.",
	"help":      "Server configuration, app updates, support and terms of service.",
	"langpack":  "Localization and language packs.",
	"messages":  "Sending, editing and fetching messages, plus reactions, stickers and drafts.",
	"payments":  "Payments, invoices, Telegram Stars and subscriptions.",
	"phone":     "Voice and video calls, including group calls.",
	"photos":    "Profile photos.",
	"premium":   "Telegram Premium features: boosts, gifts and more.",
	"smsjobs":   "SMS jobs (carrier verification integration).",
	"stats":     "Channel, message and story statistics.",
	"stickers":  "Creating and managing sticker sets.",
	"stories":   "Stories: posting, viewing, reactions and privacy.",
	"test":      "Test-only methods for the Telegram test servers.",
	"updates":   "Fetching update state and differences.",
	"upload":    "Uploading and downloading file parts.",
	"users":     "User information and full profiles.",
}
