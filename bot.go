// Package tgbotapi has functions and types used for interacting with
// the Telegram Bot API.
package tgbotapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// HTTPClient is the type needed for the bot to perform HTTP requests.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// BotAPI allows you to interact with the Telegram Bot API.
type BotAPI struct {
	Token  string `json:"token"`
	Debug  bool   `json:"debug"`
	Buffer int    `json:"buffer"`

	Self            User       `json:"-"`
	Client          HTTPClient `json:"-"`
	shutdownChannel chan interface{}

	apiEndpoint string
}

// NewBotAPI creates a new BotAPI instance.
//
// It requires a token, provided by @BotFather on Telegram.
func NewBotAPI(token string) (*BotAPI, error) {
	return NewBotAPIWithClient(token, APIEndpoint, &http.Client{})
}

// NewBotAPIWithAPIEndpoint creates a new BotAPI instance
// and allows you to pass API endpoint.
//
// It requires a token, provided by @BotFather on Telegram and API endpoint.
func NewBotAPIWithAPIEndpoint(token, apiEndpoint string) (*BotAPI, error) {
	return NewBotAPIWithClient(token, apiEndpoint, &http.Client{})
}

// NewBotAPIWithClient creates a new BotAPI instance
// and allows you to pass a http.Client.
//
// It requires a token, provided by @BotFather on Telegram and API endpoint.
func NewBotAPIWithClient(token, apiEndpoint string, client HTTPClient) (*BotAPI, error) {
	bot := &BotAPI{
		Token:           token,
		Client:          client,
		Buffer:          100,
		shutdownChannel: make(chan interface{}),

		apiEndpoint: apiEndpoint,
	}

	self, err := bot.GetMe()
	if err != nil {
		return nil, err
	}

	bot.Self = self

	return bot, nil
}

// SetAPIEndpoint changes the Telegram Bot API endpoint used by the instance.
func (bot *BotAPI) SetAPIEndpoint(apiEndpoint string) {
	bot.apiEndpoint = apiEndpoint
}

func buildParams(in Params) url.Values {
	out := url.Values{}
	for key, value := range in {
		out.Set(key, value)
	}
	return out
}

// closeBody drains any unread bytes before closing the response body so that
// net/http can return the underlying connection to the keep-alive pool.
// json.Decoder may stop short of EOF (trailing whitespace, partial errors),
// and an undrained body forces the transport to discard the connection.
func closeBody(body io.ReadCloser) {
	_, _ = io.Copy(io.Discard, body)
	body.Close()
}

