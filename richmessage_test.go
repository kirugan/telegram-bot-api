package tgbotapi

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRichTextUnmarshalForms verifies that the polymorphic RichText type
// decodes the three wire forms: plain string, array, and styled span.
func TestRichTextUnmarshalForms(t *testing.T) {
	var plain RichText
	require.NoError(t, json.Unmarshal([]byte(`"hello"`), &plain))
	assert.True(t, plain.IsPlain)
	assert.Equal(t, "hello", plain.PlainText)

	var arr RichText
	require.NoError(t, json.Unmarshal([]byte(`["a","b"]`), &arr))
	require.Len(t, arr.Parts, 2)
	assert.Equal(t, "a", arr.Parts[0].PlainText)

	var span RichText
	require.NoError(t, json.Unmarshal([]byte(`{"type":"url","text":"site","url":"https://example.com"}`), &span))
	assert.Equal(t, RichTextTypeURL, span.Type)
	assert.Equal(t, "https://example.com", span.URL)
	require.NotNil(t, span.Text)
	assert.True(t, span.Text.IsPlain)
	assert.Equal(t, "site", span.Text.PlainText)
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
		require.NoError(t, json.Unmarshal([]byte(in), &first), "unmarshal %s", in)
		out, err := json.Marshal(first)
		require.NoError(t, err, "marshal %s", in)
		var second RichText
		require.NoError(t, json.Unmarshal(out, &second), "re-unmarshal %s", out)
		// Raw differs by construction (only set on decode), so clear it.
		first.Raw, second.Raw = nil, nil
		assert.Equal(t, first, second, "round trip mismatch for %s", in)
	}
}

// TestRichBlockCaptionRouting verifies that the shared "caption" wire field is
// routed to TableCaption for a table block and to Caption for a media block.
func TestRichBlockCaptionRouting(t *testing.T) {
	var table RichBlock
	require.NoError(t, json.Unmarshal([]byte(`{"type":"table","cells":[],"caption":"a table"}`), &table))
	require.NotNil(t, table.TableCaption)
	assert.Equal(t, "a table", table.TableCaption.PlainText)
	assert.Nil(t, table.Caption)

	var photo RichBlock
	require.NoError(t, json.Unmarshal([]byte(`{"type":"photo","photo":[],"caption":{"text":"pic"}}`), &photo))
	require.NotNil(t, photo.Caption)
	assert.True(t, photo.Caption.Text.IsPlain)
	assert.Equal(t, "pic", photo.Caption.Text.PlainText)
	assert.Nil(t, photo.TableCaption)

	// The table caption must marshal back out under the shared "caption" key.
	out, err := json.Marshal(table)
	require.NoError(t, err)
	var check map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(out, &check))
	assert.Equal(t, `"a table"`, string(check["caption"]))
}

// TestRichMessageDecode verifies a full RichMessage with nested blocks decodes.
func TestRichMessageDecode(t *testing.T) {
	raw := `{"blocks":[{"type":"heading","text":"Title","size":1},{"type":"paragraph","text":["plain ",{"type":"bold","text":"bold"}]}],"is_rtl":false}`
	var rm RichMessage
	require.NoError(t, json.Unmarshal([]byte(raw), &rm))
	require.Len(t, rm.Blocks, 2)
	assert.Equal(t, RichBlockTypeHeading, rm.Blocks[0].Type)
	assert.Equal(t, 1, rm.Blocks[0].Size)
	para := rm.Blocks[1]
	assert.Equal(t, RichBlockTypeParagraph, para.Type)
	require.NotNil(t, para.Text)
	require.Len(t, para.Text.Parts, 2)
	assert.Equal(t, RichTextTypeBold, para.Text.Parts[1].Type)
}

// TestInputRichMessageParams verifies the send-side config serializes the
// rich_message parameter as JSON.
func TestInputRichMessageParams(t *testing.T) {
	cfg := NewRichMessage(123, InputRichMessage{HTML: "<b>hi</b>"})
	params, err := cfg.params()
	require.NoError(t, err)
	assert.Equal(t, "123", params["chat_id"])
	assert.Equal(t, "sendRichMessage", cfg.method())
	var got InputRichMessage
	require.NoError(t, json.Unmarshal([]byte(params["rich_message"]), &got))
	assert.Equal(t, "<b>hi</b>", got.HTML)
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
	require.NoError(t, err)
	assert.Equal(t, `{"type":"buttons","buttons":[{"text":"Press","style":"primary","callback_data":"cb"}],"align":"center"}`, string(out))

	raw := `{"type":"document","document":{"file_id":"f1","file_unique_id":"u1"},` +
		`"caption":{"text":"a file"}}`
	var doc RichBlock
	require.NoError(t, json.Unmarshal([]byte(raw), &doc))
	require.NotNil(t, doc.Document)
	assert.Equal(t, "f1", doc.Document.FileID)
	require.NotNil(t, doc.Caption)
	assert.True(t, doc.Caption.Text.IsPlain)
	assert.Equal(t, "a file", doc.Caption.Text.PlainText)

	quote := InputRichBlock{
		Type: RichBlockTypeExpandableBlockquote,
		Text: &RichText{IsPlain: true, PlainText: "long quote"},
	}
	out, err = json.Marshal(quote)
	require.NoError(t, err)
	assert.Equal(t, `{"type":"expandable_blockquote","text":"long quote"}`, string(out))

	table := InputRichBlock{Type: RichBlockTypeTable, IsCompact: true}
	out, err = json.Marshal(table)
	require.NoError(t, err)
	assert.Equal(t, `{"type":"table","is_compact":true}`, string(out))

	span := RichText{Type: RichTextTypeButton, Button: &RichMessageButton{
		Text: RichText{IsPlain: true, PlainText: "Go"},
		URL:  "https://example.com",
	}}
	out, err = json.Marshal(span)
	require.NoError(t, err)
	assert.Equal(t, `{"type":"button","button":{"text":"Go","url":"https://example.com"}}`, string(out))
}
