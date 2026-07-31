// Copyright © 2016- 2025 Wuhan Sesame Small Customer Service Network Technology Co., Ltd.

package custom_eino

import (
	"context"
	"fmt"
	"strings"

	einotool "github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/zhimaAi/go_tools/tool"
)

const KbsearchToolName = "search_knowledge"

type KbsearchLibrary struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Introduction string `json:"introduction"`
}

type KbsearchTool struct {
	libraries []KbsearchLibrary
	do        func(query string, libraryIds []string) (string, error)
}

func BuildKbsearchTool(libraries []KbsearchLibrary, do func(query string, libraryIds []string) (string, error)) einotool.BaseTool {
	return &KbsearchTool{libraries: libraries, do: do}
}

func (t *KbsearchTool) Info(_ context.Context) (*schema.ToolInfo, error) {
	libraryIds := make([]string, 0, len(t.libraries))
	for _, library := range t.libraries {
		libraryIds = append(libraryIds, library.ID)
	}
	return &schema.ToolInfo{
		Name: KbsearchToolName,
		Desc: fmt.Sprintf(
			"Search one or more selected local knowledge bases for passages relevant to a user question. Select the most relevant knowledge bases by their name and introduction. Available knowledge bases: %s",
			tool.JsonEncodeNoError(t.libraries),
		),
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"query": {
				Type:     schema.String,
				Desc:     "Search query or user question",
				Required: true,
			},
			"library_ids": {
				Type: schema.Array,
				ElemInfo: &schema.ParameterInfo{
					Type: schema.String,
					Enum: libraryIds,
				},
				Desc:     "IDs of one or more knowledge bases selected from the available knowledge bases in the tool description",
				Required: true,
			},
		}),
	}, nil
}

func (t *KbsearchTool) InvokableRun(_ context.Context, argumentsInJSON string, _ ...einotool.Option) (string, error) {
	var input struct {
		Query      string   `json:"query"`
		LibraryIds []string `json:"library_ids"`
	}
	if err := tool.JsonDecodeUseNumber(argumentsInJSON, &input); err != nil {
		return "", fmt.Errorf("parse %s arguments: %w", KbsearchToolName, err)
	}
	query := strings.TrimSpace(input.Query)
	if query == `` {
		return `query can not be empty`, nil
	}
	allowedLibraryIds := make(map[string]struct{}, len(t.libraries))
	availableLibraryIds := make([]string, 0, len(t.libraries))
	for _, library := range t.libraries {
		allowedLibraryIds[library.ID] = struct{}{}
		availableLibraryIds = append(availableLibraryIds, library.ID)
	}
	libraryIds := make([]string, 0, len(input.LibraryIds))
	selectedLibraryIds := make(map[string]struct{}, len(input.LibraryIds))
	for _, libraryId := range input.LibraryIds {
		libraryId = strings.TrimSpace(libraryId)
		if _, ok := allowedLibraryIds[libraryId]; !ok {
			return fmt.Sprintf(`library_ids must only contain available knowledge base IDs: %s`, strings.Join(availableLibraryIds, `,`)), nil
		}
		if _, ok := selectedLibraryIds[libraryId]; ok {
			continue
		}
		selectedLibraryIds[libraryId] = struct{}{}
		libraryIds = append(libraryIds, libraryId)
	}
	if len(libraryIds) == 0 {
		return `library_ids can not be empty`, nil
	}
	return t.do(query, libraryIds)
}
