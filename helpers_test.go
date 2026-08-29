package tgbotapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewWebhook(t *testing.T) {
	result, err := NewWebhook("https://example.com/token")

	require.NoError(t, err)
	assert.Equal(t, "https://example.com/token", result.URL.String())
	assert.Nil(t, result.Certificate)
	assert.Equal(t, 0, result.MaxConnections)
	assert.Empty(t, result.AllowedUpdates)
}

func TestNewWebhookWithCert(t *testing.T) {
	exampleFile := FileID("123")
	result, err := NewWebhookWithCert("https://example.com/token", exampleFile)

	require.NoError(t, err)
	assert.Equal(t, "https://example.com/token", result.URL.String())
	assert.Equal(t, RequestFileData(exampleFile), result.Certificate)
	assert.Equal(t, 0, result.MaxConnections)
	assert.Empty(t, result.AllowedUpdates)
}

func TestNewInlineQueryResultArticle(t *testing.T) {
	result := NewInlineQueryResultArticle("id", "title", "message")

	assert.Equal(t, "article", result.Type)
	assert.Equal(t, "id", result.ID)
	assert.Equal(t, "title", result.Title)
	assert.Equal(t, "message", result.InputMessageContent.(InputTextMessageContent).Text)
}

func TestNewInlineQueryResultArticleMarkdown(t *testing.T) {
	result := NewInlineQueryResultArticleMarkdown("id", "title", "*message*")

	assert.Equal(t, "article", result.Type)
	assert.Equal(t, "id", result.ID)
	assert.Equal(t, "title", result.Title)
	assert.Equal(t, "*message*", result.InputMessageContent.(InputTextMessageContent).Text)
	assert.Equal(t, "Markdown", result.InputMessageContent.(InputTextMessageContent).ParseMode)
}

func TestNewInlineQueryResultArticleHTML(t *testing.T) {
	result := NewInlineQueryResultArticleHTML("id", "title", "<b>message</b>")

	assert.Equal(t, "article", result.Type)
	assert.Equal(t, "id", result.ID)
	assert.Equal(t, "title", result.Title)
	assert.Equal(t, "<b>message</b>", result.InputMessageContent.(InputTextMessageContent).Text)
	assert.Equal(t, "HTML", result.InputMessageContent.(InputTextMessageContent).ParseMode)
}

func TestNewInlineQueryResultGIF(t *testing.T) {
	result := NewInlineQueryResultGIF("id", "google.com")

	assert.Equal(t, "gif", result.Type)
	assert.Equal(t, "id", result.ID)
	assert.Equal(t, "google.com", result.URL)
}

func TestNewInlineQueryResultMPEG4GIF(t *testing.T) {
	result := NewInlineQueryResultMPEG4GIF("id", "google.com")

	assert.Equal(t, "mpeg4_gif", result.Type)
	assert.Equal(t, "id", result.ID)
	assert.Equal(t, "google.com", result.URL)
}

func TestNewInlineQueryResultPhoto(t *testing.T) {
	result := NewInlineQueryResultPhoto("id", "google.com")

	assert.Equal(t, "photo", result.Type)
	assert.Equal(t, "id", result.ID)
	assert.Equal(t, "google.com", result.URL)
}

func TestNewInlineQueryResultPhotoWithThumbnail(t *testing.T) {
	result := NewInlineQueryResultPhotoWithThumbnail("id", "google.com", "thumbnail.com")

	assert.Equal(t, "photo", result.Type)
	assert.Equal(t, "id", result.ID)
	assert.Equal(t, "google.com", result.URL)
	assert.Equal(t, "thumbnail.com", result.ThumbnailURL)
}

func TestNewInlineQueryResultVideo(t *testing.T) {
	result := NewInlineQueryResultVideo("id", "google.com")

	assert.Equal(t, "video", result.Type)
	assert.Equal(t, "id", result.ID)
	assert.Equal(t, "google.com", result.URL)
}

func TestNewInlineQueryResultAudio(t *testing.T) {
	result := NewInlineQueryResultAudio("id", "google.com", "title")

	assert.Equal(t, "audio", result.Type)
	assert.Equal(t, "id", result.ID)
	assert.Equal(t, "google.com", result.URL)
	assert.Equal(t, "title", result.Title)
}

func TestNewInlineQueryResultVoice(t *testing.T) {
	result := NewInlineQueryResultVoice("id", "google.com", "title")

	assert.Equal(t, "voice", result.Type)
	assert.Equal(t, "id", result.ID)
	assert.Equal(t, "google.com", result.URL)
	assert.Equal(t, "title", result.Title)
}

