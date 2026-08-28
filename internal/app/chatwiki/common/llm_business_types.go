// Copyright © 2016- 2025 Wuhan Sesame Small Customer Service Network Technology Co., Ltd.

package common

import (
	"strings"

	"github.com/zhimaAi/llm_adaptor/v2/chat"
)

type ChatResponse struct {
	chat.CreateResponse
	IsValidFunctionCall bool `json:"is_valid_function_call"`
}

func NewChatResponse(response *chat.CreateResponse) ChatResponse {
	if response == nil {
		return ChatResponse{}
	}
	result := ChatResponse{CreateResponse: *response}
	result.Choices = append([]chat.Choice(nil), response.Choices...)
	return result
}

func (r ChatResponse) Result() string {
	if len(r.Choices) == 0 || r.Choices[0].Message.Content.Text == nil {
		return ``
	}
	return *r.Choices[0].Message.Content.Text
}

func (r *ChatResponse) SetResult(value string) {
	if len(r.Choices) == 0 {
		r.Choices = append(r.Choices, chat.Choice{Message: chat.Message{Role: chat.RoleAssistant}})
	}
	r.Choices[0].Message.Content = chat.TextContent(value)
}

func (r *ChatResponse) NormalizeJSONResult() {
	value := strings.TrimSpace(r.Result())
	value = strings.TrimPrefix(value, "```json")
	value = strings.TrimPrefix(value, "```")
	value = strings.TrimSuffix(value, "```")
	r.SetResult(strings.TrimSpace(value))
}

func (r ChatResponse) ToolCalls() []chat.ToolCall {
	if len(r.Choices) == 0 {
		return nil
	}
	return r.Choices[0].Message.ToolCalls
}

func (r ChatResponse) ReasoningContent() string {
	if len(r.Choices) == 0 {
		return ``
	}
	return r.Choices[0].Message.ReasoningContent
}

func (r ChatResponse) FinishReason() string {
	if len(r.Choices) == 0 {
		return ``
	}
	return r.Choices[0].FinishReason
}
