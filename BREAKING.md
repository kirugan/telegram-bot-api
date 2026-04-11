# Breaking Changes

This file lists every source-incompatible change introduced while bringing
this fork from Bot API 6.0 up to 9.6. Upgrades are grouped by topic so you
can jump straight to the area your code touches.

Each section lists **what to rename** or **what to replace** — the fastest
way to migrate is `grep` for the old identifier in your code and apply the
rewrite. Nothing here changes runtime semantics beyond what the Telegram
Bot API itself changed.

---

## `thumb` → `thumbnail` (Bot API 6.6)

The single largest mechanical rename. Applies everywhere media has a
thumbnail.

### Config fields
```go
PhotoConfig.Thumb           → PhotoConfig.Thumbnail
AudioConfig.Thumb           → AudioConfig.Thumbnail
DocumentConfig.Thumb        → DocumentConfig.Thumbnail
VideoConfig.Thumb           → VideoConfig.Thumbnail
AnimationConfig.Thumb       → AnimationConfig.Thumbnail
VideoNoteConfig.Thumb       → VideoNoteConfig.Thumbnail
VoiceConfig.Thumb           → VoiceConfig.Thumbnail
```

### `InputMedia*` types
```go
InputMediaVideo.Thumb       → InputMediaVideo.Thumbnail
InputMediaAnimation.Thumb   → InputMediaAnimation.Thumbnail
InputMediaAudio.Thumb       → InputMediaAudio.Thumbnail
InputMediaDocument.Thumb    → InputMediaDocument.Thumbnail
```

### Inline query result types
```go
InlineQueryResult*.ThumbURL    → .ThumbnailURL
InlineQueryResult*.ThumbWidth  → .ThumbnailWidth
InlineQueryResult*.ThumbHeight → .ThumbnailHeight
```
(Applies to Article, Contact, Document, Location, Venue, Photo, Video,
GIF, MPEG4GIF.)

### Helper functions
```go
NewInlineQueryResultPhotoWithThumb → NewInlineQueryResultPhotoWithThumbnail
```

### Config types
```go
SetStickerSetThumbConfig → SetStickerSetThumbnailConfig
SetStickerSetThumbConfig.Thumb → SetStickerSetThumbnailConfig.Thumbnail
```

---

## `reply_to_message_id` → `ReplyParameters` (Bot API 7.0)

`ReplyToMessageID` and `AllowSendingWithoutReply` were removed from
`BaseChat` and `MediaGroupConfig`. Replies now go through a structured
`ReplyParameters` value that also supports cross-chat replies and quoting.

### Before
```go
msg := tgbotapi.NewMessage(chatID, "hi")
msg.ReplyToMessageID = origMsgID
msg.AllowSendingWithoutReply = true
```

### After
```go
msg := tgbotapi.NewMessage(chatID, "hi")
msg.ReplyParameters = &tgbotapi.ReplyParameters{
    MessageID:                origMsgID,
    AllowSendingWithoutReply: true,
}
```

---

## `DisableWebPagePreview` → `LinkPreviewOptions` (Bot API 7.0)

### Send / edit configs
```go
MessageConfig.DisableWebPagePreview     → LinkPreviewOptions
EditMessageTextConfig.DisableWebPagePreview → LinkPreviewOptions
```

### Inline content
```go
InputTextMessageContent.DisableWebPagePreview → LinkPreviewOptions
```

### Before
```go
msg := tgbotapi.NewMessage(chatID, "see https://example.com")
msg.DisableWebPagePreview = true
```

### After
```go
msg := tgbotapi.NewMessage(chatID, "see https://example.com")
msg.LinkPreviewOptions = &tgbotapi.LinkPreviewOptions{IsDisabled: true}
```

---

## `forward_from` fields → `ForwardOrigin` (Bot API 7.0)

Removed from `Message`:
- `ForwardFrom *User`
- `ForwardFromChat *Chat`
- `ForwardFromMessageID int`
- `ForwardSignature string`
- `ForwardSenderName string`
- `ForwardDate int`

Added: `ForwardOrigin *MessageOrigin` — a flat polymorphic struct with a
`Type` discriminator ("user", "hidden_user", "chat", "channel") and all
variant fields as optional.

### Before
```go
if msg.ForwardFrom != nil {
    fmt.Println("forwarded from user", msg.ForwardFrom.UserName)
}
```

### After
```go
if msg.ForwardOrigin != nil && msg.ForwardOrigin.Type == tgbotapi.MessageOriginTypeUser {
    fmt.Println("forwarded from user", msg.ForwardOrigin.SenderUser.UserName)
}
```

---

## Sticker API rewrite (Bot API 6.6)

`createNewStickerSet`, `uploadStickerFile`, and `addStickerToSet` all
changed shape. The old per-format parameters (`png_sticker`, `tgs_sticker`,
`webm_sticker`) were replaced with the new `InputSticker` type.

