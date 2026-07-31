// Copyright © 2016- 2025 Wuhan Sesame Small Customer Service Network Technology Co., Ltd.

package common

import (
	"chatwiki/internal/app/chatwiki/define"
	"chatwiki/internal/app/chatwiki/i18n"
	"chatwiki/internal/pkg/lib_redis"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/spf13/cast"
	"github.com/zhimaAi/go_tools/logs"
	"github.com/zhimaAi/go_tools/msql"
	"github.com/zhimaAi/go_tools/tool"
)

const (
	RobotRateLimitStatusAllowed = iota
	RobotRateLimitStatusFiveMinute
	RobotRateLimitStatusDaily
)

const robotRateLimitCounterScript = `
local five_minute_ttl = tonumber(ARGV[1])
local five_minute_limit = tonumber(ARGV[2])
local daily_limit = tonumber(ARGV[3])
local daily_ttl = tonumber(ARGV[4])
local status_allowed = tonumber(ARGV[5])
local status_five_minute = tonumber(ARGV[6])
local status_daily = tonumber(ARGV[7])

local five_minute_count = redis.call('INCR', KEYS[1])
if five_minute_count == 1 then
	redis.call('EXPIRE', KEYS[1], five_minute_ttl)
end

local daily_count = tonumber(redis.call('GET', KEYS[2]) or '0')
if daily_count >= daily_limit then
	return status_daily
end

if five_minute_count > five_minute_limit then
	return status_five_minute
end

daily_count = redis.call('INCR', KEYS[2])
if daily_count == 1 then
	redis.call('EXPIRE', KEYS[2], daily_ttl)
end

return status_allowed
`

type RobotRateLimitConfCacheBuildHandler struct{ RobotKey string }

func (h *RobotRateLimitConfCacheBuildHandler) GetCacheKey() string {
	return fmt.Sprintf(`chatwiki.robot_rate_limit_conf.%s`, h.RobotKey)
}

func (h *RobotRateLimitConfCacheBuildHandler) GetCacheData() (any, error) {
	model := msql.Model(define.TableChatAiRobotRateLimitConf, define.Postgres).
		Where(`robot_key`, h.RobotKey)
	info, err := model.Find()
	if err != nil {
		logs.Error(`sql:%s,err:%s`, model.GetLastSql(), err.Error())
	}
	return info, err
}

type RobotRateLimitResult struct {
	Status       int
	ReplyType    int
	ReplyContent string
}

func GetRobotRateLimitConfInfo(robotKey string) (msql.Params, error) {
	result := make(msql.Params)
	err := lib_redis.GetCacheWithBuild(
		define.Redis,
		&RobotRateLimitConfCacheBuildHandler{RobotKey: robotKey},
		&result,
		time.Hour,
	)
	return result, err
}

func GetRobotRateLimitConf(lang string, adminUserId int, robotKey string) (*define.RobotRateLimitConfParams, error) {
	if err := checkRobotRateLimitConfRobot(lang, adminUserId, robotKey); err != nil {
		return nil, err
	}
	info, err := GetRobotRateLimitConfInfo(robotKey)
	if err != nil {
		logs.Error(err.Error())
		return nil, errors.New(i18n.Show(lang, `sys_err`))
	}
	if len(info) == 0 {
		return define.DefaultRobotRateLimitConf(robotKey), nil
	}
	return BuildRobotRateLimitConfParams(info), nil
}

func BuildRobotRateLimitConfParams(info msql.Params) *define.RobotRateLimitConfParams {
	return &define.RobotRateLimitConfParams{
		RobotKey:               info[`robot_key`],
		SwitchStatus:           cast.ToInt(info[`switch_status`]),
		FiveMinuteLimit:        cast.ToInt64(info[`five_minute_limit`]),
		FiveMinuteReplyType:    cast.ToInt(info[`five_minute_reply_type`]),
		FiveMinuteReplyContent: info[`five_minute_reply_content`],
		DailyLimit:             cast.ToInt64(info[`daily_limit`]),
		DailyReplyType:         cast.ToInt(info[`daily_reply_type`]),
		DailyReplyContent:      info[`daily_reply_content`],
	}
}

func SaveRobotRateLimitConf(lang string, adminUserId int, params define.RobotRateLimitConfParams) (*define.RobotRateLimitConfParams, error) {
	if err := checkRobotRateLimitConfRobot(lang, adminUserId, params.RobotKey); err != nil {
		return nil, err
	}
	params.FiveMinuteReplyContent = strings.TrimSpace(params.FiveMinuteReplyContent)
	params.DailyReplyContent = strings.TrimSpace(params.DailyReplyContent)
	if err := checkRobotRateLimitConfParams(lang, params); err != nil {
		return nil, err
	}
	info, err := GetRobotRateLimitConfInfo(params.RobotKey)
	if err != nil {
		logs.Error(err.Error())
		return nil, errors.New(i18n.Show(lang, `sys_err`))
	}
	data := msql.Datas{
		`admin_user_id`:             adminUserId,
		`switch_status`:             params.SwitchStatus,
		`five_minute_limit`:         params.FiveMinuteLimit,
		`five_minute_reply_type`:    params.FiveMinuteReplyType,
		`five_minute_reply_content`: params.FiveMinuteReplyContent,
		`daily_limit`:               params.DailyLimit,
		`daily_reply_type`:          params.DailyReplyType,
		`daily_reply_content`:       params.DailyReplyContent,
		`update_time`:               tool.Time2Int(),
	}
	m := msql.Model(define.TableChatAiRobotRateLimitConf, define.Postgres)
	update := func() error {
		_, updateErr := m.Where(`robot_key`, params.RobotKey).Update(data)
		return updateErr
	}
	if len(info) == 0 {
		data[`robot_key`] = params.RobotKey
		data[`create_time`] = tool.Time2Int()
		if _, err = m.Insert(data); err != nil {
			var sqlerr *pq.Error
			if errors.As(err, &sqlerr) && sqlerr.Code == `23505` { // Unique index constraint
				delete(data, `robot_key`)
				delete(data, `create_time`)
				err = update()
			}
		}
	} else {
		err = update()
	}
	if err != nil {
		logs.Error(`sql:%s,err:%s`, m.GetLastSql(), err.Error())
		return nil, errors.New(i18n.Show(lang, `sys_err`))
	}
	lib_redis.DelCacheData(define.Redis, &RobotRateLimitConfCacheBuildHandler{RobotKey: params.RobotKey})
	return GetRobotRateLimitConf(lang, adminUserId, params.RobotKey)
}