// MakeRequest makes a request to a specific endpoint with our token.
func (bot *BotAPI) MakeRequest(endpoint string, params Params) (*APIResponse, error) {
	if bot.Debug {
		log.Printf("Endpoint: %s, params: %v\n", endpoint, params)
	}

	method := fmt.Sprintf(bot.apiEndpoint, bot.Token, endpoint)

	values := buildParams(params)

	req, err := http.NewRequest(http.MethodPost, method, strings.NewReader(values.Encode()))
	if err != nil {
		return &APIResponse{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := bot.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp.Body)

	var apiResp APIResponse
	bytes, err := bot.decodeAPIResponse(resp.Body, &apiResp)
	if err != nil {
		return &apiResp, err
	}

	if bot.Debug {
		log.Printf("Endpoint: %s, response: %s\n", endpoint, string(bytes))
	}

	if !apiResp.Ok {
		var parameters ResponseParameters

		if apiResp.Parameters != nil {
			parameters = *apiResp.Parameters
		}

		return &apiResp, &Error{
			Code:               apiResp.ErrorCode,
			Message:            apiResp.Description,
			ResponseParameters: parameters,
		}
	}

	return &apiResp, nil
}

// decodeAPIResponse decode response and return slice of bytes if debug enabled.
// If debug disabled, just decode http.Response.Body stream to APIResponse struct
// for efficient memory usage
func (bot *BotAPI) decodeAPIResponse(responseBody io.Reader, resp *APIResponse) ([]byte, error) {
	if !bot.Debug {
		dec := json.NewDecoder(responseBody)
		err := dec.Decode(resp)
		return nil, err
	}

	// if debug, read response body
	data, err := io.ReadAll(responseBody)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, resp)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// UploadFiles makes a request to the API with files.
func (bot *BotAPI) UploadFiles(endpoint string, params Params, files []RequestFile) (*APIResponse, error) {
	r, w := io.Pipe()
	m := multipart.NewWriter(w)

	// This code modified from the very helpful @HirbodBehnam
	// https://github.com/go-telegram-bot-api/telegram-bot-api/issues/354#issuecomment-663856473
	go func() {
		defer w.Close()
		defer m.Close()

		for field, value := range params {
			if err := m.WriteField(field, value); err != nil {
				w.CloseWithError(err)
				return
			}
		}

		for _, file := range files {
			if file.Data.NeedsUpload() {
				name, reader, err := file.Data.UploadData()
				if err != nil {
					w.CloseWithError(err)
					return
				}

				part, err := m.CreateFormFile(file.Name, name)
				if err != nil {
					w.CloseWithError(err)
					return
				}

				if _, err := io.Copy(part, reader); err != nil {
					w.CloseWithError(err)
					return
				}

				if closer, ok := reader.(io.ReadCloser); ok {
					if err = closer.Close(); err != nil {
						w.CloseWithError(err)
						return
					}
				}
			} else {
				value := file.Data.SendData()

				if err := m.WriteField(file.Name, value); err != nil {
					w.CloseWithError(err)
					return
				}
			}
		}
	}()

	if bot.Debug {
		log.Printf("Endpoint: %s, params: %v, with %d files\n", endpoint, params, len(files))
	}

	method := fmt.Sprintf(bot.apiEndpoint, bot.Token, endpoint)

	req, err := http.NewRequest(http.MethodPost, method, r)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", m.FormDataContentType())

	resp, err := bot.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp.Body)

	var apiResp APIResponse
	bytes, err := bot.decodeAPIResponse(resp.Body, &apiResp)
	if err != nil {
		return &apiResp, err
	}

	if bot.Debug {
		log.Printf("Endpoint: %s, response: %s\n", endpoint, string(bytes))
	}

	if !apiResp.Ok {
		var parameters ResponseParameters

		if apiResp.Parameters != nil {
			parameters = *apiResp.Parameters
		}

		return &apiResp, &Error{
			Code:               apiResp.ErrorCode,
			Message:            apiResp.Description,
			ResponseParameters: parameters,
		}
	}

	return &apiResp, nil
}

// GetFileDirectURL returns direct URL to file
//
// It requires the FileID.
func (bot *BotAPI) GetFileDirectURL(fileID string) (string, error) {
	file, err := bot.GetFile(FileConfig{fileID})

	if err != nil {
		return "", err
	}

	return file.Link(bot.Token), nil
}

// GetMe fetches the currently authenticated bot.
//
// This method is called upon creation to validate the token,
// and so you may get this data from BotAPI.Self without the need for
// another request.
func (bot *BotAPI) GetMe() (User, error) {
	resp, err := bot.MakeRequest("getMe", nil)
	if err != nil {
		return User{}, err
	}

	var user User
	err = json.Unmarshal(resp.Result, &user)

	return user, err
}

// IsMessageToMe returns true if message directed to this bot.
//
// It requires the Message.
func (bot *BotAPI) IsMessageToMe(message Message) bool {
	return strings.Contains(message.Text, "@"+bot.Self.UserName)
}

func hasFilesNeedingUpload(files []RequestFile) bool {
	for _, file := range files {
		if file.Data.NeedsUpload() {
			return true
		}
	}

	return false
}

