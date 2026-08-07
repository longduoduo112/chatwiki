// Copyright © 2016- 2025 Wuhan Sesame Small Customer Service Network Technology Co., Ltd.

package manage

import (
	"chatwiki/internal/app/chatwiki/common"
	"chatwiki/internal/app/chatwiki/define"
	"chatwiki/internal/app/chatwiki/i18n"
	"chatwiki/internal/app/chatwiki/middlewares"
	"chatwiki/internal/pkg/lib_redis"
	"chatwiki/internal/pkg/lib_web"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"github.com/zhimaAi/go_tools/logs"
)

func buildGeneratedSkillInstallResult(c *gin.Context, adminUserId int, taskType string, taskId, robotId int64, robotKey string, item *common.ClawbotUserSkillItem) *common.ClawbotSkillInstallResult {
	result := common.NewClawbotSkillInstallResult(item, common.ClawbotSkillBindStatusSkipped, ``)
	if robotKey == `` {
		return result
	}
	robotSkillId, errKey, err := common.BindClawbotUserSkillToRobot(adminUserId, robotKey, item.SkillId)
	if err != nil || errKey != `` {
		bindMessage := i18n.Show(common.GetLang(c), `sys_err`)
		if errKey != `` {
			bindMessage = i18n.Show(common.GetLang(c), errKey)
		}
		logs.Error(`generated skill bind failed,type:%s,task_id:%d,robot_id:%d,skill_id:%d,err_key:%s,err:%v`, taskType, taskId, robotId, item.SkillId, errKey, err)
		result.BindStatus = common.ClawbotSkillBindStatusFailed
		result.BindMessage = bindMessage
		return result
	}
	result.RobotSkillId = robotSkillId
	result.IsSelected = 1
	result.BindStatus = common.ClawbotSkillBindStatusBound
	return result
}

func lockWebToSkillTaskOperation(c *gin.Context, id int64) func() {
	lockKey := define.LockPreKey + `WebToSkillTask.` + cast.ToString(id)
	if !lib_redis.AddLock(define.Redis, lockKey, time.Minute*10) {
		common.FmtError(c, `op_lock`)
		return nil
	}
	return func() {
		lib_redis.UnLock(define.Redis, lockKey)
	}
}

func GetWebToSkillTaskList(c *gin.Context) {
	adminUserId := GetAdminUserId(c)
	if adminUserId == 0 {
		return
	}
	filter := define.WebToSkillTaskListFilter{
		Status: -1,
		Page:   1,
		Size:   define.WebToSkillTaskDefaultPageSize,
	}
	if err := c.ShouldBindQuery(&filter); err != nil {
		common.FmtError(c, `param_err`, middlewares.GetValidateErr(filter, err, common.GetLang(c)).Error())
		return
	}
	data, err := common.GetWebToSkillTaskList(common.GetLang(c), adminUserId, filter)
	if err != nil {
		c.String(http.StatusOK, lib_web.FmtJson(nil, err))
		return
	}
	common.FmtOk(c, data)
}

func CreateWebToSkillTask(c *gin.Context) {
	adminUserId := GetAdminUserId(c)
	if adminUserId == 0 {
		return
	}
	params := define.WebToSkillTaskCreateParams{}
	if err := common.RequestParamsBind(&params, c); err != nil {
		common.FmtError(c, `param_err`, middlewares.GetValidateErr(params, err, common.GetLang(c)).Error())
		return
	}
	params.Urls = append(params.Urls, c.PostFormArray(`urls[]`)...)
	if len(params.Urls) == 0 {
		params.Urls = common.ParseWebToSkillURLText(c.PostForm(`urls`))
	}
	id, err := common.CreateWebToSkillTask(common.GetLang(c), adminUserId, params)
	if err != nil {
		c.String(http.StatusOK, lib_web.FmtJson(nil, err))
		return
	}
	common.FmtOk(c, id)
}

func StopWebToSkillTask(c *gin.Context) {
	adminUserId := GetAdminUserId(c)
	if adminUserId == 0 {
		return
	}
	params := define.WebToSkillTaskIDParams{}
	if err := common.RequestParamsBind(&params, c); err != nil {
		common.FmtError(c, `param_err`, middlewares.GetValidateErr(params, err, common.GetLang(c)).Error())
		return
	}
	unlock := lockWebToSkillTaskOperation(c, params.ID)
	if unlock == nil {
		return
	}
	defer unlock()
	data, err := common.StopWebToSkillTask(common.GetLang(c), adminUserId, params.ID)
	if err != nil {
		c.String(http.StatusOK, lib_web.FmtJson(nil, err))
		return
	}
	common.FmtOk(c, data)
}

