package tgbotapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"
)

// Telegram constants
const (
	// APIEndpoint is the endpoint for all API methods,
	// with formatting for Sprintf.
	APIEndpoint = "https://api.telegram.org/bot%s/%s"
	// FileEndpoint is the endpoint for downloading a file from Telegram.
	FileEndpoint = "https://api.telegram.org/file/bot%s/%s"
)

// Constant values for ChatActions
const (
	ChatTyping          = "typing"
	ChatUploadPhoto     = "upload_photo"
	ChatRecordVideo     = "record_video"
	ChatUploadVideo     = "upload_video"
	ChatRecordVoice     = "record_voice"
	ChatUploadVoice     = "upload_voice"
	ChatUploadDocument  = "upload_document"
	ChatChooseSticker   = "choose_sticker"
	ChatFindLocation    = "find_location"
	ChatRecordVideoNote = "record_video_note"
	ChatUploadVideoNote = "upload_video_note"
)

// API errors
const (
	// ErrAPIForbidden happens when a token is bad
	ErrAPIForbidden = "forbidden"
)

// Constant values for ParseMode in MessageConfig
const (
	ModeMarkdown   = "Markdown"
	ModeMarkdownV2 = "MarkdownV2"
	ModeHTML       = "HTML"
)

// Constant values for update types
const (
	// UpdateTypeMessage is new incoming message of any kind — text, photo, sticker, etc.
	UpdateTypeMessage = "message"

	// UpdateTypeEditedMessage is new version of a message that is known to the bot and was edited
	UpdateTypeEditedMessage = "edited_message"

	// UpdateTypeChannelPost is new incoming channel post of any kind — text, photo, sticker, etc.
	UpdateTypeChannelPost = "channel_post"

	// UpdateTypeEditedChannelPost is new version of a channel post that is known to the bot and was edited
	UpdateTypeEditedChannelPost = "edited_channel_post"

	// UpdateTypeInlineQuery is new incoming inline query
	UpdateTypeInlineQuery = "inline_query"

	// UpdateTypeChosenInlineResult i the result of an inline query that was chosen by a user and sent to their
	// chat partner. Please see the documentation on the feedback collecting for
	// details on how to enable these updates for your bot.
	UpdateTypeChosenInlineResult = "chosen_inline_result"

	// UpdateTypeCallbackQuery is new incoming callback query
	UpdateTypeCallbackQuery = "callback_query"

	// UpdateTypeShippingQuery is new incoming shipping query. Only for invoices with flexible price
	UpdateTypeShippingQuery = "shipping_query"

	// UpdateTypePreCheckoutQuery is new incoming pre-checkout query. Contains full information about checkout
	UpdateTypePreCheckoutQuery = "pre_checkout_query"

	// UpdateTypePoll is new poll state. Bots receive only updates about stopped polls and polls
	// which are sent by the bot
	UpdateTypePoll = "poll"

	// UpdateTypePollAnswer is when user changed their answer in a non-anonymous poll. Bots receive new votes
	// only in polls that were sent by the bot itself.
	UpdateTypePollAnswer = "poll_answer"

	// UpdateTypeMessageReaction is when a reaction to a message was changed by a user.
	UpdateTypeMessageReaction = "message_reaction"

	// UpdateTypeMessageReactionCount is when reactions to a message with anonymous reactions were changed.
	UpdateTypeMessageReactionCount = "message_reaction_count"

	// UpdateTypeChatBoost is when a boost was added to a chat or changed.
	UpdateTypeChatBoost = "chat_boost"

	// UpdateTypeRemovedChatBoost is when a boost was removed from a chat.
	UpdateTypeRemovedChatBoost = "removed_chat_boost"

	// UpdateTypeBusinessConnection is when the bot was connected to or
	// disconnected from a business account, or a user edited an existing
	// connection with the bot.
	UpdateTypeBusinessConnection = "business_connection"

	// UpdateTypeBusinessMessage is a new non-service message from a connected
	// business account.
	UpdateTypeBusinessMessage = "business_message"

	// UpdateTypeEditedBusinessMessage is a new version of a message from a
	// connected business account.
	UpdateTypeEditedBusinessMessage = "edited_business_message"

	// UpdateTypeDeletedBusinessMessages is when messages were deleted from a
	// connected business account.
	UpdateTypeDeletedBusinessMessages = "deleted_business_messages"

	// UpdateTypePurchasedPaidMedia is when a user purchased paid media with
	// a non-empty payload sent by the bot in a non-channel chat.
	UpdateTypePurchasedPaidMedia = "purchased_paid_media"

	// UpdateTypeMyChatMember is when the bot's chat member status was updated in a chat. For private chats, this
	// update is received only when the bot is blocked or unblocked by the user.
	UpdateTypeMyChatMember = "my_chat_member"

	// UpdateTypeChatMember is when the bot must be an administrator in the chat and must explicitly specify
	// this update in the list of allowed_updates to receive these updates.
	UpdateTypeChatMember = "chat_member"

	// UpdateTypeManagedBot is when a new bot was created to be managed by the bot,
	// or token or owner of a managed bot was changed.
	UpdateTypeManagedBot = "managed_bot"
)

// Library errors
const (
	ErrBadURL = "bad or empty url"
)

// Chattable is any config type that can be sent.
type Chattable interface {
	params() (Params, error)
	method() string
}

// Fileable is any config type that can be sent that includes a file.
type Fileable interface {
	Chattable
	files() []RequestFile
}

// RequestFile represents a file associated with a field name.
type RequestFile struct {
	// The file field name.
	Name string
	// The file data to include.
	Data RequestFileData
}

// RequestFileData represents the data to be used for a file.
type RequestFileData interface {
	// NeedsUpload shows if the file needs to be uploaded.
	NeedsUpload() bool

	// UploadData gets the file name and an `io.Reader` for the file to be uploaded. This
	// must only be called when the file needs to be uploaded.
	UploadData() (string, io.Reader, error)
	// SendData gets the file data to send when a file does not need to be uploaded. This
	// must only be called when the file does not need to be uploaded.
	SendData() string
}

// FileBytes contains information about a set of bytes to upload
// as a File.
type FileBytes struct {
	Name  string
	Bytes []byte
}

func (fb FileBytes) NeedsUpload() bool {
	return true
}

func (fb FileBytes) UploadData() (string, io.Reader, error) {
	return fb.Name, bytes.NewReader(fb.Bytes), nil
}

func (fb FileBytes) SendData() string {
	panic("FileBytes must be uploaded")
}

// FileReader contains information about a reader to upload as a File.
type FileReader struct {
	Name   string
	Reader io.Reader
}

func (fr FileReader) NeedsUpload() bool {
	return true
}

func (fr FileReader) UploadData() (string, io.Reader, error) {
	return fr.Name, fr.Reader, nil
}

func (fr FileReader) SendData() string {
	panic("FileReader must be uploaded")
}

// FilePath is a path to a local file.
type FilePath string

func (fp FilePath) NeedsUpload() bool {
	return true
}

func (fp FilePath) UploadData() (string, io.Reader, error) {
	fileHandle, err := os.Open(string(fp))
	if err != nil {
		return "", nil, err
	}

	name := fileHandle.Name()
	return name, fileHandle, err
}

func (fp FilePath) SendData() string {
	panic("FilePath must be uploaded")
}

// FileURL is a URL to use as a file for a request.
type FileURL string

func (fu FileURL) NeedsUpload() bool {
	return false
}

func (fu FileURL) UploadData() (string, io.Reader, error) {
	panic("FileURL cannot be uploaded")
}

func (fu FileURL) SendData() string {
	return string(fu)
}

// FileID is an ID of a file already uploaded to Telegram.
type FileID string

func (fi FileID) NeedsUpload() bool {
	return false
}

func (fi FileID) UploadData() (string, io.Reader, error) {
	panic("FileID cannot be uploaded")
}

func (fi FileID) SendData() string {
	return string(fi)
}

// fileAttach is an internal file type used for processed media groups.
type fileAttach string

func (fa fileAttach) NeedsUpload() bool {
	return false
}

func (fa fileAttach) UploadData() (string, io.Reader, error) {
	panic("fileAttach cannot be uploaded")
}

func (fa fileAttach) SendData() string {
	return string(fa)
}

// LogOutConfig is a request to log out of the cloud Bot API server.
//
// Note that you may not log back in for at least 10 minutes.
type LogOutConfig struct{}

func (LogOutConfig) method() string {
	return "logOut"
}

func (LogOutConfig) params() (Params, error) {
	return nil, nil
}

// CloseConfig is a request to close the bot instance on a local server.
//
// Note that you may not close an instance for the first 10 minutes after the
// bot has started.
type CloseConfig struct{}

func (CloseConfig) method() string {
	return "close"
}

func (CloseConfig) params() (Params, error) {
	return nil, nil
}

// BaseChat is base type for all chat config types.
type BaseChat struct {
	ChatID               int64 // required
	ChannelUsername      string
	BusinessConnectionID string
	MessageThreadID      int
	MessageEffectID      string
	ProtectContent       bool
	AllowPaidBroadcast   bool
	ReplyParameters      *ReplyParameters
	ReplyMarkup          interface{}
	DisableNotification  bool
}

func (chat *BaseChat) params() (Params, error) {
	params := make(Params)

	if err := params.AddFirstValid("chat_id", chat.ChatID, chat.ChannelUsername); err != nil {
		return params, err
	}
	params.AddNonEmpty("business_connection_id", chat.BusinessConnectionID)
	params.AddNonZero("message_thread_id", chat.MessageThreadID)
	params.AddNonEmpty("message_effect_id", chat.MessageEffectID)
	params.AddBool("disable_notification", chat.DisableNotification)
	params.AddBool("protect_content", chat.ProtectContent)
	params.AddBool("allow_paid_broadcast", chat.AllowPaidBroadcast)

	if err := params.AddAny("reply_parameters", chat.ReplyParameters); err != nil {
		return params, err
	}
	err := params.AddAny("reply_markup", chat.ReplyMarkup)

	return params, err
}

// BaseFile is a base type for all file config types.
type BaseFile struct {
	BaseChat
	File RequestFileData
}

func (file BaseFile) params() (Params, error) {
	return file.BaseChat.params()
}

// BaseEdit is base type of all chat edits.
type BaseEdit struct {
	BusinessConnectionID string
	ChatID               int64
	ChannelUsername      string
	MessageID            int
	InlineMessageID      string
	ReplyMarkup          *InlineKeyboardMarkup
}

func (edit BaseEdit) params() (Params, error) {
	params := make(Params)

	params.AddNonEmpty("business_connection_id", edit.BusinessConnectionID)

	if edit.InlineMessageID != "" {
		params["inline_message_id"] = edit.InlineMessageID
	} else {
		if err := params.AddFirstValid("chat_id", edit.ChatID, edit.ChannelUsername); err != nil {
			return params, err
		}
		params.AddNonZero("message_id", edit.MessageID)
	}

	err := params.AddAny("reply_markup", edit.ReplyMarkup)

	return params, err
}

// MessageConfig contains information about a SendMessage request.
type MessageConfig struct {
	BaseChat
	Text               string
	ParseMode          string
	Entities           []MessageEntity
	LinkPreviewOptions *LinkPreviewOptions
}

func (config MessageConfig) params() (Params, error) {
	params, err := config.BaseChat.params()
	if err != nil {
		return params, err
	}

	params.AddNonEmpty("text", config.Text)
	params.AddNonEmpty("parse_mode", config.ParseMode)
	if err = params.AddAny("link_preview_options", config.LinkPreviewOptions); err != nil {
		return params, err
	}
	err = params.AddAny("entities", config.Entities)

	return params, err
}

func (config MessageConfig) method() string {
	return "sendMessage"
}

// ForwardConfig contains information about a ForwardMessage request.
type ForwardConfig struct {
	BaseChat
	FromChatID          int64 // required
	FromChannelUsername string
	MessageID           int // required
}

func (config ForwardConfig) params() (Params, error) {
	params, err := config.BaseChat.params()
	if err != nil {
		return params, err
	}

	params.AddNonZero64("from_chat_id", config.FromChatID)
	params.AddNonZero("message_id", config.MessageID)

	return params, nil
}

func (config ForwardConfig) method() string {
	return "forwardMessage"
}

// CopyMessageConfig contains information about a copyMessage request.
type CopyMessageConfig struct {
	BaseChat
	FromChatID            int64
	FromChannelUsername   string
	MessageID             int
	Caption               string
	ParseMode             string
	CaptionEntities       []MessageEntity
	ShowCaptionAboveMedia bool
}

func (config CopyMessageConfig) params() (Params, error) {
	params, err := config.BaseChat.params()
	if err != nil {
		return params, err
	}

	if err = params.AddFirstValid("from_chat_id", config.FromChatID, config.FromChannelUsername); err != nil {
		return params, err
	}
	params.AddNonZero("message_id", config.MessageID)
	params.AddNonEmpty("caption", config.Caption)
	params.AddNonEmpty("parse_mode", config.ParseMode)
	params.AddBool("show_caption_above_media", config.ShowCaptionAboveMedia)
	err = params.AddAny("caption_entities", config.CaptionEntities)

	return params, err
}

func (config CopyMessageConfig) method() string {
	return "copyMessage"
}

// PhotoConfig contains information about a SendPhoto request.
type PhotoConfig struct {
	BaseFile
	Thumbnail             RequestFileData
	Caption               string
	ParseMode             string
	CaptionEntities       []MessageEntity
	ShowCaptionAboveMedia bool
	HasSpoiler            bool
}

func (config PhotoConfig) params() (Params, error) {
	params, err := config.BaseFile.params()
	if err != nil {
		return params, err
	}

	params.AddNonEmpty("caption", config.Caption)
	params.AddNonEmpty("parse_mode", config.ParseMode)
	params.AddBool("show_caption_above_media", config.ShowCaptionAboveMedia)
	params.AddBool("has_spoiler", config.HasSpoiler)
	err = params.AddAny("caption_entities", config.CaptionEntities)

	return params, err
}

