// Copyright © 2016- 2025 Wuhan Sesame Small Customer Service Network Technology Co., Ltd.

package common

import (
	"errors"
	"reflect"
	"unicode/utf8"

	"github.com/zhimaAi/go_tools/logs"
	"github.com/zhimaAi/go_tools/tool"
	llm "github.com/zhimaAi/llm_adaptor/v2"
)

const (
	llmAdaptorLogName            = `llm_adaptor`
	llmAdaptorLogJSONMaxBytes    = 16 * 1024
	llmAdaptorLogTruncatedSuffix = `...`

	llmAdaptorAPIChatCreate           = `Chat.Create`
	llmAdaptorAPIChatStream           = `Chat.Stream`
	llmAdaptorAPIEmbeddingCreate      = `Embeddings.Create`
	llmAdaptorAPIRerankCreate         = `Rerank.Create`
	llmAdaptorAPIImageGenerate        = `Images.Generate`
	llmAdaptorAPIImageEdit            = `Images.Edit`
	llmAdaptorAPIImageStream          = `Images.Stream`
	llmAdaptorAPIImageEditStream      = `Images.EditStream`
	llmAdaptorAPISpeechCreate         = `Speech.Create`
	llmAdaptorAPISpeechListVoices     = `Speech.ListVoices`
	llmAdaptorAPISpeechUploadVoice    = `Speech.UploadVoiceFile`
	llmAdaptorAPISpeechCloneVoice     = `Speech.CloneVoice`
	llmAdaptorAPISpeechCloneFromFiles = `Speech.CloneVoiceFromFiles`
)

func logLLMAdaptorError(api string, request, response any, err error) {
	if err == nil {
		return
	}
	logs.Other(
		llmAdaptorLogName,
		`api:%s,req:%s,resp:%s,err:%s`,
		api,
		truncateLLMAdaptorLogJSON(tool.JsonEncodeNoError(request)),
		truncateLLMAdaptorLogJSON(llmAdaptorResponseLog(response, err)),
		err.Error(),
	)
}

func truncateLLMAdaptorLogJSON(value string) string {
	if len(value) <= llmAdaptorLogJSONMaxBytes {
		return value
	}
	end := llmAdaptorLogJSONMaxBytes - len(llmAdaptorLogTruncatedSuffix)
	for end > 0 && !utf8.RuneStart(value[end]) {
		end--
	}
	return value[:end] + llmAdaptorLogTruncatedSuffix
}

func llmAdaptorResponseLog(response any, err error) string {
	if !isNilLLMAdaptorResponse(response) {
		return tool.JsonEncodeNoError(response)
	}
	var apiError *llm.APIError
	if errors.As(err, &apiError) && len(apiError.Raw) > 0 {
		return string(apiError.Raw)
	}
	return tool.JsonEncodeNoError(response)
}

func isNilLLMAdaptorResponse(response any) bool {
	if response == nil {
		return true
	}
	value := reflect.ValueOf(response)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}