### `UploadStickerConfig`
```go
// Before
cfg := tgbotapi.UploadStickerConfig{
    UserID:     userID,
    PNGSticker: tgbotapi.FilePath("sticker.png"),
}

// After
cfg := tgbotapi.UploadStickerConfig{
    UserID:        userID,
    Sticker:       tgbotapi.FilePath("sticker.png"),
    StickerFormat: tgbotapi.StickerFormatStatic,
}
```

### `NewStickerSetConfig`
```go
// Before
cfg := tgbotapi.NewStickerSetConfig{
    UserID:     userID,
    Name:       "my_pack",
    Title:      "My Pack",
    PNGSticker: tgbotapi.FilePath("s.png"),
    Emojis:     "😀",
}

// After — Stickers is a slice of InputSticker, each with its own Format (7.2)
cfg := tgbotapi.NewStickerSetConfig{
    UserID: userID,
    Name:   "my_pack",
    Title:  "My Pack",
    Stickers: []tgbotapi.InputSticker{{
        Sticker:   tgbotapi.FilePath("s.png"),
        Format:    tgbotapi.StickerFormatStatic,
        EmojiList: []string{"😀"},
    }},
}
```
Dropped fields: `PNGSticker`, `TGSSticker`, `Emojis`, `MaskPosition`,
`StickerFormat` (moved per-sticker in 7.2), `ContainsMasks` (deprecated 6.2).

### `AddStickerConfig`
```go
// Before
cfg := tgbotapi.AddStickerConfig{
    UserID:     userID,
    Name:       "my_pack",
    PNGSticker: tgbotapi.FilePath("new.png"),
    Emojis:     "🎉",
}

// After
cfg := tgbotapi.AddStickerConfig{
    UserID: userID,
    Name:   "my_pack",
    Sticker: tgbotapi.InputSticker{
        Sticker:   tgbotapi.FilePath("new.png"),
        Format:    tgbotapi.StickerFormatStatic,
        EmojiList: []string{"🎉"},
    },
}
```

### Removed fields
```go
StickerSet.IsAnimated      // (removed 7.2; mixed-format packs)
StickerSet.IsVideo         // (removed 7.2)
StickerSet.ContainsMasks   // (removed 6.2; use StickerType instead)
```

---

## Chat permissions: granular media (Bot API 6.5)

`CanSendMediaMessages` was replaced with six per-type fields on both
`ChatPermissions` and `ChatMember`:

```go
CanSendMediaMessages →
    CanSendAudios +
    CanSendDocuments +
    CanSendPhotos +
    CanSendVideos +
    CanSendVideoNotes +
    CanSendVoiceNotes
```

---

## `UserShared` → `UsersShared` (Bot API 7.0, 7.2)

### 7.0 — rename + slice of IDs
```go
KeyboardButtonRequestUser  → KeyboardButtonRequestUsers
UserShared (type)          → UsersShared
KeyboardButton.RequestUser → KeyboardButton.RequestUsers
Message.UserShared         → Message.UsersShared
UsersShared.UserID (int64) → UsersShared.UserIDs ([]int64)
```

### 7.2 — slice of user IDs becomes slice of `SharedUser`
```go
UsersShared.UserIDs ([]int64) → UsersShared.Users ([]SharedUser)
```

`SharedUser` has the basic profile info (first/last name, username, photo)
when the bot requested it via `KeyboardButtonRequestUsers`.

---

## `BusinessConnection.CanReply` → `Rights` (Bot API 9.0)

`BusinessConnection.CanReply bool` was replaced with
`BusinessConnection.Rights *BusinessBotRights`, which carries ~14
granular permissions (including a new `CanReply` field inside the
sub-struct).

### Before
```go
if bc.CanReply {
    ...
}
```

### After
```go
if bc.Rights != nil && bc.Rights.CanReply {
    ...
}
```

---

## `ChatFullInfo.CanSendGift` → `AcceptedGiftTypes` (Bot API 9.0)

The single-bool `can_send_gift` was replaced with a struct describing which
specific kinds of gifts a chat accepts:

```go
ChatFullInfo.CanSendGift (bool) → ChatFullInfo.AcceptedGiftTypes (AcceptedGiftTypes)
```

`AcceptedGiftTypes` has four bools: `UnlimitedGifts`, `LimitedGifts`,
`UniqueGifts`, `PremiumSubscription` (plus `GiftsFromChannels` added in 9.3).

---

## `Poll.CorrectOptionID` → `CorrectOptionIDs` (Bot API 9.6)

Multi-answer quizzes meant the singular field had to become a slice:

```go
Poll.CorrectOptionID (int)        → Poll.CorrectOptionIDs ([]int)
SendPollConfig.CorrectOptionID    → SendPollConfig.CorrectOptionIDs
```

