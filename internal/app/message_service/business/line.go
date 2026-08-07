// Copyright © 2016- 2025 Wuhan Sesame Small Customer Service Network Technology Co., Ltd.

package business

import (
	"chatwiki/internal/app/message_service/common"
	"chatwiki/internal/app/message_service/define"
	"chatwiki/internal/pkg/lib_define"
	"chatwiki/internal/pkg/wechat/line"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
	"github.com/spf13/cast"
	"github.com/zhimaAi/go_tools/logs"
	"github.com/zhimaAi/go_tools/tool"
)

func LinePush(c *gin.Context) {
	accessKey := strings.TrimSpace(c.Param(`access_key`))
	appInfo, err := common.GetWechatAppInfo(`access_key`, accessKey)
	if err != nil {
		logs.Error(err.Error())
		c.String(http.StatusOK, lib_define.SUCCESS)
		return
	}
	if len(appInfo) == 0 || appInfo[`app_type`] != lib_define.AppLine {
		c.String(http.StatusOK, lib_define.SUCCESS)
		return
	}
	// GET: LINE webhook verification needs only 200 OK
	if c.Request.Method == http.MethodGet {
		c.String(http.StatusOK, lib_define.SUCCESS)
		return
	}

	channelSecret := appInfo[`app_secret`]
	cb, err := webhook.ParseRequest(channelSecret, c.Request)
	if err != nil {
		logs.Error(`line webhook parse err:%s`, err.Error())
		c.String(http.StatusOK, lib_define.SUCCESS)
		return
	}

	channelId := appInfo[`app_id`]
	adminUserId := cast.ToInt(appInfo[`admin_user_id`])
	for _, event := range cb.Events {
		// Dedup: LINE redelivers the same event when our 200 is slow/lost; SetNX on the ULID
		// webhookEventId ensures we process it only once.
		if eventId := lineWebhookEventId(event); len(eventId) > 0 {
			ok, dErr := define.Redis.SetNX(context.Background(),
				define.LineInboundDedup+eventId, `1`, 10*time.Minute).Result()
			if dErr != nil {
				logs.Error(`line dedup setnx failed event_id:%s err:%s`, eventId, dErr.Error())
			} else if !ok {
				logs.Info(`line skip duplicate event_id:%s`, eventId)
				continue
			}
		}
		go func(event webhook.EventInterface) {
			nsqMsg := buildLineNSQMsg(channelId, channelSecret, adminUserId, event)
			if nsqMsg == nil {
				return
			}
			common.PushNSQ(nsqMsg)
		}(event)
	}
	c.String(http.StatusOK, lib_define.SUCCESS)
}

// lineWebhookEventId extracts the ULID webhookEventId from LINE message events.
func lineWebhookEventId(event webhook.EventInterface) string {
	switch e := event.(type) {
	case webhook.MessageEvent:
		return e.WebhookEventId
	}
	return ``
}

// buildLineNSQMsg converts a LINE message event into a unified NSQ message.
func buildLineNSQMsg(channelId, channelSecret string, adminUserId int, event webhook.EventInterface) map[string]any {
	switch e := event.(type) {
	case webhook.MessageEvent:
		userId := lineSourceUserId(e.Source)
		if len(userId) == 0 {
			return nil
		}
		msgType, content, ossUrl, msgId := parseLineMessage(e.Message, channelId, channelSecret, adminUserId)
		if len(msgType) == 0 {
			return nil
		}
		nsqMsg := map[string]any{
			`appid`:        channelId,
			`ToUserName`:   channelId,
			`FromUserName`: userId,
			`CreateTime`:   e.Timestamp,
			`MsgId`:        msgId,
			`MsgType`:      msgType,
			`Content`:      content,
			`replyToken`:   e.ReplyToken,
		}
		if len(ossUrl) > 0 {
			nsqMsg[`oss_url`] = ossUrl
		}
		return nsqMsg
	}
	return nil
}

// lineSourceUserId extracts userId from a LINE event source.
func lineSourceUserId(source webhook.SourceInterface) string {
	if s, ok := source.(webhook.UserSource); ok {
		return s.UserId
	}
	return ``
}

// parseLineMessage normalizes basic LINE messages on the receiving side.
func parseLineMessage(message webhook.MessageContentInterface, channelId, channelSecret string, adminUserId int) (msgType, content, ossUrl, msgId string) {
	switch m := message.(type) {
	case webhook.TextMessageContent:
		return lib_define.MsgTypeText, m.Text, ``, m.Id
	case webhook.ImageMessageContent:
		ossUrl = lineContentProviderURL(m.ContentProvider)
		if len(ossUrl) == 0 {
			ossUrl = downloadLineMedia(m.Id, channelId, channelSecret, adminUserId, `.jpg`)
		}
		if len(ossUrl) == 0 {
			return lib_define.MsgTypeText, `[media message]`, ``, m.Id
		}
		return lib_define.MsgTypeImage, ``, ossUrl, m.Id
	case webhook.VideoMessageContent:
		ossUrl = lineContentProviderURL(m.ContentProvider)
		if len(ossUrl) == 0 {
			ossUrl = downloadLineMedia(m.Id, channelId, channelSecret, adminUserId, `.mp4`)
		}
		if len(ossUrl) == 0 {
			return lib_define.MsgTypeText, `[media message]`, ``, m.Id
		}
		return lib_define.MsgTypeVideo, ``, ossUrl, m.Id
	case webhook.AudioMessageContent:
		// LINE audio -> normalized to system voice type
		ossUrl = lineContentProviderURL(m.ContentProvider)
		if len(ossUrl) == 0 {
			ossUrl = downloadLineMedia(m.Id, channelId, channelSecret, adminUserId, `.m4a`)
		}
		if len(ossUrl) == 0 {
			return lib_define.MsgTypeText, `[media message]`, ``, m.Id
		}
		return lib_define.MsgTypeVoice, ``, ossUrl, m.Id
	}
	logs.Warning(`line unsupported message content type:%T`, message)
	return `unsupported`, `unsupported message type`, ``, ``
}

