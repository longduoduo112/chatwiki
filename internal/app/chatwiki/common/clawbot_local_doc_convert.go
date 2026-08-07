// Copyright © 2016- 2025 Wuhan Sesame Small Customer Service Network Technology Co., Ltd.

package common

import (
	"chatwiki/internal/app/chatwiki/define"
	"chatwiki/internal/pkg/lib_redis"
	"chatwiki/internal/pkg/llm_runner"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/zhimaAi/go_tools/logs"
	"github.com/zhimaAi/go_tools/tool"
)

const (
	clawbotLocalDocConvertLockTTL = 2 * time.Hour
	clawbotLocalDocIndexLockTTL   = 5 * time.Minute
)

type clawbotLocalDocConvertRPCResult struct {
	Success      bool                           `json:"success"`
	SourcePath   string                         `json:"source_path"`
	MarkdownPath string                         `json:"markdown_path"`
	ImagePaths   []string                       `json:"image_paths"`
	Error        clawbotLocalDocConvertRPCError `json:"error"`
}

type clawbotLocalDocConvertRPCError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type clawbotLocalDocValidatedOutput struct {
	MarkdownPath string
	ImagePaths   []string
}

type clawbotLocalDocFileSnapshot struct {
	Exists  bool
	Content []byte
}

func EnqueueClawbotLocalDocConvert(robotKey string, sourceName string) (*define.ClawbotLocalDocConvertTaskResult, error) {
	sourceName, ok := NormalizeClawbotLocalDocName(sourceName)
	if !ok {
		return nil, errors.New(`file name is invalid`)
	}
	ext := strings.ToLower(path.Ext(sourceName))
	if ext != `.docx` && ext != `.pdf` {
		return nil, errors.New(`only docx and pdf can be converted`)
	}
	privateDir := filepath.Clean(filepath.FromSlash(strings.ReplaceAll(define.PrivateFileDir, `<robot_key>`, robotKey)))
	sourcePath := filepath.Join(privateDir, sourceName)
	sourceHash, err := clawbotLocalDocFileSHA256(sourcePath)
	if err != nil {
		return nil, err
	}
	indexItem, found, err := GetClawbotLocalDocIndexItem(robotKey, sourceName)
	if err != nil {
		return nil, err
	}
	if !found {
		indexItem = define.ClawbotLocalDocIndexItem{}
	}

	taskBatch := uuid.NewString()
	sourceLock := clawbotLocalDocConvertSourceLockKey(robotKey, sourceName)
	if !reserveClawbotLocalDocConvertLock(sourceLock, taskBatch) {
		return nil, errors.New(`conversion task already submitted`)
	}
	releaseLocks := true
	defer func() {
		if !releaseLocks {
			return
		}
		releaseClawbotLocalDocConvertLock(sourceLock, taskBatch)
	}()

	task := define.ClawbotLocalDocConvertTask{
		TaskBatch:   taskBatch,
		RobotKey:    robotKey,
		SourceName:  sourceName,
		SourceHash:  sourceHash,
		Description: indexItem.Description,
		Keywords:    indexItem.Keywords,
		SourceLock:  sourceLock,
	}
	if err = AddJobs(define.ClawbotLocalDocConvertTaskTopic, tool.JsonEncodeNoError(task)); err != nil {
		return nil, err
	}
	releaseLocks = false
	return &define.ClawbotLocalDocConvertTaskResult{TaskBatch: taskBatch}, nil
}

