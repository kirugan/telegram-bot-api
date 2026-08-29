package tgbotapi

import (
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	TestToken               = "153667468:AAHlSHlMqSt1f_uFmVRJbm5gntu2HI4WW8I"
	ChatID                  = 76918703
	Channel                 = "@tgbotapitest"
	SupergroupChatID        = -1001120141283
	ReplyToMessageID        = 35
	ExistingPhotoFileID     = "AgACAgIAAxkDAAEBFUZhIALQ9pZN4BUe8ZSzUU_2foSo1AACnrMxG0BucEhezsBWOgcikQEAAwIAA20AAyAE"
	ExistingDocumentFileID  = "BQADAgADOQADjMcoCcioX1GrDvp3Ag"
	ExistingAudioFileID     = "BQADAgADRgADjMcoCdXg3lSIN49lAg"
	ExistingVoiceFileID     = "AwADAgADWQADjMcoCeul6r_q52IyAg"
	ExistingVideoFileID     = "BAADAgADZgADjMcoCav432kYe0FRAg"
	ExistingVideoNoteFileID = "DQADAgADdQAD70cQSUK41dLsRMqfAg"
	ExistingStickerFileID   = "BQADAgADcwADjMcoCbdl-6eB--YPAg"
)

type testLogger struct {
	t *testing.T
}

func (t testLogger) Println(v ...interface{}) {
	t.t.Log(v...)
}

func (t testLogger) Printf(format string, v ...interface{}) {
	t.t.Logf(format, v...)
}

func getBot(t *testing.T) *BotAPI {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		token = TestToken
	}

	bot, err := NewBotAPI(token)
	if err != nil {
		t.Skipf("skipping: NewBotAPI failed (set TELEGRAM_BOT_TOKEN to a real token to run): %v", err)
	}

	bot.Debug = true
	SetLogger(testLogger{t})

	return bot
}

func TestNewBotAPI_notoken(t *testing.T) {
	_, err := NewBotAPI("")

	assert.Error(t, err)
}

func TestGetUpdates(t *testing.T) {
	bot := getBot(t)

	u := NewUpdate(0)

	_, err := bot.GetUpdates(u)

	require.NoError(t, err)
}

func TestSendWithMessage(t *testing.T) {
	bot := getBot(t)

	msg := NewMessage(ChatID, "A test message from the test library in telegram-bot-api")
	msg.ParseMode = ModeMarkdown
	_, err := bot.Send(msg)

	require.NoError(t, err)
}

func TestSendWithMessageReply(t *testing.T) {
	bot := getBot(t)

	msg := NewMessage(ChatID, "A test message from the test library in telegram-bot-api")
	msg.ReplyParameters = &ReplyParameters{MessageID: ReplyToMessageID}
	_, err := bot.Send(msg)

	require.NoError(t, err)
}

func TestSendWithMessageForward(t *testing.T) {
	bot := getBot(t)

	msg := NewForward(ChatID, ChatID, ReplyToMessageID)
	_, err := bot.Send(msg)

	require.NoError(t, err)
}

func TestCopyMessage(t *testing.T) {
	bot := getBot(t)

	msg := NewMessage(ChatID, "A test message from the test library in telegram-bot-api")
	message, err := bot.Send(msg)
	require.NoError(t, err)

	copyMessageConfig := NewCopyMessage(SupergroupChatID, message.Chat.ID, message.MessageID)
	messageID, err := bot.CopyMessage(copyMessageConfig)
	require.NoError(t, err)

	assert.NotEqual(t, message.MessageID, messageID.MessageID, "copied message ID was the same as original message")
}

func TestSendWithNewPhoto(t *testing.T) {
	bot := getBot(t)

	msg := NewPhoto(ChatID, FilePath("tests/image.jpg"))
	msg.Caption = "Test"
	_, err := bot.Send(msg)

	require.NoError(t, err)
}

func TestSendWithNewPhotoWithFileBytes(t *testing.T) {
	bot := getBot(t)

	data, _ := os.ReadFile("tests/image.jpg")
	b := FileBytes{Name: "image.jpg", Bytes: data}

	msg := NewPhoto(ChatID, b)
	msg.Caption = "Test"
	_, err := bot.Send(msg)

	require.NoError(t, err)
}

