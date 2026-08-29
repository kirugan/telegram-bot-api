package tgbotapi

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInputRichBlockCaptionRouting verifies that the shared "caption" wire
// field is routed to TableCaption for a table block and to Caption for a media
// block, mirroring RichBlock.
func TestInputRichBlockCaptionRouting(t *testing.T) {
	var table InputRichBlock
	require.NoError(t, json.Unmarshal([]byte(`{"type":"table","cells":[],"caption":"a table"}`), &table))
	require.NotNil(t, table.TableCaption)
	assert.Equal(t, "a table", table.TableCaption.PlainText)
	assert.Nil(t, table.Caption)

	var photo InputRichBlock
	require.NoError(t, json.Unmarshal([]byte(`{"type":"photo","caption":{"text":"a photo"}}`), &photo))
	require.NotNil(t, photo.Caption)
	assert.True(t, photo.Caption.Text.IsPlain)
	assert.Equal(t, "a photo", photo.Caption.Text.PlainText)
	assert.Nil(t, photo.TableCaption)
}

// TestInputRichBlockRoundTrip verifies that the block variants survive a
// decode/encode/decode cycle unchanged.
func TestInputRichBlockRoundTrip(t *testing.T) {
	cases := []string{
		`{"type":"paragraph","text":"hello"}`,
		`{"type":"heading","text":"title","size":2}`,
		`{"type":"pre","text":"x := 1","language":"go"}`,
		`{"type":"divider"}`,
		`{"type":"mathematical_expression","expression":"e^{i\\pi}+1=0"}`,
		`{"type":"anchor","name":"intro"}`,
		`{"type":"list","items":[{"blocks":[{"type":"paragraph","text":"a"}],"value":1,"type":"1"}]}`,
		`{"type":"blockquote","blocks":[{"type":"paragraph","text":"q"}],"credit":"someone"}`,
		`{"type":"details","summary":"more","blocks":[{"type":"paragraph","text":"d"}],"is_open":true}`,
		`{"type":"table","cells":[[{"text":"c","align":"left","valign":"top"}]],"is_bordered":true,"caption":"cap"}`,
		`{"type":"map","location":{"latitude":1.5,"longitude":2.5},"zoom":10,"width":600,"height":400}`,
		`{"type":"thinking","text":"working"}`,
	}

	for _, in := range cases {
		var first InputRichBlock
		require.NoError(t, json.Unmarshal([]byte(in), &first), "unmarshal %s", in)
		out, err := json.Marshal(first)
		require.NoError(t, err, "marshal %s", in)
		var second InputRichBlock
		require.NoError(t, json.Unmarshal(out, &second), "re-unmarshal %s", out)
		assert.Equal(t, first, second, "round trip mismatch for %s", in)
	}
}

// TestInputRichBlockMediaMarshal verifies the encoding of the media-bearing
// block variants. These are marshal-only: the embedded InputMedia* types hold
// the file as a RequestFileData interface, which cannot be decoded, so they
// are not part of the round-trip test above.
func TestInputRichBlockMediaMarshal(t *testing.T) {
	photo := NewInputMediaPhoto(FileID("file_id"))
	block := InputRichBlock{
		Type:    RichBlockTypePhoto,
		Photo:   &photo,
		Caption: &RichBlockCaption{Text: RichText{IsPlain: true, PlainText: "pic"}},
	}
	out, err := json.Marshal(block)
	require.NoError(t, err)
	assert.Equal(t, `{"type":"photo","photo":{"type":"photo","media":"file_id"},"caption":{"text":"pic"}}`, string(out))

	voice := NewInputMediaVoiceNote(FileID("file_id"))
	voice.Duration = 5
	out, err = json.Marshal(InputRichBlock{Type: RichBlockTypeVoiceNote, VoiceNote: &voice})
	require.NoError(t, err)
	assert.Equal(t, `{"type":"voice_note","voice_note":{"type":"voice_note","media":"file_id","duration":5}}`, string(out))
}

// TestInputRichMessageBlocks verifies that a block-formatted rich message and
// a markdown rich message with explicit media both encode as expected.
func TestInputRichMessageBlocks(t *testing.T) {
	msg := InputRichMessage{
		Blocks: []InputRichBlock{
			{Type: RichBlockTypeParagraph, Text: &RichText{IsPlain: true, PlainText: "hi"}},
		},
	}
	out, err := json.Marshal(msg)
	require.NoError(t, err)
	assert.Equal(t, `{"blocks":[{"type":"paragraph","text":"hi"}]}`, string(out))

	withMedia := InputRichMessage{
		Markdown: "![pic](tg://photo?id=p1)",
		Media: []InputRichMessageMedia{
			{ID: "p1", Media: NewInputMediaPhoto(FileID("file_id"))},
		},
	}
	out, err = json.Marshal(withMedia)
	require.NoError(t, err)
	assert.Equal(t, `{"markdown":"![pic](tg://photo?id=p1)","media":[{"id":"p1","media":{"type":"photo","media":"file_id"}}]}`, string(out))
}

