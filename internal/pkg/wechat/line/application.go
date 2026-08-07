// Copyright © 2016- 2025 Wuhan Sesame Small Customer Service Network Technology Co., Ltd.

package line

import (
	"chatwiki/internal/pkg/lib_define"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ArtisanCloud/PowerWeChat/v3/src/kernel/response"
	openresponse "github.com/ArtisanCloud/PowerWeChat/v3/src/openPlatform/authorizer/miniProgram/account/response"
	"github.com/line/line-bot-sdk-go/v8/linebot/channel_access_token"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/spf13/cast"
	"github.com/zhimaAi/go_tools/logs"
)

var ErrNotSupported = errors.New(`not supported`)

// ImageUploadDomain is the public domain used to build sendable media URLs.
// LINE requires https URLs for image/video/audio messages.
var ImageUploadDomain string

var APIBase = "https://api.line.me"
var DataAPIBase = "https://api-data.line.me"

const defaultVideoPreviewPath = `/upload/default/line_avatar.png`

func SetAPIBase(base string) {
	base = strings.TrimRight(strings.TrimSpace(base), `/`)
	if len(base) > 0 {
		APIBase = base
	}
}

func SetDataAPIBase(base string) {
	base = strings.TrimRight(strings.TrimSpace(base), `/`)
	if len(base) > 0 {
		DataAPIBase = base
	}
}

// SetImageUploadDomain sets the public domain for outbound media URLs.
func SetImageUploadDomain(domain string) {
	domain = strings.TrimRight(strings.TrimSpace(domain), `/`)
	if len(domain) > 0 {
		ImageUploadDomain = domain
	}
}

type Application struct {
	ChannelID     string
	ChannelSecret string
}

// GetToken returns not-supported because LINE OAuth token type is incompatible
// with PowerWeChat's ResponseGetToken. Credential validation is done separately
// in the management save flow via connectivity check.
func (a *Application) GetToken(refresh bool) (*response.ResponseGetToken, int, error) {
	return nil, 0, ErrNotSupported
}

// getAccessToken issues (or reads from cache) a short-lived channel access token,
// shared by the send side and the inbound media download.
func getAccessToken(channelID, channelSecret string) (string, error) {
	redisKey := fmt.Sprintf(lib_define.RedisPrefixLineAccessToken, channelID)
	if lib_define.Redis != nil {
		if token, err := lib_define.Redis.Get(context.Background(), redisKey).Result(); err == nil && len(token) > 0 {
			return token, nil
		}
	}
	token, expiresIn, err := IssueChannelToken(channelID, channelSecret)
	if err != nil {
		return ``, err
	}
	if lib_define.Redis != nil && expiresIn > 0 {
		ttl := time.Duration(expiresIn/2) * time.Second
		if _, err := lib_define.Redis.Set(context.Background(), redisKey, token, ttl).Result(); err != nil {
			logs.Error(err.Error())
		}
	}
	return token, nil
}

// GetChannelToken returns a cached channel access token for callers outside
// the Application struct (e.g. inbound media download in message_service).
func GetChannelToken(channelID, channelSecret string) (string, error) {
	return getAccessToken(channelID, channelSecret)
}

func (a *Application) getAccessToken() (string, error) {
	return getAccessToken(a.ChannelID, a.ChannelSecret)
}

// getClient returns a messaging api client authenticated with channel access token.
func (a *Application) getClient() (*messaging_api.MessagingApiAPI, error) {
	token, err := a.getAccessToken()
	if err != nil {
		return nil, err
	}
	return messaging_api.NewMessagingApiAPI(token, messaging_api.WithEndpoint(APIBase))
}

// reply tries the free Reply API first using replyToken, falling back to the
// paid Push API when the token is missing or rejected.
func (a *Application) reply(customer string, push *lib_define.PushMessage, messages []messaging_api.MessageInterface) (int, error) {
	client, err := a.getClient()
	if err != nil {
		return 0, err
	}
	replyToken := ``
	if push != nil && push.Message != nil {
		replyToken = cast.ToString(push.Message[`replyToken`])
	}
	if len(replyToken) > 0 {
		if _, err = client.ReplyMessage(&messaging_api.ReplyMessageRequest{
			ReplyToken: replyToken,
			Messages:   messages,
		}); err == nil {
			return 0, nil
		}
		logs.Error(`line reply message failed, fallback to push: %s`, err.Error())
	}
	if _, err = client.PushMessage(&messaging_api.PushMessageRequest{
		To:       customer,
		Messages: messages,
	}, ``); err != nil {
		return 0, err
	}
	return 0, nil
}

func (a *Application) SendText(customer, content string, push *lib_define.PushMessage) (int, error) {
	return a.reply(customer, push, []messaging_api.MessageInterface{
		messaging_api.TextMessage{Text: content},
	})
}

func (a *Application) SendImage(customer, filePath string, push *lib_define.PushMessage) (int, error) {
	url := buildMediaURL(filePath)
	if len(url) == 0 {
		return 0, errors.New(`line image url is empty`)
	}
	return a.reply(customer, push, []messaging_api.MessageInterface{
		messaging_api.ImageMessage{OriginalContentUrl: url, PreviewImageUrl: url},
	})
}