// Request sends a Chattable to Telegram, and returns the APIResponse.
func (bot *BotAPI) Request(c Chattable) (*APIResponse, error) {
	params, err := c.params()
	if err != nil {
		return nil, err
	}

	if t, ok := c.(Fileable); ok {
		files := t.files()

		// If we have files that need to be uploaded, we should delegate the
		// request to UploadFile.
		if hasFilesNeedingUpload(files) {
			return bot.UploadFiles(t.method(), params, files)
		}

		// However, if there are no files to be uploaded, there's likely things
		// that need to be turned into params instead.
		for _, file := range files {
			params[file.Name] = file.Data.SendData()
		}
	}

	return bot.MakeRequest(c.method(), params)
}

// Send will send a Chattable item to Telegram and provides the
// returned Message.
func (bot *BotAPI) Send(c Chattable) (Message, error) {
	resp, err := bot.Request(c)
	if err != nil {
		return Message{}, err
	}

	var message Message
	err = json.Unmarshal(resp.Result, &message)

	return message, err
}

// SendMediaGroup sends a media group and returns the resulting messages.
func (bot *BotAPI) SendMediaGroup(config MediaGroupConfig) ([]Message, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return nil, err
	}

	var messages []Message
	err = json.Unmarshal(resp.Result, &messages)

	return messages, err
}

// GetUserProfilePhotos gets a user's profile photos.
//
// It requires UserID.
// Offset and Limit are optional.
func (bot *BotAPI) GetUserProfilePhotos(config UserProfilePhotosConfig) (UserProfilePhotos, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return UserProfilePhotos{}, err
	}

	var profilePhotos UserProfilePhotos
	err = json.Unmarshal(resp.Result, &profilePhotos)

	return profilePhotos, err
}

// GetFile returns a File which can download a file from Telegram.
//
// Requires FileID.
func (bot *BotAPI) GetFile(config FileConfig) (File, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return File{}, err
	}

	var file File
	err = json.Unmarshal(resp.Result, &file)

	return file, err
}

// GetUpdates fetches updates.
// If a WebHook is set, this will not return any data!
//
// Offset, Limit, Timeout, and AllowedUpdates are optional.
// To avoid stale items, set Offset to one higher than the previous item.
// Set Timeout to a large number to reduce requests, so you can get updates
// instantly instead of having to wait between requests.
func (bot *BotAPI) GetUpdates(config UpdateConfig) ([]Update, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return []Update{}, err
	}

	var updates []Update
	err = json.Unmarshal(resp.Result, &updates)

	return updates, err
}

// GetWebhookInfo allows you to fetch information about a webhook and if
// one currently is set, along with pending update count and error messages.
func (bot *BotAPI) GetWebhookInfo() (WebhookInfo, error) {
	resp, err := bot.MakeRequest("getWebhookInfo", nil)
	if err != nil {
		return WebhookInfo{}, err
	}

	var info WebhookInfo
	err = json.Unmarshal(resp.Result, &info)

	return info, err
}

// GetUpdatesChan starts and returns a channel for getting updates.
func (bot *BotAPI) GetUpdatesChan(config UpdateConfig) UpdatesChannel {
	ch := make(chan Update, bot.Buffer)

	go func() {
		for {
			select {
			case <-bot.shutdownChannel:
				close(ch)
				return
			default:
			}

			updates, err := bot.GetUpdates(config)
			if err != nil {
				log.Println(err)
				log.Println("Failed to get updates, retrying in 3 seconds...")
				time.Sleep(time.Second * 3)

				continue
			}

			for _, update := range updates {
				if update.UpdateID >= config.Offset {
					config.Offset = update.UpdateID + 1
					ch <- update
				}
			}
		}
	}()

	return ch
}

// StopReceivingUpdates stops the go routine which receives updates
func (bot *BotAPI) StopReceivingUpdates() {
	if bot.Debug {
		log.Println("Stopping the update receiver routine...")
	}
	close(bot.shutdownChannel)
}

// ListenForWebhook registers a http handler for a webhook.
func (bot *BotAPI) ListenForWebhook(pattern string) UpdatesChannel {
	ch := make(chan Update, bot.Buffer)

	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		update, err := bot.HandleUpdate(r)
		if err != nil {
			errMsg, _ := json.Marshal(map[string]string{"error": err.Error()})
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(errMsg)
			return
		}

		ch <- *update
	})

	return ch
}