func (config PhotoConfig) method() string {
	return "sendPhoto"
}

func (config PhotoConfig) files() []RequestFile {
	files := []RequestFile{{
		Name: "photo",
		Data: config.File,
	}}

	if config.Thumbnail != nil {
		files = append(files, RequestFile{
			Name: "thumbnail",
			Data: config.Thumbnail,
		})
	}

	return files
}

// AudioConfig contains information about a SendAudio request.
type AudioConfig struct {
	BaseFile
	Thumbnail       RequestFileData
	Caption         string
	ParseMode       string
	CaptionEntities []MessageEntity
	Duration        int
	Performer       string
	Title           string
}

func (config AudioConfig) params() (Params, error) {
	params, err := config.BaseChat.params()
	if err != nil {
		return params, err
	}

	params.AddNonZero("duration", config.Duration)
	params.AddNonEmpty("performer", config.Performer)
	params.AddNonEmpty("title", config.Title)
	params.AddNonEmpty("caption", config.Caption)
	params.AddNonEmpty("parse_mode", config.ParseMode)
	err = params.AddInterface("caption_entities", config.CaptionEntities)

	return params, err
}

func (config AudioConfig) method() string {
	return "sendAudio"
}

func (config AudioConfig) files() []RequestFile {
	files := []RequestFile{{
		Name: "audio",
		Data: config.File,
	}}

	if config.Thumbnail != nil {
		files = append(files, RequestFile{
			Name: "thumbnail",
			Data: config.Thumbnail,
		})
	}

	return files
}

// DocumentConfig contains information about a SendDocument request.
type DocumentConfig struct {
	BaseFile
	Thumbnail                   RequestFileData
	Caption                     string
	ParseMode                   string
	CaptionEntities             []MessageEntity
	DisableContentTypeDetection bool
}

func (config DocumentConfig) params() (Params, error) {
	params, err := config.BaseFile.params()

	params.AddNonEmpty("caption", config.Caption)
	params.AddNonEmpty("parse_mode", config.ParseMode)
	params.AddBool("disable_content_type_detection", config.DisableContentTypeDetection)

	return params, err
}

func (config DocumentConfig) method() string {
	return "sendDocument"
}

func (config DocumentConfig) files() []RequestFile {
	files := []RequestFile{{
		Name: "document",
		Data: config.File,
	}}

	if config.Thumbnail != nil {
		files = append(files, RequestFile{
			Name: "thumbnail",
			Data: config.Thumbnail,
		})
	}

	return files
}

// StickerConfig contains information about a SendSticker request.
type StickerConfig struct {
	BaseFile
	// Emoji associated with the sticker; only for just uploaded stickers.
	Emoji string
}

func (config StickerConfig) params() (Params, error) {
	params, err := config.BaseChat.params()
	if err != nil {
		return params, err
	}

	params.AddNonEmpty("emoji", config.Emoji)

	return params, nil
}

func (config StickerConfig) method() string {
	return "sendSticker"
}

func (config StickerConfig) files() []RequestFile {
	return []RequestFile{{
		Name: "sticker",
		Data: config.File,
	}}
}

// VideoConfig contains information about a SendVideo request.
type VideoConfig struct {
	BaseFile
	Thumbnail             RequestFileData
	Duration              int
	Caption               string
	ParseMode             string
	CaptionEntities       []MessageEntity
	ShowCaptionAboveMedia bool
	SupportsStreaming     bool
	HasSpoiler            bool
}

func (config VideoConfig) params() (Params, error) {
	params, err := config.BaseChat.params()
	if err != nil {
		return params, err
	}

	params.AddNonZero("duration", config.Duration)
	params.AddNonEmpty("caption", config.Caption)
	params.AddNonEmpty("parse_mode", config.ParseMode)
	params.AddBool("show_caption_above_media", config.ShowCaptionAboveMedia)
	params.AddBool("supports_streaming", config.SupportsStreaming)
	params.AddBool("has_spoiler", config.HasSpoiler)
	err = params.AddAny("caption_entities", config.CaptionEntities)

	return params, err
}

func (config VideoConfig) method() string {
	return "sendVideo"
}

func (config VideoConfig) files() []RequestFile {
	files := []RequestFile{{
		Name: "video",
		Data: config.File,
	}}

	if config.Thumbnail != nil {
		files = append(files, RequestFile{
			Name: "thumbnail",
			Data: config.Thumbnail,
		})
	}

	return files
}

// AnimationConfig contains information about a SendAnimation request.
type AnimationConfig struct {
	BaseFile
	Duration              int
	Thumbnail             RequestFileData
	Caption               string
	ParseMode             string
	CaptionEntities       []MessageEntity
	ShowCaptionAboveMedia bool
	HasSpoiler            bool
}

func (config AnimationConfig) params() (Params, error) {
	params, err := config.BaseChat.params()
	if err != nil {
		return params, err
	}

	params.AddNonZero("duration", config.Duration)
	params.AddNonEmpty("caption", config.Caption)
	params.AddNonEmpty("parse_mode", config.ParseMode)
	params.AddBool("show_caption_above_media", config.ShowCaptionAboveMedia)
	params.AddBool("has_spoiler", config.HasSpoiler)
	err = params.AddAny("caption_entities", config.CaptionEntities)

	return params, err
}

func (config AnimationConfig) method() string {
	return "sendAnimation"
}

func (config AnimationConfig) files() []RequestFile {
	files := []RequestFile{{
		Name: "animation",
		Data: config.File,
	}}

	if config.Thumbnail != nil {
		files = append(files, RequestFile{
			Name: "thumbnail",
			Data: config.Thumbnail,
		})
	}

	return files
}

// VideoNoteConfig contains information about a SendVideoNote request.
type VideoNoteConfig struct {
	BaseFile
	Thumbnail RequestFileData
	Duration  int
	Length    int
}

func (config VideoNoteConfig) params() (Params, error) {
	params, err := config.BaseChat.params()

	params.AddNonZero("duration", config.Duration)
	params.AddNonZero("length", config.Length)

	return params, err
}

func (config VideoNoteConfig) method() string {
	return "sendVideoNote"
}

func (config VideoNoteConfig) files() []RequestFile {
	files := []RequestFile{{
		Name: "video_note",
		Data: config.File,
	}}

	if config.Thumbnail != nil {
		files = append(files, RequestFile{
			Name: "thumbnail",
			Data: config.Thumbnail,
		})
	}

	return files
}

// VoiceConfig contains information about a SendVoice request.
type VoiceConfig struct {
	BaseFile
	Thumbnail       RequestFileData
	Caption         string
	ParseMode       string
	CaptionEntities []MessageEntity
	Duration        int
}

func (config VoiceConfig) params() (Params, error) {
	params, err := config.BaseChat.params()
	if err != nil {
		return params, err
	}

	params.AddNonZero("duration", config.Duration)
	params.AddNonEmpty("caption", config.Caption)
	params.AddNonEmpty("parse_mode", config.ParseMode)
	err = params.AddInterface("caption_entities", config.CaptionEntities)

	return params, err
}

func (config VoiceConfig) method() string {
	return "sendVoice"
}

func (config VoiceConfig) files() []RequestFile {
	files := []RequestFile{{
		Name: "voice",
		Data: config.File,
	}}

	if config.Thumbnail != nil {
		files = append(files, RequestFile{
			Name: "thumbnail",
			Data: config.Thumbnail,
		})
	}

	return files
}

// LocationConfig contains information about a SendLocation request.
type LocationConfig struct {
	BaseChat
	Latitude             float64 // required
	Longitude            float64 // required
	HorizontalAccuracy   float64 // optional
	LivePeriod           int     // optional
	Heading              int     // optional
	ProximityAlertRadius int     // optional
}

func (config LocationConfig) params() (Params, error) {
	params, err := config.BaseChat.params()

	params.AddNonZeroFloat("latitude", config.Latitude)
	params.AddNonZeroFloat("longitude", config.Longitude)
	params.AddNonZeroFloat("horizontal_accuracy", config.HorizontalAccuracy)
	params.AddNonZero("live_period", config.LivePeriod)
	params.AddNonZero("heading", config.Heading)
	params.AddNonZero("proximity_alert_radius", config.ProximityAlertRadius)

	return params, err
}

func (config LocationConfig) method() string {
	return "sendLocation"
}

// EditMessageLiveLocationConfig allows you to update a live location.
type EditMessageLiveLocationConfig struct {
	BaseEdit
	Latitude             float64 // required
	Longitude            float64 // required
	LivePeriod           int     // optional; pass 0x7FFFFFFF to keep live indefinitely
	HorizontalAccuracy   float64 // optional
	Heading              int     // optional
	ProximityAlertRadius int     // optional
}

func (config EditMessageLiveLocationConfig) params() (Params, error) {
	params, err := config.BaseEdit.params()

	params.AddNonZeroFloat("latitude", config.Latitude)
	params.AddNonZeroFloat("longitude", config.Longitude)
	params.AddNonZero("live_period", config.LivePeriod)
	params.AddNonZeroFloat("horizontal_accuracy", config.HorizontalAccuracy)
	params.AddNonZero("heading", config.Heading)
	params.AddNonZero("proximity_alert_radius", config.ProximityAlertRadius)

	return params, err
}

func (config EditMessageLiveLocationConfig) method() string {
	return "editMessageLiveLocation"
}

// StopMessageLiveLocationConfig stops updating a live location.
type StopMessageLiveLocationConfig struct {
	BaseEdit
}

func (config StopMessageLiveLocationConfig) params() (Params, error) {
	return config.BaseEdit.params()
}

func (config StopMessageLiveLocationConfig) method() string {
	return "stopMessageLiveLocation"
}

// VenueConfig contains information about a SendVenue request.
type VenueConfig struct {
	BaseChat
	Latitude        float64 // required
	Longitude       float64 // required
	Title           string  // required
	Address         string  // required
	FoursquareID    string
	FoursquareType  string
	GooglePlaceID   string
	GooglePlaceType string
}

func (config VenueConfig) params() (Params, error) {
	params, err := config.BaseChat.params()

	params.AddNonZeroFloat("latitude", config.Latitude)
	params.AddNonZeroFloat("longitude", config.Longitude)
	params["title"] = config.Title
	params["address"] = config.Address
	params.AddNonEmpty("foursquare_id", config.FoursquareID)
	params.AddNonEmpty("foursquare_type", config.FoursquareType)
	params.AddNonEmpty("google_place_id", config.GooglePlaceID)
	params.AddNonEmpty("google_place_type", config.GooglePlaceType)

	return params, err
}

func (config VenueConfig) method() string {
	return "sendVenue"
}

// ContactConfig allows you to send a contact.
type ContactConfig struct {
	BaseChat
	PhoneNumber string
	FirstName   string
	LastName    string
	VCard       string
}

func (config ContactConfig) params() (Params, error) {
	params, err := config.BaseChat.params()

	params["phone_number"] = config.PhoneNumber
	params["first_name"] = config.FirstName

	params.AddNonEmpty("last_name", config.LastName)
	params.AddNonEmpty("vcard", config.VCard)

	return params, err
}

func (config ContactConfig) method() string {
	return "sendContact"
}

// SendPollConfig allows you to send a poll.
type SendPollConfig struct {
	BaseChat
	Question              string
	QuestionParseMode     string
	QuestionEntities      []MessageEntity
	Options               []InputPollOption
	IsAnonymous           bool
	Type                  string
	AllowsMultipleAnswers bool
	CorrectOptionID       int64
	Explanation           string
	ExplanationParseMode  string
	ExplanationEntities   []MessageEntity
	OpenPeriod            int
	CloseDate             int
	IsClosed              bool
}

func (config SendPollConfig) params() (Params, error) {
	params, err := config.BaseChat.params()
	if err != nil {
		return params, err
	}

	params["question"] = config.Question
	params.AddNonEmpty("question_parse_mode", config.QuestionParseMode)
	if err = params.AddAny("question_entities", config.QuestionEntities); err != nil {
		return params, err
	}
	if err = params.AddAny("options", config.Options); err != nil {
		return params, err
	}
	params["is_anonymous"] = strconv.FormatBool(config.IsAnonymous)
	params.AddNonEmpty("type", config.Type)
	params["allows_multiple_answers"] = strconv.FormatBool(config.AllowsMultipleAnswers)
	params["correct_option_id"] = strconv.FormatInt(config.CorrectOptionID, 10)
	params.AddBool("is_closed", config.IsClosed)
	params.AddNonEmpty("explanation", config.Explanation)
	params.AddNonEmpty("explanation_parse_mode", config.ExplanationParseMode)
	params.AddNonZero("open_period", config.OpenPeriod)
	params.AddNonZero("close_date", config.CloseDate)
	err = params.AddAny("explanation_entities", config.ExplanationEntities)

	return params, err
}

func (SendPollConfig) method() string {
	return "sendPoll"
}

// GameConfig allows you to send a game.
type GameConfig struct {
	BaseChat
	GameShortName string
}

func (config GameConfig) params() (Params, error) {
	params, err := config.BaseChat.params()

	params["game_short_name"] = config.GameShortName

	return params, err
}

func (config GameConfig) method() string {
	return "sendGame"
}

// SetGameScoreConfig allows you to update the game score in a chat.
type SetGameScoreConfig struct {
	UserID             int64
	Score              int
	Force              bool
	DisableEditMessage bool
	ChatID             int64
	ChannelUsername    string
	MessageID          int
	InlineMessageID    string
}

func (config SetGameScoreConfig) params() (Params, error) {
	params := make(Params)

	params.AddNonZero64("user_id", config.UserID)
	params.AddNonZero("scrore", config.Score)
	params.AddBool("disable_edit_message", config.DisableEditMessage)

	if config.InlineMessageID != "" {
		params["inline_message_id"] = config.InlineMessageID
	} else {
		if err := params.AddFirstValid("chat_id", config.ChatID, config.ChannelUsername); err != nil {
			return params, err
		}
		params.AddNonZero("message_id", config.MessageID)
	}

	return params, nil
}

