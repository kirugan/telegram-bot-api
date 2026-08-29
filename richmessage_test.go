package tgbotapi

import (
	"encoding/json"
	"reflect"
	"testing"
)

// TestRichTextUnmarshalForms verifies that the polymorphic RichText type
// decodes the three wire forms: plain string, array, and styled span.
func TestRichTextUnmarshalForms(t *testing.T) {
	var plain RichText
	if err := json.Unmarshal([]byte(`"hello"`), &plain); err != nil {
		t.Fatalf("plain unmarshal: %v", err)
	}
	if !plain.IsPlain || plain.PlainText != "hello" {
		t.Fatalf("plain form not decoded: %+v", plain)
	}

	var arr RichText
	if err := json.Unmarshal([]byte(`["a","b"]`), &arr); err != nil {
		t.Fatalf("array unmarshal: %v", err)
	}
	if len(arr.Parts) != 2 || arr.Parts[0].PlainText != "a" {
		t.Fatalf("array form not decoded: %+v", arr)
	}

	var span RichText
	if err := json.Unmarshal([]byte(`{"type":"url","text":"site","url":"https://example.com"}`), &span); err != nil {
		t.Fatalf("span unmarshal: %v", err)
	}
	if span.Type != RichTextTypeURL || span.URL != "https://example.com" {
		t.Fatalf("span form not decoded: %+v", span)
	}
	if span.Text == nil || !span.Text.IsPlain || span.Text.PlainText != "site" {
		t.Fatalf("nested span text not decoded: %+v", span.Text)
	}
}

// TestRichTextMarshalRoundTrip verifies that each RichText form survives a
// decode/encode/decode cycle unchanged.
func TestRichTextMarshalRoundTrip(t *testing.T) {
	cases := []string{
		`"plain"`,
		`["a","b"]`,
		`{"type":"bold","text":"x"}`,
		`{"type":"date_time","text":"now","unix_time":1700000000,"date_time_format":"d MMM"}`,
		`{"type":"text_mention","text":"u","user":{"id":1,"is_bot":true,"first_name":"A"}}`,
	}
	for _, in := range cases {
		var first RichText
		if err := json.Unmarshal([]byte(in), &first); err != nil {
			t.Fatalf("unmarshal %s: %v", in, err)
		}
		out, err := json.Marshal(first)
		if err != nil {
			t.Fatalf("marshal %s: %v", in, err)
		}
		var second RichText
		if err := json.Unmarshal(out, &second); err != nil {
			t.Fatalf("re-unmarshal %s: %v", out, err)
		}
		// Raw differs by construction (only set on decode), so clear it.
		first.Raw, second.Raw = nil, nil
		if !reflect.DeepEqual(first, second) {
			t.Errorf("round trip mismatch:\n  in:  %s\n  out: %s", in, out)
		}
	}
}

// TestRichBlockCaptionRouting verifies that the shared "caption" wire field is
// routed to TableCaption for a table block and to Caption for a media block.
func TestRichBlockCaptionRouting(t *testing.T) {
	var table RichBlock
	if err := json.Unmarshal([]byte(`{"type":"table","cells":[],"caption":"a table"}`), &table); err != nil {
		t.Fatalf("table unmarshal: %v", err)
	}
	if table.TableCaption == nil || table.TableCaption.PlainText != "a table" {
		t.Fatalf("table caption not routed to TableCaption: %+v", table)
	}
	if table.Caption != nil {
		t.Fatalf("table caption wrongly set Caption: %+v", table.Caption)
	}

	var photo RichBlock
	if err := json.Unmarshal([]byte(`{"type":"photo","photo":[],"caption":{"text":"pic"}}`), &photo); err != nil {
		t.Fatalf("photo unmarshal: %v", err)
	}
	if photo.Caption == nil || !photo.Caption.Text.IsPlain || photo.Caption.Text.PlainText != "pic" {
		t.Fatalf("photo caption not routed to Caption: %+v", photo)
	}
	if photo.TableCaption != nil {
		t.Fatalf("photo caption wrongly set TableCaption: %+v", photo.TableCaption)
	}

	// The table caption must marshal back out under the shared "caption" key.
	out, err := json.Marshal(table)
	if err != nil {
		t.Fatalf("table marshal: %v", err)
	}
	var check map[string]json.RawMessage
	if err := json.Unmarshal(out, &check); err != nil {
		t.Fatalf("remarshal check: %v", err)
	}
	if string(check["caption"]) != `"a table"` {
		t.Errorf("table caption not emitted under caption key: %s", out)
	}
}