// ListenForWebhookRespReqFormat registers a http handler for a single incoming webhook.
func (bot *BotAPI) ListenForWebhookRespReqFormat(w http.ResponseWriter, r *http.Request) UpdatesChannel {
	ch := make(chan Update, bot.Buffer)

	func(w http.ResponseWriter, r *http.Request) {
		defer close(ch)

		update, err := bot.HandleUpdate(r)
		if err != nil {
			errMsg, _ := json.Marshal(map[string]string{"error": err.Error()})
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(errMsg)
			return
		}

		ch <- *update
	}(w, r)

	return ch
}

// HandleUpdate parses and returns update received via webhook
func (bot *BotAPI) HandleUpdate(r *http.Request) (*Update, error) {
	if r.Method != http.MethodPost {
		err := errors.New("wrong HTTP method required POST")
		return nil, err
	}

	var update Update
	err := json.NewDecoder(r.Body).Decode(&update)
	if err != nil {
		return nil, err
	}

	return &update, nil
}

// WriteToHTTPResponse writes the request to the HTTP ResponseWriter.
//
// It doesn't support uploading files.
//
// See https://core.telegram.org/bots/api#making-requests-when-getting-updates
// for details.
func WriteToHTTPResponse(w http.ResponseWriter, c Chattable) error {
	params, err := c.params()
	if err != nil {
		return err
	}

	if t, ok := c.(Fileable); ok {
		if hasFilesNeedingUpload(t.files()) {
			return errors.New("unable to use http response to upload files")
		}
	}

	values := buildParams(params)
	values.Set("method", c.method())

	w.Header().Set("Content-Type", "application/x-www-form-urlencoded")
	_, err = w.Write([]byte(values.Encode()))
	return err
}

// GetChat gets full information about a chat.
//
// As of Bot API 7.3 the return type is ChatFullInfo, which embeds Chat and
// adds the fields that are only populated by getChat.
func (bot *BotAPI) GetChat(config ChatInfoConfig) (ChatFullInfo, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return ChatFullInfo{}, err
	}

	var info ChatFullInfo
	err = json.Unmarshal(resp.Result, &info)

	return info, err
}

// GetChatAdministrators gets a list of administrators in the chat.
//
// If none have been appointed, only the creator will be returned.
// Bots are not shown, even if they are an administrator.
func (bot *BotAPI) GetChatAdministrators(config ChatAdministratorsConfig) ([]ChatMember, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return []ChatMember{}, err
	}

	var members []ChatMember
	err = json.Unmarshal(resp.Result, &members)

	return members, err
}

// GetChatMembersCount gets the number of users in a chat.
func (bot *BotAPI) GetChatMembersCount(config ChatMemberCountConfig) (int, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return -1, err
	}

	var count int
	err = json.Unmarshal(resp.Result, &count)

	return count, err
}

// GetChatMember gets a specific chat member.
func (bot *BotAPI) GetChatMember(config GetChatMemberConfig) (ChatMember, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return ChatMember{}, err
	}

	var member ChatMember
	err = json.Unmarshal(resp.Result, &member)

	return member, err
}

// GetGameHighScores allows you to get the high scores for a game.
func (bot *BotAPI) GetGameHighScores(config GetGameHighScoresConfig) ([]GameHighScore, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return []GameHighScore{}, err
	}

	var highScores []GameHighScore
	err = json.Unmarshal(resp.Result, &highScores)

	return highScores, err
}

// GetInviteLink get InviteLink for a chat
func (bot *BotAPI) GetInviteLink(config ChatInviteLinkConfig) (string, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return "", err
	}

	var inviteLink string
	err = json.Unmarshal(resp.Result, &inviteLink)

	return inviteLink, err
}

// GetStickerSet returns a StickerSet.
func (bot *BotAPI) GetStickerSet(config GetStickerSetConfig) (StickerSet, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return StickerSet{}, err
	}

	var stickers StickerSet
	err = json.Unmarshal(resp.Result, &stickers)

	return stickers, err
}