`SendPollConfig.CorrectOptionID` also changed from `int64` to `int` as part
of the rewrite.

---

## `SendPollConfig.Options` type change (Bot API 7.3)

```go
SendPollConfig.Options ([]string) → SendPollConfig.Options ([]InputPollOption)
```

The `NewPoll` helper still takes `options ...string` and wraps them
internally, so callers that use the helper don't need to change.
Direct struct literal callers do:

### Before
```go
SendPollConfig{Options: []string{"yes", "no"}}
```

### After
```go
SendPollConfig{Options: []tgbotapi.InputPollOption{
    {Text: "yes"},
    {Text: "no"},
}}
```

---

## `UniqueGiftInfo.LastResaleStarCount` → currency+amount (Bot API 9.3)

The Stars-only resale field was generalized when TON payments arrived:

```go
UniqueGiftInfo.LastResaleStarCount (int) →
    UniqueGiftInfo.LastResaleCurrency (string) +
    UniqueGiftInfo.LastResaleAmount (int)
```

---

## `GetBusinessAccountGiftsConfig.ExcludeLimited` split (Bot API 9.3)

```go
ExcludeLimited (bool) →
    ExcludeLimitedUpgradable (bool) +
    ExcludeLimitedNonUpgradable (bool)
```

Plus `ExcludeFromBlockchain (bool)` was added in the same version.

---

## `switch_pm_*` → `InlineQueryResultsButton` (Bot API 6.7)

```go
InlineConfig.SwitchPMText      → InlineConfig.Button.Text
InlineConfig.SwitchPMParameter → InlineConfig.Button.StartParameter
```

`InlineConfig.Button` is a `*InlineQueryResultsButton` which additionally
supports launching a Web App:

```go
cfg := tgbotapi.InlineConfig{
    // ...
    Button: &tgbotapi.InlineQueryResultsButton{
        Text:           "Show help",
        StartParameter: "help",
    },
}
```

---

## Removed fields (no direct replacement)

- **`InlineQueryResultArticle.HideURL`** (Bot API 8.2) — pass an empty
  string as the `URL` instead.
- **`Gift.ContainsMasks`** / **`StickerSet.ContainsMasks`** —
  use the `StickerType` / `Type` field (`"regular"`, `"mask"`,
  `"custom_emoji"`) instead.

---

## `InvoiceConfig.ProviderData` type change (Bot API 6.1 era)

```go
InvoiceConfig.ProviderData     (string) → json.RawMessage
InvoiceLinkConfig.ProviderData (string) → json.RawMessage
```

If you had a pre-marshalled JSON string, wrap it with `json.RawMessage(s)`.

---

## Method return type change: `GetChat` (Bot API 7.3)

```go
bot.GetChat(cfg) (Chat, error) → (ChatFullInfo, error)
```

`ChatFullInfo` embeds `Chat`, so field access via `.ID`, `.Title`, `.Bio`,
etc. still works through Go's field promotion. The only thing that changes
is the variable's declared type:

```go
// Before
var c tgbotapi.Chat
c, err = bot.GetChat(cfg)

// After
var c tgbotapi.ChatFullInfo
c, err = bot.GetChat(cfg)
```

---

## Types that gained fields (soft-breaking)

A few types changed from empty structs to carrying fields. Anyone who used
them via pointer (`*Story`) is unaffected; callers who constructed them by
value may need to initialize new fields.

- **`Story`** — 6.8 empty struct → 7.1 `{Chat, ID}`
- **`GiveawayCreated`** — 7.0 empty struct → 7.10 `{PrizeStarCount}`
- **`WriteAccessAllowed`** — 6.4 empty struct → grew `WebAppName` (6.7),
  `FromRequest`, `FromAttachmentMenu` (6.9)

---

## Internal helper `AddFirstValid` error propagation

Not API-breaking for callers of the public surface, but if you wrote
custom `Chattable` implementations and copied the old pattern of
ignoring the `Params.AddFirstValid` return value, know that the fork
now consistently captures and returns that error from `params()`. Your
custom configs should do the same.

---

## Checklist for upgrading

1. Global find-replace for the `Thumb` → `Thumbnail` renames (Section 1).
2. Search for `ReplyToMessageID` and migrate each to `ReplyParameters`.
3. Search for `DisableWebPagePreview` and migrate to `LinkPreviewOptions`.
4. Search for `ForwardFrom` / `ForwardDate` and switch to `ForwardOrigin`.
5. Anything touching sticker set creation needs the `InputSticker` rewrite.
6. If you read `ChatMember.CanSendMediaMessages` or
   `ChatPermissions.CanSendMediaMessages`, switch to the six granular
   `CanSend*` fields.
7. If you handle business connections, check `Rights` instead of `CanReply`.
8. If you read quiz correctness, use `CorrectOptionIDs[0]` (or loop for
   multi-answer).

Everything else is additive and should compile unchanged.
