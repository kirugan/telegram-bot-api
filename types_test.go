package tgbotapi

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUserStringWith(t *testing.T) {
	user := User{
		ID:           0,
		FirstName:    "Test",
		LastName:     "Test",
		UserName:     "",
		LanguageCode: "en",
		IsBot:        false,
	}

	assert.Equal(t, "Test Test", user.String())
}

func TestUserStringWithUserName(t *testing.T) {
	user := User{
		ID:           0,
		FirstName:    "Test",
		LastName:     "Test",
		UserName:     "@test",
		LanguageCode: "en",
	}

	assert.Equal(t, "@test", user.String())
}

func TestMessageTime(t *testing.T) {
	message := Message{Date: 0}

	assert.Equal(t, time.Unix(0, 0), message.Time())
}

func TestMessageIsCommandWithCommand(t *testing.T) {
	message := Message{Text: "/command"}
	message.Entities = []MessageEntity{{Type: "bot_command", Offset: 0, Length: 8}}

	assert.True(t, message.IsCommand())
}

func TestIsCommandWithText(t *testing.T) {
	message := Message{Text: "some text"}

	assert.False(t, message.IsCommand())
}

func TestIsCommandWithEmptyText(t *testing.T) {
	message := Message{Text: ""}

	assert.False(t, message.IsCommand())
}

func TestCommandWithCommand(t *testing.T) {
	message := Message{Text: "/command"}
	message.Entities = []MessageEntity{{Type: "bot_command", Offset: 0, Length: 8}}

	assert.Equal(t, "command", message.Command())
}

func TestCommandWithEmptyText(t *testing.T) {
	message := Message{Text: ""}

	assert.Equal(t, "", message.Command())
}

func TestCommandWithNonCommand(t *testing.T) {
	message := Message{Text: "test text"}

	assert.Equal(t, "", message.Command())
}

func TestCommandWithBotName(t *testing.T) {
	message := Message{Text: "/command@testbot"}
	message.Entities = []MessageEntity{{Type: "bot_command", Offset: 0, Length: 16}}

	assert.Equal(t, "command", message.Command())
}

func TestCommandWithAtWithBotName(t *testing.T) {
	message := Message{Text: "/command@testbot"}
	message.Entities = []MessageEntity{{Type: "bot_command", Offset: 0, Length: 16}}

	assert.Equal(t, "command@testbot", message.CommandWithAt())
}

func TestMessageCommandArgumentsWithArguments(t *testing.T) {
	message := Message{Text: "/command with arguments"}
	message.Entities = []MessageEntity{{Type: "bot_command", Offset: 0, Length: 8}}
	assert.Equal(t, "with arguments", message.CommandArguments())
}

func TestMessageCommandArgumentsWithMalformedArguments(t *testing.T) {
	message := Message{Text: "/command-without argument space"}
	message.Entities = []MessageEntity{{Type: "bot_command", Offset: 0, Length: 8}}
	assert.Equal(t, "without argument space", message.CommandArguments())
}

func TestMessageCommandArgumentsWithoutArguments(t *testing.T) {
	message := Message{Text: "/command"}
	assert.Equal(t, "", message.CommandArguments())
}

func TestMessageCommandArgumentsForNonCommand(t *testing.T) {
	message := Message{Text: "test text"}
	assert.Equal(t, "", message.CommandArguments())
}

func TestMessageEntityParseURLGood(t *testing.T) {
	entity := MessageEntity{URL: "https://www.google.com"}

	_, err := entity.ParseURL()
	assert.NoError(t, err)
}

func TestMessageEntityParseURLBad(t *testing.T) {
	entity := MessageEntity{URL: ""}

	_, err := entity.ParseURL()
	assert.Error(t, err)
}

func TestChatIsPrivate(t *testing.T) {
	chat := Chat{ID: 10, Type: "private"}

	assert.True(t, chat.IsPrivate())
}

func TestChatIsGroup(t *testing.T) {
	chat := Chat{ID: 10, Type: "group"}

	assert.True(t, chat.IsGroup())
}

func TestChatIsChannel(t *testing.T) {
	chat := Chat{ID: 10, Type: "channel"}

	assert.True(t, chat.IsChannel())
}

func TestChatIsSuperGroup(t *testing.T) {
	chat := Chat{ID: 10, Type: "supergroup"}

	assert.True(t, chat.IsSuperGroup())
}

func TestMessageEntityIsMention(t *testing.T) {
	entity := MessageEntity{Type: "mention"}

	assert.True(t, entity.IsMention())
}

func TestMessageEntityIsHashtag(t *testing.T) {
	entity := MessageEntity{Type: "hashtag"}

	assert.True(t, entity.IsHashtag())
}

func TestMessageEntityIsBotCommand(t *testing.T) {
	entity := MessageEntity{Type: "bot_command"}

	assert.True(t, entity.IsCommand())
}

func TestMessageEntityIsUrl(t *testing.T) {
	entity := MessageEntity{Type: "url"}

	assert.True(t, entity.IsURL())
}

func TestMessageEntityIsEmail(t *testing.T) {
	entity := MessageEntity{Type: "email"}

	assert.True(t, entity.IsEmail())
}

func TestMessageEntityIsBold(t *testing.T) {
	entity := MessageEntity{Type: "bold"}

	assert.True(t, entity.IsBold())
}

func TestMessageEntityIsItalic(t *testing.T) {
	entity := MessageEntity{Type: "italic"}

	assert.True(t, entity.IsItalic())
}

func TestMessageEntityIsCode(t *testing.T) {
	entity := MessageEntity{Type: "code"}

	assert.True(t, entity.IsCode())
}