func TestSendWithNewPhotoWithFileReader(t *testing.T) {
	bot := getBot(t)

	f, _ := os.Open("tests/image.jpg")
	reader := FileReader{Name: "image.jpg", Reader: f}

	msg := NewPhoto(ChatID, reader)
	msg.Caption = "Test"
	_, err := bot.Send(msg)

	require.NoError(t, err)
}

func TestSendWithNewPhotoReply(t *testing.T) {
	bot := getBot(t)

	msg := NewPhoto(ChatID, FilePath("tests/image.jpg"))
	msg.ReplyParameters = &ReplyParameters{MessageID: ReplyToMessageID}

	_, err := bot.Send(msg)

	require.NoError(t, err)
}

func TestSendNewPhotoToChannel(t *testing.T) {
	bot := getBot(t)

	msg := NewPhotoToChannel(Channel, FilePath("tests/image.jpg"))
	msg.Caption = "Test"
	_, err := bot.Send(msg)

	require.NoError(t, err)
}

func TestSendNewPhotoToChannelFileBytes(t *testing.T) {
	bot := getBot(t)

	data, _ := os.ReadFile("tests/image.jpg")
	b := FileBytes{Name: "image.jpg", Bytes: data}

	msg := NewPhotoToChannel(Channel, b)
	msg.Caption = "Test"
	_, err := bot.Send(msg)

	require.NoError(t, err)
}

func TestSendNewPhotoToChannelFileReader(t *testing.T) {
	bot := getBot(t)

	f, _ := os.Open("tests/image.jpg")
	reader := FileReader{Name: "image.jpg", Reader: f}

	msg := NewPhotoToChannel(Channel, reader)
	msg.Caption = "Test"
	_, err := bot.Send(msg)

	require.NoError(t, err)
}

func TestSendWithExistingPhoto(t *testing.T) {
	bot := getBot(t)

	msg := NewPhoto(ChatID, FileID(ExistingPhotoFileID))
	msg.Caption = "Test"
	_, err := bot.Send(msg)

	require.NoError(t, err)
}

func TestSendWithNewDocument(t *testing.T) {
	bot := getBot(t)

	msg := NewDocument(ChatID, FilePath("tests/image.jpg"))
	_, err := bot.Send(msg)

	require.NoError(t, err)
}

func TestSendWithNewDocumentAndThumbnail(t *testing.T) {
	bot := getBot(t)

	msg := NewDocument(ChatID, FilePath("tests/voice.ogg"))
	msg.Thumbnail = FilePath("tests/image.jpg")
	_, err := bot.Send(msg)

	require.NoError(t, err)
}

func TestSendWithExistingDocument(t *testing.T) {
	bot := getBot(t)

	msg := NewDocument(ChatID, FileID(ExistingDocumentFileID))
	_, err := bot.Send(msg)

	require.NoError(t, err)
}

func TestSendWithNewAudio(t *testing.T) {
	bot := getBot(t)

	msg := NewAudio(ChatID, FilePath("tests/audio.mp3"))
	msg.Title = "TEST"
	msg.Duration = 10
	msg.Performer = "TEST"
	_, err := bot.Send(msg)

	require.NoError(t, err)
}

func TestSendWithExistingAudio(t *testing.T) {
	bot := getBot(t)

	msg := NewAudio(ChatID, FileID(ExistingAudioFileID))
	msg.Title = "TEST"
	msg.Duration = 10
	msg.Performer = "TEST"

	_, err := bot.Send(msg)

	require.NoError(t, err)
}

func TestSendWithNewVoice(t *testing.T) {
	bot := getBot(t)

	msg := NewVoice(ChatID, FilePath("tests/voice.ogg"))
	msg.Duration = 10
	_, err := bot.Send(msg)

	require.NoError(t, err)
}

func TestSendWithExistingVoice(t *testing.T) {
	bot := getBot(t)

	msg := NewVoice(ChatID, FileID(ExistingVoiceFileID))
	msg.Duration = 10
	_, err := bot.Send(msg)

	require.NoError(t, err)
}

func TestSendWithContact(t *testing.T) {
	bot := getBot(t)

	contact := NewContact(ChatID, "5551234567", "Test")

	_, err := bot.Send(contact)
	require.NoError(t, err)
}

