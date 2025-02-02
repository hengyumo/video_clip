package web

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"video-clipper/pkg/ffmpeg"

	"github.com/gin-gonic/gin"
)

type VideoHandler struct {
    videoDir  string
    processor  *ffmpeg.VideoProcessor
}

func NewVideoHandler(videoDir string, processor *ffmpeg.VideoProcessor) *VideoHandler {
    return &VideoHandler{
        videoDir: videoDir,
        processor: processor,
    }
}

func (h *VideoHandler) UpdateVideoDir(c *gin.Context) {
    var request struct {
        VideoDir string `json:"videoDir"`
    }

    if err := c.ShouldBindJSON(&request); err != nil {
        c.JSON(400, gin.H{"error": "无效的请求数据"})
        return
    }

    // 检查路径是否存在
    if _, err := os.Stat(request.VideoDir); os.IsNotExist(err) {
        c.JSON(400, gin.H{"error": "路径不存在"})
        return
    }

    // 更新视频目录
    h.videoDir = request.VideoDir
    fmt.Println(h.videoDir)
    c.JSON(200, gin.H{"message": "视频目录已更新", "videoDir": h.videoDir})
}

func (h *VideoHandler) ListVideos(c *gin.Context) {
    files, err := os.ReadDir(h.videoDir)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error(), "path": h.videoDir})
        return
    }

    var videoFiles []string
    for _, file := range files {
        if !file.IsDir() {
            ext := filepath.Ext(file.Name())
            switch ext {
            case ".mp4":
                videoFiles = append(videoFiles, file.Name())
            }
        }
    }

    c.JSON(200, gin.H{
       "videos": videoFiles,
       "videoDir": h.videoDir,
    })
}

func (h *VideoHandler) GetVideo(c *gin.Context) {
    encodedFilename := c.Param("filename")
    filename, err := url.QueryUnescape(encodedFilename)
    if err != nil {
        c.JSON(400, gin.H{"error": "文件名解码失败"})
        return
    }

    videoPath := filepath.Join(h.videoDir, filename)

    if _, err := os.Stat(videoPath); os.IsNotExist(err) {
        c.JSON(404, gin.H{"error": "视频文件不存在"})
        return
    }

    // 根据文件扩展名设置Content-Type
    ext := filepath.Ext(filename)
    switch ext {
    case ".mp4":
        c.Header("Content-Type", "video/mp4")
    case ".mkv":
        c.Header("Content-Type", "video/x-matroska")
    case ".avi":
        c.Header("Content-Type", "video/x-msvideo")
    case ".mov":
        c.Header("Content-Type", "video/quicktime")
    default:
        c.Header("Content-Type", "application/octet-stream")
    }
    c.File(videoPath)
}

func (h *VideoHandler) CutVideo(c *gin.Context) {
    var request struct {
        TaskID int64  `json:"taskID"`
        Input  string `json:"input"`
        Output string `json:"output"`
        Start  string `json:"start"`
        End    string `json:"end"`
    }

    if err := c.ShouldBindJSON(&request); err != nil {
        c.JSON(400, gin.H{"error": "无效的请求数据"})
        return
    }

    inputPath := filepath.Join(h.videoDir, request.Input)
    outputPath := filepath.Join(h.videoDir, request.Output)

    err := h.processor.Cut(inputPath, outputPath, request.Start, request.End, request.TaskID)
    if err != nil {
        c.JSON(500, gin.H{"error": "视频剪切失败: " + err.Error()})
        return
    }

    c.JSON(200, gin.H{"taskID": request.TaskID, "message": "视频剪切成功: " + request.Output})
}

func (h *VideoHandler) MergeVideos(c *gin.Context) {
    var request struct {
        TaskID int64  `json:"taskID"`
        Videos []string `json:"videos"`
        Output string   `json:"output"`
    }

    if err := c.ShouldBindJSON(&request); err != nil {
        c.JSON(400, gin.H{"error": "无效的请求数据"})
        return
    }

    if len(request.Videos) < 2 {
        c.JSON(400, gin.H{"error": "至少需要两个视频进行合并"})
        return
    }

    // 拼接完整路径
    var fullVideoPaths []string
    for _, video := range request.Videos {
        fullVideoPaths = append(fullVideoPaths, filepath.Join(h.videoDir, video))
    }
    outputPath := filepath.Join(h.videoDir, request.Output)

    err := h.processor.Merge(fullVideoPaths, outputPath, request.TaskID)
    if err != nil {
        c.JSON(500, gin.H{"error": "视频合并失败: " + err.Error()})
        return
    }

    c.JSON(200, gin.H{"taskID": request.TaskID, "message": "视频合并成功: " + request.Output})
}

// 获取所有任务信息的接口
func (h *VideoHandler) ListTasks(c *gin.Context) {
    var tasks = h.processor.GetAllTasks() 
    c.JSON(http.StatusOK, gin.H{"tasks": tasks})
}

// 获取某个任务信息的接口
func (h *VideoHandler) GetTask(c *gin.Context) {
    taskIDStr := c.Param("taskID")
    taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "无效的任务ID"})
        return
    }

    taskInfo, err := h.processor.GetTaskInfo(taskID)
    if err == nil { 
        c.JSON(http.StatusOK, gin.H{"task": taskInfo})
    } else {
        c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
    }
}