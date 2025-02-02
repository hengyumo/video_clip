package ffmpeg

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type TaskInfo struct {
	TaskID       int64
	Progress     int
	TotalSeconds int64
	StartTime    time.Time
	Status       string // running, completed, failed
	InputFiles   []string
	OutputFile   string
	TaskType     string // cut, merge, convert
	Description  string
}

var (
	taskProgressMap = sync.Map{} // map[int64]*TaskInfo
)

type ProgressCallback func(progress int)

type VideoProcessor struct {
	ffmpegPath string
}

func NewVideoProcessor() (*VideoProcessor, error) {
	path, err := exec.LookPath("ffmpeg")
	if err != nil {
		return nil, fmt.Errorf("请先安装ffmpeg(%w)", err)
	}

	return &VideoProcessor{
		ffmpegPath: path,
	}, nil
}

func (v *VideoProcessor) Cut(input, output, start, end string, taskID int64) error {
	// 检查输入文件是否存在
	if _, err := os.Stat(input); os.IsNotExist(err) {
		return fmt.Errorf("输入文件不存在: %s", input)
	}

	// 根据输出文件扩展名决定编码方式
	ext := filepath.Ext(output)

	// 获取输出文件的绝对路径
	absOutput, err := filepath.Abs(output)
	if err != nil {
		return fmt.Errorf("无法解析输出路径: %v", err)
	}

	args := []string{
		"-i", input,
		"-ss", start,
		"-to", end,
	}

	// 添加编码参数
	switch ext {
	case ".mp4":
		args = append(args, "-c:v", "libx264", "-c:a", "aac")
	case ".mkv":
		args = append(args, "-c:v", "libx264", "-c:a", "aac")
	case ".avi":
		args = append(args, "-c:v", "mpeg4", "-c:a", "mp3")
	case ".mov":
		args = append(args, "-c:v", "libx264", "-c:a", "aac")
	default:
		args = append(args, "-c", "copy")
	}

	args = append(args, absOutput)

	cmd := exec.Command(v.ffmpegPath, args...)

	// 获取标准错误输出
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("无法获取stderr管道: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("ffmpeg启动失败: %v", err)
	}

	startTime := parseDuration(start)
	endTime := parseDuration(end)
	totalDuration := endTime - startTime
	fileName := filepath.Base(input)
	taskInfo := &TaskInfo{
		TaskID:      taskID,
		TaskType:    "cut",
		StartTime:   time.Now(),
		Status:      "running",
		InputFiles:  []string{input},
		OutputFile:  output,
		Description: fmt.Sprintf("剪切 %s 从 %v 到 %v", fileName, start, end),
	}
	taskProgressMap.Store(taskID, taskInfo)

	// 解析进度
	go parseCmdProgress(taskID, totalDuration, stderr)

	if err := cmd.Wait(); err != nil {
		taskInfo.Status = "failed"
		return fmt.Errorf("ffmpeg执行失败: %v", err)
	}

	taskInfo.Status = "completed"
	return nil
}

// GetVideoDuration 获取视频时长（秒）
func (v *VideoProcessor) GetVideoDuration(input string) (int64, error) {
	cmd := exec.Command(v.ffmpegPath, "-i", input)
	output, _ := cmd.CombinedOutput()

	// 查找持续时间信息
	outputStr := string(output)
	durationIndex := strings.Index(outputStr, "Duration: ")
	if durationIndex == -1 {
		return 0, fmt.Errorf("无法解析视频时长")
	}

	// 解析持续时间
	durationStr := outputStr[durationIndex+10 : durationIndex+21]
	return parseDuration(durationStr), nil
}

func (v *VideoProcessor) Merge(inputs []string, output string, taskID int64) error {
	if len(inputs) < 2 {
		return errors.New("至少需要两个视频文件")
	}

	// 计算总时长
	var totalDuration int64
	for _, input := range inputs {
		duration, err := v.GetVideoDuration(input)
		if err != nil {
			return fmt.Errorf("获取视频%s时长失败: %v", input, err)
		}
		totalDuration += duration
	}
	// 生成文件列表
	listFile := "concat.txt"
	content := ""
	fileNames := make([]string, 0)
	for _, input := range inputs {
		content += fmt.Sprintf("file '%s'\n", input)
		fileNames = append(fileNames, filepath.Base(input))
	}

	// 创建临时文件
	err := os.WriteFile(listFile, []byte(content), 0644)
	if err != nil {
		return err
	}
	defer os.Remove(listFile)

	cmd := exec.Command(v.ffmpegPath,
		"-f", "concat",
		"-safe", "0",
		"-i", listFile,
		"-c", "copy",
		output,
	)
	taskInfo := &TaskInfo{
		TaskID:      taskID,
		TaskType:    "merge",
		StartTime:   time.Now(),
		Status:      "running",
		InputFiles:  inputs,
		OutputFile:  output,
		Description: fmt.Sprintf("合并视频: %v", fileNames),
	}
	taskProgressMap.Store(taskID, taskInfo)

	// 获取标准错误输出
	stderr, err := cmd.StderrPipe()
	if err != nil {
		taskInfo.Status = "failed"
		return fmt.Errorf("无法获取stderr管道: %v", err)
	}

	if err := cmd.Start(); err != nil {
		taskInfo.Status = "failed"
		return fmt.Errorf("ffmpeg启动失败: %v", err)
	}

	// 解析进度
	go parseCmdProgress(taskID, totalDuration, stderr)

	if err := cmd.Wait(); err != nil {
		taskInfo.Status = "failed"
		return fmt.Errorf("ffmpeg执行失败: %v", err)
	}

	taskInfo.Status = "completed"
	return nil
}

