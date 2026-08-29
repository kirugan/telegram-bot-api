package tgbotapi

import (
	"encoding/json"
	"reflect"
	"testing"
)

// TestInputRichBlockCaptionRouting verifies that the shared "caption" wire
// field is routed to TableCaption for a table block and to Caption for a media
// block, mirroring RichBlock.
func TestInputRichBlockCaptionRouting(t *testing.T) {
	var table InputRichBlock
	if err := json.Unmarshal([]byte(`{"type":"table","cells":[],"caption":"a table"}`), &table); err != nil {
		t.Fatalf("table unmarshal: %v", err)
	}
	if table.TableCaption == nil || table.TableCaption.PlainText != "a table" {
		t.Fatalf("table caption not routed to TableCaption: %+v", table)
	}
	if table.Caption != nil {
		t.Fatalf("table caption also set Caption: %+v", table.Caption)
	}

	var photo InputRichBlock
	if err := json.Unmarshal([]byte(`{"type":"photo","caption":{"text":"a photo"}}`), &photo); err != nil {
		t.Fatalf("photo unmarshal: %v", err)
	}
	if photo.Caption == nil || !photo.Caption.Text.IsPlain || photo.Caption.Text.PlainText != "a photo" {
		t.Fatalf("photo caption not routed to Caption: %+v", photo)
	}
	if photo.TableCaption != nil {
		t.Fatalf("photo caption also set TableCaption: %+v", photo.TableCaption)
	}
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
		if err := json.Unmarshal([]byte(in), &first); err != nil {
			t.Fatalf("unmarshal %s: %v", in, err)
		}
		out, err := json.Marshal(first)
		if err != nil {
			t.Fatalf("marshal %s: %v", in, err)
		}
		var second InputRichBlock
		if err := json.Unmarshal(out, &second); err != nil {
			t.Fatalf("re-unmarshal %s: %v", out, err)
		}
		if !reflect.DeepEqual(first, second) {
			t.Errorf("round trip mismatch:\n  in:  %s\n  out: %s", in, out)
		}
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
	if err != nil {
		t.Fatalf("marshal photo block: %v", err)
	}
	want := `{"type":"photo","photo":{"type":"photo","media":"file_id"},"caption":{"text":"pic"}}`
	if string(out) != want {
		t.Errorf("photo block encoding:\n  got:  %s\n  want: %s", out, want)
	}

	voice := NewInputMediaVoiceNote(FileID("file_id"))
	voice.Duration = 5
	out, err = json.Marshal(InputRichBlock{Type: RichBlockTypeVoiceNote, VoiceNote: &voice})
	if err != nil {
		t.Fatalf("marshal voice note block: %v", err)
	}
	want = `{"type":"voice_note","voice_note":{"type":"voice_note","media":"file_id","duration":5}}`
	if string(out) != want {
		t.Errorf("voice note block encoding:\n  got:  %s\n  want: %s", out, want)
	}
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
	if err != nil {
		t.Fatalf("marshal blocks: %v", err)
	}
	if got, want := string(out), `{"blocks":[{"type":"paragraph","text":"hi"}]}`; got != want {
		t.Errorf("blocks encoding:\n  got:  %s\n  want: %s", got, want)
	}

	withMedia := InputRichMessage{
		Markdown: "![pic](tg://photo?id=p1)",
		Media: []InputRichMessageMedia{
			{ID: "p1", Media: NewInputMediaPhoto(FileID("file_id"))},
		},
	}
	out, err = json.Marshal(withMedia)
	if err != nil {
		t.Fatalf("marshal media: %v", err)
	}
	want := `{"markdown":"![pic](tg://photo?id=p1)","media":[{"id":"p1","media":{"type":"photo","media":"file_id"}}]}`
	if got := string(out); got != want {
		t.Errorf("media encoding:\n  got:  %s\n  want: %s", got, want)
	}
}

// TestEphemeralSendParams verifies that the ephemeral parameters are only
// emitted when set, and that they reach the wire for a send config.
func TestEphemeralSendParams(t *testing.T) {
	plain := NewMessage(12345, "hello")
	params, err := plain.params()
	if err != nil {
		t.Fatalf("plain params: %v", err)
	}
	if _, ok := params["ephemeral_message_parameters"]; ok {
		t.Error("ephemeral_message_parameters emitted for a non-ephemeral message")
	}

	ephemeral := NewMessage(12345, "hello")
	ephemeral.ReceiverUserID = 777
	ephemeral.CallbackQueryID = "cbq"
	ephemeral.ReplaceCallbackQueryMessage = true
	params, err = ephemeral.params()
	if err != nil {
		t.Fatalf("ephemeral params: %v", err)
	}
	want := `{"receiver_user_id":777,"callback_query_id":"cbq","replace_callback_query_message":true}`
	if params["ephemeral_message_parameters"] != want {
		t.Errorf("ephemeral_message_parameters = %q, want %q", params["ephemeral_message_parameters"], want)
	}
}