func TestNewInlineQueryResultDocument(t *testing.T) {
	result := NewInlineQueryResultDocument("id", "google.com", "title", "mime/type")

	assert.Equal(t, "document", result.Type)
	assert.Equal(t, "id", result.ID)
	assert.Equal(t, "google.com", result.URL)
	assert.Equal(t, "title", result.Title)
	assert.Equal(t, "mime/type", result.MimeType)
}

func TestNewInlineQueryResultLocation(t *testing.T) {
	result := NewInlineQueryResultLocation("id", "name", 40, 50)

	assert.Equal(t, "location", result.Type)
	assert.Equal(t, "id", result.ID)
	assert.Equal(t, "name", result.Title)
	assert.Equal(t, float64(40), result.Latitude)
	assert.Equal(t, float64(50), result.Longitude)
}

func TestNewInlineKeyboardButtonLoginURL(t *testing.T) {
	result := NewInlineKeyboardButtonLoginURL("text", LoginURL{
		URL:                "url",
		ForwardText:        "ForwardText",
		BotUsername:        "username",
		RequestWriteAccess: false,
	})

	assert.Equal(t, "text", result.Text)
	assert.Equal(t, "url", result.LoginURL.URL)
	assert.Equal(t, "ForwardText", result.LoginURL.ForwardText)
	assert.Equal(t, "username", result.LoginURL.BotUsername)
	assert.False(t, result.LoginURL.RequestWriteAccess)
}

func TestNewEditMessageText(t *testing.T) {
	edit := NewEditMessageText(ChatID, ReplyToMessageID, "new text")

	assert.Equal(t, "new text", edit.Text)
	assert.Equal(t, int64(ChatID), edit.BaseEdit.ChatID)
	assert.Equal(t, ReplyToMessageID, edit.BaseEdit.MessageID)
}

func TestNewEditMessageCaption(t *testing.T) {
	edit := NewEditMessageCaption(ChatID, ReplyToMessageID, "new caption")

	assert.Equal(t, "new caption", edit.Caption)
	assert.Equal(t, int64(ChatID), edit.BaseEdit.ChatID)
	assert.Equal(t, ReplyToMessageID, edit.BaseEdit.MessageID)
}

func TestNewEditMessageReplyMarkup(t *testing.T) {
	markup := InlineKeyboardMarkup{
		InlineKeyboard: [][]InlineKeyboardButton{
			{
				{Text: "test"},
			},
		},
	}

	edit := NewEditMessageReplyMarkup(ChatID, ReplyToMessageID, markup)

	assert.Equal(t, "test", edit.ReplyMarkup.InlineKeyboard[0][0].Text)
	assert.Equal(t, int64(ChatID), edit.BaseEdit.ChatID)
	assert.Equal(t, ReplyToMessageID, edit.BaseEdit.MessageID)
}

func TestNewDice(t *testing.T) {
	dice := NewDice(42)

	assert.Equal(t, int64(42), dice.ChatID)
	assert.Equal(t, "", dice.Emoji)
}

func TestNewDiceWithEmoji(t *testing.T) {
	dice := NewDiceWithEmoji(42, "🏀")

	assert.Equal(t, int64(42), dice.ChatID)
	assert.Equal(t, "🏀", dice.Emoji)
}

func TestValidateWebAppData(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		token := "5473903189:AAFnHnISQMP5UQQ5MEaoEWvxeiwNgz2CN2U"
		initData := "query_id=AAG1bpMJAAAAALVukwmZ_H2t&user=%7B%22id%22%3A160657077%2C%22first_name%22%3A%22Yury%20R%22%2C%22last_name%22%3A%22%22%2C%22username%22%3A%22crashiura%22%2C%22language_code%22%3A%22en%22%7D&auth_date=1656804462&hash=8d6960760a573d3212deb05e20d1a34959c83d24c1bc44bb26dde49a42aa9b34"
		result, err := ValidateWebAppData(token, initData)
		require.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("error", func(t *testing.T) {
		token := "5473903189:AAFnHnISQMP5UQQ5MEaoEWvxeiwNgz2CN2U"
		initData := "asdfasdfasdfasdfasdf"
		result, err := ValidateWebAppData(token, initData)
		assert.Error(t, err)
		assert.False(t, result)
	})
}