func parseCmdProgress2(taskID, totalSeconds int64, stderr io.ReadCloser) {
	defer stderr.Close()

	scanner := bufio.NewScanner(stderr)
	re := regexp.MustCompile(`time=(\d+):(\d+):(\d+.\d+)`)

	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println("line: ", line)
		matches := re.FindStringSubmatch(line)
		if len(matches) == 4 {
			hours, _ := strconv.Atoi(matches[1])
			minutes, _ := strconv.Atoi(matches[2])
			seconds, _ := strconv.ParseFloat(matches[3], 64)
			currentSeconds := int64(hours*3600+minutes*60) + int64(seconds)

			if value, ok := taskProgressMap.Load(taskID); ok {
				taskInfo := value.(*TaskInfo)
				progress := int((float64(currentSeconds) / float64(totalSeconds)) * 100)
				if progress < taskInfo.Progress {
					progress = taskInfo.Progress
				}
				fmt.Println("task: ", taskID, ", progress: ", progress)
				taskInfo.Progress = progress
				taskProgressMap.Store(taskID, taskInfo)
			}
		} else {
			if value, ok := taskProgressMap.Load(taskID); ok {
				taskInfo := value.(*TaskInfo)
				if taskInfo.Progress < 60 {
					taskInfo.Progress++
				}
				taskProgressMap.Store(taskID, taskInfo)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		log.Printf("解析进度时出错: %v", err)
	}
}

func parseCmdProgress(taskID, totalDuration int64, stderr io.ReadCloser) {
	buf := make([]byte, 1024)
	strbuf := bytes.Buffer{}
	for {
		n, err := stderr.Read(buf)
		if err != nil {
			break
		}
		_, _ = strbuf.Write(buf)
		output := string(buf[:n])
		fmt.Println(output)
		progress := parseProgress(strbuf.String(), totalDuration)

		if value, ok := taskProgressMap.Load(taskID); ok {
			taskInfo := value.(*TaskInfo)
			fmt.Println("task: ", taskID, ", progress: ", progress)
			if progress != -1 && progress > taskInfo.Progress {
				taskInfo.Progress = progress
			} else if taskInfo.Progress < 20 {
				taskInfo.Progress++
			}
			taskProgressMap.Store(taskID, taskInfo)
		}
	}
	if value, ok := taskProgressMap.Load(taskID); ok {
		taskInfo := value.(*TaskInfo)
		taskInfo.Progress = 100
		taskProgressMap.Store(taskID, taskInfo)
	}
}

// parseProgress 从ffmpeg输出中解析进度百分比
func parseProgress(output string, totalDuration int64) int {
	// 查找时间进度信息，格式如：time=00:00:12.34
	timeIndex := strings.LastIndex(output, "time=")
	if timeIndex == -1 {
		return -1
	}

	// 解析当前时间
	timeStr := output[timeIndex+5 : timeIndex+16]
	currentTime := parseDuration(timeStr)

	// 计算进度百分比
	if totalDuration == 0 {
		return -1
	}
	progress := int((float64(currentTime) / float64(totalDuration)) * 100)
	if progress > 100 {
		progress = 100
	}
	return progress
}

func (v *VideoProcessor) GetVideoInfo(input string) (string, error) {
	cmd := exec.Command(v.ffmpegPath,
		"-i", input,
	)

	outputBytes, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("ffmpeg error: %s\n%s", err, string(outputBytes))
	}

	return string(outputBytes), nil
}

// parseDuration 将HH:MM:SS格式的时间转换为秒数
func parseDuration(timeStr string) int64 {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 3 {
		return 0
	}
	hours, _ := strconv.Atoi(parts[0])
	minutes, _ := strconv.Atoi(parts[1])
	seconds, _ := strconv.ParseFloat(parts[2], 64)
	return int64(hours*3600+minutes*60) + int64(seconds)
}

func (v *VideoProcessor) Convert(input, output string, taskID int64) error {
	cmd := exec.Command(v.ffmpegPath,
		"-i", input,
		"-c:v", "libx264",
		"-c:a", "aac",
		"-strict", "experimental",
		output,
	)

	// 获取标准错误输出
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("无法获取stderr管道: %v", err)
	}

	// 创建任务信息
	taskInfo := &TaskInfo{
		TaskID:      taskID,
		StartTime:   time.Now(),
		Status:      "running",
		InputFiles:  []string{input},
		OutputFile:  output,
		TaskType:    "convert",
		Description: fmt.Sprintf("转换视频 %s 为 %s", input, output),
	}
	taskProgressMap.Store(taskID, taskInfo)

	if err := cmd.Start(); err != nil {
		taskInfo.Status = "failed"
		return fmt.Errorf("ffmpeg启动失败: %v", err)
	}

	// 解析进度
	go parseCmdProgress(taskID, 0, stderr)

	if err := cmd.Wait(); err != nil {
		taskInfo.Status = "failed"
		return fmt.Errorf("ffmpeg执行失败: %v", err)
	}

	taskInfo.Status = "completed"
	return nil
}

// GetAllTasks 获取所有任务信息
func (v *VideoProcessor) GetAllTasks() []*TaskInfo {
	var tasks []*TaskInfo
	taskProgressMap.Range(func(key, value interface{}) bool {
		if taskInfo, ok := value.(*TaskInfo); ok {
			tasks = append(tasks, taskInfo)
		}
		return true
	})
	return tasks
}

// GetTaskInfo 获取指定任务信息
func (v *VideoProcessor) GetTaskInfo(taskID int64) (*TaskInfo, error) {
	value, ok := taskProgressMap.Load(taskID)
	if !ok {
		return nil, fmt.Errorf("任务ID %d 不存在", taskID)
	}

	taskInfo, ok := value.(*TaskInfo)
	if !ok {
		return nil, fmt.Errorf("任务信息格式错误")
	}

	return taskInfo, nil
}