func TestSendWithLocation(t *testing.T) {
	bot := getBot(t)

	_, err := bot.Send(NewLocation(ChatID, 40, 40))

	require.NoError(t, err)
}

func TestSendWithVenue(t *testing.T) {
	bot := getBot(t)

	venue := NewVenue(ChatID, "A Test Location", "123 Test Street", 40, 40)

	_, err := bot.Send(venue)
	require.NoError(t, err)
}

func TestSendWithNewVideo(t *testing.T) {
	bot := getBot(t)

	msg := NewVideo(ChatID, FilePath("tests/video.mp4"))
	msg.Duration = 10
	msg.Caption = "TEST"

	_, err := bot.Send(msg)

	require.NoError(t, err)
}

func TestSendWithExistingVideo(t *testing.T) {
	bot := getBot(t)

	msg := NewVideo(ChatID, FileID(ExistingVideoFileID))
	msg.Duration = 10
	msg.Caption = "TEST"

	_, err := bot.Send(msg)

	require.NoError(t, err)
}

func TestSendWithNewVideoNote(t *testing.T) {
	bot := getBot(t)

	msg := NewVideoNote(ChatID, 240, FilePath("tests/videonote.mp4"))
	msg.Duration = 10

	_, err := bot.Send(msg)

	require.NoError(t, err)
}

func TestSendWithExistingVideoNote(t *testing.T) {
	bot := getBot(t)

	msg := NewVideoNote(ChatID, 240, FileID(ExistingVideoNoteFileID))
	msg.Duration = 10

	_, err := bot.Send(msg)

	require.NoError(t, err)
}

func TestSendWithNewSticker(t *testing.T) {
	bot := getBot(t)

	msg := NewSticker(ChatID, FilePath("tests/image.jpg"))

	_, err := bot.Send(msg)

	require.NoError(t, err)
}

func TestSendWithExistingSticker(t *testing.T) {
	bot := getBot(t)

	msg := NewSticker(ChatID, FileID(ExistingStickerFileID))

	_, err := bot.Send(msg)

	require.NoError(t, err)
}

func TestSendWithNewStickerAndKeyboardHide(t *testing.T) {
	bot := getBot(t)

	msg := NewSticker(ChatID, FilePath("tests/image.jpg"))
	msg.ReplyMarkup = ReplyKeyboardRemove{
		RemoveKeyboard: true,
		Selective:      false,
	}
	_, err := bot.Send(msg)

	require.NoError(t, err)
}

func TestSendWithExistingStickerAndKeyboardHide(t *testing.T) {
	bot := getBot(t)

	msg := NewSticker(ChatID, FileID(ExistingStickerFileID))
	msg.ReplyMarkup = ReplyKeyboardRemove{
		RemoveKeyboard: true,
		Selective:      false,
	}

	_, err := bot.Send(msg)

	require.NoError(t, err)
}

func TestSendWithDice(t *testing.T) {
	bot := getBot(t)

	msg := NewDice(ChatID)
	_, err := bot.Send(msg)

	require.NoError(t, err)
}

func TestSendWithDiceWithEmoji(t *testing.T) {
	bot := getBot(t)

	msg := NewDiceWithEmoji(ChatID, "🏀")
	_, err := bot.Send(msg)

	require.NoError(t, err)
}

func TestGetFile(t *testing.T) {
	bot := getBot(t)

	file := FileConfig{
		FileID: ExistingPhotoFileID,
	}

	_, err := bot.GetFile(file)

	require.NoError(t, err)
}

func TestSendChatConfig(t *testing.T) {
	bot := getBot(t)

	_, err := bot.Request(NewChatAction(ChatID, ChatTyping))

	require.NoError(t, err)
}

// TODO: identify why this isn't working
// func TestSendEditMessage(t *testing.T) {
// 	bot := getBot(t)

// 	msg, err := bot.Send(NewMessage(ChatID, "Testing editing."))
// 	require.NoError(t, err)

// 	edit := EditMessageTextConfig{
// 		BaseEdit: BaseEdit{
// 			ChatID:    ChatID,
// 			MessageID: msg.MessageID,
// 		},
// 		Text: "Updated text.",
// 	}

// 	_, err = bot.Send(edit)
// 	require.NoError(t, err)
// }