func (config SetGameScoreConfig) method() string {
	return "setGameScore"
}

// GetGameHighScoresConfig allows you to fetch the high scores for a game.
type GetGameHighScoresConfig struct {
	UserID          int64
	ChatID          int64
	ChannelUsername string
	MessageID       int
	InlineMessageID string
}

func (config GetGameHighScoresConfig) params() (Params, error) {
	params := make(Params)

	params.AddNonZero64("user_id", config.UserID)

	if config.InlineMessageID != "" {
		params["inline_message_id"] = config.InlineMessageID
	} else {
		if err := params.AddFirstValid("chat_id", config.ChatID, config.ChannelUsername); err != nil {
			return params, err
		}
		params.AddNonZero("message_id", config.MessageID)
	}

	return params, nil
}

func (config GetGameHighScoresConfig) method() string {
	return "getGameHighScores"
}

// ChatActionConfig contains information about a SendChatAction request.
type ChatActionConfig struct {
	BaseChat
	Action string // required
}

func (config ChatActionConfig) params() (Params, error) {
	params, err := config.BaseChat.params()

	params["action"] = config.Action

	return params, err
}

func (config ChatActionConfig) method() string {
	return "sendChatAction"
}

// EditMessageTextConfig allows you to modify the text in a message.
type EditMessageTextConfig struct {
	BaseEdit
	Text               string
	ParseMode          string
	Entities           []MessageEntity
	LinkPreviewOptions *LinkPreviewOptions
}

func (config EditMessageTextConfig) params() (Params, error) {
	params, err := config.BaseEdit.params()
	if err != nil {
		return params, err
	}

	params["text"] = config.Text
	params.AddNonEmpty("parse_mode", config.ParseMode)
	if err = params.AddAny("link_preview_options", config.LinkPreviewOptions); err != nil {
		return params, err
	}
	err = params.AddAny("entities", config.Entities)

	return params, err
}

func (config EditMessageTextConfig) method() string {
	return "editMessageText"
}

// EditMessageCaptionConfig allows you to modify the caption of a message.
type EditMessageCaptionConfig struct {
	BaseEdit
	Caption               string
	ParseMode             string
	CaptionEntities       []MessageEntity
	ShowCaptionAboveMedia bool
}

func (config EditMessageCaptionConfig) params() (Params, error) {
	params, err := config.BaseEdit.params()
	if err != nil {
		return params, err
	}

	params["caption"] = config.Caption
	params.AddNonEmpty("parse_mode", config.ParseMode)
	params.AddBool("show_caption_above_media", config.ShowCaptionAboveMedia)
	err = params.AddAny("caption_entities", config.CaptionEntities)

	return params, err
}

func (config EditMessageCaptionConfig) method() string {
	return "editMessageCaption"
}

// EditMessageMediaConfig allows you to make an editMessageMedia request.
type EditMessageMediaConfig struct {
	BaseEdit

	Media interface{}
}

func (EditMessageMediaConfig) method() string {
	return "editMessageMedia"
}

func (config EditMessageMediaConfig) params() (Params, error) {
	params, err := config.BaseEdit.params()
	if err != nil {
		return params, err
	}

	err = params.AddInterface("media", prepareInputMediaParam(config.Media, 0))

	return params, err
}

func (config EditMessageMediaConfig) files() []RequestFile {
	return prepareInputMediaFile(config.Media, 0)
}

// EditMessageReplyMarkupConfig allows you to modify the reply markup
// of a message.
type EditMessageReplyMarkupConfig struct {
	BaseEdit
}

func (config EditMessageReplyMarkupConfig) params() (Params, error) {
	return config.BaseEdit.params()
}

func (config EditMessageReplyMarkupConfig) method() string {
	return "editMessageReplyMarkup"
}

// StopPollConfig allows you to stop a poll sent by the bot.
type StopPollConfig struct {
	BaseEdit
}

func (config StopPollConfig) params() (Params, error) {
	return config.BaseEdit.params()
}

func (StopPollConfig) method() string {
	return "stopPoll"
}

// UserProfilePhotosConfig contains information about a
// GetUserProfilePhotos request.
type UserProfilePhotosConfig struct {
	UserID int64
	Offset int
	Limit  int
}

func (UserProfilePhotosConfig) method() string {
	return "getUserProfilePhotos"
}

func (config UserProfilePhotosConfig) params() (Params, error) {
	params := make(Params)

	params.AddNonZero64("user_id", config.UserID)
	params.AddNonZero("offset", config.Offset)
	params.AddNonZero("limit", config.Limit)

	return params, nil
}

// FileConfig has information about a file hosted on Telegram.
type FileConfig struct {
	FileID string
}

func (FileConfig) method() string {
	return "getFile"
}

func (config FileConfig) params() (Params, error) {
	params := make(Params)

	params["file_id"] = config.FileID

	return params, nil
}

// UpdateConfig contains information about a GetUpdates request.
type UpdateConfig struct {
	Offset         int
	Limit          int
	Timeout        int
	AllowedUpdates []string
}

func (UpdateConfig) method() string {
	return "getUpdates"
}

func (config UpdateConfig) params() (Params, error) {
	params := make(Params)

	params.AddNonZero("offset", config.Offset)
	params.AddNonZero("limit", config.Limit)
	params.AddNonZero("timeout", config.Timeout)
	params.AddInterface("allowed_updates", config.AllowedUpdates)

	return params, nil
}

// WebhookConfig contains information about a SetWebhook request.
type WebhookConfig struct {
	URL                *url.URL
	Certificate        RequestFileData
	IPAddress          string
	MaxConnections     int
	AllowedUpdates     []string
	DropPendingUpdates bool
	SecretToken        string
}

func (config WebhookConfig) method() string {
	return "setWebhook"
}

func (config WebhookConfig) params() (Params, error) {
	params := make(Params)

	if config.URL != nil {
		params["url"] = config.URL.String()
	}

	params.AddNonEmpty("ip_address", config.IPAddress)
	params.AddNonZero("max_connections", config.MaxConnections)
	err := params.AddInterface("allowed_updates", config.AllowedUpdates)
	params.AddBool("drop_pending_updates", config.DropPendingUpdates)
	params.AddNonEmpty("secret_token", config.SecretToken)

	return params, err
}

func (config WebhookConfig) files() []RequestFile {
	if config.Certificate != nil {
		return []RequestFile{{
			Name: "certificate",
			Data: config.Certificate,
		}}
	}

	return nil
}

// DeleteWebhookConfig is a helper to delete a webhook.
type DeleteWebhookConfig struct {
	DropPendingUpdates bool
}

func (config DeleteWebhookConfig) method() string {
	return "deleteWebhook"
}

func (config DeleteWebhookConfig) params() (Params, error) {
	params := make(Params)

	params.AddBool("drop_pending_updates", config.DropPendingUpdates)

	return params, nil
}

// InlineConfig contains information on making an InlineQuery response.
type InlineConfig struct {
	InlineQueryID string                    `json:"inline_query_id"`
	Results       []interface{}             `json:"results"`
	CacheTime     int                       `json:"cache_time"`
	IsPersonal    bool                      `json:"is_personal"`
	NextOffset    string                    `json:"next_offset"`
	Button        *InlineQueryResultsButton `json:"button,omitempty"`
}

func (config InlineConfig) method() string {
	return "answerInlineQuery"
}

func (config InlineConfig) params() (Params, error) {
	params := make(Params)

	params["inline_query_id"] = config.InlineQueryID
	params.AddNonZero("cache_time", config.CacheTime)
	params.AddBool("is_personal", config.IsPersonal)
	params.AddNonEmpty("next_offset", config.NextOffset)
	if err := params.AddAny("button", config.Button); err != nil {
		return params, err
	}
	err := params.AddAny("results", config.Results)

	return params, err
}

// AnswerWebAppQueryConfig is used to set the result of an interaction with a
// Web App and send a corresponding message on behalf of the user to the chat
// from which the query originated.
type AnswerWebAppQueryConfig struct {
	// WebAppQueryID is the unique identifier for the query to be answered.
	WebAppQueryID string `json:"web_app_query_id"`
	// Result is an InlineQueryResult object describing the message to be sent.
	Result interface{} `json:"result"`
}

func (config AnswerWebAppQueryConfig) method() string {
	return "answerWebAppQuery"
}

func (config AnswerWebAppQueryConfig) params() (Params, error) {
	params := make(Params)

	params["web_app_query_id"] = config.WebAppQueryID
	err := params.AddInterface("result", config.Result)

	return params, err
}

// CallbackConfig contains information on making a CallbackQuery response.
type CallbackConfig struct {
	CallbackQueryID string `json:"callback_query_id"`
	Text            string `json:"text"`
	ShowAlert       bool   `json:"show_alert"`
	URL             string `json:"url"`
	CacheTime       int    `json:"cache_time"`
}

func (config CallbackConfig) method() string {
	return "answerCallbackQuery"
}

func (config CallbackConfig) params() (Params, error) {
	params := make(Params)

	params["callback_query_id"] = config.CallbackQueryID
	params.AddNonEmpty("text", config.Text)
	params.AddBool("show_alert", config.ShowAlert)
	params.AddNonEmpty("url", config.URL)
	params.AddNonZero("cache_time", config.CacheTime)

	return params, nil
}

// ChatMemberConfig contains information about a user in a chat for use
// with administrative functions such as kicking or unbanning a user.
type ChatMemberConfig struct {
	ChatID             int64
	SuperGroupUsername string
	ChannelUsername    string
	UserID             int64
}

// UnbanChatMemberConfig allows you to unban a user.
type UnbanChatMemberConfig struct {
	ChatMemberConfig
	OnlyIfBanned bool
}

func (config UnbanChatMemberConfig) method() string {
	return "unbanChatMember"
}

func (config UnbanChatMemberConfig) params() (Params, error) {
	params := make(Params)

	if err := params.AddFirstValid("chat_id", config.ChatID, config.SuperGroupUsername, config.ChannelUsername); err != nil {
		return params, err
	}
	params.AddNonZero64("user_id", config.UserID)
	params.AddBool("only_if_banned", config.OnlyIfBanned)

	return params, nil
}

// BanChatMemberConfig contains extra fields to kick user.
type BanChatMemberConfig struct {
	ChatMemberConfig
	UntilDate      int64
	RevokeMessages bool
}

func (config BanChatMemberConfig) method() string {
	return "banChatMember"
}

func (config BanChatMemberConfig) params() (Params, error) {
	params := make(Params)

	if err := params.AddFirstValid("chat_id", config.ChatID, config.SuperGroupUsername); err != nil {
		return params, err
	}
	params.AddNonZero64("user_id", config.UserID)
	params.AddNonZero64("until_date", config.UntilDate)
	params.AddBool("revoke_messages", config.RevokeMessages)

	return params, nil
}

// KickChatMemberConfig contains extra fields to ban user.
//
// This was renamed to BanChatMember in later versions of the Telegram Bot API.
type KickChatMemberConfig = BanChatMemberConfig

// RestrictChatMemberConfig contains fields to restrict members of chat
type RestrictChatMemberConfig struct {
	ChatMemberConfig
	UntilDate                     int64
	Permissions                   *ChatPermissions
	UseIndependentChatPermissions bool
}

func (config RestrictChatMemberConfig) method() string {
	return "restrictChatMember"
}

func (config RestrictChatMemberConfig) params() (Params, error) {
	params := make(Params)

	if err := params.AddFirstValid("chat_id", config.ChatID, config.SuperGroupUsername, config.ChannelUsername); err != nil {
		return params, err
	}
	params.AddNonZero64("user_id", config.UserID)

	err := params.AddInterface("permissions", config.Permissions)
	params.AddBool("use_independent_chat_permissions", config.UseIndependentChatPermissions)
	params.AddNonZero64("until_date", config.UntilDate)

	return params, err
}

// PromoteChatMemberConfig contains fields to promote members of chat
type PromoteChatMemberConfig struct {
	ChatMemberConfig
	IsAnonymous         bool
	CanManageChat       bool
	CanChangeInfo       bool
	CanPostMessages     bool
	CanEditMessages     bool
	CanDeleteMessages   bool
	CanManageVideoChats bool
	CanInviteUsers      bool
	CanRestrictMembers  bool
	CanPinMessages      bool
	CanPromoteMembers   bool
	CanPostStories      bool
	CanEditStories      bool
	CanDeleteStories    bool
	CanManageTopics     bool
}

func (config PromoteChatMemberConfig) method() string {
	return "promoteChatMember"
}

func (config PromoteChatMemberConfig) params() (Params, error) {
	params := make(Params)

	if err := params.AddFirstValid("chat_id", config.ChatID, config.SuperGroupUsername, config.ChannelUsername); err != nil {
		return params, err
	}
	params.AddNonZero64("user_id", config.UserID)

	params.AddBool("is_anonymous", config.IsAnonymous)
	params.AddBool("can_manage_chat", config.CanManageChat)
	params.AddBool("can_change_info", config.CanChangeInfo)
	params.AddBool("can_post_messages", config.CanPostMessages)
	params.AddBool("can_edit_messages", config.CanEditMessages)
	params.AddBool("can_delete_messages", config.CanDeleteMessages)
	params.AddBool("can_manage_video_chats", config.CanManageVideoChats)
	params.AddBool("can_invite_users", config.CanInviteUsers)
	params.AddBool("can_restrict_members", config.CanRestrictMembers)
	params.AddBool("can_pin_messages", config.CanPinMessages)
	params.AddBool("can_promote_members", config.CanPromoteMembers)
	params.AddBool("can_post_stories", config.CanPostStories)
	params.AddBool("can_edit_stories", config.CanEditStories)
	params.AddBool("can_delete_stories", config.CanDeleteStories)
	params.AddBool("can_manage_topics", config.CanManageTopics)

	return params, nil
}