func RunClawbotLocalDocConvertTask(message string) error {
	task := define.ClawbotLocalDocConvertTask{}
	if err := tool.JsonDecodeUseNumber(message, &task); err != nil {
		return fmt.Errorf(`decode local document conversion task: %w`, err)
	}
	if _, err := uuid.Parse(task.TaskBatch); err != nil {
		return fmt.Errorf(`invalid task batch: %w`, err)
	}
	sourceName, ok := NormalizeClawbotLocalDocName(task.SourceName)
	if !ok || sourceName != task.SourceName {
		return errors.New(`invalid source file name`)
	}
	ext := strings.ToLower(path.Ext(sourceName))
	if ext != `.docx` && ext != `.pdf` {
		return errors.New(`invalid source file extension`)
	}
	if task.SourceHash == `` {
		return errors.New(`source file hash is required`)
	}
	if task.SourceLock != clawbotLocalDocConvertSourceLockKey(task.RobotKey, task.SourceName) {
		return errors.New(`invalid source lock`)
	}
	defer releaseClawbotLocalDocConvertLock(task.SourceLock, task.TaskBatch)

	workRoot := filepath.Clean(filepath.FromSlash(strings.ReplaceAll(define.ClawbotLocalDocConvertWorkDir, `<task_batch>`, ``)))
	workDir := filepath.Clean(filepath.Join(workRoot, task.TaskBatch))
	if err := ensureClawbotLocalDocPathWithin(workRoot, workDir); err != nil {
		return err
	}
	if err := os.RemoveAll(workDir); err != nil {
		return err
	}
	if err := tool.MkDirAll(workDir); err != nil {
		return err
	}
	defer func() {
		if cleanupErr := os.RemoveAll(workDir); cleanupErr != nil {
			logs.Error(`cleanup local document conversion directory:%s,err:%s`, workDir, cleanupErr.Error())
		}
	}()

	privateDir := filepath.Clean(filepath.FromSlash(strings.ReplaceAll(define.PrivateFileDir, `<robot_key>`, task.RobotKey)))
	sourcePath := filepath.Join(privateDir, task.SourceName)
	currentHash, err := clawbotLocalDocFileSHA256(sourcePath)
	if err != nil {
		return err
	}
	if !strings.EqualFold(currentHash, task.SourceHash) {
		return errors.New(`source file changed after conversion task was created`)
	}

	targetName := clawbotLocalDocConvertTargetName(task.SourceName, tool.Date(`YmdHis`))
	if tool.IsFile(filepath.Join(privateDir, targetName)) {
		return nil
	}

	inputName := strings.TrimSuffix(targetName, `.md`) + strings.ToLower(path.Ext(task.SourceName))
	localInputPath := filepath.Join(workDir, inputName)
	if err = copyFile(sourcePath, localInputPath); err != nil {
		return err
	}
	copiedHash, err := clawbotLocalDocFileSHA256(localInputPath)
	if err != nil {
		return err
	}
	if !strings.EqualFold(copiedHash, task.SourceHash) {
		return errors.New(`copied source file hash mismatch`)
	}

	containerWorkDir := filepath.ToSlash(workDir)
	containerInputPath := path.Join(containerWorkDir, inputName)
	command := `python3 ` + clawbotLocalDocShellQuote(define.ClawbotLocalDocConvertScript) +
		` --input ` + clawbotLocalDocShellQuote(containerInputPath)
	resp := llm_runner.RpcExecuteRun(define.Config.WebService[`llm_runner_host`], ``, command)
	if resp.IsError {
		return fmt.Errorf(`llm runner error: %s`, resp.ErrorMsg)
	}
	if resp.ExitCode != 0 {
		return fmt.Errorf(`convert command exit code %d: %s`, resp.ExitCode, resp.Output)
	}
	result := clawbotLocalDocConvertRPCResult{}
	if err = tool.JsonDecodeUseNumber(strings.TrimSpace(resp.Output), &result); err != nil {
		return fmt.Errorf(`decode convert result: %w`, err)
	}
	if !result.Success {
		if result.Error.Message != `` {
			return errors.New(result.Error.Message)
		}
		return errors.New(`document conversion failed`)
	}
	validated, err := validateClawbotLocalDocConvertRPCResult(
		result,
		workDir,
		containerWorkDir,
		containerInputPath,
		targetName,
	)
	if err != nil {
		return err
	}
	if err = publishClawbotLocalDocConvertResult(task, targetName, privateDir, validated); err != nil {
		return err
	}
	return nil
}