func TestGetUserProfilePhotos(t *testing.T) {
	bot := getBot(t)

	_, err := bot.GetUserProfilePhotos(NewUserProfilePhotos(ChatID))
	require.NoError(t, err)
}

func TestSetWebhookWithCert(t *testing.T) {
	bot := getBot(t)

	time.Sleep(time.Second * 2)

	bot.Request(DeleteWebhookConfig{})

	wh, err := NewWebhookWithCert("https://example.com/tgbotapi-test/"+bot.Token, FilePath("tests/cert.pem"))

	require.NoError(t, err)

	_, err = bot.Request(wh)

	require.NoError(t, err)

	_, err = bot.GetWebhookInfo()

	require.NoError(t, err)

	bot.Request(DeleteWebhookConfig{})
}

func TestSetWebhookWithoutCert(t *testing.T) {
	bot := getBot(t)

	time.Sleep(time.Second * 2)

	bot.Request(DeleteWebhookConfig{})

	wh, err := NewWebhook("https://example.com/tgbotapi-test/" + bot.Token)

	require.NoError(t, err)

	_, err = bot.Request(wh)

	require.NoError(t, err)

	info, err := bot.GetWebhookInfo()

	require.NoError(t, err)
	assert.NotZero(t, info.MaxConnections, "Expected maximum connections to be greater than 0")
	assert.Zero(t, info.LastErrorDate, "failed to set webhook: %s", info.LastErrorMessage)

	bot.Request(DeleteWebhookConfig{})
}

func TestSendWithMediaGroupPhotoVideo(t *testing.T) {
	bot := getBot(t)

	cfg := NewMediaGroup(ChatID, []interface{}{
		NewInputMediaPhoto(FileURL("https://github.com/go-telegram-bot-api/telegram-bot-api/raw/0a3a1c8716c4cd8d26a262af9f12dcbab7f3f28c/tests/image.jpg")),
		NewInputMediaPhoto(FilePath("tests/image.jpg")),
		NewInputMediaVideo(FilePath("tests/video.mp4")),
	})

	messages, err := bot.SendMediaGroup(cfg)
	require.NoError(t, err)
	require.NotNil(t, messages, "No received messages")
	assert.Len(t, messages, len(cfg.Media))
}

func TestSendWithMediaGroupDocument(t *testing.T) {
	bot := getBot(t)

	cfg := NewMediaGroup(ChatID, []interface{}{
		NewInputMediaDocument(FileURL("https://i.imgur.com/unQLJIb.jpg")),
		NewInputMediaDocument(FilePath("tests/image.jpg")),
	})

	messages, err := bot.SendMediaGroup(cfg)
	require.NoError(t, err)
	require.NotNil(t, messages, "No received messages")
	assert.Len(t, messages, len(cfg.Media))
}

func TestSendWithMediaGroupAudio(t *testing.T) {
	bot := getBot(t)

	cfg := NewMediaGroup(ChatID, []interface{}{
		NewInputMediaAudio(FilePath("tests/audio.mp3")),
		NewInputMediaAudio(FilePath("tests/audio.mp3")),
	})

	messages, err := bot.SendMediaGroup(cfg)
	require.NoError(t, err)
	require.NotNil(t, messages, "No received messages")
	assert.Len(t, messages, len(cfg.Media))
}