// SetChatAdministratorCustomTitle sets the title of an administrative user
// promoted by the bot for a chat.
type SetChatAdministratorCustomTitle struct {
	ChatMemberConfig
	CustomTitle string
}

func (SetChatAdministratorCustomTitle) method() string {
	return "setChatAdministratorCustomTitle"
}

func (config SetChatAdministratorCustomTitle) params() (Params, error) {
	params := make(Params)

	if err := params.AddFirstValid("chat_id", config.ChatID, config.SuperGroupUsername, config.ChannelUsername); err != nil {
		return params, err
	}
	params.AddNonZero64("user_id", config.UserID)
	params.AddNonEmpty("custom_title", config.CustomTitle)

	return params, nil
}

// BanChatSenderChatConfig bans a channel chat in a supergroup or a channel. The
// owner of the chat will not be able to send messages and join live streams on
// behalf of the chat, unless it is unbanned first. The bot must be an
// administrator in the supergroup or channel for this to work and must have the
// appropriate administrator rights.
type BanChatSenderChatConfig struct {
	ChatID          int64
	ChannelUsername string
	SenderChatID    int64
	UntilDate       int
}

func (config BanChatSenderChatConfig) method() string {
	return "banChatSenderChat"
}

func (config BanChatSenderChatConfig) params() (Params, error) {
	params := make(Params)

	if err := params.AddFirstValid("chat_id", config.ChatID, config.ChannelUsername); err != nil {
		return params, err
	}
	params.AddNonZero64("sender_chat_id", config.SenderChatID)
	params.AddNonZero("until_date", config.UntilDate)

	return params, nil
}

// UnbanChatSenderChatConfig unbans a previously banned channel chat in a
// supergroup or channel. The bot must be an administrator for this to work and
// must have the appropriate administrator rights.
type UnbanChatSenderChatConfig struct {
	ChatID          int64
	ChannelUsername string
	SenderChatID    int64
}

func (config UnbanChatSenderChatConfig) method() string {
	return "unbanChatSenderChat"
}

func (config UnbanChatSenderChatConfig) params() (Params, error) {
	params := make(Params)

	if err := params.AddFirstValid("chat_id", config.ChatID, config.ChannelUsername); err != nil {
		return params, err
	}
	params.AddNonZero64("sender_chat_id", config.SenderChatID)

	return params, nil
}

// ChatConfig contains information about getting information on a chat.
type ChatConfig struct {
	ChatID             int64
	SuperGroupUsername string
}

func (config ChatConfig) params() (Params, error) {
	params := make(Params)

	if err := params.AddFirstValid("chat_id", config.ChatID, config.SuperGroupUsername); err != nil {
		return params, err
	}

	return params, nil
}

// ChatInfoConfig contains information about getting chat information.
type ChatInfoConfig struct {
	ChatConfig
}

func (ChatInfoConfig) method() string {
	return "getChat"
}

// ChatMemberCountConfig contains information about getting the number of users in a chat.
type ChatMemberCountConfig struct {
	ChatConfig
}

func (ChatMemberCountConfig) method() string {
	return "getChatMembersCount"
}

// ChatAdministratorsConfig contains information about getting chat administrators.
type ChatAdministratorsConfig struct {
	ChatConfig
}

func (ChatAdministratorsConfig) method() string {
	return "getChatAdministrators"
}

// SetChatPermissionsConfig allows you to set default permissions for the
// members in a group. The bot must be an administrator and have rights to
// restrict members.
type SetChatPermissionsConfig struct {
	ChatConfig
	Permissions                   *ChatPermissions
	UseIndependentChatPermissions bool
}

func (SetChatPermissionsConfig) method() string {
	return "setChatPermissions"
}

func (config SetChatPermissionsConfig) params() (Params, error) {
	params := make(Params)

	if err := params.AddFirstValid("chat_id", config.ChatID, config.SuperGroupUsername); err != nil {
		return params, err
	}
	err := params.AddInterface("permissions", config.Permissions)
	params.AddBool("use_independent_chat_permissions", config.UseIndependentChatPermissions)

	return params, err
}

// ChatInviteLinkConfig contains information about getting a chat link.
//
// Note that generating a new link will revoke any previous links.
type ChatInviteLinkConfig struct {
	ChatConfig
}

func (ChatInviteLinkConfig) method() string {
	return "exportChatInviteLink"
}

func (config ChatInviteLinkConfig) params() (Params, error) {
	params := make(Params)

	if err := params.AddFirstValid("chat_id", config.ChatID, config.SuperGroupUsername); err != nil {
		return params, err
	}

	return params, nil
}

// CreateChatInviteLinkConfig allows you to create an additional invite link for
// a chat. The bot must be an administrator in the chat for this to work and
// must have the appropriate admin rights. The link can be revoked using the
// RevokeChatInviteLinkConfig.
type CreateChatInviteLinkConfig struct {
	ChatConfig
	Name               string
	ExpireDate         int
	MemberLimit        int
	CreatesJoinRequest bool
}

func (CreateChatInviteLinkConfig) method() string {
	return "createChatInviteLink"
}

func (config CreateChatInviteLinkConfig) params() (Params, error) {
	params := make(Params)

	params.AddNonEmpty("name", config.Name)
	if err := params.AddFirstValid("chat_id", config.ChatID, config.SuperGroupUsername); err != nil {
		return params, err
	}
	params.AddNonZero("expire_date", config.ExpireDate)
	params.AddNonZero("member_limit", config.MemberLimit)
	params.AddBool("creates_join_request", config.CreatesJoinRequest)

	return params, nil
}

// EditChatInviteLinkConfig allows you to edit a non-primary invite link created
// by the bot. The bot must be an administrator in the chat for this to work and
// must have the appropriate admin rights.
type EditChatInviteLinkConfig struct {
	ChatConfig
	InviteLink         string
	Name               string
	ExpireDate         int
	MemberLimit        int
	CreatesJoinRequest bool
}

func (EditChatInviteLinkConfig) method() string {
	return "editChatInviteLink"
}

func (config EditChatInviteLinkConfig) params() (Params, error) {
	params := make(Params)

	if err := params.AddFirstValid("chat_id", config.ChatID, config.SuperGroupUsername); err != nil {
		return params, err
	}
	params.AddNonEmpty("name", config.Name)
	params["invite_link"] = config.InviteLink
	params.AddNonZero("expire_date", config.ExpireDate)
	params.AddNonZero("member_limit", config.MemberLimit)
	params.AddBool("creates_join_request", config.CreatesJoinRequest)

	return params, nil
}

// CreateChatSubscriptionInviteLinkConfig creates a subscription invite link
// for a channel chat. The bot must have the can_invite_users administrator
// rights. The link can be edited using EditChatSubscriptionInviteLinkConfig
// or revoked using RevokeChatInviteLinkConfig.
type CreateChatSubscriptionInviteLinkConfig struct {
	ChatConfig
	Name               string
	SubscriptionPeriod int // required, in seconds; currently must be 2592000 (30 days)
	SubscriptionPrice  int // required, 1-10000 Telegram Stars
}

func (CreateChatSubscriptionInviteLinkConfig) method() string {
	return "createChatSubscriptionInviteLink"
}

func (config CreateChatSubscriptionInviteLinkConfig) params() (Params, error) {
	params := make(Params)

	if err := params.AddFirstValid("chat_id", config.ChatID, config.SuperGroupUsername); err != nil {
		return params, err
	}
	params.AddNonEmpty("name", config.Name)
	params.AddNonZero("subscription_period", config.SubscriptionPeriod)
	params.AddNonZero("subscription_price", config.SubscriptionPrice)

	return params, nil
}

// EditChatSubscriptionInviteLinkConfig edits a subscription invite link
// created by the bot. Only the Name field can be edited.
type EditChatSubscriptionInviteLinkConfig struct {
	ChatConfig
	InviteLink string
	Name       string
}

func (EditChatSubscriptionInviteLinkConfig) method() string {
	return "editChatSubscriptionInviteLink"
}

func (config EditChatSubscriptionInviteLinkConfig) params() (Params, error) {
	params := make(Params)

	if err := params.AddFirstValid("chat_id", config.ChatID, config.SuperGroupUsername); err != nil {
		return params, err
	}
	params["invite_link"] = config.InviteLink
	params.AddNonEmpty("name", config.Name)

	return params, nil
}

// RevokeChatInviteLinkConfig allows you to revoke an invite link created by the
// bot. If the primary link is revoked, a new link is automatically generated.
// The bot must be an administrator in the chat for this to work and must have
// the appropriate admin rights.
type RevokeChatInviteLinkConfig struct {
	ChatConfig
	InviteLink string
}

func (RevokeChatInviteLinkConfig) method() string {
	return "revokeChatInviteLink"
}

func (config RevokeChatInviteLinkConfig) params() (Params, error) {
	params := make(Params)

	if err := params.AddFirstValid("chat_id", config.ChatID, config.SuperGroupUsername); err != nil {
		return params, err
	}
	params["invite_link"] = config.InviteLink

	return params, nil
}

// ApproveChatJoinRequestConfig allows you to approve a chat join request.
type ApproveChatJoinRequestConfig struct {
	ChatConfig
	UserID int64
}

func (ApproveChatJoinRequestConfig) method() string {
	return "approveChatJoinRequest"
}

func (config ApproveChatJoinRequestConfig) params() (Params, error) {
	params := make(Params)

	if err := params.AddFirstValid("chat_id", config.ChatID, config.SuperGroupUsername); err != nil {
		return params, err
	}
	params.AddNonZero("user_id", int(config.UserID))

	return params, nil
}

// DeclineChatJoinRequest allows you to decline a chat join request.
type DeclineChatJoinRequest struct {
	ChatConfig
	UserID int64
}

func (DeclineChatJoinRequest) method() string {
	return "declineChatJoinRequest"
}

func (config DeclineChatJoinRequest) params() (Params, error) {
	params := make(Params)

	if err := params.AddFirstValid("chat_id", config.ChatID, config.SuperGroupUsername); err != nil {
		return params, err
	}
	params.AddNonZero("user_id", int(config.UserID))

	return params, nil
}

// LeaveChatConfig allows you to leave a chat.
type LeaveChatConfig struct {
	ChatID          int64
	ChannelUsername string
}

func (config LeaveChatConfig) method() string {
	return "leaveChat"
}

func (config LeaveChatConfig) params() (Params, error) {
	params := make(Params)

	if err := params.AddFirstValid("chat_id", config.ChatID, config.ChannelUsername); err != nil {
		return params, err
	}

	return params, nil
}

// ChatConfigWithUser contains information about a chat and a user.
type ChatConfigWithUser struct {
	ChatID             int64
	SuperGroupUsername string
	UserID             int64
}

func (config ChatConfigWithUser) params() (Params, error) {
	params := make(Params)

	if err := params.AddFirstValid("chat_id", config.ChatID, config.SuperGroupUsername); err != nil {
		return params, err
	}
	params.AddNonZero64("user_id", config.UserID)

	return params, nil
}

// GetChatMemberConfig is information about getting a specific member in a chat.
type GetChatMemberConfig struct {
	ChatConfigWithUser
}

func (GetChatMemberConfig) method() string {
	return "getChatMember"
}

// InvoiceConfig contains information for sendInvoice request.
type InvoiceConfig struct {
	BaseChat
	Title                     string         // required
	Description               string         // required
	Payload                   string         // required
	ProviderToken             string         // omit for payments in Telegram Stars
	Currency                  string         // required ("XTR" for Telegram Stars)
	Prices                    []LabeledPrice // required
	MaxTipAmount              int
	SuggestedTipAmounts       []int
	StartParameter            string
	ProviderData              json.RawMessage
	PhotoURL                  string
	PhotoSize                 int
	PhotoWidth                int
	PhotoHeight               int
	NeedName                  bool
	NeedPhoneNumber           bool
	NeedEmail                 bool
	NeedShippingAddress       bool
	SendPhoneNumberToProvider bool
	SendEmailToProvider       bool
	IsFlexible                bool
}

func (config InvoiceConfig) params() (Params, error) {
	params, err := config.BaseChat.params()
	if err != nil {
		return params, err
	}

	params["title"] = config.Title
	params["description"] = config.Description
	params["payload"] = config.Payload
	params.AddNonEmpty("provider_token", config.ProviderToken)
	params["currency"] = config.Currency
	if err = params.AddAny("prices", config.Prices); err != nil {
		return params, err
	}

	params.AddNonZero("max_tip_amount", config.MaxTipAmount)
	err = params.AddAny("suggested_tip_amounts", config.SuggestedTipAmounts)
	params.AddNonEmpty("start_parameter", config.StartParameter)
	if len(config.ProviderData) > 0 {
		params["provider_data"] = string(config.ProviderData)
	}
	params.AddNonEmpty("photo_url", config.PhotoURL)
	params.AddNonZero("photo_size", config.PhotoSize)
	params.AddNonZero("photo_width", config.PhotoWidth)
	params.AddNonZero("photo_height", config.PhotoHeight)
	params.AddBool("need_name", config.NeedName)
	params.AddBool("need_phone_number", config.NeedPhoneNumber)
	params.AddBool("need_email", config.NeedEmail)
	params.AddBool("need_shipping_address", config.NeedShippingAddress)
	params.AddBool("is_flexible", config.IsFlexible)
	params.AddBool("send_phone_number_to_provider", config.SendPhoneNumberToProvider)
	params.AddBool("send_email_to_provider", config.SendEmailToProvider)

	return params, err
}

func (config InvoiceConfig) method() string {
	return "sendInvoice"
}