func reserveClawbotLocalDocConvertLock(key string, taskBatch string) bool {
	if strings.TrimSpace(key) == `` || strings.TrimSpace(taskBatch) == `` {
		return false
	}
	ok, err := define.Redis.SetNX(context.Background(), key, taskBatch, clawbotLocalDocConvertLockTTL).Result()
	if err != nil {
		logs.Error(`reserve clawbot local doc convert lock:%s,err:%s`, key, err.Error())
		return false
	}
	return ok
}

func releaseClawbotLocalDocConvertLock(key string, taskBatch string) {
	if !strings.HasPrefix(key, define.LockPreKey+`ClawbotLocalDocConvert`) || strings.TrimSpace(taskBatch) == `` {
		return
	}
	const script = `if redis.call("get", KEYS[1]) == ARGV[1] then return redis.call("del", KEYS[1]) else return 0 end`
	if err := define.Redis.Eval(context.Background(), script, []string{key}, taskBatch).Err(); err != nil {
		logs.Error(`release clawbot local doc convert lock:%s,err:%s`, key, err.Error())
	}
}

func clawbotLocalDocConvertSourceLockKey(robotKey string, sourceName string) string {
	return define.LockPreKey + `ClawbotLocalDocConvertSource.` + clawbotLocalDocLockDigest(robotKey, sourceName)
}

func clawbotLocalDocLockDigest(robotKey string, name string) string {
	sum := sha256.Sum256([]byte(robotKey + "\x00" + name))
	return hex.EncodeToString(sum[:])
}

func clawbotLocalDocConvertTargetName(sourceName string, timestamp string) string {
	sourceExt := path.Ext(sourceName)
	sourceStem := strings.TrimSuffix(sourceName, sourceExt)
	// Keep same-stem DOCX and PDF conversion tasks on distinct output paths.
	return sourceStem + `-` + strings.TrimPrefix(strings.ToLower(sourceExt), `.`) + `-` + timestamp + `.md`
}

func validateClawbotLocalDocConvertRPCResult(
	result clawbotLocalDocConvertRPCResult,
	workDir string,
	containerWorkDir string,
	containerInputPath string,
	targetName string,
) (clawbotLocalDocValidatedOutput, error) {
	containerWorkDir = path.Clean(containerWorkDir)
	containerInputPath = path.Clean(containerInputPath)
	if path.IsAbs(containerWorkDir) || path.IsAbs(containerInputPath) ||
		containerWorkDir == `.` || containerWorkDir == `..` ||
		strings.HasPrefix(containerWorkDir, `../`) ||
		containerInputPath != path.Join(containerWorkDir, path.Base(containerInputPath)) {
		return clawbotLocalDocValidatedOutput{}, errors.New(`invalid container conversion path`)
	}
	if !clawbotLocalDocContainerPathMatches(result.SourcePath, containerInputPath) {
		return clawbotLocalDocValidatedOutput{}, errors.New(`convert result source path mismatch`)
	}
	resultWorkDir := path.Dir(path.Clean(result.SourcePath))
	expectedMarkdownRemote := path.Join(resultWorkDir, targetName)
	if path.Clean(result.MarkdownPath) != path.Clean(expectedMarkdownRemote) {
		return clawbotLocalDocValidatedOutput{}, errors.New(`convert result markdown path mismatch`)
	}

	markdownPath := filepath.Join(workDir, targetName)
	markdownInfo, err := os.Stat(markdownPath)
	if err != nil || !markdownInfo.Mode().IsRegular() {
		return clawbotLocalDocValidatedOutput{}, errors.New(`converted markdown file is missing`)
	}

	remoteImageRoot := path.Join(resultWorkDir, `assets`, targetName)
	localImageRoot := filepath.Join(workDir, `assets`, targetName)
	validated := clawbotLocalDocValidatedOutput{
		MarkdownPath: markdownPath,
		ImagePaths:   make([]string, 0, len(result.ImagePaths)),
	}
	for _, imagePath := range result.ImagePaths {
		imagePath = path.Clean(strings.TrimSpace(imagePath))
		if !strings.HasPrefix(imagePath, remoteImageRoot+`/`) {
			return clawbotLocalDocValidatedOutput{}, errors.New(`convert result image path mismatch`)
		}
		relativePath := strings.TrimPrefix(imagePath, resultWorkDir+`/`)
		localPath := filepath.Join(workDir, filepath.FromSlash(relativePath))
		if err = ensureClawbotLocalDocPathWithin(localImageRoot, localPath); err != nil {
			return clawbotLocalDocValidatedOutput{}, errors.New(`invalid converted image path`)
		}
		info, statErr := os.Stat(localPath)
		if statErr != nil || !info.Mode().IsRegular() {
			return clawbotLocalDocValidatedOutput{}, errors.New(`converted image file is missing`)
		}
		validated.ImagePaths = append(validated.ImagePaths, localPath)
	}
	return validated, nil
}