func TestMessageEntityIsPre(t *testing.T) {
	entity := MessageEntity{Type: "pre"}

	assert.True(t, entity.IsPre())
}

func TestMessageEntityIsTextLink(t *testing.T) {
	entity := MessageEntity{Type: "text_link"}

	assert.True(t, entity.IsTextLink())
}

func TestFileLink(t *testing.T) {
	file := File{FilePath: "test/test.txt"}

	assert.Equal(t, "https://api.telegram.org/file/bottoken/test/test.txt", file.Link("token"))
}

// Ensure all configs are sendable
var (
	_ Chattable = AnimationConfig{}
	_ Chattable = AnswerWebAppQueryConfig{}
	_ Chattable = AudioConfig{}
	_ Chattable = BanChatMemberConfig{}
	_ Chattable = BanChatSenderChatConfig{}
	_ Chattable = CallbackConfig{}
	_ Chattable = ChatActionConfig{}
	_ Chattable = ChatAdministratorsConfig{}
	_ Chattable = ChatInfoConfig{}
	_ Chattable = ChatInviteLinkConfig{}
	_ Chattable = CloseConfig{}
	_ Chattable = ContactConfig{}
	_ Chattable = CopyMessageConfig{}
	_ Chattable = CreateChatInviteLinkConfig{}
	_ Chattable = DeleteChatPhotoConfig{}
	_ Chattable = DeleteChatStickerSetConfig{}
	_ Chattable = DeleteMessageConfig{}
	_ Chattable = DeleteMyCommandsConfig{}
	_ Chattable = DeleteWebhookConfig{}
	_ Chattable = DocumentConfig{}
	_ Chattable = EditChatInviteLinkConfig{}
	_ Chattable = EditMessageCaptionConfig{}
	_ Chattable = EditMessageLiveLocationConfig{}
	_ Chattable = EditMessageMediaConfig{}
	_ Chattable = EditMessageReplyMarkupConfig{}
	_ Chattable = EditMessageTextConfig{}
	_ Chattable = FileConfig{}
	_ Chattable = ForwardConfig{}
	_ Chattable = GameConfig{}
	_ Chattable = GetChatMemberConfig{}
	_ Chattable = GetChatMenuButtonConfig{}
	_ Chattable = GetGameHighScoresConfig{}
	_ Chattable = GetMyDefaultAdministratorRightsConfig{}
	_ Chattable = InlineConfig{}
	_ Chattable = InvoiceConfig{}
	_ Chattable = KickChatMemberConfig{}
	_ Chattable = LeaveChatConfig{}
	_ Chattable = LocationConfig{}
	_ Chattable = LogOutConfig{}
	_ Chattable = MediaGroupConfig{}
	_ Chattable = MessageConfig{}
	_ Chattable = PhotoConfig{}
	_ Chattable = PinChatMessageConfig{}
	_ Chattable = PreCheckoutConfig{}
	_ Chattable = PromoteChatMemberConfig{}
	_ Chattable = RestrictChatMemberConfig{}
	_ Chattable = RevokeChatInviteLinkConfig{}
	_ Chattable = SendPollConfig{}
	_ Chattable = SetChatDescriptionConfig{}
	_ Chattable = SetChatMenuButtonConfig{}
	_ Chattable = SetChatPhotoConfig{}
	_ Chattable = SetChatTitleConfig{}
	_ Chattable = SetGameScoreConfig{}
	_ Chattable = SetMyDefaultAdministratorRightsConfig{}
	_ Chattable = ShippingConfig{}
	_ Chattable = StickerConfig{}
	_ Chattable = StopMessageLiveLocationConfig{}
	_ Chattable = StopPollConfig{}
	_ Chattable = UnbanChatMemberConfig{}
	_ Chattable = UnbanChatSenderChatConfig{}
	_ Chattable = UnpinChatMessageConfig{}
	_ Chattable = UpdateConfig{}
	_ Chattable = UserProfilePhotosConfig{}
	_ Chattable = VenueConfig{}
	_ Chattable = VideoConfig{}
	_ Chattable = VideoNoteConfig{}
	_ Chattable = VoiceConfig{}
	_ Chattable = WebhookConfig{}
)

// Ensure all Fileable types are correct.
var (
	_ Fileable = (*PhotoConfig)(nil)
	_ Fileable = (*AudioConfig)(nil)
	_ Fileable = (*DocumentConfig)(nil)
	_ Fileable = (*StickerConfig)(nil)
	_ Fileable = (*VideoConfig)(nil)
	_ Fileable = (*AnimationConfig)(nil)
	_ Fileable = (*VideoNoteConfig)(nil)
	_ Fileable = (*VoiceConfig)(nil)
	_ Fileable = (*SetChatPhotoConfig)(nil)
	_ Fileable = (*EditMessageMediaConfig)(nil)
	_ Fileable = (*SetChatPhotoConfig)(nil)
	_ Fileable = (*UploadStickerConfig)(nil)
	_ Fileable = (*NewStickerSetConfig)(nil)
	_ Fileable = (*AddStickerConfig)(nil)
	_ Fileable = (*MediaGroupConfig)(nil)
	_ Fileable = (*WebhookConfig)(nil)
	_ Fileable = (*SetStickerSetThumbnailConfig)(nil)
)

// Ensure all RequestFileData types are correct.
var (
	_ RequestFileData = (*FilePath)(nil)
	_ RequestFileData = (*FileBytes)(nil)
	_ RequestFileData = (*FileReader)(nil)
	_ RequestFileData = (*FileURL)(nil)
	_ RequestFileData = (*FileID)(nil)
	_ RequestFileData = (*fileAttach)(nil)
)