// GetCustomEmojiStickers returns information about custom emoji stickers by their identifiers.
func (bot *BotAPI) GetCustomEmojiStickers(config GetCustomEmojiStickersConfig) ([]Sticker, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return nil, err
	}

	var stickers []Sticker
	err = json.Unmarshal(resp.Result, &stickers)

	return stickers, err
}

// CreateForumTopic creates a topic in a forum supergroup chat and returns
// information about the created topic.
func (bot *BotAPI) CreateForumTopic(config CreateForumTopicConfig) (ForumTopic, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return ForumTopic{}, err
	}

	var topic ForumTopic
	err = json.Unmarshal(resp.Result, &topic)

	return topic, err
}

// GetForumTopicIconStickers returns custom emoji stickers, which can be used
// as a forum topic icon by any user.
func (bot *BotAPI) GetForumTopicIconStickers() ([]Sticker, error) {
	resp, err := bot.Request(GetForumTopicIconStickersConfig{})
	if err != nil {
		return nil, err
	}

	var stickers []Sticker
	err = json.Unmarshal(resp.Result, &stickers)

	return stickers, err
}

// GetUserChatBoosts returns the list of boosts added to a chat by a user.
func (bot *BotAPI) GetUserChatBoosts(config GetUserChatBoostsConfig) (UserChatBoosts, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return UserChatBoosts{}, err
	}

	var boosts UserChatBoosts
	err = json.Unmarshal(resp.Result, &boosts)

	return boosts, err
}

// ForwardMessages forwards multiple messages of any kind and returns the
// identifiers of the sent messages.
func (bot *BotAPI) ForwardMessages(config ForwardMessagesConfig) ([]MessageID, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return nil, err
	}

	var ids []MessageID
	err = json.Unmarshal(resp.Result, &ids)

	return ids, err
}

// CopyMessages copies messages of any kind and returns the identifiers of
// the sent messages.
func (bot *BotAPI) CopyMessages(config CopyMessagesConfig) ([]MessageID, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return nil, err
	}

	var ids []MessageID
	err = json.Unmarshal(resp.Result, &ids)

	return ids, err
}

// GetAvailableGifts returns the list of gifts that can be sent by the bot
// to users.
func (bot *BotAPI) GetAvailableGifts() (Gifts, error) {
	resp, err := bot.Request(GetAvailableGiftsConfig{})
	if err != nil {
		return Gifts{}, err
	}

	var gifts Gifts
	err = json.Unmarshal(resp.Result, &gifts)

	return gifts, err
}

// SavePreparedInlineMessage stores a message that can be sent by a user of
// a Mini App and returns a PreparedInlineMessage.
func (bot *BotAPI) SavePreparedInlineMessage(config SavePreparedInlineMessageConfig) (PreparedInlineMessage, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return PreparedInlineMessage{}, err
	}

	var msg PreparedInlineMessage
	err = json.Unmarshal(resp.Result, &msg)

	return msg, err
}

// GetStarTransactions returns the bot's Telegram Star transactions in
// chronological order.
func (bot *BotAPI) GetStarTransactions(config GetStarTransactionsConfig) (StarTransactions, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return StarTransactions{}, err
	}

	var st StarTransactions
	err = json.Unmarshal(resp.Result, &st)

	return st, err
}

// GetUserProfileAudios fetches a list of audios added to the profile of
// a user.
func (bot *BotAPI) GetUserProfileAudios(config GetUserProfileAudiosConfig) (UserProfileAudios, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return UserProfileAudios{}, err
	}

	var audios UserProfileAudios
	err = json.Unmarshal(resp.Result, &audios)

	return audios, err
}

// GetMyStarBalance returns the current Telegram Stars balance of the bot.
func (bot *BotAPI) GetMyStarBalance() (StarAmount, error) {
	resp, err := bot.Request(GetMyStarBalanceConfig{})
	if err != nil {
		return StarAmount{}, err
	}

	var amount StarAmount
	err = json.Unmarshal(resp.Result, &amount)

	return amount, err
}