// TestEphemeralSendParams verifies that the ephemeral parameters are only
// emitted when set, and that they reach the wire for a send config.
func TestEphemeralSendParams(t *testing.T) {
	plain := NewMessage(12345, "hello")
	params, err := plain.params()
	require.NoError(t, err)
	assert.NotContains(t, params, "ephemeral_message_parameters")

	ephemeral := NewMessage(12345, "hello")
	ephemeral.ReceiverUserID = 777
	ephemeral.CallbackQueryID = "cbq"
	ephemeral.ReplaceCallbackQueryMessage = true
	params, err = ephemeral.params()
	require.NoError(t, err)
	want := `{"receiver_user_id":777,"callback_query_id":"cbq","replace_callback_query_message":true}`
	assert.Equal(t, want, params["ephemeral_message_parameters"])
}

// TestEphemeralEditConfigs verifies the params and methods of the ephemeral
// edit and delete configs.
func TestEphemeralEditConfigs(t *testing.T) {
	edit := NewEditEphemeralMessageText(12345, 777, 42, "updated")
	params, err := edit.params()
	require.NoError(t, err)
	for key, want := range map[string]string{
		"chat_id":              "12345",
		"receiver_user_id":     "777",
		"ephemeral_message_id": "42",
		"text":                 "updated",
	} {
		assert.Equal(t, want, params[key], key)
	}
	assert.Equal(t, "editEphemeralMessageText", edit.method())

	del := NewDeleteEphemeralMessage(12345, 777, 42)
	params, err = del.params()
	require.NoError(t, err)
	assert.NotContains(t, params, "reply_markup", "deleteEphemeralMessage must not send reply_markup")
	assert.Equal(t, "deleteEphemeralMessage", del.method())
}

// TestReplyParametersEphemeral verifies that ReplyParameters can address an
// ephemeral message and that message_id is omitted when unset.
func TestReplyParametersEphemeral(t *testing.T) {
	out, err := json.Marshal(ReplyParameters{EphemeralMessageID: 42})
	require.NoError(t, err)
	assert.Equal(t, `{"ephemeral_message_id":42}`, string(out))
}

// TestPrepareInputMediaAnimation verifies that InputMediaAnimation is handled
// by the input media preparation helpers. It was previously missing from both
// switches, so animations were silently dropped from the request.
func TestPrepareInputMediaAnimation(t *testing.T) {
	byID := NewInputMediaAnimation(FileID("file_id"))
	require.NotNil(t, prepareInputMediaParam(byID, 0), "prepareInputMediaParam dropped InputMediaAnimation")

	upload := NewInputMediaAnimation(FilePath("/tmp/anim.gif"))
	upload.Thumbnail = FilePath("/tmp/thumb.jpg")

	param := prepareInputMediaParam(upload, 3)
	animation, ok := param.(InputMediaAnimation)
	require.True(t, ok, "prepareInputMediaParam returned %T, want InputMediaAnimation", param)
	assert.Equal(t, "attach://file-3", animation.Media.SendData())
	assert.Equal(t, "attach://file-3-thumbnail", animation.Thumbnail.SendData())

	files := prepareInputMediaFile(upload, 3)
	require.Len(t, files, 2)
	assert.Equal(t, "file-3", files[0].Name)
	assert.Equal(t, "file-3-thumbnail", files[1].Name)
}

// TestUpdateSentFromSubscription verifies that SentFrom resolves the user of a
// subscription update.
func TestUpdateSentFromSubscription(t *testing.T) {
	update := Update{
		Subscription: &BotSubscriptionUpdated{
			User:  User{ID: 5, FirstName: "A"},
			State: BotSubscriptionStateCanceled,
		},
	}
	from := update.SentFrom()
	require.NotNil(t, from, "SentFrom returned nil for a subscription update")
	assert.Equal(t, int64(5), from.ID)
}

// TestCommunityServiceMessages verifies decoding of the community service
// messages and the subscription update.
func TestCommunityServiceMessages(t *testing.T) {
	var msg Message
	raw := `{"message_id":1,"date":1,"chat":{"id":1,"type":"supergroup"},` +
		`"community_chat_added":{"community":{"id":99,"name":"Gophers"}}}`
	require.NoError(t, json.Unmarshal([]byte(raw), &msg))
	require.NotNil(t, msg.CommunityChatAdded, "community_chat_added not decoded")
	assert.Equal(t, int64(99), msg.CommunityChatAdded.Community.ID)
	assert.Equal(t, "Gophers", msg.CommunityChatAdded.Community.Name)

	var update Update
	rawUpdate := `{"update_id":1,"subscription":{"user":{"id":5,"is_bot":false,"first_name":"A"},` +
		`"invoice_payload":"p","state":"active"}}`
	require.NoError(t, json.Unmarshal([]byte(rawUpdate), &update))
	require.NotNil(t, update.Subscription, "subscription not decoded")
	assert.Equal(t, BotSubscriptionStateActive, update.Subscription.State)
	assert.Equal(t, int64(5), update.Subscription.User.ID)
}