// TestRichMessageDecode verifies a full RichMessage with nested blocks decodes.
func TestRichMessageDecode(t *testing.T) {
	raw := `{"blocks":[{"type":"heading","text":"Title","size":1},{"type":"paragraph","text":["plain ",{"type":"bold","text":"bold"}]}],"is_rtl":false}`
	var rm RichMessage
	if err := json.Unmarshal([]byte(raw), &rm); err != nil {
		t.Fatalf("rich message unmarshal: %v", err)
	}
	if len(rm.Blocks) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(rm.Blocks))
	}
	if rm.Blocks[0].Type != RichBlockTypeHeading || rm.Blocks[0].Size != 1 {
		t.Errorf("heading block not decoded: %+v", rm.Blocks[0])
	}
	para := rm.Blocks[1]
	if para.Type != RichBlockTypeParagraph || para.Text == nil || len(para.Text.Parts) != 2 {
		t.Fatalf("paragraph block not decoded: %+v", para)
	}
	if para.Text.Parts[1].Type != RichTextTypeBold {
		t.Errorf("nested bold span not decoded: %+v", para.Text.Parts[1])
	}
}

// TestInputRichMessageParams verifies the send-side config serializes the
// rich_message parameter as JSON.
func TestInputRichMessageParams(t *testing.T) {
	cfg := NewRichMessage(123, InputRichMessage{HTML: "<b>hi</b>"})
	params, err := cfg.params()
	if err != nil {
		t.Fatalf("params: %v", err)
	}
	if params["chat_id"] != "123" {
		t.Errorf("chat_id = %q", params["chat_id"])
	}
	if cfg.method() != "sendRichMessage" {
		t.Errorf("method = %q", cfg.method())
	}
	var got InputRichMessage
	if err := json.Unmarshal([]byte(params["rich_message"]), &got); err != nil {
		t.Fatalf("rich_message param not valid JSON: %v", err)
	}
	if got.HTML != "<b>hi</b>" {
		t.Errorf("rich_message html = %q", got.HTML)
	}
}

// TestRichBlockButtonsAndDocument verifies the Bot API 10.3 block types:
// "buttons", "document", and "expandable_blockquote", plus the "button" span
// and the is_compact table flag.
func TestRichBlockButtonsAndDocument(t *testing.T) {
	buttons := InputRichBlock{
		Type:  RichBlockTypeButtons,
		Align: "center",
		Buttons: []RichMessageButton{{
			Text:         RichText{IsPlain: true, PlainText: "Press"},
			Style:        "primary",
			CallbackData: "cb",
		}},
	}
	out, err := json.Marshal(buttons)
	if err != nil {
		t.Fatalf("marshal buttons block: %v", err)
	}
	want := `{"type":"buttons","buttons":[{"text":"Press","style":"primary","callback_data":"cb"}],"align":"center"}`
	if string(out) != want {
		t.Errorf("buttons block encoding:\n  got:  %s\n  want: %s", out, want)
	}

	raw := `{"type":"document","document":{"file_id":"f1","file_unique_id":"u1"},` +
		`"caption":{"text":"a file"}}`
	var doc RichBlock
	if err := json.Unmarshal([]byte(raw), &doc); err != nil {
		t.Fatalf("unmarshal document block: %v", err)
	}
	if doc.Document == nil || doc.Document.FileID != "f1" {
		t.Errorf("document not decoded: %+v", doc.Document)
	}
	if doc.Caption == nil || !doc.Caption.Text.IsPlain || doc.Caption.Text.PlainText != "a file" {
		t.Errorf("document caption not routed: %+v", doc.Caption)
	}

	quote := InputRichBlock{
		Type: RichBlockTypeExpandableBlockquote,
		Text: &RichText{IsPlain: true, PlainText: "long quote"},
	}
	out, err = json.Marshal(quote)
	if err != nil {
		t.Fatalf("marshal expandable blockquote: %v", err)
	}
	if got, want := string(out), `{"type":"expandable_blockquote","text":"long quote"}`; got != want {
		t.Errorf("expandable blockquote encoding:\n  got:  %s\n  want: %s", got, want)
	}

	table := InputRichBlock{Type: RichBlockTypeTable, IsCompact: true}
	out, err = json.Marshal(table)
	if err != nil {
		t.Fatalf("marshal table: %v", err)
	}
	if got, want := string(out), `{"type":"table","is_compact":true}`; got != want {
		t.Errorf("compact table encoding:\n  got:  %s\n  want: %s", got, want)
	}

	span := RichText{Type: RichTextTypeButton, Button: &RichMessageButton{
		Text: RichText{IsPlain: true, PlainText: "Go"},
		URL:  "https://example.com",
	}}
	out, err = json.Marshal(span)
	if err != nil {
		t.Fatalf("marshal button span: %v", err)
	}
	if got, want := string(out), `{"type":"button","button":{"text":"Go","url":"https://example.com"}}`; got != want {
		t.Errorf("button span encoding:\n  got:  %s\n  want: %s", got, want)
	}
}
