// Copyright © 2016- 2025 Wuhan Sesame Small Customer Service Network Technology Co., Ltd.

package custom_eino

import (
	"encoding/json"

	"github.com/cloudwego/eino/schema"
	"github.com/eino-contrib/jsonschema"
	orderedmap "github.com/wk8/go-ordered-map/v2"
	"github.com/zhimaAi/go_tools/tool"
	"github.com/zhimaAi/llm_adaptor/v2/chat"
)

func StructConvert[In, Out any](in In) (out Out, err error) {
	var jsonStr string
	jsonStr, err = tool.JsonEncode(in)
	if err != nil {
		return
	}
	err = tool.JsonDecodeUseNumber(jsonStr, &out)
	return
}

func ConvertMessage(message schema.Message) (chat.Message, error) {
	return StructConvert[schema.Message, chat.Message](message)
}

func ConvertTools(info schema.ToolInfo) (result chat.Tool, err error) {
	params, err := info.ParamsOneOf.ToJSONSchema()
	if err != nil {
		return
	}
	result.Type = `function`
	result.Function.Name = info.Name
	result.Function.Description = info.Desc
	if params != nil {
		result.Function.Parameters = json.RawMessage(tool.JsonEncodeNoError(params))
	}
	return
}

func ConvertToolCalls(toolCalls []chat.ToolCall) ([]schema.ToolCall, error) {
	return StructConvert[[]chat.ToolCall, []schema.ToolCall](toolCalls)
}

func ConvertChatResp(response chat.CreateResponse) *schema.Message {
	var choice chat.Choice
	if len(response.Choices) > 0 {
		choice = response.Choices[0]
	}
	result := ``
	if choice.Message.Content.Text != nil {
		result = *choice.Message.Content.Text
	}
	var toolCalls []schema.ToolCall
	if len(choice.Message.ToolCalls) > 0 {
		toolCalls, _ = ConvertToolCalls(choice.Message.ToolCalls)
	}
	totalTokens := response.Usage.TotalTokens
	if totalTokens == 0 {
		totalTokens = response.Usage.PromptTokens + response.Usage.CompletionTokens
	}
	msg := schema.AssistantMessage(result, toolCalls)
	msg.ReasoningContent = choice.Message.ReasoningContent
	msg.ResponseMeta = &schema.ResponseMeta{
		FinishReason: choice.FinishReason,
		Usage: &schema.TokenUsage{
			PromptTokens:     response.Usage.PromptTokens,
			CompletionTokens: response.Usage.CompletionTokens,
			TotalTokens:      totalTokens,
		},
	}
	return msg
}

func ConvertStreamChunk(chunk *chat.StreamChunk) *schema.Message {
	if chunk == nil {
		return schema.AssistantMessage(``, nil)
	}
	if len(chunk.Choices) == 0 {
		message := schema.AssistantMessage(``, nil)
		if chunk.Usage != nil {
			message.ResponseMeta = &schema.ResponseMeta{Usage: &schema.TokenUsage{
				PromptTokens: chunk.Usage.PromptTokens, CompletionTokens: chunk.Usage.CompletionTokens, TotalTokens: chunk.Usage.TotalTokens,
			}}
		}
		return message
	}
	response := chat.CreateResponse{
		Choices: []chat.Choice{{
			Index:        chunk.Choices[0].Index,
			Message:      chunk.Choices[0].Delta,
			FinishReason: chunk.Choices[0].FinishReason,
			LogProbs:     chunk.Choices[0].LogProbs,
		}},
	}
	if chunk.Usage != nil {
		response.Usage = *chunk.Usage
	}
	return ConvertChatResp(response)
}

func ConvertProperties(properties any) (*orderedmap.OrderedMap[string, *jsonschema.Schema], error) {
	return StructConvert[any, *orderedmap.OrderedMap[string, *jsonschema.Schema]](properties)
}
