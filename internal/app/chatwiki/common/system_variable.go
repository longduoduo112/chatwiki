// Copyright © 2016- 2025 Wuhan Sesame Small Customer Service Network Technology Co., Ltd.

package common

import (
	"strings"
	"time"

	"chatwiki/internal/app/chatwiki/i18n"
	"github.com/zhimaAi/go_tools/logs"
	"github.com/zhimaAi/go_tools/tool"
)

const (
	// SystemVarKeyCurrentTime is the backend key of the built-in "current time" system variable.
	// It is an English identifier so that the frontend can localize the display name via i18n.
	SystemVarKeyCurrentTime = "current_time"
	// SystemVarKeyCurrentDate is the backend key of the built-in "current date" system variable.
	SystemVarKeyCurrentDate = "current_date"
	// SystemVarPlaceholderCurrentTime is the literal placeholder inserted into the prompt.
	// It is an English protocol string exchanged between backend and frontend through the
	// GetChatVariables interface; the frontend inserts it as-is and the backend replaces it.
	// Keep it identical to what the frontend receives from the placeholder field.
	SystemVarPlaceholderCurrentTime = "【system_var:current_time】"
	// SystemVarPlaceholderCurrentDate is the literal placeholder of the "current date" system variable.
	SystemVarPlaceholderCurrentDate = "【system_var:current_date】"
)

// SystemVariable describes a built-in system variable injected by the backend.
// Unlike chat variables, system variables require no user input and are not persisted.
// Name/Key are English identifiers; the localized display name is rendered by the frontend.
type SystemVariable struct {
	Key         string
	Name        string
	Type        string // fixed to system_time
	Placeholder string // literal placeholder exchanged with the frontend
}

// GetBuiltinSystemVariables returns all built-in system variables (currently "current time" and "current date").
// Extend this slice to add more system variables in the future (e.g. random seed)
// without touching the replacement call sites. The display Name is localized by lang via i18n;
// Key/Type are stable English identifiers used by the frontend and the replacement logic.
func GetBuiltinSystemVariables(lang string) []SystemVariable {
	return []SystemVariable{
		{Key: SystemVarKeyCurrentTime, Name: i18n.Show(lang, "system_var_current_time"), Type: "system_time", Placeholder: SystemVarPlaceholderCurrentTime},
		{Key: SystemVarKeyCurrentDate, Name: i18n.Show(lang, "system_var_current_date"), Type: "system_time", Placeholder: SystemVarPlaceholderCurrentDate},
	}
}

// buildCurrentTimeValue generates the current time in a language-dependent format.
// The format layout is resolved per lang via i18n (key "system_var_time_format"),
// e.g. zh-CN: "2026年10月12日 16时10分", en-US: "October 12, 2026 16:10".
// The layout follows the Go reference time (2006-01-02 15:04:05).
func buildCurrentTimeValue(lang string) string {
	return time.Now().Format(i18n.Show(lang, `system_var_time_format`))
}

// buildCurrentDateValue generates the current date in a language-dependent format.
// The format layout is resolved per lang via i18n (key "system_var_date_format"),
// e.g. zh-CN: "2026年10月12日", en-US: "October 12, 2026".
func buildCurrentDateValue(lang string) string {
	return time.Now().Format(i18n.Show(lang, `system_var_date_format`))
}

// systemVariableReplaces returns the placeholder -> value mapping for all built-in system variables.
func systemVariableReplaces(lang string) map[string]string {
	return map[string]string{
		SystemVarPlaceholderCurrentTime: buildCurrentTimeValue(lang),
		SystemVarPlaceholderCurrentDate: buildCurrentDateValue(lang),
	}
}

// ReplaceSystemVariables replaces system variable placeholders in both the custom prompt
// and the structured prompt_struct with real-time computed values.
// It is independent from ReplaceChatVariables (which handles user-fillable session variables),
// keeping each function focused on a single variable domain.
func ReplaceSystemVariables(lang string, prompt *string, promptStruct *string) {
	replaces := systemVariableReplaces(lang)
	// custom prompt
	for k, v := range replaces {
		*prompt = strings.ReplaceAll(*prompt, k, v)
	}
	// structured prompt
	sp := StructPrompt{}
	if err := tool.JsonDecodeUseNumber(*promptStruct, &sp); err != nil {
		logs.Error("promptStruct:%s,err:%v", *promptStruct, err)
		return
	}
	for k, v := range replaces {
		sp.Role.Describe = strings.ReplaceAll(sp.Role.Describe, k, v)
		sp.Task.Describe = strings.ReplaceAll(sp.Task.Describe, k, v)
		sp.Constraints.Describe = strings.ReplaceAll(sp.Constraints.Describe, k, v)
		sp.Skill.Describe = strings.ReplaceAll(sp.Skill.Describe, k, v)
		sp.Output.Describe = strings.ReplaceAll(sp.Output.Describe, k, v)
		sp.Tone.Describe = strings.ReplaceAll(sp.Tone.Describe, k, v)
		for i := range sp.Custom {
			sp.Custom[i].Describe = strings.ReplaceAll(sp.Custom[i].Describe, k, v)
		}
	}
	*promptStruct = tool.JsonEncodeNoError(sp)
}

// GetSystemVariablePlaceholder exposes the placeholder literal of a given system variable
// to the frontend, avoiding hard-coded divergence between frontend insertion and backend replacement.
func GetSystemVariablePlaceholder(key string) string {
	placeholders := map[string]string{
		SystemVarKeyCurrentTime: SystemVarPlaceholderCurrentTime,
		SystemVarKeyCurrentDate: SystemVarPlaceholderCurrentDate,
	}
	return placeholders[key]
}
