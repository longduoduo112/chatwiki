// Copyright © 2016- 2025 Wuhan Sesame Small Customer Service Network Technology Co., Ltd.

package define

const TableChatAiRobotRateLimitConf = `chat_ai_robot_rate_limit_conf`

const (
	RobotRateLimitFiveMinuteDefault       int64 = 10
	RobotRateLimitDailyDefault            int64 = 100
	RobotRateLimitFiveMinuteWindowSeconds       = 300

	RobotRateLimitReplyTypeNone      = 0
	RobotRateLimitReplyTypeSpecified = 1
)

type RobotRateLimitConfGetParams struct {
	RobotKey string `form:"robot_key" json:"robot_key" binding:"required"`
}

type RobotRateLimitConfParams struct {
	RobotKey               string `form:"robot_key" json:"robot_key" binding:"required"`
	SwitchStatus           int    `form:"switch_status" json:"switch_status" binding:"oneof=0 1"`
	FiveMinuteLimit        int64  `form:"five_minute_limit" json:"five_minute_limit" binding:"gt=0"`
	FiveMinuteReplyType    int    `form:"five_minute_reply_type" json:"five_minute_reply_type" binding:"oneof=0 1"`
	FiveMinuteReplyContent string `form:"five_minute_reply_content" json:"five_minute_reply_content"`
	DailyLimit             int64  `form:"daily_limit" json:"daily_limit" binding:"gt=0"`
	DailyReplyType         int    `form:"daily_reply_type" json:"daily_reply_type" binding:"oneof=0 1"`
	DailyReplyContent      string `form:"daily_reply_content" json:"daily_reply_content"`
}

func DefaultRobotRateLimitConf(robotKey string) *RobotRateLimitConfParams {
	return &RobotRateLimitConfParams{
		RobotKey:            robotKey,
		SwitchStatus:        SwitchOff,
		FiveMinuteLimit:     RobotRateLimitFiveMinuteDefault,
		FiveMinuteReplyType: RobotRateLimitReplyTypeSpecified,
		DailyLimit:          RobotRateLimitDailyDefault,
		DailyReplyType:      RobotRateLimitReplyTypeNone,
	}
}
