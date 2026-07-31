// Copyright © 2016- 2025 Wuhan Sesame Small Customer Service Network Technology Co., Ltd.

package manage

import (
	"chatwiki/internal/app/chatwiki/common"
	"chatwiki/internal/app/chatwiki/define"
	"chatwiki/internal/app/chatwiki/middlewares"
	"chatwiki/internal/pkg/lib_web"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetRobotRateLimitConf(c *gin.Context) {
	adminUserId := GetAdminUserId(c)
	if adminUserId == 0 {
		return
	}
	params := define.RobotRateLimitConfGetParams{}
	if err := common.RequestParamsBind(&params, c); err != nil {
		common.FmtError(c, `param_err`, middlewares.GetValidateErr(params, err, common.GetLang(c)).Error())
		return
	}
	data, err := common.GetRobotRateLimitConf(common.GetLang(c), adminUserId, params.RobotKey)
	c.String(http.StatusOK, lib_web.FmtJson(data, err))
}

func SaveRobotRateLimitConf(c *gin.Context) {
	adminUserId := GetAdminUserId(c)
	if adminUserId == 0 {
		return
	}
	params := define.RobotRateLimitConfParams{}
	if err := common.RequestParamsBind(&params, c); err != nil {
		common.FmtError(c, `param_err`, middlewares.GetValidateErr(params, err, common.GetLang(c)).Error())
		return
	}
	data, err := common.SaveRobotRateLimitConf(common.GetLang(c), adminUserId, params)
	c.String(http.StatusOK, lib_web.FmtJson(data, err))
}