// InvoiceLinkConfig contains information for createInvoiceLink request.
type InvoiceLinkConfig struct {
	BusinessConnectionID      string
	Title                     string         // required
	Description               string         // required
	Payload                   string         // required
	ProviderToken             string         // omit for payments in Telegram Stars
	Currency                  string         // required ("XTR" for Telegram Stars)
	Prices                    []LabeledPrice // required
	SubscriptionPeriod        int            // in seconds; currently must be 2592000 (30 days) if set
	MaxTipAmount              int
	SuggestedTipAmounts       []int
	ProviderData              json.RawMessage
	PhotoURL                  string
	PhotoSize                 int
	PhotoWidth                int
	PhotoHeight               int
	NeedName                  bool
	NeedPhoneNumber           bool
	NeedEmail                 bool
	NeedShippingAddress       bool
	SendPhoneNumberToProvider bool
	SendEmailToProvider       bool
	IsFlexible                bool
}

func (config InvoiceLinkConfig) params() (Params, error) {
	params := make(Params)

	params.AddNonEmpty("business_connection_id", config.BusinessConnectionID)
	params["title"] = config.Title
	params["description"] = config.Description
	params["payload"] = config.Payload
	params.AddNonEmpty("provider_token", config.ProviderToken)
	params["currency"] = config.Currency
	params.AddNonZero("subscription_period", config.SubscriptionPeriod)
	if err := params.AddAny("prices", config.Prices); err != nil {
		return params, err
	}

	params.AddNonZero("max_tip_amount", config.MaxTipAmount)
	if err := params.AddAny("suggested_tip_amounts", config.SuggestedTipAmounts); err != nil {
		return params, err
	}
	if len(config.ProviderData) > 0 {
		params["provider_data"] = string(config.ProviderData)
	}
	params.AddNonEmpty("photo_url", config.PhotoURL)
	params.AddNonZero("photo_size", config.PhotoSize)
	params.AddNonZero("photo_width", config.PhotoWidth)
	params.AddNonZero("photo_height", config.PhotoHeight)
	params.AddBool("need_name", config.NeedName)
	params.AddBool("need_phone_number", config.NeedPhoneNumber)
	params.AddBool("need_email", config.NeedEmail)
	params.AddBool("need_shipping_address", config.NeedShippingAddress)
	params.AddBool("send_phone_number_to_provider", config.SendPhoneNumberToProvider)
	params.AddBool("send_email_to_provider", config.SendEmailToProvider)
	params.AddBool("is_flexible", config.IsFlexible)

	return params, nil
}

func (config InvoiceLinkConfig) method() string {
	return "createInvoiceLink"
}

// GetAvailableGiftsConfig returns the list of gifts that can be sent by the
// bot to users.
type GetAvailableGiftsConfig struct{}

func (GetAvailableGiftsConfig) method() string {
	return "getAvailableGifts"
}

func (GetAvailableGiftsConfig) params() (Params, error) {
	return make(Params), nil
}

// SendGiftConfig sends a gift to a user.
type SendGiftConfig struct {
	UserID        int64
	GiftID        string
	Text          string
	TextParseMode string
	TextEntities  []MessageEntity
}

func (SendGiftConfig) method() string {
	return "sendGift"
}

func (config SendGiftConfig) params() (Params, error) {
	params := make(Params)

	params.AddNonZero64("user_id", config.UserID)
	params["gift_id"] = config.GiftID
	params.AddNonEmpty("text", config.Text)
	params.AddNonEmpty("text_parse_mode", config.TextParseMode)
	err := params.AddAny("text_entities", config.TextEntities)

	return params, err
}

// SetUserEmojiStatusConfig changes the emoji status for a given user that
// previously allowed the bot to manage their emoji status.
type SetUserEmojiStatusConfig struct {
	UserID                    int64
	EmojiStatusCustomEmojiID  string
	EmojiStatusExpirationDate int
}

func (SetUserEmojiStatusConfig) method() string {
	return "setUserEmojiStatus"
}

func (config SetUserEmojiStatusConfig) params() (Params, error) {
	params := make(Params)

	params.AddNonZero64("user_id", config.UserID)
	params.AddNonEmpty("emoji_status_custom_emoji_id", config.EmojiStatusCustomEmojiID)
	params.AddNonZero("emoji_status_expiration_date", config.EmojiStatusExpirationDate)

	return params, nil
}

// EditUserStarSubscriptionConfig cancels or reenables an active Telegram
// Star subscription.
type EditUserStarSubscriptionConfig struct {
	UserID                  int64
	TelegramPaymentChargeID string
	IsCanceled              bool
}

func (EditUserStarSubscriptionConfig) method() string {
	return "editUserStarSubscription"
}

func (config EditUserStarSubscriptionConfig) params() (Params, error) {
	params := make(Params)

	params.AddNonZero64("user_id", config.UserID)
	params["telegram_payment_charge_id"] = config.TelegramPaymentChargeID
	params.AddBool("is_canceled", config.IsCanceled)

	return params, nil
}

// SavePreparedInlineMessageConfig stores a message that can be sent by a
// user of a Mini App.
type SavePreparedInlineMessageConfig struct {
	UserID            int64
	Result            interface{} // InlineQueryResult
	AllowUserChats    bool
	AllowBotChats     bool
	AllowGroupChats   bool
	AllowChannelChats bool
}

func (SavePreparedInlineMessageConfig) method() string {
	return "savePreparedInlineMessage"
}

func (config SavePreparedInlineMessageConfig) params() (Params, error) {
	params := make(Params)

	params.AddNonZero64("user_id", config.UserID)
	if err := params.AddAny("result", config.Result); err != nil {
		return params, err
	}
	params.AddBool("allow_user_chats", config.AllowUserChats)
	params.AddBool("allow_bot_chats", config.AllowBotChats)
	params.AddBool("allow_group_chats", config.AllowGroupChats)
	params.AddBool("allow_channel_chats", config.AllowChannelChats)

	return params, nil
}

// GetStarTransactionsConfig returns the bot's Telegram Star transactions in
// chronological order.
type GetStarTransactionsConfig struct {
	// Offset is the number of transactions to skip in the response.
	Offset int
	// Limit is the maximum number of transactions to be retrieved; 1-100.
	// Defaults to 100.
	Limit int
}

func (config GetStarTransactionsConfig) method() string {
	return "getStarTransactions"
}

func (config GetStarTransactionsConfig) params() (Params, error) {
	params := make(Params)

	params.AddNonZero("offset", config.Offset)
	params.AddNonZero("limit", config.Limit)

	return params, nil
}

// RefundStarPaymentConfig refunds a successful payment in Telegram Stars.
type RefundStarPaymentConfig struct {
	UserID                  int64
	TelegramPaymentChargeID string
}

func (config RefundStarPaymentConfig) method() string {
	return "refundStarPayment"
}

func (config RefundStarPaymentConfig) params() (Params, error) {
	params := make(Params)

	params.AddNonZero64("user_id", config.UserID)
	params["telegram_payment_charge_id"] = config.TelegramPaymentChargeID

	return params, nil
}

// ShippingConfig contains information for answerShippingQuery request.
type ShippingConfig struct {
	ShippingQueryID string // required
	OK              bool   // required
	ShippingOptions []ShippingOption
	ErrorMessage    string
}

func (config ShippingConfig) method() string {
	return "answerShippingQuery"
}

func (config ShippingConfig) params() (Params, error) {
	params := make(Params)

	params["shipping_query_id"] = config.ShippingQueryID
	params.AddBool("ok", config.OK)
	err := params.AddInterface("shipping_options", config.ShippingOptions)
	params.AddNonEmpty("error_message", config.ErrorMessage)

	return params, err
}

// PreCheckoutConfig contains information for answerPreCheckoutQuery request.
type PreCheckoutConfig struct {
	PreCheckoutQueryID string // required
	OK                 bool   // required
	ErrorMessage       string
}

func (config PreCheckoutConfig) method() string {
	return "answerPreCheckoutQuery"
}

func (config PreCheckoutConfig) params() (Params, error) {
	params := make(Params)

	params["pre_checkout_query_id"] = config.PreCheckoutQueryID
	params.AddBool("ok", config.OK)
	params.AddNonEmpty("error_message", config.ErrorMessage)

	return params, nil
}

// DeleteMessageConfig contains information of a message in a chat to delete.
type DeleteMessageConfig struct {
	ChannelUsername string
	ChatID          int64
	MessageID       int
}

func (config DeleteMessageConfig) method() string {
	return "deleteMessage"
}

func (config DeleteMessageConfig) params() (Params, error) {
	params := make(Params)

	if err := params.AddFirstValid("chat_id", config.ChatID, config.ChannelUsername); err != nil {
		return params, err
	}
	params.AddNonZero("message_id", config.MessageID)

	return params, nil
}

// DeleteMessagesConfig deletes multiple messages simultaneously. If some of
// the specified messages can't be found, they are skipped.
type DeleteMessagesConfig struct {
	ChatID          int64
	ChannelUsername string
	MessageIDs      []int // 1-100 message identifiers
}

func (config DeleteMessagesConfig) method() string {
	return "deleteMessages"
}

func (config DeleteMessagesConfig) params() (Params, error) {
	params := make(Params)

	if err := params.AddFirstValid("chat_id", config.ChatID, config.ChannelUsername); err != nil {
		return params, err
	}
	err := params.AddAny("message_ids", config.MessageIDs)

	return params, err
}

// ForwardMessagesConfig forwards multiple messages of any kind. If some of the
// specified messages can't be found or forwarded, they are skipped. Service
// messages and messages with protected content can't be forwarded. Album
// grouping is kept for forwarded messages.
type ForwardMessagesConfig struct {
	ChatID              int64
	ChannelUsername     string
	MessageThreadID     int
	FromChatID          int64
	FromChannelUsername string
	MessageIDs          []int // 1-100 message identifiers, must be in strictly increasing order
	DisableNotification bool
	ProtectContent      bool
}

func (config ForwardMessagesConfig) method() string {
	return "forwardMessages"
}

func (config ForwardMessagesConfig) params() (Params, error) {
	params := make(Params)

	if err := params.AddFirstValid("chat_id", config.ChatID, config.ChannelUsername); err != nil {
		return params, err
	}
	if err := params.AddFirstValid("from_chat_id", config.FromChatID, config.FromChannelUsername); err != nil {
		return params, err
	}
	params.AddNonZero("message_thread_id", config.MessageThreadID)
	params.AddBool("disable_notification", config.DisableNotification)
	params.AddBool("protect_content", config.ProtectContent)
	err := params.AddAny("message_ids", config.MessageIDs)

	return params, err
}

// CopyMessagesConfig copies messages of any kind. If some of the specified
// messages can't be found or copied, they are skipped. Service messages,
// giveaway messages, giveaway winners messages, and invoice messages can't
// be copied. Album grouping is kept for copied messages.
type CopyMessagesConfig struct {
	ChatID              int64
	ChannelUsername     string
	MessageThreadID     int
	FromChatID          int64
	FromChannelUsername string
	MessageIDs          []int // 1-100 message identifiers, must be in strictly increasing order
	DisableNotification bool
	ProtectContent      bool
	RemoveCaption       bool
}

func (config CopyMessagesConfig) method() string {
	return "copyMessages"
}

func (config CopyMessagesConfig) params() (Params, error) {
	params := make(Params)

	if err := params.AddFirstValid("chat_id", config.ChatID, config.ChannelUsername); err != nil {
		return params, err
	}
	if err := params.AddFirstValid("from_chat_id", config.FromChatID, config.FromChannelUsername); err != nil {
		return params, err
	}
	params.AddNonZero("message_thread_id", config.MessageThreadID)
	params.AddBool("disable_notification", config.DisableNotification)
	params.AddBool("protect_content", config.ProtectContent)
	params.AddBool("remove_caption", config.RemoveCaption)
	err := params.AddAny("message_ids", config.MessageIDs)

	return params, err
}

// SetMessageReactionConfig sets the bot's reaction to a message. Pass an
// empty Reaction to remove all reactions from the message.
type SetMessageReactionConfig struct {
	ChatID          int64
	ChannelUsername string
	MessageID       int
	Reaction        []ReactionType
	IsBig           bool
}

func (config SetMessageReactionConfig) method() string {
	return "setMessageReaction"
}

func (config SetMessageReactionConfig) params() (Params, error) {
	params := make(Params)

	if err := params.AddFirstValid("chat_id", config.ChatID, config.ChannelUsername); err != nil {
		return params, err
	}
	params.AddNonZero("message_id", config.MessageID)
	params.AddBool("is_big", config.IsBig)

	err := params.AddAny("reaction", config.Reaction)

	return params, err
}

// GetUserChatBoostsConfig returns the list of boosts added to a chat by a
// user. The bot must be an administrator in the chat.
//
// Provide the target chat via either ChatID (numeric identifier) or
// ChannelUsername ("@channelusername"); the first non-zero / non-empty
// value is used.
type GetUserChatBoostsConfig struct {
	ChatID          int64
	ChannelUsername string
	UserID          int64
}

func (config GetUserChatBoostsConfig) method() string {
	return "getUserChatBoosts"
}

func (config GetUserChatBoostsConfig) params() (Params, error) {
	params := make(Params)

	if err := params.AddFirstValid("chat_id", config.ChatID, config.ChannelUsername); err != nil {
		return params, err
	}
	params.AddNonZero64("user_id", config.UserID)

	return params, nil
}

// PinChatMessageConfig contains information of a message in a chat to pin.
//
// Provide the target chat via either ChatID (numeric identifier) or
// ChannelUsername ("@channelusername"); the first non-zero / non-empty
// value is used.
type PinChatMessageConfig struct {
	BusinessConnectionID string
	ChatID               int64
	ChannelUsername      string
	MessageID            int
	DisableNotification  bool
}

func (config PinChatMessageConfig) method() string {
	return "pinChatMessage"
}

func (config PinChatMessageConfig) params() (Params, error) {
	params := make(Params)

	params.AddNonEmpty("business_connection_id", config.BusinessConnectionID)
	if err := params.AddFirstValid("chat_id", config.ChatID, config.ChannelUsername); err != nil {
		return params, err
	}
	params.AddNonZero("message_id", config.MessageID)
	params.AddBool("disable_notification", config.DisableNotification)

	return params, nil
}