// GetBusinessAccountStarBalance returns the amount of Telegram Stars owned
// by a managed business account.
func (bot *BotAPI) GetBusinessAccountStarBalance(config GetBusinessAccountStarBalanceConfig) (StarAmount, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return StarAmount{}, err
	}

	var amount StarAmount
	err = json.Unmarshal(resp.Result, &amount)

	return amount, err
}

// GetUserGifts returns the list of gifts received and owned by a user.
func (bot *BotAPI) GetUserGifts(config GetUserGiftsConfig) (OwnedGifts, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return OwnedGifts{}, err
	}

	var gifts OwnedGifts
	err = json.Unmarshal(resp.Result, &gifts)

	return gifts, err
}

// GetChatGifts returns the list of gifts received and owned by a chat.
func (bot *BotAPI) GetChatGifts(config GetChatGiftsConfig) (OwnedGifts, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return OwnedGifts{}, err
	}

	var gifts OwnedGifts
	err = json.Unmarshal(resp.Result, &gifts)

	return gifts, err
}

// RepostStory reposts a story across different business accounts managed
// by the bot and returns the newly posted Story.
func (bot *BotAPI) RepostStory(config RepostStoryConfig) (Story, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return Story{}, err
	}

	var story Story
	err = json.Unmarshal(resp.Result, &story)

	return story, err
}

// GetBusinessAccountGifts returns the gifts received and owned by a managed
// business account.
func (bot *BotAPI) GetBusinessAccountGifts(config GetBusinessAccountGiftsConfig) (OwnedGifts, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return OwnedGifts{}, err
	}

	var gifts OwnedGifts
	err = json.Unmarshal(resp.Result, &gifts)

	return gifts, err
}

// PostStory posts a story on behalf of a managed business account and
// returns the posted Story.
func (bot *BotAPI) PostStory(config PostStoryConfig) (Story, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return Story{}, err
	}

	var story Story
	err = json.Unmarshal(resp.Result, &story)

	return story, err
}

// EditStory edits a story previously posted by the bot and returns the
// edited Story.
func (bot *BotAPI) EditStory(config EditStoryConfig) (Story, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return Story{}, err
	}

	var story Story
	err = json.Unmarshal(resp.Result, &story)

	return story, err
}

// GetBusinessConnection returns information about the connection of the bot
// with a business account.
func (bot *BotAPI) GetBusinessConnection(config GetBusinessConnectionConfig) (BusinessConnection, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return BusinessConnection{}, err
	}

	var bc BusinessConnection
	err = json.Unmarshal(resp.Result, &bc)

	return bc, err
}

// GetMyName returns the current bot name for the given user language.
func (bot *BotAPI) GetMyName(config GetMyNameConfig) (BotName, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return BotName{}, err
	}

	var name BotName
	err = json.Unmarshal(resp.Result, &name)

	return name, err
}

// GetMyDescription returns the current bot description for the given user language.
func (bot *BotAPI) GetMyDescription(config GetMyDescriptionConfig) (BotDescription, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return BotDescription{}, err
	}

	var desc BotDescription
	err = json.Unmarshal(resp.Result, &desc)

	return desc, err
}

// GetMyShortDescription returns the current bot short description for the
// given user language.
func (bot *BotAPI) GetMyShortDescription(config GetMyShortDescriptionConfig) (BotShortDescription, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return BotShortDescription{}, err
	}

	var desc BotShortDescription
	err = json.Unmarshal(resp.Result, &desc)

	return desc, err
}

// StopPoll stops a poll and returns the result.
func (bot *BotAPI) StopPoll(config StopPollConfig) (Poll, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return Poll{}, err
	}

	var poll Poll
	err = json.Unmarshal(resp.Result, &poll)

	return poll, err
}

// GetMyCommands gets the currently registered commands.
func (bot *BotAPI) GetMyCommands() ([]BotCommand, error) {
	return bot.GetMyCommandsWithConfig(GetMyCommandsConfig{})
}