func (a *Application) SendVideo(customer, filePath string, push *lib_define.PushMessage) (int, error) {
	url := buildMediaURL(filePath)
	if len(url) == 0 {
		return 0, errors.New(`line video url is empty`)
	}
	previewUrl := buildMediaURL(defaultVideoPreviewPath)
	if len(previewUrl) == 0 {
		return 0, errors.New(`line video preview url is empty`)
	}
	return a.reply(customer, push, []messaging_api.MessageInterface{
		messaging_api.VideoMessage{OriginalContentUrl: url, PreviewImageUrl: previewUrl},
	})
}

func (a *Application) SendVoice(customer, filePath string, push *lib_define.PushMessage) (int, error) {
	duration := lineAudioDuration(filePath)
	url := buildMediaURL(filePath)
	if len(url) == 0 {
		return 0, errors.New(`line voice url is empty`)
	}
	return a.reply(customer, push, []messaging_api.MessageInterface{
		messaging_api.AudioMessage{OriginalContentUrl: url, Duration: duration},
	})
}

func (a *Application) GetFileByMedia(mediaId string, push *lib_define.PushMessage) ([]byte, http.Header, int, error) {
	return nil, nil, 0, ErrNotSupported
}

func (a *Application) SetTyping(customer, command string) (int, error) {
	return 0, ErrNotSupported
}

func (a *Application) SendMsgOnEvent(code, content string) (int, error) {
	return 0, ErrNotSupported
}

func (a *Application) GetCustomerInfo(customer string) (map[string]any, int, error) {
	return nil, 0, ErrNotSupported
}

func (a *Application) UploadTempImage(filePath string) (string, int, error) {
	return ``, 0, ErrNotSupported
}

func (a *Application) SendUrl(customer, url, title string, push *lib_define.PushMessage) (int, error) {
	return 0, ErrNotSupported
}

func (a *Application) SendMiniProgramPage(customer, appid, title, pagePath, localThumbURL string, push *lib_define.PushMessage) (int, error) {
	return 0, ErrNotSupported
}

func (a *Application) SendImageTextLink(customer, url, title, description, localThumbURL, picurl string, push *lib_define.PushMessage) (int, error) {
	return 0, ErrNotSupported
}

func (a *Application) SendSmartMenu(customer string, smartMenu lib_define.SmartMenu, push *lib_define.PushMessage) (int, error) {
	return 0, ErrNotSupported
}

func (a *Application) GetAccountBasicInfo() (*openresponse.ResponseGetBasicInfo, int, error) {
	return nil, 0, ErrNotSupported
}

// buildMediaURL converts a local upload path into a public https url.
func buildMediaURL(filePath string) string {
	filePath = strings.TrimSpace(filePath)
	filePath = stripLineMediaMeta(filePath)
	if len(filePath) == 0 {
		return ``
	}
	if isURL(filePath) {
		if !isHTTPSURL(filePath) {
			logs.Error(`line media url must use https:%s`, filePath)
			return ``
		}
		return filePath
	}
	domain := strings.TrimRight(ImageUploadDomain, `/`)
	if len(domain) == 0 {
		logs.Error(`line media url domain is empty`)
		return ``
	}
	url := domain + `/` + strings.TrimLeft(filePath, `/`)
	if !isHTTPSURL(url) {
		logs.Error(`line media url must use https:%s`, url)
		return ``
	}
	return url
}

func lineAudioDuration(filePath string) int64 {
	parts := strings.SplitN(filePath, `#duration=`, 2)
	if len(parts) != 2 {
		logs.Error(`line audio duration missing, fallback to 60000ms:%s`, filePath)
		return 60000
	}
	duration := cast.ToInt64(parts[1])
	if duration <= 0 {
		logs.Error(`line audio duration invalid, fallback to 60000ms:%s`, filePath)
		return 60000
	}
	return duration
}

func stripLineMediaMeta(filePath string) string {
	if idx := strings.Index(filePath, `#duration=`); idx >= 0 {
		return filePath[:idx]
	}
	return filePath
}

func isURL(s string) bool {
	return len(s) >= 7 && (s[:7] == `http://` || (len(s) >= 8 && s[:8] == `https://`))
}

func isHTTPSURL(s string) bool {
	return strings.HasPrefix(s, `https://`)
}

// IssueChannelToken issues a short-lived channel access token via Channel ID/Secret.
func IssueChannelToken(channelID, channelSecret string) (string, int, error) {
	client, err := channel_access_token.NewChannelAccessTokenAPI(channel_access_token.WithEndpoint(APIBase))
	if err != nil {
		return ``, 0, err
	}
	resp, err := client.IssueChannelToken(`client_credentials`, channelID, channelSecret)
	if err != nil {
		return ``, 0, err
	}
	return resp.AccessToken, int(resp.ExpiresIn), nil
}

// SetWebhookEndpoint registers the webhook endpoint URL on the LINE channel.
func SetWebhookEndpoint(channelID, channelSecret, endpoint string) error {
	token, _, err := IssueChannelToken(channelID, channelSecret)
	if err != nil {
		return err
	}
	client, err := messaging_api.NewMessagingApiAPI(token, messaging_api.WithEndpoint(APIBase))
	if err != nil {
		return err
	}
	_, err = client.SetWebhookEndpoint(&messaging_api.SetWebhookEndpointRequest{Endpoint: endpoint})
	return err
}
