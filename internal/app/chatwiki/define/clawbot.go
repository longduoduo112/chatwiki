// Copyright © 2016- 2025 Wuhan Sesame Small Customer Service Network Technology Co., Ltd.

package define

import (
	"regexp"
	"time"
)

const (
	SystemSkillsDir  = `clawbot/skills_system/<biz_type>`
	PublicSkillsDir  = `clawbot/skills_public`
	UserSkillsDir    = `clawbot/skills_user/<admin_user_id>`
	PrivateSkillsDir = `clawbot/skills_robot/<robot_key>`
	PrivateFileDir   = `clawbot/skills_robot/<robot_key>/query-local-docs/references`
	PrivateWorkDir   = `clawbot/working_dir/<robot_key>`

	ClawbotLocalDocConvertWorkDir = `clawbot/working_dir/query-local-docs/<task_batch>`
	ClawbotLocalDocConvertScript  = `clawbot/working_dir/query-local-docs/convert_docs.py`
)

// skill management constants
const (
	SkillMdFileName      = `SKILL.md`
	MaxSkillZipSize      = 100 * 1024 * 1024
	SkillReservedName    = `query-local-docs`
	SkillTmpDir          = `.skill_tmp`
	SkillUploadKeyExpire = 10 * time.Minute
)

// skill source type, maps to chat_ai_clawbot_skill.source_type
const (
	SkillSourceTypeUpload = 1
)

var SkillNameRegexp = regexp.MustCompile(`^[A-Za-z0-9_-]{1,50}$`)

var SkillUploadKeyRegexp = regexp.MustCompile(`^skup_[0-9]{14}_[A-Za-z0-9]{16}$`)

var ClawbotLocalDocAllowExt = []string{`docx`, `md`, `txt`, `pdf`, `csv`}

type ClawbotLocalDocInfo struct {
	Name string    `json:"name"`
	Size int64     `json:"size"`
	Time time.Time `json:"time"`
	Ext  string    `json:"ext"`
}

type ClawbotLocalDocIndexItem struct {
	File        string   `yaml:"file"`
	Type        string   `yaml:"type"`
	Description string   `yaml:"description"`
	Keywords    []string `yaml:"keywords"`
	Updated     string   `yaml:"updated"`
}

type ClawbotLocalDocConvertTask struct {
	TaskBatch   string   `json:"task_batch"`
	RobotKey    string   `json:"robot_key"`
	SourceName  string   `json:"source_name"`
	SourceHash  string   `json:"source_hash"`
	Description string   `json:"description"`
	Keywords    []string `json:"keywords"`
	SourceLock  string   `json:"source_lock"`
}

type ClawbotLocalDocConvertTaskResult struct {
	TaskBatch string `json:"task_batch"`
}