// GetMyCommandsWithConfig gets the currently registered commands with a config.
func (bot *BotAPI) GetMyCommandsWithConfig(config GetMyCommandsConfig) ([]BotCommand, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return nil, err
	}

	var commands []BotCommand
	err = json.Unmarshal(resp.Result, &commands)

	return commands, err
}

// CopyMessage copy messages of any kind. The method is analogous to the method
// forwardMessage, but the copied message doesn't have a link to the original
// message. Returns the MessageID of the sent message on success.
func (bot *BotAPI) CopyMessage(config CopyMessageConfig) (MessageID, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return MessageID{}, err
	}

	var messageID MessageID
	err = json.Unmarshal(resp.Result, &messageID)

	return messageID, err
}

// AnswerWebAppQuery sets the result of an interaction with a Web App and send a
// corresponding message on behalf of the user to the chat from which the query originated.
func (bot *BotAPI) AnswerWebAppQuery(config AnswerWebAppQueryConfig) (SentWebAppMessage, error) {
	var sentWebAppMessage SentWebAppMessage

	resp, err := bot.Request(config)
	if err != nil {
		return sentWebAppMessage, err
	}

	err = json.Unmarshal(resp.Result, &sentWebAppMessage)
	return sentWebAppMessage, err
}

// GetMyDefaultAdministratorRights gets the current default administrator rights of the bot.
func (bot *BotAPI) GetMyDefaultAdministratorRights(config GetMyDefaultAdministratorRightsConfig) (ChatAdministratorRights, error) {
	var rights ChatAdministratorRights

	resp, err := bot.Request(config)
	if err != nil {
		return rights, err
	}

	err = json.Unmarshal(resp.Result, &rights)
	return rights, err
}

// GetManagedBotToken returns the token of a managed bot.
func (bot *BotAPI) GetManagedBotToken(config GetManagedBotTokenConfig) (string, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return "", err
	}

	var token string
	err = json.Unmarshal(resp.Result, &token)

	return token, err
}

// ReplaceManagedBotToken revokes the current token of a managed bot
// and generates a new one. Returns the new token.
func (bot *BotAPI) ReplaceManagedBotToken(config ReplaceManagedBotTokenConfig) (string, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return "", err
	}

	var token string
	err = json.Unmarshal(resp.Result, &token)

	return token, err
}

// SavePreparedKeyboardButton stores a keyboard button that can be used
// by a user within a Mini App. Returns a PreparedKeyboardButton.
func (bot *BotAPI) SavePreparedKeyboardButton(config SavePreparedKeyboardButtonConfig) (PreparedKeyboardButton, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return PreparedKeyboardButton{}, err
	}

	var button PreparedKeyboardButton
	err = json.Unmarshal(resp.Result, &button)

	return button, err
}

// EscapeText takes an input text and escape Telegram markup symbols.
// In this way we can send a text without being afraid of having to escape the characters manually.
// Note that you don't have to include the formatting style in the input text, or it will be escaped too.
// If there is an error, an empty string will be returned.
//
// parseMode is the text formatting mode (ModeMarkdown, ModeMarkdownV2 or ModeHTML)
// text is the input string that will be escaped
func EscapeText(parseMode string, text string) string {
	var replacer *strings.Replacer

	if parseMode == ModeHTML {
		replacer = strings.NewReplacer("<", "&lt;", ">", "&gt;", "&", "&amp;")
	} else if parseMode == ModeMarkdown {
		replacer = strings.NewReplacer("_", "\\_", "*", "\\*", "`", "\\`", "[", "\\[")
	} else if parseMode == ModeMarkdownV2 {
		replacer = strings.NewReplacer(
			"_", "\\_", "*", "\\*", "[", "\\[", "]", "\\]", "(",
			"\\(", ")", "\\)", "~", "\\~", "`", "\\`", ">", "\\>",
			"#", "\\#", "+", "\\+", "-", "\\-", "=", "\\=", "|",
			"\\|", "{", "\\{", "}", "\\}", ".", "\\.", "!", "\\!",
		)
	} else {
		return ""
	}

	return replacer.Replace(text)
}
