package tgbotapi

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRichTextBuilders verifies that each span builder encodes to the correct
// wire form.
func TestRichTextBuilders(t *testing.T) {
	cases := []struct {
		name string
		in   interface{}
		want string
	}{
		{"plain", PlainText("hi"), `"hi"`},
		{"parts", TextParts(PlainText("a"), Bold("b")), `["a",{"type":"bold","text":"b"}]`},
		{"bold", Bold("x"), `{"type":"bold","text":"x"}`},
		{"italic", Italic("x"), `{"type":"italic","text":"x"}`},
		{"underline", Underline("x"), `{"type":"underline","text":"x"}`},
		{"strikethrough", Strikethrough("x"), `{"type":"strikethrough","text":"x"}`},
		{"spoiler", Spoiler("x"), `{"type":"spoiler","text":"x"}`},
		{"code", Code("x"), `{"type":"code","text":"x"}`},
		{"link", LinkText("site", "https://example.com"), `{"type":"url","text":"site","url":"https://example.com"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := json.Marshal(tc.in)
			require.NoError(t, err)
			assert.Equal(t, tc.want, string(out))
		})
	}
}

// TestRichBlockBuilders verifies that each block builder encodes to the
// correct wire form.
func TestRichBlockBuilders(t *testing.T) {
	cases := []struct {
		name string
		in   InputRichBlock
		want string
	}{
		{"paragraph single", Paragraph(PlainText("hi")), `{"type":"paragraph","text":"hi"}`},
		{"paragraph parts", Paragraph(PlainText("a"), Bold("b")), `{"type":"paragraph","text":["a",{"type":"bold","text":"b"}]}`},
		{"heading", Heading(2, "Title"), `{"type":"heading","text":"Title","size":2}`},
		{"pre", Pre("x := 1", "go"), `{"type":"pre","text":"x := 1","language":"go"}`},
		{"divider", Divider(), `{"type":"divider"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := json.Marshal(tc.in)
			require.NoError(t, err)
			assert.Equal(t, tc.want, string(out))
		})
	}
}

// TestTableRowBuilders verifies header and body rows, including the default
// alignment.
func TestTableRowBuilders(t *testing.T) {
	table := InputRichBlock{
		Type: RichBlockTypeTable,
		Cells: [][]RichBlockTableCell{
			TableHeaderRow(PlainText("Name")),
			TableRow(PlainText("Go")),
		},
	}
	out, err := json.Marshal(table)
	require.NoError(t, err)
	want := `{"type":"table","cells":[` +
		`[{"text":"Name","is_header":true,"align":"left","valign":"top"}],` +
		`[{"text":"Go","align":"left","valign":"top"}]]}`
	assert.Equal(t, want, string(out))
}