func CheckRobotRateLimit(robotKey, openid string) (RobotRateLimitResult, error) {
	result := RobotRateLimitResult{Status: RobotRateLimitStatusAllowed}
	confInfo, err := GetRobotRateLimitConfInfo(robotKey)
	if err != nil || len(confInfo) == 0 || cast.ToInt(confInfo[`switch_status`]) != define.SwitchOn {
		return result, err
	}
	conf := BuildRobotRateLimitConfParams(confInfo)
	fiveMinuteKey := fmt.Sprintf(`chatwiki.robot_rate_limit.5m.%s.%s`, robotKey, openid)
	dailyKey := fmt.Sprintf(`chatwiki.robot_rate_limit.daily.%s.%s.%s`, robotKey, openid, tool.Date(`Ymd`))
	result.Status, err = define.Redis.Eval(
		context.Background(),
		robotRateLimitCounterScript,
		[]string{fiveMinuteKey, dailyKey},
		define.RobotRateLimitFiveMinuteWindowSeconds,
		conf.FiveMinuteLimit,
		conf.DailyLimit,
		86400,
		RobotRateLimitStatusAllowed,
		RobotRateLimitStatusFiveMinute,
		RobotRateLimitStatusDaily,
	).Int()
	if err != nil {
		return result, err
	}
	switch result.Status {
	case RobotRateLimitStatusFiveMinute:
		result.ReplyType = conf.FiveMinuteReplyType
		result.ReplyContent = conf.FiveMinuteReplyContent
	case RobotRateLimitStatusDaily:
		result.ReplyType = conf.DailyReplyType
		result.ReplyContent = conf.DailyReplyContent
	case RobotRateLimitStatusAllowed:
	default:
		return RobotRateLimitResult{Status: RobotRateLimitStatusAllowed},
			fmt.Errorf(`invalid robot rate limit status: %d`, result.Status)
	}
	return result, nil
}

func DeleteRobotRateLimitConf(robotKey string) error {
	if len(robotKey) == 0 {
		return nil
	}
	m := msql.Model(define.TableChatAiRobotRateLimitConf, define.Postgres)
	_, err := m.Where(`robot_key`, robotKey).Delete()
	if err != nil {
		logs.Error(`sql:%s,err:%s`, m.GetLastSql(), err.Error())
		return err
	}
	lib_redis.DelCacheData(define.Redis, &RobotRateLimitConfCacheBuildHandler{RobotKey: robotKey})
	return nil
}

func checkRobotRateLimitConfRobot(lang string, adminUserId int, robotKey string) error {
	if !CheckRobotKey(robotKey) {
		return errors.New(i18n.Show(lang, `param_invalid`, `robot_key`))
	}
	robot, err := GetRobotInfo(robotKey)
	if err != nil {
		logs.Error(err.Error())
		return errors.New(i18n.Show(lang, `sys_err`))
	}
	if len(robot) == 0 || cast.ToInt(robot[`admin_user_id`]) != adminUserId {
		return errors.New(i18n.Show(lang, `no_data`))
	}
	return nil
}

func checkRobotRateLimitConfParams(lang string, params define.RobotRateLimitConfParams) error {
	if params.SwitchStatus != define.SwitchOff && params.SwitchStatus != define.SwitchOn {
		return errors.New(i18n.Show(lang, `param_invalid`, `switch_status`))
	}
	if params.FiveMinuteLimit <= 0 {
		return errors.New(i18n.Show(lang, `param_invalid`, `five_minute_limit`))
	}
	if params.DailyLimit <= 0 {
		return errors.New(i18n.Show(lang, `param_invalid`, `daily_limit`))
	}
	if params.FiveMinuteReplyType != define.RobotRateLimitReplyTypeNone &&
		params.FiveMinuteReplyType != define.RobotRateLimitReplyTypeSpecified {
		return errors.New(i18n.Show(lang, `param_invalid`, `five_minute_reply_type`))
	}
	if params.DailyReplyType != define.RobotRateLimitReplyTypeNone &&
		params.DailyReplyType != define.RobotRateLimitReplyTypeSpecified {
		return errors.New(i18n.Show(lang, `param_invalid`, `daily_reply_type`))
	}
	if params.SwitchStatus == define.SwitchOn &&
		params.FiveMinuteReplyType == define.RobotRateLimitReplyTypeSpecified &&
		len(params.FiveMinuteReplyContent) == 0 {
		return errors.New(i18n.Show(lang, `param_empty`, `five_minute_reply_content`))
	}
	if params.SwitchStatus == define.SwitchOn &&
		params.DailyReplyType == define.RobotRateLimitReplyTypeSpecified &&
		len(params.DailyReplyContent) == 0 {
		return errors.New(i18n.Show(lang, `param_empty`, `daily_reply_content`))
	}
	return nil
}
