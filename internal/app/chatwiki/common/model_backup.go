// Copyright © 2016- 2025 Wuhan Sesame Small Customer Service Network Technology Co., Ltd.

package common

import (
	"chatwiki/internal/app/chatwiki/define"
	"chatwiki/internal/pkg/lib_redis"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cast"
	"github.com/zhimaAi/go_tools/logs"
	"github.com/zhimaAi/go_tools/msql"
	"github.com/zhimaAi/go_tools/tool"
	"github.com/zhimaAi/llm_adaptor/v2/chat"
)

// ErrModelUnavailable indicates self-owned model availability failure
// such as integral exhausted or model offline. This type of error
// is recoverable via backup model and should be treated as provider error.
var ErrModelUnavailable = errors.New(`self-owned model unavailable`)

// WrapModelUnavailable wraps self-owned availability error
func WrapModelUnavailable(detail string) error {
	return fmt.Errorf(`%w: %s`, ErrModelUnavailable, detail)
}

// IsModelUnavailable checks if error is self-owned availability failure
func IsModelUnavailable(err error) bool {
	return errors.Is(err, ErrModelUnavailable)
}

// precheckErrStage maps handler-build error to error stage
func precheckErrStage(err error) ModelErrStage {
	if IsModelUnavailable(err) {
		return ModelErrProvider
	}
	return ModelErrPrecheck
}

type BackupModelConfigCacheBuildHandler struct{ AdminUserId int }

func (h *BackupModelConfigCacheBuildHandler) GetCacheKey() string {
	return fmt.Sprintf(`chatwiki.model_backup_config.%d`, h.AdminUserId)
}

func (h *BackupModelConfigCacheBuildHandler) GetCacheData() (any, error) {
	return msql.Model(`model_backup_config`, define.Postgres).
		Where(`admin_user_id`, cast.ToString(h.AdminUserId)).Find()
}

func GetBackupModelConfig(adminUserId int) (msql.Params, error) {
	result := make(msql.Params)
	err := lib_redis.GetCacheWithBuild(define.Redis, &BackupModelConfigCacheBuildHandler{AdminUserId: adminUserId}, &result, time.Hour*12)
	return result, err
}

func DelBackupModelConfigCache(adminUserId int) {
	lib_redis.DelCacheData(define.Redis, &BackupModelConfigCacheBuildHandler{AdminUserId: adminUserId})
}

func SetBackupModelConfig(adminUserId, modelConfigId int, useModel string) error {
	defer DelBackupModelConfigCache(adminUserId)
	if modelConfigId == 0 {
		_, err := msql.Model(`model_backup_config`, define.Postgres).
			Where(`admin_user_id`, cast.ToString(adminUserId)).Delete()
		if err != nil {
			logs.Error(err.Error())
		}
		return err
	}
	nowTime := tool.Time2Int()
	sql := `INSERT INTO model_backup_config
		(admin_user_id,model_config_id,use_model,create_time,update_time)
		VALUES ($1,$2,$3,$4,$4)
		ON CONFLICT (admin_user_id)
		DO UPDATE SET model_config_id = EXCLUDED.model_config_id, use_model = EXCLUDED.use_model, update_time = EXCLUDED.update_time`
	if _, err := msql.RawExec(define.Postgres, sql, nil,
		adminUserId, modelConfigId, useModel, nowTime); err != nil {
		logs.Error(err.Error())
		return err
	}
	return nil
}

type ModelErrStage int

const (
	ModelErrNone ModelErrStage = iota
	ModelErrPrecheck
	ModelErrProvider
	ModelErrStreamRead
)

func logModelError(lang string, adminUserId, modelConfigId int, useModel string, robot msql.Params, errMsg string) {
	config, err := GetModelConfigInfo(modelConfigId, adminUserId)
	if err != nil || len(config) == 0 {
		logs.Error(`logModelError: model config not found, modelConfigId=` + cast.ToString(modelConfigId))
		return
	}
	LlmLogModelError(lang, adminUserId, config, useModel, robot, errMsg)
}

func getUsableBackupModel(lang string, adminUserId, primaryConfigId int, primaryUseModel string, messages []chat.Message, functionTools []chat.Tool) (int, string, bool) {
	backup, err := GetBackupModelConfig(adminUserId)
	if err != nil || len(backup) == 0 {
		return 0, ``, false
	}
	backupConfigId := cast.ToInt(backup[`model_config_id`])
	backupUseModel := backup[`use_model`]
	if backupConfigId == 0 || len(backupUseModel) == 0 {
		return 0, ``, false
	}
	if backupConfigId == primaryConfigId && backupUseModel == primaryUseModel {
		return 0, ``, false
	}
	modelInfo, ok := GetModelInfoByConfig(lang, adminUserId, backupConfigId)
	if !ok {
		return 0, ``, false
	}
	var llmCfg *UseModelConfig
	for i := range modelInfo.UseModelConfigs {
		if modelInfo.UseModelConfigs[i].ModelType == Llm && modelInfo.UseModelConfigs[i].UseModelName == backupUseModel {
			llmCfg = &modelInfo.UseModelConfigs[i]
			break
		}
	}
	if llmCfg == nil {
		return 0, ``, false
	}
	if !backupModelCapable(llmCfg, messages, functionTools) {
		return 0, ``, false
	}
	return backupConfigId, backupUseModel, true
}

func backupModelCapable(cfg *UseModelConfig, messages []chat.Message, functionTools []chat.Tool) bool {
	if len(functionTools) > 0 && cfg.FunctionCall == 0 {
		return false
	}
	var needImage, needVoice, needVideo bool
	for _, msg := range messages {
		if msg.Role != chat.RoleUser || msg.Content.Text == nil {
			continue
		}
		questionMultiple, ok := ParseInputQuestion(*msg.Content.Text)
		if !ok {
			continue
		}
		for _, item := range questionMultiple {
			switch item.Type {
			case chat.ContentPartImageURL:
				needImage = true
			case chat.ContentPartInputAudio:
				needVoice = true
			case chat.ContentPartVideoURL:
				needVideo = true
			}
		}
	}
	if needImage && cfg.InputImage == 0 {
		return false
	}
	if needVoice && cfg.InputVoice == 0 {
		return false
	}
	if needVideo && cfg.InputVideo == 0 {
		return false
	}
	return true
}
