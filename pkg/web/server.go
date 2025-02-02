package web

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"video-clipper/pkg/ffmpeg"
)

type Server struct {
	videoDir string
	processor *ffmpeg.VideoProcessor
	router *gin.Engine
}

func NewServer(videoDir string) (*Server, error) {
	processor, err := ffmpeg.NewVideoProcessor()
	if err != nil {
		return nil, err
	}

	router := gin.Default()
	
	return &Server{
		videoDir: videoDir,
		processor: processor,
		router: router,
	}, nil
}

func (s *Server) setupRoutes() {
	handler := NewVideoHandler(s.videoDir, s.processor)
	
	api := s.router.Group("/api")
	{
		videos := api.Group("/videos")
		{
			videos.GET("", handler.ListVideos)
			videos.GET("/:filename", handler.GetVideo)
			videos.POST("/cut", handler.CutVideo)
			videos.POST("/merge", handler.MergeVideos)
			videos.POST("/dir", handler.UpdateVideoDir)
		}

		tasks := api.Group("/tasks")
		{
			tasks.GET("", handler.ListTasks)
			tasks.GET("/:taskID", handler.GetTask)
		}
	}
	
	// 静态文件服务
	s.router.Static("/static", "./static")
}

func (s *Server) StartServer(port string) {
	s.setupRoutes() 
	if err := s.router.Run(":" + port); err != nil {
		panic(fmt.Sprintf("failed to start server: %v", err))
	}
}