// UnpinChatMessageConfig contains information of a chat message to unpin.
//
// If MessageID is not specified, it will unpin the most recent pin.
//
// Provide the target chat via either ChatID (numeric identifier) or
// ChannelUsername ("@channelusername"); the first non-zero / non-empty
// value is used.
type UnpinChatMessageConfig struct {
	BusinessConnectionID string
	ChatID               int64
	ChannelUsername      string
	MessageID            int
}

func (config UnpinChatMessageConfig) method() string {
	return "unpinChatMessage"
}

func (config UnpinChatMessageConfig) params() (Params, error) {
	params := make(Params)

	params.AddNonEmpty("business_connection_id", config.BusinessConnectionID)
	if err := params.AddFirstValid("chat_id", config.ChatID, config.ChannelUsername); err != nil {
		return params, err
	}
	params.AddNonZero("message_id", config.MessageID)

	return params, nil
}

// UnpinAllChatMessagesConfig contains information of all messages to unpin in
// a chat.
type UnpinAllChatMessagesConfig struct {
	ChatID          int64
	ChannelUsername string
}

func (config UnpinAllChatMessagesConfig) method() string {
	return "unpinAllChatMessages"
}

func (config UnpinAllChatMessagesConfig) params() (Params, error) {
	params := make(Params)

	if err := params.AddFirstValid("chat_id", config.ChatID, config.ChannelUsername); err != nil {
		return params, err
	}

	return params, nil
}

// CreateForumTopicConfig creates a topic in a forum supergroup chat.
// The bot must have the can_manage_topics administrator rights.
// Returns information about the created topic as a ForumTopic object.
type CreateForumTopicConfig struct {
	ChatID             int64
	SuperGroupUsername string
	Name               string // required, 1-128 characters
	IconColor          int
	IconCustomEmojiID  string
}

func (config CreateForumTopicConfig) method() string {
	return "createForumTopic"
}

func (config CreateForumTopicConfig) params() (Params, error) {
	params := make(Params)

	if err := params.AddFirstValid("chat_id", config.ChatID, config.SuperGroupUsername); err != nil {
		return params, err
	}
	params["name"] = config.Name
	params.AddNonZero("icon_color", config.IconColor)
	params.AddNonEmpty("icon_custom_emoji_id", config.IconCustomEmojiID)

	return params, nil
}

// EditForumTopicConfig edits the name and icon of a topic in a forum
// supergroup chat. The bot must have the can_manage_topics administrator
// rights, unless it is the creator of the topic.
type EditForumTopicConfig struct {
	ChatID             int64
	SuperGroupUsername string
	MessageThreadID    int // required
	Name               string
	IconCustomEmojiID  string
}

func (config EditForumTopicConfig) method() string {
	return "editForumTopic"
}

func (config EditForumTopicConfig) params() (Params, error) {
	params := make(Params)

	if err := params.AddFirstValid("chat_id", config.ChatID, config.SuperGroupUsername); err != nil {
		return params, err
	}
	params.AddNonZero("message_thread_id", config.MessageThreadID)
	params.AddNonEmpty("name", config.Name)
	params.AddNonEmpty("icon_custom_emoji_id", config.IconCustomEmojiID)

	return params, nil
}

// CloseForumTopicConfig closes an open topic in a forum supergroup chat.
// The bot must have the can_manage_topics administrator rights, unless it
// is the creator of the topic.
type CloseForumTopicConfig struct {
	ChatID             int64
	SuperGroupUsername string
	MessageThreadID    int // required
}

func (config CloseForumTopicConfig) method() string {
	return "closeForumTopic"
}

func (config CloseForumTopicConfig) params() (Params, error) {
	params := make(Params)

	if err := params.AddFirstValid("chat_id", config.ChatID, config.SuperGroupUsername); err != nil {
		return params, err
	}
	params.AddNonZero("message_thread_id", config.MessageThreadID)

	return params, nil
}

// ReopenForumTopicConfig reopens a closed topic in a forum supergroup chat.
// The bot must have the can_manage_topics administrator rights, unless it
// is the creator of the topic.
type ReopenForumTopicConfig struct {
	ChatID             int64
	SuperGroupUsername string
	MessageThreadID    int // required
}

func (config ReopenForumTopicConfig) method() string {
	return "reopenForumTopic"
}

func (config ReopenForumTopicConfig) params() (Params, error) {
	params := make(Params)

	if err := params.AddFirstValid("chat_id", config.ChatID, config.SuperGroupUsername); err != nil {
		return params, err
	}
	params.AddNonZero("message_thread_id", config.MessageThreadID)

	return params, nil
}

// DeleteForumTopicConfig deletes a forum topic along with all its messages
// in a forum supergroup chat. The bot must have the can_delete_messages
// administrator rights.
type DeleteForumTopicConfig struct {
	ChatID             int64
	SuperGroupUsername string
	MessageThreadID    int // required
}

func (config DeleteForumTopicConfig) method() string {
	return "deleteForumTopic"
}

func (config DeleteForumTopicConfig) params() (Params, error) {
	params := make(Params)

	if err := params.AddFirstValid("chat_id", config.ChatID, config.SuperGroupUsername); err != nil {
		return params, err
	}
	params.AddNonZero("message_thread_id", config.MessageThreadID)

	return params, nil
}

// UnpinAllForumTopicMessagesConfig clears the list of pinned messages in a
// forum topic. The bot must have the can_pin_messages administrator right
// in the supergroup.
type UnpinAllForumTopicMessagesConfig struct {
	ChatID             int64
	SuperGroupUsername string
	MessageThreadID    int // required
}

func (config UnpinAllForumTopicMessagesConfig) method() string {
	return "unpinAllForumTopicMessages"
}

func (config UnpinAllForumTopicMessagesConfig) params() (Params, error) {
	params := make(Params)

	if err := params.AddFirstValid("chat_id", config.ChatID, config.SuperGroupUsername); err != nil {
		return params, err
	}
	params.AddNonZero("message_thread_id", config.MessageThreadID)

	return params, nil
}

// GetForumTopicIconStickersConfig gets custom emoji stickers, which can be
// used as a forum topic icon by any user. Requires no parameters.
// Returns an Array of Sticker objects.
type GetForumTopicIconStickersConfig struct{}

func (config GetForumTopicIconStickersConfig) method() string {
	return "getForumTopicIconStickers"
}

func (config GetForumTopicIconStickersConfig) params() (Params, error) {
	return make(Params), nil
}

// EditGeneralForumTopicConfig edits the name of the 'General' topic in a
// forum supergroup chat. The bot must have the can_manage_topics
// administrator rights.
type EditGeneralForumTopicConfig struct {
	ChatID             int64
	SuperGroupUsername string
	Name               string // required, 1-128 characters
}

func (config EditGeneralForumTopicConfig) method() string {
	return "editGeneralForumTopic"
}

func (config EditGeneralForumTopicConfig) params() (Params, error) {
	params := make(Params)

	if err := params.AddFirstValid("chat_id", config.ChatID, config.SuperGroupUsername); err != nil {
		return params, err
	}
	params["name"] = config.Name

	return params, nil
}

// CloseGeneralForumTopicConfig closes an open 'General' topic in a forum
// supergroup chat. The bot must have the can_manage_topics administrator rights.
type CloseGeneralForumTopicConfig struct {
	ChatID             int64
	SuperGroupUsername string
}

func (config CloseGeneralForumTopicConfig) method() string {
	return "closeGeneralForumTopic"
}

func (config CloseGeneralForumTopicConfig) params() (Params, error) {
	params := make(Params)

	if err := params.AddFirstValid("chat_id", config.ChatID, config.SuperGroupUsername); err != nil {
		return params, err
	}

	return params, nil
}

// ReopenGeneralForumTopicConfig reopens a closed 'General' topic in a forum
// supergroup chat. The bot must have the can_manage_topics administrator
// rights. The topic will be automatically unhidden if it was hidden.
type ReopenGeneralForumTopicConfig struct {
	ChatID             int64
	SuperGroupUsername string
}

func (config ReopenGeneralForumTopicConfig) method() string {
	return "reopenGeneralForumTopic"
}

func (config ReopenGeneralForumTopicConfig) params() (Params, error) {
	params := make(Params)

	if err := params.AddFirstValid("chat_id", config.ChatID, config.SuperGroupUsername); err != nil {
		return params, err
	}

	return params, nil
}

// HideGeneralForumTopicConfig hides the 'General' topic in a forum supergroup
// chat. The bot must have the can_manage_topics administrator rights. The
// topic will be automatically closed if it was open.
type HideGeneralForumTopicConfig struct {
	ChatID             int64
	SuperGroupUsername string
}

func (config HideGeneralForumTopicConfig) method() string {
	return "hideGeneralForumTopic"
}

func (config HideGeneralForumTopicConfig) params() (Params, error) {
	params := make(Params)

	if err := params.AddFirstValid("chat_id", config.ChatID, config.SuperGroupUsername); err != nil {
		return params, err
	}

	return params, nil
}

// UnhideGeneralForumTopicConfig unhides the 'General' topic in a forum
// supergroup chat. The bot must have the can_manage_topics administrator rights.
type UnhideGeneralForumTopicConfig struct {
	ChatID             int64
	SuperGroupUsername string
}

func (config UnhideGeneralForumTopicConfig) method() string {
	return "unhideGeneralForumTopic"
}

func (config UnhideGeneralForumTopicConfig) params() (Params, error) {
	params := make(Params)

	if err := params.AddFirstValid("chat_id", config.ChatID, config.SuperGroupUsername); err != nil {
		return params, err
	}

	return params, nil
}

// UnpinAllGeneralForumTopicMessagesConfig clears the list of pinned messages
// in the General forum topic. The bot must be an administrator in the chat
// with the can_pin_messages administrator right in the supergroup.
type UnpinAllGeneralForumTopicMessagesConfig struct {
	ChatID             int64
	SuperGroupUsername string
}

func (config UnpinAllGeneralForumTopicMessagesConfig) method() string {
	return "unpinAllGeneralForumTopicMessages"
}

func (config UnpinAllGeneralForumTopicMessagesConfig) params() (Params, error) {
	params := make(Params)

	if err := params.AddFirstValid("chat_id", config.ChatID, config.SuperGroupUsername); err != nil {
		return params, err
	}

	return params, nil
}

// SetChatPhotoConfig allows you to set a group, supergroup, or channel's photo.
type SetChatPhotoConfig struct {
	BaseFile
}

func (config SetChatPhotoConfig) method() string {
	return "setChatPhoto"
}

func (config SetChatPhotoConfig) files() []RequestFile {
	return []RequestFile{{
		Name: "photo",
		Data: config.File,
	}}
}

// DeleteChatPhotoConfig allows you to delete a group, supergroup, or channel's photo.
type DeleteChatPhotoConfig struct {
	ChatID          int64
	ChannelUsername string
}

func (config DeleteChatPhotoConfig) method() string {
	return "deleteChatPhoto"
}

func (config DeleteChatPhotoConfig) params() (Params, error) {
	params := make(Params)

	if err := params.AddFirstValid("chat_id", config.ChatID, config.ChannelUsername); err != nil {
		return params, err
	}

	return params, nil
}

// SetChatTitleConfig allows you to set the title of something other than a private chat.
type SetChatTitleConfig struct {
	ChatID          int64
	ChannelUsername string

	Title string
}

func (config SetChatTitleConfig) method() string {
	return "setChatTitle"
}

func (config SetChatTitleConfig) params() (Params, error) {
	params := make(Params)

	if err := params.AddFirstValid("chat_id", config.ChatID, config.ChannelUsername); err != nil {
		return params, err
	}
	params["title"] = config.Title

	return params, nil
}

// SetChatDescriptionConfig allows you to set the description of a supergroup or channel.
type SetChatDescriptionConfig struct {
	ChatID          int64
	ChannelUsername string

	Description string
}

func (config SetChatDescriptionConfig) method() string {
	return "setChatDescription"
}

func (config SetChatDescriptionConfig) params() (Params, error) {
	params := make(Params)

	if err := params.AddFirstValid("chat_id", config.ChatID, config.ChannelUsername); err != nil {
		return params, err
	}
	params["description"] = config.Description

	return params, nil
}

// GetStickerSetConfig allows you to get the stickers in a set.
type GetStickerSetConfig struct {
	Name string
}

func (config GetStickerSetConfig) method() string {
	return "getStickerSet"
}

func (config GetStickerSetConfig) params() (Params, error) {
	params := make(Params)

	params["name"] = config.Name

	return params, nil
}

// GetCustomEmojiStickersConfig get information about custom emoji stickers
// by their identifiers.
type GetCustomEmojiStickersConfig struct {
	CustomEmojiIDs []string
}

func (config GetCustomEmojiStickersConfig) method() string {
	return "getCustomEmojiStickers"
}

func (config GetCustomEmojiStickersConfig) params() (Params, error) {
	params := make(Params)

	err := params.AddInterface("custom_emoji_ids", config.CustomEmojiIDs)

	return params, err
}

// UploadStickerConfig uploads a sticker file for later use in a sticker set.
type UploadStickerConfig struct {
	UserID        int64
	Sticker       RequestFileData // required
	StickerFormat string          // required, one of StickerFormatStatic, StickerFormatAnimated, StickerFormatVideo
}

func (config UploadStickerConfig) method() string {
	return "uploadStickerFile"
}

func (config UploadStickerConfig) params() (Params, error) {
	params := make(Params)

	params.AddNonZero64("user_id", config.UserID)
	params.AddNonEmpty("sticker_format", config.StickerFormat)

	return params, nil
}

func (config UploadStickerConfig) files() []RequestFile {
	return []RequestFile{{
		Name: "sticker",
		Data: config.Sticker,
	}}
}

// NewStickerSetConfig creates a new sticker set owned by a user.
//
// Each sticker's format is specified on the InputSticker itself via its
// Format field, allowing mixed-format sticker packs.
type NewStickerSetConfig struct {
	UserID          int64
	Name            string
	Title           string
	Stickers        []InputSticker
	StickerType     string // one of StickerTypeRegular, StickerTypeMask, StickerTypeCustomEmoji
	NeedsRepainting bool
}