func ExampleNewBotAPI() {
	bot, err := NewBotAPI("MyAwesomeBotToken")
	if err != nil {
		panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	// Optional: wait for updates and clear them if you don't want to handle
	// a large backlog of old messages
	time.Sleep(time.Millisecond * 500)
	updates.Clear()

	for update := range updates {
		if update.Message == nil {
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		msg := NewMessage(update.Message.Chat.ID, update.Message.Text)
		msg.ReplyParameters = &ReplyParameters{MessageID: update.Message.MessageID}

		bot.Send(msg)
	}
}

func ExampleNewWebhook() {
	bot, err := NewBotAPI("MyAwesomeBotToken")
	if err != nil {
		panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	wh, err := NewWebhookWithCert("https://www.google.com:8443/"+bot.Token, FilePath("cert.pem"))

	if err != nil {
		panic(err)
	}

	_, err = bot.Request(wh)

	if err != nil {
		panic(err)
	}

	info, err := bot.GetWebhookInfo()

	if err != nil {
		panic(err)
	}

	if info.LastErrorDate != 0 {
		log.Printf("failed to set webhook: %s", info.LastErrorMessage)
	}

	updates := bot.ListenForWebhook("/" + bot.Token)
	go http.ListenAndServeTLS("0.0.0.0:8443", "cert.pem", "key.pem", nil)

	for update := range updates {
		log.Printf("%+v\n", update)
	}
}

func ExampleNewWebhookWithCert() {
	bot, err := NewBotAPI("MyAwesomeBotToken")
	if err != nil {
		panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	wh, err := NewWebhookWithCert("https://www.google.com:8443/"+bot.Token, FilePath("cert.pem"))

	if err != nil {
		panic(err)
	}

	_, err = bot.Request(wh)
	if err != nil {
		panic(err)
	}
	info, err := bot.GetWebhookInfo()
	if err != nil {
		panic(err)
	}
	if info.LastErrorDate != 0 {
		log.Printf("[Telegram callback failed]%s", info.LastErrorMessage)
	}

	http.HandleFunc("/"+bot.Token, func(w http.ResponseWriter, r *http.Request) {
		update, err := bot.HandleUpdate(r)
		if err != nil {
			log.Printf("%+v\n", err.Error())
		} else {
			log.Printf("%+v\n", *update)
		}
	})

	go http.ListenAndServeTLS("0.0.0.0:8443", "cert.pem", "key.pem", nil)
}

func ExampleInlineConfig() {
	bot, err := NewBotAPI("MyAwesomeBotToken") // create new bot
	if err != nil {
		panic(err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.InlineQuery == nil { // if no inline query, ignore it
			continue
		}

		article := NewInlineQueryResultArticle(update.InlineQuery.ID, "Echo", update.InlineQuery.Query)
		article.Description = update.InlineQuery.Query

		inlineConf := InlineConfig{
			InlineQueryID: update.InlineQuery.ID,
			IsPersonal:    true,
			CacheTime:     0,
			Results:       []interface{}{article},
		}

		if _, err := bot.Request(inlineConf); err != nil {
			log.Println(err)
		}
	}
}

func TestDeleteMessage(t *testing.T) {
	bot := getBot(t)

	msg := NewMessage(ChatID, "A test message from the test library in telegram-bot-api")
	msg.ParseMode = ModeMarkdown
	message, _ := bot.Send(msg)

	deleteMessageConfig := DeleteMessageConfig{
		ChatID:    message.Chat.ID,
		MessageID: message.MessageID,
	}
	_, err := bot.Request(deleteMessageConfig)

	require.NoError(t, err)
}

func TestPinChatMessage(t *testing.T) {
	bot := getBot(t)

	msg := NewMessage(SupergroupChatID, "A test message from the test library in telegram-bot-api")
	msg.ParseMode = ModeMarkdown
	message, _ := bot.Send(msg)

	pinChatMessageConfig := PinChatMessageConfig{
		ChatID:              message.Chat.ID,
		MessageID:           message.MessageID,
		DisableNotification: false,
	}
	_, err := bot.Request(pinChatMessageConfig)

	require.NoError(t, err)
}

func TestUnpinChatMessage(t *testing.T) {
	bot := getBot(t)

	msg := NewMessage(SupergroupChatID, "A test message from the test library in telegram-bot-api")
	msg.ParseMode = ModeMarkdown
	message, _ := bot.Send(msg)

	// We need pin message to unpin something
	pinChatMessageConfig := PinChatMessageConfig{
		ChatID:              message.Chat.ID,
		MessageID:           message.MessageID,
		DisableNotification: false,
	}

	_, err := bot.Request(pinChatMessageConfig)
	require.NoError(t, err)

	unpinChatMessageConfig := UnpinChatMessageConfig{
		ChatID:    message.Chat.ID,
		MessageID: message.MessageID,
	}

	_, err = bot.Request(unpinChatMessageConfig)
	require.NoError(t, err)
}

func TestUnpinAllChatMessages(t *testing.T) {
	bot := getBot(t)

	msg := NewMessage(SupergroupChatID, "A test message from the test library in telegram-bot-api")
	msg.ParseMode = ModeMarkdown
	message, _ := bot.Send(msg)

	pinChatMessageConfig := PinChatMessageConfig{
		ChatID:              message.Chat.ID,
		MessageID:           message.MessageID,
		DisableNotification: true,
	}

	_, err := bot.Request(pinChatMessageConfig)
	require.NoError(t, err)

	unpinAllChatMessagesConfig := UnpinAllChatMessagesConfig{
		ChatID: message.Chat.ID,
	}

	_, err = bot.Request(unpinAllChatMessagesConfig)
	require.NoError(t, err)
}

func TestPolls(t *testing.T) {
	bot := getBot(t)

	poll := NewPoll(SupergroupChatID, "Are polls working?", "Yes", "No")

	msg, err := bot.Send(poll)
	require.NoError(t, err)

	result, err := bot.StopPoll(NewStopPoll(SupergroupChatID, msg.MessageID))
	require.NoError(t, err)

	assert.Equal(t, "Are polls working?", result.Question, "Poll question did not match")
	assert.True(t, result.IsClosed, "Poll did not end")

	assert.Equal(t, "Yes", result.Options[0].Text, "Poll options were incorrect")
	assert.Zero(t, result.Options[0].VoterCount, "Poll options were incorrect")
	assert.Equal(t, "No", result.Options[1].Text, "Poll options were incorrect")
	assert.Zero(t, result.Options[1].VoterCount, "Poll options were incorrect")
}

func TestSendDice(t *testing.T) {
	bot := getBot(t)

	dice := NewDice(ChatID)

	msg, err := bot.Send(dice)
	require.NoError(t, err, "Unable to send dice roll")

	assert.NotNil(t, msg.Dice, "Dice roll was not received")
}

func TestCommands(t *testing.T) {
	bot := getBot(t)

	setCommands := NewSetMyCommands(BotCommand{
		Command:     "test",
		Description: "a test command",
	})

	_, err := bot.Request(setCommands)
	require.NoError(t, err, "Unable to set commands")

	commands, err := bot.GetMyCommands()
	require.NoError(t, err, "Unable to get commands")

	require.Len(t, commands, 1, "Incorrect number of commands returned")
	assert.Equal(t, "test", commands[0].Command, "Commands were incorrectly set")
	assert.Equal(t, "a test command", commands[0].Description, "Commands were incorrectly set")

	setCommands = NewSetMyCommandsWithScope(NewBotCommandScopeAllPrivateChats(), BotCommand{
		Command:     "private",
		Description: "a private command",
	})

	_, err = bot.Request(setCommands)
	require.NoError(t, err, "Unable to set commands")

	commands, err = bot.GetMyCommandsWithConfig(NewGetMyCommandsWithScope(NewBotCommandScopeAllPrivateChats()))
	require.NoError(t, err, "Unable to get commands")

	require.Len(t, commands, 1, "Incorrect number of commands returned")
	assert.Equal(t, "private", commands[0].Command, "Commands were incorrectly set")
	assert.Equal(t, "a private command", commands[0].Description, "Commands were incorrectly set")
}

// TODO: figure out why test is failing
//
// func TestEditMessageMedia(t *testing.T) {
// 	bot := getBot(t)

// 	msg := NewPhoto(ChatID, "tests/image.jpg")
// 	msg.Caption = "Test"
// 	m, err := bot.Send(msg)

// 	require.NoError(t, err)

// 	edit := EditMessageMediaConfig{
// 		BaseEdit: BaseEdit{
// 			ChatID:    ChatID,
// 			MessageID: m.MessageID,
// 		},
// 		Media: NewInputMediaVideo(FilePath("tests/video.mp4")),
// 	}

// 	_, err = bot.Request(edit)
// 	require.NoError(t, err)
// }

func TestPrepareInputMediaForParams(t *testing.T) {
	media := []interface{}{
		NewInputMediaPhoto(FilePath("tests/image.jpg")),
		NewInputMediaVideo(FileID("test")),
	}

	prepared := prepareInputMediaForParams(media)

	assert.Equal(t, RequestFileData(FilePath("tests/image.jpg")), media[0].(InputMediaPhoto).Media, "Original media was changed")
	assert.Equal(t, RequestFileData(fileAttach("attach://file-0")), prepared[0].(InputMediaPhoto).Media, "New media was not replaced")
	assert.Equal(t, RequestFileData(FileID("test")), prepared[1].(InputMediaVideo).Media, "Passthrough value was not the same")
}
