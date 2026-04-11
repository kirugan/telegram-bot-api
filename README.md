# Golang bindings for the Telegram Bot API

[![Go Reference](https://pkg.go.dev/badge/github.com/kirugan/telegram-bot-api/v5.svg)](https://pkg.go.dev/github.com/kirugan/telegram-bot-api/v5)

> **This is a maintained fork of
> [`go-telegram-bot-api/telegram-bot-api`](https://github.com/go-telegram-bot-api/telegram-bot-api).**
> The upstream project stopped updating at Bot API 6.0 (April 2022). This fork
> is brought up to **Bot API 9.6** (the latest release as of this writing) and
> tracks the [official changelog](https://core.telegram.org/bots/api-changelog)
> going forward.

All methods are fairly self-explanatory, and reading the
[godoc](https://pkg.go.dev/github.com/kirugan/telegram-bot-api/v5) page should
explain everything.

The scope of this project is just to provide a wrapper around the API without
any additional features. There are other projects for creating something with
plugins and command handlers without having to design all that yourself.

---

## Installing

```sh
go get github.com/kirugan/telegram-bot-api/v5
```

```go
import tgbotapi "github.com/kirugan/telegram-bot-api/v5"
```

Note: because the module path changed from `go-telegram-bot-api` to `kirugan`,
a Go `replace` directive in `go.mod` **will not** silently redirect existing
code — Go requires the replacement's module path to match the original. Either
update your imports, or keep using upstream on 6.0.

## Migrating from upstream

If you're coming from `go-telegram-bot-api/telegram-bot-api` v5.5.x and want to
jump to the latest Bot API, read [**BREAKING.md**](./BREAKING.md). It lists
every source-incompatible change introduced while bringing the fork from 6.0
to 9.6, grouped by topic, with before/after snippets for each one. The fastest
way to migrate is `grep` your codebase for the old identifier and apply the
rewrite.

### Migrating with Claude (or any LLM)

`BREAKING.md` was written so it can be fed directly to a coding assistant.
A prompt like:

> Here is `BREAKING.md` from a Go library I'm upgrading. Here are the files in
> my project that use `tgbotapi`. Please rewrite them to compile against the
> new API, preserving behavior.

handles most of the mechanical rewrites (`Thumb` → `Thumbnail`,
`ReplyToMessageID` → `ReplyParameters`, `DisableWebPagePreview` →
`LinkPreviewOptions`, sticker creation, etc.) in one pass. Review the diff,
run `go build ./...`, done.

---

## Example

This is a very simple bot that just displays any gotten updates, then replies
it to that chat.

```go
package main

import (
	"log"

	tgbotapi "github.com/kirugan/telegram-bot-api/v5"
)

func main() {
	bot, err := tgbotapi.NewBotAPI("MyAwesomeBotToken")
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil { // If we got a message
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
			msg.ReplyParameters = &tgbotapi.ReplyParameters{
				MessageID: update.Message.MessageID,
			}

			bot.Send(msg)
		}
	}
}
```

If you need to use webhooks, you may use a slightly different method.

```go
package main

import (
	"log"
	"net/http"

	tgbotapi "github.com/kirugan/telegram-bot-api/v5"
)

func main() {
	bot, err := tgbotapi.NewBotAPI("MyAwesomeBotToken")
	if err != nil {
		log.Fatal(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	wh, _ := tgbotapi.NewWebhookWithCert("https://www.example.com:8443/"+bot.Token, tgbotapi.FilePath("cert.pem"))

	_, err = bot.Request(wh)
	if err != nil {
		log.Fatal(err)
	}

	info, err := bot.GetWebhookInfo()
	if err != nil {
		log.Fatal(err)
	}

	if info.LastErrorDate != 0 {
		log.Printf("Telegram callback failed: %s", info.LastErrorMessage)
	}

	updates := bot.ListenForWebhook("/" + bot.Token)
	go http.ListenAndServeTLS("0.0.0.0:8443", "cert.pem", "key.pem", nil)

	for update := range updates {
		log.Printf("%+v\n", update)
	}
}
```

If you need, you may generate a self-signed certificate, as this requires
HTTPS / TLS. The above example tells Telegram that this is your certificate
and that it should be trusted, even though it is not properly signed.

    openssl req -x509 -newkey rsa:2048 -keyout key.pem -out cert.pem -days 3560 -subj "//O=Org\CN=Test" -nodes

Now that [Let's Encrypt](https://letsencrypt.org) is available, you may wish
to generate your free TLS certificate there.

---

## What this fork adds

### Bot API versions

Full support for every Bot API version from **6.1 through 9.6**. See git log
for the per-version commits; each `Full support of API X` commit message is
the authoritative trail for what that version added.

### Upstream issues fixed

Real bugs and gaps reported against the original repository that are now
closed in this fork:

- [#781](https://github.com/go-telegram-bot-api/telegram-bot-api/issues/781)
  / [#745](https://github.com/go-telegram-bot-api/telegram-bot-api/issues/745)
  — `SetGameScoreConfig` serialized the score under the key `scrore`; game
  scores never reached Telegram.
- [#628](https://github.com/go-telegram-bot-api/telegram-bot-api/issues/628)
  — `GetUpdatesChan` logged raw `http.Post` errors, which embed the full
  request URL including the bot token. The token is now redacted before
  logging.
- [#683](https://github.com/go-telegram-bot-api/telegram-bot-api/issues/683)
  — `FileEndpoint` was a hard-coded package constant, so
  `GetFileDirectURL` still pointed at `api.telegram.org` when you were
  running a local Bot API server. There is now a `SetFileEndpoint` /
  `FileLink` pair and the field is per-bot.
- [#740](https://github.com/go-telegram-bot-api/telegram-bot-api/issues/740)
  — `InlineConfig.CacheTime = 0` was silently dropped by `AddNonZero`, so
  Telegram applied its 300s default instead of disabling the cache.
  `cache_time` is now serialized unconditionally.
- [#639](https://github.com/go-telegram-bot-api/telegram-bot-api/issues/639)
  / [#705](https://github.com/go-telegram-bot-api/telegram-bot-api/issues/705)
  — `Send()` tried to unmarshal the bare `true` returned by methods like
  `banChatMember` / `setChatTitle` / `sendChatAction`, producing
  `json: cannot unmarshal bool into Message`. It now returns a zero
  `Message, nil` for those shapes. (Prefer `Request` for methods whose
  documented return type is not a `Message`.)
- [#624](https://github.com/go-telegram-bot-api/telegram-bot-api/issues/624)
  — README webhook example passed `"cert.pem"` as a `string` to
  `NewWebhookWithCert`, which takes a `RequestFileData`. Example now uses
  `tgbotapi.FilePath("cert.pem")`.

### Drive-by improvements

Small enhancements that didn't correspond to a filed issue but were worth
doing while the code was already open:

- **Multipart upload name collision fix.** `prepareInputMediaFile` was using
  `file-%d` for *both* the main media and the thumbnail on `InputMediaAudio`
  / `InputMediaDocument`. The two multipart fields collided in the same form
  body; any audio or document upload with a thumbnail probably didn't work on
  the wire. Thumbnails now use `file-%d-thumbnail`, matching the existing
  `InputMediaVideo` pattern.
- **`closeBody` drains response body before close.** `json.Decoder` can stop
  short of EOF, and `net/http` will discard the underlying TCP connection if
  the body isn't fully consumed — no keep-alive reuse. The fork drains the
  remainder before `Close()`.
- **`Params.AddAny(key, value any) error`.** Replaces the older
  `AddInterface` — same behavior, but uses `any` and has a clearer name.
  `AddInterface` is kept as an alias for existing callers.
- **`Params.AddFirstValid` errors are propagated.** The upstream pattern
  ignored its return value inside `params()` methods, so JSON marshaling
  errors in complex fields were silently swallowed. The fork consistently
  captures and returns it.
- **`json.RawMessage` for JSON-serialized string fields.** Fields the
  Telegram docs describe as "JSON-serialized object" (e.g. `provider_data`
  on `InvoiceConfig` / `InvoiceLinkConfig`) are now `json.RawMessage`
  instead of `string`, so you can hand them an already-marshaled payload
  without double-encoding.