func (config NewStickerSetConfig) method() string {
	return "createNewStickerSet"
}

func (config NewStickerSetConfig) params() (Params, error) {
	params := make(Params)

	params.AddNonZero64("user_id", config.UserID)
	params["name"] = config.Name
	params["title"] = config.Title
	params.AddNonEmpty("sticker_type", config.StickerType)
	params.AddBool("needs_repainting", config.NeedsRepainting)

	err := params.AddAny("stickers", prepareInputStickersForParams(config.Stickers))

	return params, err
}

func (config NewStickerSetConfig) files() []RequestFile {
	return prepareInputStickersForFiles(config.Stickers)
}

// AddStickerConfig adds a new sticker to an existing sticker set.
type AddStickerConfig struct {
	UserID  int64
	Name    string
	Sticker InputSticker
}

func (config AddStickerConfig) method() string {
	return "addStickerToSet"
}

func (config AddStickerConfig) params() (Params, error) {
	params := make(Params)

	params.AddNonZero64("user_id", config.UserID)
	params["name"] = config.Name

	err := params.AddAny("sticker", prepareInputStickerForParams(config.Sticker, 0))

	return params, err
}

func (config AddStickerConfig) files() []RequestFile {
	return prepareInputStickerForFiles(config.Sticker, 0)
}

// prepareInputStickerForParams returns a copy of the sticker with the
// Sticker field replaced by an attach:// reference if it needs uploading.
func prepareInputStickerForParams(s InputSticker, idx int) InputSticker {
	if s.Sticker != nil && s.Sticker.NeedsUpload() {
		s.Sticker = fileAttach(fmt.Sprintf("attach://sticker-%d", idx))
	}
	return s
}

// prepareInputStickerForFiles returns the upload entries for a single sticker.
func prepareInputStickerForFiles(s InputSticker, idx int) []RequestFile {
	if s.Sticker != nil && s.Sticker.NeedsUpload() {
		return []RequestFile{{
			Name: fmt.Sprintf("sticker-%d", idx),
			Data: s.Sticker,
		}}
	}
	return nil
}

// prepareInputStickersForParams applies prepareInputStickerForParams to a slice.
func prepareInputStickersForParams(stickers []InputSticker) []InputSticker {
	out := make([]InputSticker, len(stickers))
	for i, s := range stickers {
		out[i] = prepareInputStickerForParams(s, i)
	}
	return out
}

// prepareInputStickersForFiles flattens the upload entries for a slice of stickers.
func prepareInputStickersForFiles(stickers []InputSticker) []RequestFile {
	var files []RequestFile
	for i, s := range stickers {
		files = append(files, prepareInputStickerForFiles(s, i)...)
	}
	return files
}

// SetStickerPositionConfig allows you to change the position of a sticker in a set.
type SetStickerPositionConfig struct {
	Sticker  string
	Position int
}

func (config SetStickerPositionConfig) method() string {
	return "setStickerPositionInSet"
}

func (config SetStickerPositionConfig) params() (Params, error) {
	params := make(Params)

	params["sticker"] = config.Sticker
	params.AddNonZero("position", config.Position)

	return params, nil
}

// DeleteStickerConfig allows you to delete a sticker from a set.
type DeleteStickerConfig struct {
	Sticker string
}

func (config DeleteStickerConfig) method() string {
	return "deleteStickerFromSet"
}

func (config DeleteStickerConfig) params() (Params, error) {
	params := make(Params)

	params["sticker"] = config.Sticker

	return params, nil
}

// SetStickerSetThumbnailConfig sets the thumbnail of a sticker set.
type SetStickerSetThumbnailConfig struct {
	Name      string
	UserID    int64
	Thumbnail RequestFileData
}

func (config SetStickerSetThumbnailConfig) method() string {
	return "setStickerSetThumbnail"
}

func (config SetStickerSetThumbnailConfig) params() (Params, error) {
	params := make(Params)

	params["name"] = config.Name
	params.AddNonZero64("user_id", config.UserID)

	return params, nil
}

func (config SetStickerSetThumbnailConfig) files() []RequestFile {
	return []RequestFile{{
		Name: "thumbnail",
		Data: config.Thumbnail,
	}}
}

// SetCustomEmojiStickerSetThumbnailConfig sets the thumbnail of a custom
// emoji sticker set. The bot must own the sticker set.
type SetCustomEmojiStickerSetThumbnailConfig struct {
	Name          string
	CustomEmojiID string // pass an empty string to drop the thumbnail
}

func (config SetCustomEmojiStickerSetThumbnailConfig) method() string {
	return "setCustomEmojiStickerSetThumbnail"
}

func (config SetCustomEmojiStickerSetThumbnailConfig) params() (Params, error) {
	params := make(Params)

	params["name"] = config.Name
	params.AddNonEmpty("custom_emoji_id", config.CustomEmojiID)

	return params, nil
}

// SetStickerSetTitleConfig sets the title of a sticker set created by the bot.
type SetStickerSetTitleConfig struct {
	Name  string
	Title string
}

func (config SetStickerSetTitleConfig) method() string {
	return "setStickerSetTitle"
}

func (config SetStickerSetTitleConfig) params() (Params, error) {
	params := make(Params)

	params["name"] = config.Name
	params["title"] = config.Title

	return params, nil
}

// DeleteStickerSetConfig deletes a sticker set created by the bot.
type DeleteStickerSetConfig struct {
	Name string
}

func (config DeleteStickerSetConfig) method() string {
	return "deleteStickerSet"
}

func (config DeleteStickerSetConfig) params() (Params, error) {
	params := make(Params)

	params["name"] = config.Name

	return params, nil
}

// SetStickerEmojiListConfig changes the list of emoji assigned to a regular
// or custom emoji sticker. The sticker must belong to a sticker set created
// by the bot.
type SetStickerEmojiListConfig struct {
	Sticker   string // file identifier of the sticker
	EmojiList []string
}

func (config SetStickerEmojiListConfig) method() string {
	return "setStickerEmojiList"
}

func (config SetStickerEmojiListConfig) params() (Params, error) {
	params := make(Params)

	params["sticker"] = config.Sticker
	err := params.AddAny("emoji_list", config.EmojiList)

	return params, err
}

// SetStickerKeywordsConfig changes the search keywords assigned to a regular
// or custom emoji sticker. The sticker must belong to a sticker set created
// by the bot.
type SetStickerKeywordsConfig struct {
	Sticker  string // file identifier of the sticker
	Keywords []string
}

func (config SetStickerKeywordsConfig) method() string {
	return "setStickerKeywords"
}

func (config SetStickerKeywordsConfig) params() (Params, error) {
	params := make(Params)

	params["sticker"] = config.Sticker
	err := params.AddAny("keywords", config.Keywords)

	return params, err
}

// SetStickerMaskPositionConfig changes the mask position of a mask sticker.
// The sticker must belong to a sticker set created by the bot.
type SetStickerMaskPositionConfig struct {
	Sticker      string // file identifier of the sticker
	MaskPosition *MaskPosition
}

func (config SetStickerMaskPositionConfig) method() string {
	return "setStickerMaskPosition"
}

func (config SetStickerMaskPositionConfig) params() (Params, error) {
	params := make(Params)

	params["sticker"] = config.Sticker
	err := params.AddAny("mask_position", config.MaskPosition)

	return params, err
}

// ReplaceStickerInSetConfig replaces an existing sticker in a sticker set
// with a new one. The sticker set must have been created by the bot.
type ReplaceStickerInSetConfig struct {
	UserID     int64
	Name       string
	OldSticker string // file identifier of the replaced sticker
	Sticker    InputSticker
}

func (config ReplaceStickerInSetConfig) method() string {
	return "replaceStickerInSet"
}

func (config ReplaceStickerInSetConfig) params() (Params, error) {
	params := make(Params)

	params.AddNonZero64("user_id", config.UserID)
	params["name"] = config.Name
	params["old_sticker"] = config.OldSticker

	err := params.AddAny("sticker", prepareInputStickerForParams(config.Sticker, 0))

	return params, err
}

func (config ReplaceStickerInSetConfig) files() []RequestFile {
	return prepareInputStickerForFiles(config.Sticker, 0)
}

// SetChatStickerSetConfig allows you to set the sticker set for a supergroup.
type SetChatStickerSetConfig struct {
	ChatID             int64
	SuperGroupUsername string

	StickerSetName string
}

func (config SetChatStickerSetConfig) method() string {
	return "setChatStickerSet"
}

func (config SetChatStickerSetConfig) params() (Params, error) {
	params := make(Params)

	if err := params.AddFirstValid("chat_id", config.ChatID, config.SuperGroupUsername); err != nil {
		return params, err
	}
	params["sticker_set_name"] = config.StickerSetName

	return params, nil
}

// DeleteChatStickerSetConfig allows you to remove a supergroup's sticker set.
type DeleteChatStickerSetConfig struct {
	ChatID             int64
	SuperGroupUsername string
}

func (config DeleteChatStickerSetConfig) method() string {
	return "deleteChatStickerSet"
}

func (config DeleteChatStickerSetConfig) params() (Params, error) {
	params := make(Params)

	if err := params.AddFirstValid("chat_id", config.ChatID, config.SuperGroupUsername); err != nil {
		return params, err
	}

	return params, nil
}

// SendPaidMediaConfig sends paid media to a channel or private chat.
type SendPaidMediaConfig struct {
	BaseChat
	// StarCount is the number of Telegram Stars that must be paid to buy
	// access to the media; 1-2500.
	StarCount int
	// Media is the list of media to be sent; 1-10 items.
	Media []InputPaidMedia
	// Payload is the bot-defined paid media payload, 0-128 bytes. Received
	// back in a PurchasedPaidMedia update and in TransactionPartner.
	Payload string
	// Caption of the media to be sent, 0-1024 characters after entities parsing.
	Caption string
	// ParseMode mode for parsing entities in the caption.
	ParseMode string
	// CaptionEntities is a list of special entities that appear in the caption.
	CaptionEntities []MessageEntity
	// ShowCaptionAboveMedia pass True if the caption must be shown above the message media.
	ShowCaptionAboveMedia bool
}

func (config SendPaidMediaConfig) method() string {
	return "sendPaidMedia"
}

func (config SendPaidMediaConfig) params() (Params, error) {
	params, err := config.BaseChat.params()
	if err != nil {
		return params, err
	}

	params.AddNonZero("star_count", config.StarCount)
	params.AddNonEmpty("payload", config.Payload)
	params.AddNonEmpty("caption", config.Caption)
	params.AddNonEmpty("parse_mode", config.ParseMode)
	params.AddBool("show_caption_above_media", config.ShowCaptionAboveMedia)
	if err = params.AddAny("caption_entities", config.CaptionEntities); err != nil {
		return params, err
	}
	err = params.AddAny("media", prepareInputPaidMediaForParams(config.Media))

	return params, err
}

func (config SendPaidMediaConfig) files() []RequestFile {
	return prepareInputPaidMediaForFiles(config.Media)
}

// prepareInputPaidMediaForParams rewrites InputPaidMedia entries whose Media
// or Thumbnail need uploading to attach:// references, mirroring
// prepareInputMediaForParams for regular media groups.
func prepareInputPaidMediaForParams(items []InputPaidMedia) []InputPaidMedia {
	out := make([]InputPaidMedia, len(items))
	for i, m := range items {
		if m.Media != nil && m.Media.NeedsUpload() {
			m.Media = fileAttach(fmt.Sprintf("attach://paid-media-%d", i))
		}
		if m.Thumbnail != nil && m.Thumbnail.NeedsUpload() {
			m.Thumbnail = fileAttach(fmt.Sprintf("attach://paid-media-%d-thumbnail", i))
		}
		out[i] = m
	}
	return out
}

// prepareInputPaidMediaForFiles returns the upload entries for items in the
// slice whose Media or Thumbnail need uploading.
func prepareInputPaidMediaForFiles(items []InputPaidMedia) []RequestFile {
	var files []RequestFile
	for i, m := range items {
		if m.Media != nil && m.Media.NeedsUpload() {
			files = append(files, RequestFile{
				Name: fmt.Sprintf("paid-media-%d", i),
				Data: m.Media,
			})
		}
		if m.Thumbnail != nil && m.Thumbnail.NeedsUpload() {
			files = append(files, RequestFile{
				Name: fmt.Sprintf("paid-media-%d-thumbnail", i),
				Data: m.Thumbnail,
			})
		}
	}
	return files
}

// MediaGroupConfig allows you to send a group of media.
//
// Media consist of InputMedia items (InputMediaPhoto, InputMediaVideo).
type MediaGroupConfig struct {
	ChatID               int64
	ChannelUsername      string
	BusinessConnectionID string
	MessageThreadID      int
	MessageEffectID      string

	Media               []interface{}
	DisableNotification bool
	ProtectContent      bool
	AllowPaidBroadcast  bool
	ReplyParameters     *ReplyParameters
}

func (config MediaGroupConfig) method() string {
	return "sendMediaGroup"
}

func (config MediaGroupConfig) params() (Params, error) {
	params := make(Params)

	if err := params.AddFirstValid("chat_id", config.ChatID, config.ChannelUsername); err != nil {
		return params, err
	}
	params.AddNonEmpty("business_connection_id", config.BusinessConnectionID)
	params.AddNonZero("message_thread_id", config.MessageThreadID)
	params.AddNonEmpty("message_effect_id", config.MessageEffectID)
	params.AddBool("disable_notification", config.DisableNotification)
	params.AddBool("protect_content", config.ProtectContent)
	params.AddBool("allow_paid_broadcast", config.AllowPaidBroadcast)
	if err := params.AddAny("reply_parameters", config.ReplyParameters); err != nil {
		return params, err
	}

	err := params.AddAny("media", prepareInputMediaForParams(config.Media))

	return params, err
}