func clawbotLocalDocContainerPathMatches(actual string, expectedRelative string) bool {
	actual = path.Clean(strings.TrimSpace(actual))
	expectedRelative = path.Clean(strings.TrimSpace(expectedRelative))
	if expectedRelative == `.` || expectedRelative == `..` || path.IsAbs(expectedRelative) ||
		strings.HasPrefix(expectedRelative, `../`) {
		return false
	}
	if actual == expectedRelative {
		return true
	}
	return path.IsAbs(actual) && strings.HasSuffix(actual, `/`+expectedRelative)
}

func publishClawbotLocalDocConvertResult(
	task define.ClawbotLocalDocConvertTask,
	targetName string,
	privateDir string,
	output clawbotLocalDocValidatedOutput,
) error {
	stagingRoot := filepath.Join(privateDir, `.local-doc-convert-`+task.TaskBatch)
	if err := ensureClawbotLocalDocPathWithin(privateDir, stagingRoot); err != nil {
		return err
	}
	if err := os.RemoveAll(stagingRoot); err != nil {
		return err
	}
	if err := tool.MkDirAll(stagingRoot); err != nil {
		return err
	}
	defer func() {
		if err := os.RemoveAll(stagingRoot); err != nil {
			logs.Error(`cleanup local document staging directory:%s,err:%s`, stagingRoot, err.Error())
		}
	}()

	stagedMarkdown := filepath.Join(stagingRoot, targetName)
	if err := copyFile(output.MarkdownPath, stagedMarkdown); err != nil {
		return err
	}
	workDir := filepath.Dir(output.MarkdownPath)
	for _, imagePath := range output.ImagePaths {
		relativePath, err := filepath.Rel(workDir, imagePath)
		if err != nil {
			return err
		}
		if err = copyFile(imagePath, filepath.Join(stagingRoot, relativePath)); err != nil {
			return err
		}
	}
	return commitClawbotLocalDocConvertResult(task, targetName, privateDir, stagingRoot)
}

