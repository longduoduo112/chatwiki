// Copyright © 2016- 2025 Wuhan Sesame Small Customer Service Network Technology Co., Ltd.

package common

import (
	"chatwiki/internal/app/chatwiki/define"
	"chatwiki/internal/app/chatwiki/i18n"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/spf13/cast"
	"github.com/zhimaAi/go_tools/tool"
)

type goodsWechatCardTagPayload struct {
	Title    string `json:"title"`
	Appid    string `json:"appid"`
	PagePath string `json:"page_path"`
	ThumbURL string `json:"thumb_url"`
}

// NormalizeGoodsWechatCard trims and validates one goods WeChat mini-program card.
func NormalizeGoodsWechatCard(lang string, card *define.GoodsWechatCard) error {
	if card == nil {
		return nil
	}
	card.Appid = strings.TrimSpace(card.Appid)
	card.Path = strings.TrimSpace(card.Path)
	card.Title = strings.TrimSpace(card.Title)
	card.Image = strings.TrimSpace(card.Image)
	if IsEmptyGoodsWechatCard(*card) {
		return nil
	}
	if card.Appid == `` {
		return goodsWechatCardFieldError(lang, `goods_import_header_card_appid`)
	}
	if card.Path == `` {
		return goodsWechatCardFieldError(lang, `goods_import_header_card_path`)
	}
	if card.Title == `` {
		return goodsWechatCardFieldError(lang, `goods_import_header_card_title`)
	}
	if card.Image == `` {
		return goodsWechatCardFieldError(lang, `goods_import_header_card_image`)
	}
	if utf8.RuneCountInString(card.Appid) > define.GoodsWechatCardAppidMaxLength {
		return goodsWechatCardFieldError(lang, `goods_import_header_card_appid`)
	}
	if utf8.RuneCountInString(card.Path) > define.GoodsWechatCardPathMaxLength {
		return goodsWechatCardFieldError(lang, `goods_import_header_card_path`)
	}
	if utf8.RuneCountInString(card.Title) > define.GoodsWechatCardTitleMaxLength {
		return goodsWechatCardFieldError(lang, `goods_import_header_card_title`)
	}
	if utf8.RuneCountInString(card.Image) > define.GoodsWechatCardImageMaxLength {
		return goodsWechatCardFieldError(lang, `goods_import_header_card_image`)
	}
	images, err := NormalizeGoodsLibImages(lang, []string{card.Image})
	if err != nil || len(images) != 1 {
		return goodsWechatCardFieldError(lang, `goods_import_header_card_image`)
	}
	card.Image = images[0]
	return nil
}

func goodsWechatCardFieldError(lang, field string) error {
	return errors.New(i18n.Show(lang, `param_invalid`, i18n.Show(lang, field)))
}

// IsEmptyGoodsWechatCard reports whether all card fields are empty.
func IsEmptyGoodsWechatCard(card define.GoodsWechatCard) bool {
	return card.Appid == `` && card.Path == `` && card.Title == `` && card.Image == ``
}

func isCompleteGoodsWechatCard(card define.GoodsWechatCard) bool {
	return card.Appid != `` && card.Path != `` && card.Title != `` && card.Image != ``
}

func encodeGoodsWechatCard(card *define.GoodsWechatCard) string {
	if card == nil || IsEmptyGoodsWechatCard(*card) {
		return `{}`
	}
	return tool.JsonEncodeNoError(card)
}

func formatGoodsWechatCard(raw string) map[string]string {
	card := parseGoodsWechatCard(raw)
	if IsEmptyGoodsWechatCard(card) {
		return map[string]string{}
	}
	return map[string]string{
		`appid`: card.Appid,
		`path`:  card.Path,
		`title`: card.Title,
		`image`: card.Image,
	}
}

func parseGoodsWechatCard(value any) define.GoodsWechatCard {
	card := define.GoodsWechatCard{}
	switch data := value.(type) {
	case define.GoodsWechatCard:
		return data
	case *define.GoodsWechatCard:
		if data != nil {
			return *data
		}
		return card
	case map[string]string:
		card.Appid = strings.TrimSpace(data[`appid`])
		card.Path = strings.TrimSpace(data[`path`])
		card.Title = strings.TrimSpace(data[`title`])
		card.Image = strings.TrimSpace(data[`image`])
		return card
	case map[string]any:
		card.Appid = strings.TrimSpace(cast.ToString(data[`appid`]))
		card.Path = strings.TrimSpace(cast.ToString(data[`path`]))
		card.Title = strings.TrimSpace(cast.ToString(data[`title`]))
		card.Image = strings.TrimSpace(cast.ToString(data[`image`]))
		return card
	case string:
		_ = json.Unmarshal([]byte(data), &card)
		card.Appid = strings.TrimSpace(card.Appid)
		card.Path = strings.TrimSpace(card.Path)
		card.Title = strings.TrimSpace(card.Title)
		card.Image = strings.TrimSpace(card.Image)
	}
	return card
}

// ExtractGoodsWechatCardsForReply returns valid cards in goods-result order and deduplicates goods IDs.
func ExtractGoodsWechatCardsForReply(list []map[string]any) []define.GoodsWechatCard {
	result := make([]define.GoodsWechatCard, 0)
	seenGoods := make(map[int64]struct{})
	for _, item := range list {
		goodsID := cast.ToInt64(item[`id`])
		if goodsID > 0 {
			if _, ok := seenGoods[goodsID]; ok {
				continue
			}
		}
		card := parseGoodsWechatCard(item[`goods_wechat_card`])
		if !isCompleteGoodsWechatCard(card) {
			continue
		}
		if goodsID > 0 {
			seenGoods[goodsID] = struct{}{}
		}
		result = append(result, card)
	}
	return result
}

// AppendGoodsWechatCardTags appends goods mini-program card tags to a reply.
func AppendGoodsWechatCardTags(content string, cards []define.GoodsWechatCard) (string, string) {
	tags := make([]string, 0, len(cards))
	for _, card := range cards {
		if !isCompleteGoodsWechatCard(card) {
			continue
		}
		payload := goodsWechatCardTagPayload{
			Title:    card.Title,
			Appid:    card.Appid,
			PagePath: card.Path,
			ThumbURL: card.Image,
		}
		tags = append(tags, fmt.Sprintf("[wx_mini_card]%s[/wx_mini_card]", tool.JsonEncodeNoError(payload)))
	}
	if len(tags) == 0 {
		return content, ``
	}
	appendContent := "\n" + strings.Join(tags, "\n")
	return content + appendContent, appendContent
}