func lineContentProviderURL(provider *webhook.ContentProvider) string {
	if provider == nil || provider.Type != webhook.ContentProviderTYPE_EXTERNAL {
		return ``
	}
	if len(provider.OriginalContentUrl) > 0 {
		return provider.OriginalContentUrl
	}
	return provider.PreviewImageUrl
}

// downloadLineMedia downloads LINE media via the blob API and stores it locally, returning the
// public url or empty string on failure. LINE answers 202 while video/audio is still transcoding
// (retried up to ~2min), while network/other errors fail fast so a flaky image does not block the
// chat. Runs in the webhook goroutine, so waiting only delays the NSQ push.
func downloadLineMedia(messageId, channelId, channelSecret string, adminUserId int, defaultExt string) string {
	token, err := line.GetChannelToken(channelId, channelSecret)
	if err != nil {
		logs.Error(`line get channel token err:%s`, err.Error())
		return ``
	}
	// no whole-request Client.Timeout: it also covers body streaming and cuts big videos off
	// mid-download, only connect and first-byte latency are bounded here
	httpClient := &http.Client{
		Transport: &http.Transport{
			DialContext:           (&net.Dialer{Timeout: 10 * time.Second}).DialContext,
			ResponseHeaderTimeout: 30 * time.Second,
		},
	}
	client, err := messaging_api.NewMessagingApiBlobAPI(token,
		messaging_api.WithBlobEndpoint(line.DataAPIBase),
		messaging_api.WithBlobHTTPClient(httpClient))
	if err != nil {
		logs.Error(`line blob api err:%s`, err.Error())
		return ``
	}

	const transcodeMax = 40 // 202 transcoding: 40 x 3s ~ 2min
	const transcodeInterval = 3 * time.Second
	const netErrMax = 3 // network/non-202 errors: fail fast
	const netErrInterval = 1 * time.Second
	transcodeAttempts, netErrAttempts := 0, 0
	var resp *http.Response
	for {
		resp, err = client.GetMessageContent(messageId)
		if err != nil {
			netErrAttempts++
			if netErrAttempts > netErrMax {
				logs.Error(`line download media failed (network) after %d attempts, message_id:%s err:%s`, netErrMax, messageId, err.Error())
				return ``
			}
			logs.Warning(`line get content net err attempt:%d err:%s message_id:%s`, netErrAttempts, err.Error(), messageId)
			time.Sleep(netErrInterval)
			continue
		}
		if resp.StatusCode == http.StatusOK {
			break
		}
		// non-200: 202 = transcoding (long wait), anything else = fail-fast transient error
		resp.Body.Close()
		if resp.StatusCode == http.StatusAccepted {
			transcodeAttempts++
			if transcodeAttempts > transcodeMax {
				logs.Error(`line download media failed (transcoding) after %d attempts, message_id:%s`, transcodeMax, messageId)
				return ``
			}
			if transcodeAttempts%5 == 0 { // log every 15s
				logs.Info(`line message content transcoding, retry attempt:%d message_id:%s`, transcodeAttempts, messageId)
			}
			time.Sleep(transcodeInterval)
			continue
		}
		netErrAttempts++
		if netErrAttempts > netErrMax {
			logs.Error(`line download media failed (status:%d) after %d attempts, message_id:%s`, resp.StatusCode, netErrMax, messageId)
			return ``
		}
		logs.Warning(`line get content status:%d attempt:%d message_id:%s`, resp.StatusCode, netErrAttempts, messageId)
		time.Sleep(netErrInterval)
	}
	defer resp.Body.Close()

	ext := extFromContentType(resp.Header.Get(`Content-Type`), defaultExt)
	objectKey := fmt.Sprintf("chat_ai/%d/received_message_images/%s/%s%s", adminUserId, tool.Date(`Ym`), tool.MD5(messageId), ext)
	localFile := `internal/app/chatwiki/upload/` + objectKey
	if tool.IsFile(localFile) {
		return `【image_domain】/upload/` + objectKey
	}
	if err := os.MkdirAll(filepath.Dir(localFile), 0755); err != nil {
		logs.Error(`create line media dir err:%s, path:%s`, err.Error(), filepath.Dir(localFile))
		return ``
	}
	f, err := os.Create(localFile)
	if err != nil {
		logs.Error(`create line media file err:%s`, err.Error())
		return ``
	}
	defer f.Close()
	if _, err := io.Copy(f, resp.Body); err != nil {
		logs.Error(`write line media err:%s`, err.Error())
		_ = os.Remove(localFile)
		return ``
	}
	return `【image_domain】/upload/` + objectKey
}

// extFromContentType infers a file extension from a Content-Type header.
func extFromContentType(contentType, defaultExt string) string {
	contentType = strings.ToLower(strings.TrimSpace(contentType))
	if idx := strings.Index(contentType, `;`); idx >= 0 {
		contentType = strings.TrimSpace(contentType[:idx])
	}
	switch contentType {
	case `image/jpeg`, `image/jpg`:
		return `.jpg`
	case `image/png`:
		return `.png`
	case `image/gif`:
		return `.gif`
	case `image/webp`:
		return `.webp`
	case `video/mp4`:
		return `.mp4`
	case `audio/mp4`, `audio/x-m4a`:
		return `.m4a`
	case `audio/aac`:
		return `.aac`
	case `audio/mpeg`:
		return `.mp3`
	}
	return defaultExt
}