func (config MediaGroupConfig) files() []RequestFile {
	return prepareInputMediaForFiles(config.Media)
}

// DiceConfig contains information about a sendDice request.
type DiceConfig struct {
	BaseChat
	// Emoji on which the dice throw animation is based.
	// Currently, must be one of 🎲, 🎯, 🏀, ⚽, 🎳, or 🎰.
	// Dice can have values 1-6 for 🎲, 🎯, and 🎳, values 1-5 for 🏀 and ⚽,
	// and values 1-64 for 🎰.
	// Defaults to “🎲”
	Emoji string
}

func (config DiceConfig) method() string {
	return "sendDice"
}

func (config DiceConfig) params() (Params, error) {
	params, err := config.BaseChat.params()
	if err != nil {
		return params, err
	}

	params.AddNonEmpty("emoji", config.Emoji)

	return params, err
}

// GetMyCommandsConfig gets a list of the currently registered commands.
type GetMyCommandsConfig struct {
	Scope        *BotCommandScope
	LanguageCode string
}

func (config GetMyCommandsConfig) method() string {
	return "getMyCommands"
}

func (config GetMyCommandsConfig) params() (Params, error) {
	params := make(Params)

	err := params.AddInterface("scope", config.Scope)
	params.AddNonEmpty("language_code", config.LanguageCode)

	return params, err
}

// SetMyCommandsConfig sets a list of commands the bot understands.
type SetMyCommandsConfig struct {
	Commands     []BotCommand
	Scope        *BotCommandScope
	LanguageCode string
}

func (config SetMyCommandsConfig) method() string {
	return "setMyCommands"
}

func (config SetMyCommandsConfig) params() (Params, error) {
	params := make(Params)

	if err := params.AddInterface("commands", config.Commands); err != nil {
		return params, err
	}
	err := params.AddInterface("scope", config.Scope)
	params.AddNonEmpty("language_code", config.LanguageCode)

	return params, err
}

type DeleteMyCommandsConfig struct {
	Scope        *BotCommandScope
	LanguageCode string
}

func (config DeleteMyCommandsConfig) method() string {
	return "deleteMyCommands"
}

func (config DeleteMyCommandsConfig) params() (Params, error) {
	params := make(Params)

	err := params.AddInterface("scope", config.Scope)
	params.AddNonEmpty("language_code", config.LanguageCode)

	return params, err
}

// SetMyNameConfig changes the bot's name. Different names can be set for
// different user languages.
type SetMyNameConfig struct {
	Name         string
	LanguageCode string
}

func (config SetMyNameConfig) method() string {
	return "setMyName"
}

func (config SetMyNameConfig) params() (Params, error) {
	params := make(Params)

	params.AddNonEmpty("name", config.Name)
	params.AddNonEmpty("language_code", config.LanguageCode)

	return params, nil
}

// GetMyNameConfig returns the current bot name for the given user language.
type GetMyNameConfig struct {
	LanguageCode string
}

func (config GetMyNameConfig) method() string {
	return "getMyName"
}

func (config GetMyNameConfig) params() (Params, error) {
	params := make(Params)

	params.AddNonEmpty("language_code", config.LanguageCode)

	return params, nil
}

// SetMyDescriptionConfig changes the bot's description, which is shown in the
// chat with the bot if the chat is empty.
type SetMyDescriptionConfig struct {
	Description  string
	LanguageCode string
}

func (config SetMyDescriptionConfig) method() string {
	return "setMyDescription"
}

func (config SetMyDescriptionConfig) params() (Params, error) {
	params := make(Params)

	params.AddNonEmpty("description", config.Description)
	params.AddNonEmpty("language_code", config.LanguageCode)

	return params, nil
}

// GetMyDescriptionConfig returns the current bot description for the given
// user language.
type GetMyDescriptionConfig struct {
	LanguageCode string
}

func (config GetMyDescriptionConfig) method() string {
	return "getMyDescription"
}

func (config GetMyDescriptionConfig) params() (Params, error) {
	params := make(Params)

	params.AddNonEmpty("language_code", config.LanguageCode)

	return params, nil
}

// SetMyShortDescriptionConfig changes the bot's short description, which is
// shown on the bot's profile page and is sent together with the link when
// users share the bot.
type SetMyShortDescriptionConfig struct {
	ShortDescription string
	LanguageCode     string
}

func (config SetMyShortDescriptionConfig) method() string {
	return "setMyShortDescription"
}

func (config SetMyShortDescriptionConfig) params() (Params, error) {
	params := make(Params)

	params.AddNonEmpty("short_description", config.ShortDescription)
	params.AddNonEmpty("language_code", config.LanguageCode)

	return params, nil
}

// GetBusinessConnectionConfig returns information about the connection of
// the bot with a business account.
type GetBusinessConnectionConfig struct {
	BusinessConnectionID string
}

func (config GetBusinessConnectionConfig) method() string {
	return "getBusinessConnection"
}

func (config GetBusinessConnectionConfig) params() (Params, error) {
	params := make(Params)

	params["business_connection_id"] = config.BusinessConnectionID

	return params, nil
}

// GetMyShortDescriptionConfig returns the current bot short description for
// the given user language.
type GetMyShortDescriptionConfig struct {
	LanguageCode string
}

func (config GetMyShortDescriptionConfig) method() string {
	return "getMyShortDescription"
}

func (config GetMyShortDescriptionConfig) params() (Params, error) {
	params := make(Params)

	params.AddNonEmpty("language_code", config.LanguageCode)

	return params, nil
}

// SetChatMenuButtonConfig changes the bot's menu button in a private chat,
// or the default menu button.
type SetChatMenuButtonConfig struct {
	ChatID          int64
	ChannelUsername string

	MenuButton *MenuButton
}

func (config SetChatMenuButtonConfig) method() string {
	return "setChatMenuButton"
}

func (config SetChatMenuButtonConfig) params() (Params, error) {
	params := make(Params)

	if err := params.AddFirstValid("chat_id", config.ChatID, config.ChannelUsername); err != nil {
		return params, err
	}
	err := params.AddInterface("menu_button", config.MenuButton)

	return params, err
}

type GetChatMenuButtonConfig struct {
	ChatID          int64
	ChannelUsername string
}

func (config GetChatMenuButtonConfig) method() string {
	return "getChatMenuButton"
}

func (config GetChatMenuButtonConfig) params() (Params, error) {
	params := make(Params)

	err := params.AddFirstValid("chat_id", config.ChatID, config.ChannelUsername)

	return params, err
}

type SetMyDefaultAdministratorRightsConfig struct {
	Rights      ChatAdministratorRights
	ForChannels bool
}

func (config SetMyDefaultAdministratorRightsConfig) method() string {
	return "setMyDefaultAdministratorRights"
}

func (config SetMyDefaultAdministratorRightsConfig) params() (Params, error) {
	params := make(Params)

	err := params.AddInterface("rights", config.Rights)
	params.AddBool("for_channels", config.ForChannels)

	return params, err
}

type GetMyDefaultAdministratorRightsConfig struct {
	ForChannels bool
}

func (config GetMyDefaultAdministratorRightsConfig) method() string {
	return "getMyDefaultAdministratorRights"
}

func (config GetMyDefaultAdministratorRightsConfig) params() (Params, error) {
	params := make(Params)

	params.AddBool("for_channels", config.ForChannels)

	return params, nil
}

// prepareInputMediaParam evaluates a single InputMedia and determines if it
// needs to be modified for a successful upload. If it returns nil, then the
// value does not need to be included in the params. Otherwise, it will return
// the same type as was originally provided.
//
// The idx is used to calculate the file field name. If you only have a single
// file, 0 may be used. It is formatted into "attach://file-%d" for the primary
// media and "attach://file-%d-thumb" for thumbnails.
//
// It is expected to be used in conjunction with prepareInputMediaFile.
// GetManagedBotTokenConfig contains the parameters for the getManagedBotToken method.
type GetManagedBotTokenConfig struct {
	// UserID is the user identifier of the managed bot whose token will be returned.
	UserID int64
}

func (GetManagedBotTokenConfig) method() string {
	return "getManagedBotToken"
}

func (config GetManagedBotTokenConfig) params() (Params, error) {
	params := make(Params)

	params.AddNonZero64("user_id", config.UserID)

	return params, nil
}

// ReplaceManagedBotTokenConfig contains the parameters for the replaceManagedBotToken method.
type ReplaceManagedBotTokenConfig struct {
	// UserID is the user identifier of the managed bot whose token will be replaced.
	UserID int64
}

func (ReplaceManagedBotTokenConfig) method() string {
	return "replaceManagedBotToken"
}

func (config ReplaceManagedBotTokenConfig) params() (Params, error) {
	params := make(Params)

	params.AddNonZero64("user_id", config.UserID)

	return params, nil
}

// SavePreparedKeyboardButtonConfig contains the parameters for the savePreparedKeyboardButton method.
type SavePreparedKeyboardButtonConfig struct {
	// UserID is the unique identifier of the target user that can use the button.
	UserID int64
	// Button is a KeyboardButton describing the button to be saved.
	// The button must be of the type request_users, request_chat, or request_managed_bot.
	Button KeyboardButton
}

func (SavePreparedKeyboardButtonConfig) method() string {
	return "savePreparedKeyboardButton"
}

func (config SavePreparedKeyboardButtonConfig) params() (Params, error) {
	params := make(Params)

	params.AddNonZero64("user_id", config.UserID)
	err := params.AddInterface("button", config.Button)

	return params, err
}

func prepareInputMediaParam(inputMedia interface{}, idx int) interface{} {
	switch m := inputMedia.(type) {
	case InputMediaPhoto:
		if m.Media.NeedsUpload() {
			m.Media = fileAttach(fmt.Sprintf("attach://file-%d", idx))
		}

		return m
	case InputMediaVideo:
		if m.Media.NeedsUpload() {
			m.Media = fileAttach(fmt.Sprintf("attach://file-%d", idx))
		}

		if m.Thumbnail != nil && m.Thumbnail.NeedsUpload() {
			m.Thumbnail = fileAttach(fmt.Sprintf("attach://file-%d-thumbnail", idx))
		}

		return m
	case InputMediaAudio:
		if m.Media.NeedsUpload() {
			m.Media = fileAttach(fmt.Sprintf("attach://file-%d", idx))
		}

		if m.Thumbnail != nil && m.Thumbnail.NeedsUpload() {
			m.Thumbnail = fileAttach(fmt.Sprintf("attach://file-%d-thumbnail", idx))
		}

		return m
	case InputMediaDocument:
		if m.Media.NeedsUpload() {
			m.Media = fileAttach(fmt.Sprintf("attach://file-%d", idx))
		}

		if m.Thumbnail != nil && m.Thumbnail.NeedsUpload() {
			m.Thumbnail = fileAttach(fmt.Sprintf("attach://file-%d-thumbnail", idx))
		}

		return m
	}

	return nil
}

// prepareInputMediaFile generates an array of RequestFile to provide for
// Fileable's files method. It returns an array as a single InputMedia may have
// multiple files, for the primary media and a thumbnail.
//
// The idx parameter is used to generate file field names. It uses the names
// "file-%d" for the main file and "file-%d-thumb" for the thumbnail.
//
// It is expected to be used in conjunction with prepareInputMediaParam.
func prepareInputMediaFile(inputMedia interface{}, idx int) []RequestFile {
	files := []RequestFile{}

	switch m := inputMedia.(type) {
	case InputMediaPhoto:
		if m.Media.NeedsUpload() {
			files = append(files, RequestFile{
				Name: fmt.Sprintf("file-%d", idx),
				Data: m.Media,
			})
		}
	case InputMediaVideo:
		if m.Media.NeedsUpload() {
			files = append(files, RequestFile{
				Name: fmt.Sprintf("file-%d", idx),
				Data: m.Media,
			})
		}

		if m.Thumbnail != nil && m.Thumbnail.NeedsUpload() {
			files = append(files, RequestFile{
				Name: fmt.Sprintf("file-%d", idx),
				Data: m.Thumbnail,
			})
		}
	case InputMediaDocument:
		if m.Media.NeedsUpload() {
			files = append(files, RequestFile{
				Name: fmt.Sprintf("file-%d", idx),
				Data: m.Media,
			})
		}

		if m.Thumbnail != nil && m.Thumbnail.NeedsUpload() {
			files = append(files, RequestFile{
				Name: fmt.Sprintf("file-%d", idx),
				Data: m.Thumbnail,
			})
		}
	case InputMediaAudio:
		if m.Media.NeedsUpload() {
			files = append(files, RequestFile{
				Name: fmt.Sprintf("file-%d", idx),
				Data: m.Media,
			})
		}

		if m.Thumbnail != nil && m.Thumbnail.NeedsUpload() {
			files = append(files, RequestFile{
				Name: fmt.Sprintf("file-%d", idx),
				Data: m.Thumbnail,
			})
		}
	}

	return files
}

// prepareInputMediaForParams calls prepareInputMediaParam for each item
// provided and returns a new array with the correct params for a request.
//
// It is expected that files will get data from the associated function,
// prepareInputMediaForFiles.
func prepareInputMediaForParams(inputMedia []interface{}) []interface{} {
	newMedia := make([]interface{}, len(inputMedia))
	copy(newMedia, inputMedia)

	for idx, media := range inputMedia {
		if param := prepareInputMediaParam(media, idx); param != nil {
			newMedia[idx] = param
		}
	}

	return newMedia
}

// prepareInputMediaForFiles calls prepareInputMediaFile for each item
// provided and returns a new array with the correct files for a request.
//
// It is expected that params will get data from the associated function,
// prepareInputMediaForParams.
func prepareInputMediaForFiles(inputMedia []interface{}) []RequestFile {
	files := []RequestFile{}

	for idx, media := range inputMedia {
		if file := prepareInputMediaFile(media, idx); file != nil {
			files = append(files, file...)
		}
	}

	return files
}