// TestEphemeralEditConfigs verifies the params and methods of the ephemeral
// edit and delete configs.
func TestEphemeralEditConfigs(t *testing.T) {
	edit := NewEditEphemeralMessageText(12345, 777, 42, "updated")
	params, err := edit.params()
	if err != nil {
		t.Fatalf("edit params: %v", err)
	}
	for key, want := range map[string]string{
		"chat_id":              "12345",
		"receiver_user_id":     "777",
		"ephemeral_message_id": "42",
		"text":                 "updated",
	} {
		if params[key] != want {
			t.Errorf("%s = %q, want %q", key, params[key], want)
		}
	}
	if edit.method() != "editEphemeralMessageText" {
		t.Errorf("method = %q", edit.method())
	}

	del := NewDeleteEphemeralMessage(12345, 777, 42)
	params, err = del.params()
	if err != nil {
		t.Fatalf("delete params: %v", err)
	}
	if _, ok := params["reply_markup"]; ok {
		t.Error("deleteEphemeralMessage must not send reply_markup")
	}
	if del.method() != "deleteEphemeralMessage" {
		t.Errorf("method = %q", del.method())
	}
}

// TestReplyParametersEphemeral verifies that ReplyParameters can address an
// ephemeral message and that message_id is omitted when unset.
func TestReplyParametersEphemeral(t *testing.T) {
	out, err := json.Marshal(ReplyParameters{EphemeralMessageID: 42})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if got, want := string(out), `{"ephemeral_message_id":42}`; got != want {
		t.Errorf("encoding:\n  got:  %s\n  want: %s", got, want)
	}
}

// TestPrepareInputMediaAnimation verifies that InputMediaAnimation is handled
// by the input media preparation helpers. It was previously missing from both
// switches, so animations were silently dropped from the request.
func TestPrepareInputMediaAnimation(t *testing.T) {
	byID := NewInputMediaAnimation(FileID("file_id"))
	if param := prepareInputMediaParam(byID, 0); param == nil {
		t.Fatal("prepareInputMediaParam dropped InputMediaAnimation")
	}

	upload := NewInputMediaAnimation(FilePath("/tmp/anim.gif"))
	upload.Thumbnail = FilePath("/tmp/thumb.jpg")

	param := prepareInputMediaParam(upload, 3)
	animation, ok := param.(InputMediaAnimation)
	if !ok {
		t.Fatalf("prepareInputMediaParam returned %T, want InputMediaAnimation", param)
	}
	if got := animation.Media.SendData(); got != "attach://file-3" {
		t.Errorf("media = %q, want attach://file-3", got)
	}
	if got := animation.Thumbnail.SendData(); got != "attach://file-3-thumbnail" {
		t.Errorf("thumbnail = %q, want attach://file-3-thumbnail", got)
	}

	files := prepareInputMediaFile(upload, 3)
	if len(files) != 2 {
		t.Fatalf("prepareInputMediaFile returned %d files, want 2", len(files))
	}
	if files[0].Name != "file-3" || files[1].Name != "file-3-thumbnail" {
		t.Errorf("file names = %q, %q", files[0].Name, files[1].Name)
	}
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
	if from == nil {
		t.Fatal("SentFrom returned nil for a subscription update")
	}
	if from.ID != 5 {
		t.Errorf("SentFrom().ID = %d, want 5", from.ID)
	}
}

// TestCommunityServiceMessages verifies decoding of the community service
// messages and the subscription update.
func TestCommunityServiceMessages(t *testing.T) {
	var msg Message
	raw := `{"message_id":1,"date":1,"chat":{"id":1,"type":"supergroup"},` +
		`"community_chat_added":{"community":{"id":99,"name":"Gophers"}}}`
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatalf("unmarshal message: %v", err)
	}
	if msg.CommunityChatAdded == nil {
		t.Fatal("community_chat_added not decoded")
	}
	if msg.CommunityChatAdded.Community.ID != 99 || msg.CommunityChatAdded.Community.Name != "Gophers" {
		t.Errorf("community not decoded: %+v", msg.CommunityChatAdded.Community)
	}

	var update Update
	rawUpdate := `{"update_id":1,"subscription":{"user":{"id":5,"is_bot":false,"first_name":"A"},` +
		`"invoice_payload":"p","state":"active"}}`
	if err := json.Unmarshal([]byte(rawUpdate), &update); err != nil {
		t.Fatalf("unmarshal update: %v", err)
	}
	if update.Subscription == nil {
		t.Fatal("subscription not decoded")
	}
	if update.Subscription.State != BotSubscriptionStateActive || update.Subscription.User.ID != 5 {
		t.Errorf("subscription not decoded: %+v", update.Subscription)
	}
}