func RegenerateWebToSkillTask(c *gin.Context) {
	adminUserId := GetAdminUserId(c)
	if adminUserId == 0 {
		return
	}
	params := define.WebToSkillTaskIDParams{}
	if err := common.RequestParamsBind(&params, c); err != nil {
		common.FmtError(c, `param_err`, middlewares.GetValidateErr(params, err, common.GetLang(c)).Error())
		return
	}
	unlock := lockWebToSkillTaskOperation(c, params.ID)
	if unlock == nil {
		return
	}
	defer unlock()
	id, err := common.RegenerateWebToSkillTask(common.GetLang(c), adminUserId, params.ID)
	if err != nil {
		c.String(http.StatusOK, lib_web.FmtJson(nil, err))
		return
	}
	common.FmtOk(c, id)
}

func UpdateWebToSkillTask(c *gin.Context) {
	adminUserId := GetAdminUserId(c)
	if adminUserId == 0 {
		return
	}
	params := define.WebToSkillTaskIDParams{}
	if err := common.RequestParamsBind(&params, c); err != nil {
		common.FmtError(c, `param_err`, middlewares.GetValidateErr(params, err, common.GetLang(c)).Error())
		return
	}
	unlock := lockWebToSkillTaskOperation(c, params.ID)
	if unlock == nil {
		return
	}
	defer unlock()
	id, err := common.UpdateWebToSkillTask(common.GetLang(c), adminUserId, params.ID)
	if err != nil {
		c.String(http.StatusOK, lib_web.FmtJson(nil, err))
		return
	}
	common.FmtOk(c, id)
}

func DeleteWebToSkillTask(c *gin.Context) {
	adminUserId := GetAdminUserId(c)
	if adminUserId == 0 {
		return
	}
	params := define.WebToSkillTaskIDParams{}
	if err := common.RequestParamsBind(&params, c); err != nil {
		common.FmtError(c, `param_err`, middlewares.GetValidateErr(params, err, common.GetLang(c)).Error())
		return
	}
	unlock := lockWebToSkillTaskOperation(c, params.ID)
	if unlock == nil {
		return
	}
	defer unlock()
	if err := common.DeleteWebToSkillTask(common.GetLang(c), adminUserId, params.ID); err != nil {
		c.String(http.StatusOK, lib_web.FmtJson(nil, err))
		return
	}
	common.FmtOk(c, nil)
}

func GetWebToSkillTaskInfo(c *gin.Context) {
	adminUserId := GetAdminUserId(c)
	if adminUserId == 0 {
		return
	}
	params := define.WebToSkillTaskIDParams{}
	if err := c.ShouldBindQuery(&params); err != nil {
		common.FmtError(c, `param_err`, middlewares.GetValidateErr(params, err, common.GetLang(c)).Error())
		return
	}
	data, err := common.GetWebToSkillTaskDetail(common.GetLang(c), adminUserId, params.ID)
	if err != nil {
		c.String(http.StatusOK, lib_web.FmtJson(nil, err))
		return
	}
	common.FmtOk(c, data)
}

func DownloadWebToSkillFile(c *gin.Context) {
	adminUserId := GetAdminUserId(c)
	if adminUserId == 0 {
		return
	}
	params := define.WebToSkillTaskIDParams{}
	if err := c.ShouldBindQuery(&params); err != nil {
		common.FmtError(c, `param_err`, middlewares.GetValidateErr(params, err, common.GetLang(c)).Error())
		return
	}
	unlock := lockWebToSkillTaskOperation(c, params.ID)
	if unlock == nil {
		return
	}
	defer unlock()
	file, fileName, err := common.GetWebToSkillTaskDownloadFile(common.GetLang(c), adminUserId, params.ID)
	if err != nil {
		c.String(http.StatusOK, lib_web.FmtJson(nil, err))
		return
	}
	c.FileAttachment(file, fileName)
}

func InstallWebToSkill(c *gin.Context) {
	adminUserId := GetAdminUserId(c)
	if adminUserId == 0 {
		return
	}
	params := define.WebToSkillTaskInstallParams{}
	if err := common.RequestParamsBind(&params, c); err != nil {
		common.FmtError(c, `param_err`, middlewares.GetValidateErr(params, err, common.GetLang(c)).Error())
		return
	}
	robotKey := ``
	if params.RobotID > 0 {
		var ok bool
		if robotKey, ok = common.GetClawbotRobotKey(c, adminUserId, params.RobotID); !ok {
			return
		}
	}
	unlock := lockWebToSkillTaskOperation(c, params.ID)
	if unlock == nil {
		return
	}
	defer unlock()
	lockKey := define.LockPreKey + `ClawbotUserSkill.` + cast.ToString(adminUserId)
	if !lib_redis.AddLock(define.Redis, lockKey, time.Minute*5) {
		common.FmtError(c, `op_lock`)
		return
	}
	defer lib_redis.UnLock(define.Redis, lockKey)

	data, err := common.InstallWebToSkillTask(common.GetLang(c), adminUserId, params.ID, params.Overwrite)
	if err != nil {
		c.String(http.StatusOK, lib_web.FmtJson(nil, err))
		return
	}
	common.FmtOk(c, buildGeneratedSkillInstallResult(c, adminUserId, `web`, params.ID, params.RobotID, robotKey, data))
}