func commitClawbotLocalDocConvertResult(
	task define.ClawbotLocalDocConvertTask,
	targetName string,
	privateDir string,
	stagingRoot string,
) error {
	lockKey := define.LockPreKey + `ClawbotLocalDocIndex.` + task.RobotKey
	if err := waitClawbotLocalDocIndexLock(lockKey, 30*time.Second); err != nil {
		return err
	}
	defer lib_redis.UnLock(define.Redis, lockKey)

	sourcePath := filepath.Join(privateDir, task.SourceName)
	currentHash, err := clawbotLocalDocFileSHA256(sourcePath)
	if err != nil {
		return err
	}
	if !strings.EqualFold(currentHash, task.SourceHash) {
		return errors.New(`source file changed before conversion result was committed`)
	}

	finalMarkdown := filepath.Join(privateDir, targetName)
	finalAssets := filepath.Join(privateDir, `assets`, targetName)
	if _, err = os.Stat(finalMarkdown); err == nil || !os.IsNotExist(err) {
		return errors.New(`converted markdown already exists`)
	}
	if _, err = os.Stat(finalAssets); err == nil || !os.IsNotExist(err) {
		return errors.New(`converted markdown assets already exist`)
	}

	stagedMarkdown := filepath.Join(stagingRoot, targetName)
	stagedAssets := filepath.Join(stagingRoot, `assets`, targetName)

	indexPath := filepath.Join(privateDir, `index.yaml`)
	skillPath := filepath.Clean(filepath.Join(privateDir, `..`, `SKILL.md`))
	indexSnapshot, err := snapshotClawbotLocalDocFile(indexPath)
	if err != nil {
		return err
	}
	skillSnapshot, err := snapshotClawbotLocalDocFile(skillPath)
	if err != nil {
		return err
	}

	rollback := func() {
		_ = os.Remove(finalMarkdown)
		_ = os.RemoveAll(finalAssets)
		if restoreErr := restoreClawbotLocalDocFile(indexPath, indexSnapshot); restoreErr != nil {
			logs.Error(`restore clawbot local doc index:%s`, restoreErr.Error())
		}
		if restoreErr := restoreClawbotLocalDocFile(skillPath, skillSnapshot); restoreErr != nil {
			logs.Error(`restore query-local-docs skill:%s`, restoreErr.Error())
		}
	}

	if err = os.Rename(stagedMarkdown, finalMarkdown); err != nil {
		rollback()
		return err
	}
	if info, statErr := os.Stat(stagedAssets); statErr == nil && info.IsDir() {
		if err = tool.MkDirAll(filepath.Dir(finalAssets)); err != nil {
			rollback()
			return err
		}
		if err = os.Rename(stagedAssets, finalAssets); err != nil {
			rollback()
			return err
		}
	} else if statErr != nil && !os.IsNotExist(statErr) {
		rollback()
		return statErr
	}

	info, err := os.Stat(finalMarkdown)
	if err != nil {
		rollback()
		return err
	}
	docInfo := define.ClawbotLocalDocInfo{Name: targetName, Size: info.Size(), Ext: `md`}
	if err = SaveClawbotLocalDocIndex(task.RobotKey, docInfo, task.Description, task.Keywords); err != nil {
		rollback()
		return err
	}
	if err = DeleteClawbotLocalDocIndex(task.RobotKey, task.SourceName); err != nil {
		rollback()
		return err
	}
	if err = os.Remove(sourcePath); err != nil {
		rollback()
		return err
	}
	return nil
}

func waitClawbotLocalDocIndexLock(lockKey string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		if lib_redis.AddLock(define.Redis, lockKey, clawbotLocalDocIndexLockTTL) {
			return nil
		}
		if time.Now().After(deadline) {
			return errors.New(`local document index is busy`)
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func snapshotClawbotLocalDocFile(filePath string) (clawbotLocalDocFileSnapshot, error) {
	content, err := os.ReadFile(filePath)
	if err == nil {
		return clawbotLocalDocFileSnapshot{Exists: true, Content: content}, nil
	}
	if os.IsNotExist(err) {
		return clawbotLocalDocFileSnapshot{}, nil
	}
	return clawbotLocalDocFileSnapshot{}, err
}

func restoreClawbotLocalDocFile(filePath string, snapshot clawbotLocalDocFileSnapshot) error {
	if !snapshot.Exists {
		if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	if err := tool.MkDirAll(filepath.Dir(filePath)); err != nil {
		return err
	}
	return os.WriteFile(filePath, snapshot.Content, 0644)
}

func clawbotLocalDocFileSHA256(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return ``, err
	}
	defer func() { _ = file.Close() }()
	digest := sha256.New()
	if _, err = io.Copy(digest, file); err != nil {
		return ``, err
	}
	return hex.EncodeToString(digest.Sum(nil)), nil
}

func clawbotLocalDocShellQuote(value string) string {
	return `'` + strings.ReplaceAll(value, `'`, `'"'"'`) + `'`
}
