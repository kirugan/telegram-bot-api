package tgbotapi

// This file holds short builders for rich message content. They remove the
// struct boilerplate from RichText spans and InputRichBlock blocks:
//
//	tgbotapi.NewRichMessage(chatID, tgbotapi.InputRichMessage{
//		Blocks: []tgbotapi.InputRichBlock{
//			tgbotapi.Heading(1, "Title"),
//			tgbotapi.Paragraph(tgbotapi.PlainText("Hello, "), tgbotapi.Bold("world")),
//		},
//	})

// PlainText returns a RichText that holds bare text without style.
func PlainText(text string) RichText {
	return RichText{IsPlain: true, PlainText: text}
}

// TextParts returns a RichText that concatenates the given parts.
func TextParts(parts ...RichText) RichText {
	return RichText{Parts: parts}
}

func styledText(styleType, text string) RichText {
	inner := PlainText(text)
	return RichText{Type: styleType, Text: &inner}
}

// Bold returns a RichText span with bold style.
func Bold(text string) RichText {
	return styledText(RichTextTypeBold, text)
}

// Italic returns a RichText span with italic style.
func Italic(text string) RichText {
	return styledText(RichTextTypeItalic, text)
}

// Underline returns a RichText span with underline style.
func Underline(text string) RichText {
	return styledText(RichTextTypeUnderline, text)
}

// Strikethrough returns a RichText span with strikethrough style.
func Strikethrough(text string) RichText {
	return styledText(RichTextTypeStrikethrough, text)
}

// Spoiler returns a RichText span that is hidden until tapped.
func Spoiler(text string) RichText {
	return styledText(RichTextTypeSpoiler, text)
}

// Code returns a RichText span with monospace style.
func Code(text string) RichText {
	return styledText(RichTextTypeCode, text)
}

// LinkText returns a RichText span that opens the given URL.
func LinkText(text, url string) RichText {
	span := styledText(RichTextTypeURL, text)
	span.URL = url
	return span
}

// Paragraph returns a "paragraph" input block. Pass one part or several;
// several parts are concatenated.
func Paragraph(parts ...RichText) InputRichBlock {
	var text RichText
	if len(parts) == 1 {
		text = parts[0]
	} else {
		text = TextParts(parts...)
	}
	return InputRichBlock{Type: RichBlockTypeParagraph, Text: &text}
}

// Heading returns a "heading" input block. Size is 1-6; 1 is largest.
func Heading(size int, text string) InputRichBlock {
	t := PlainText(text)
	return InputRichBlock{Type: RichBlockTypeHeading, Size: size, Text: &t}
}

// Pre returns a "pre" input block with source code. Language may be empty.
func Pre(code, language string) InputRichBlock {
	t := PlainText(code)
	return InputRichBlock{Type: RichBlockTypePre, Text: &t, Language: language}
}

// Divider returns a "divider" input block.
func Divider() InputRichBlock {
	return InputRichBlock{Type: RichBlockTypeDivider}
}

func tableRow(isHeader bool, cells []RichText) []RichBlockTableCell {
	row := make([]RichBlockTableCell, len(cells))
	for i := range cells {
		row[i] = RichBlockTableCell{
			Text:     &cells[i],
			IsHeader: isHeader,
			Align:    "left",
			Valign:   "top",
		}
	}
	return row
}

// TableRow returns a table row of body cells, aligned left and top.
func TableRow(cells ...RichText) []RichBlockTableCell {
	return tableRow(false, cells)
}

// TableHeaderRow returns a table row of header cells, aligned left and top.
func TableHeaderRow(cells ...RichText) []RichBlockTableCell {
	return tableRow(true, cells)
}
