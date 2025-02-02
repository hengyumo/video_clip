package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync/atomic"
	"video-clipper/pkg/config"
	"video-clipper/pkg/ffmpeg"
	"video-clipper/pkg/web"
)

var currentTaskID = atomic.Int64{}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	command := os.Args[1]
	switch command {
	case "cut":
		handleCutCommand()
	case "merge":
		handleMergeCommand()
	case "web":
		startWebServer()
	case "convert":
		handleConvertCommand()
	case "convertdir":
		handleConvertDirCommand()
	default:
		fmt.Println("未知命令:", command)
		printUsage()
	}
}

func startWebServer() {
	webCmd := flag.NewFlagSet("web", flag.ExitOnError)
	port := webCmd.String("p", "", "端口")
	webCmd.Parse(os.Args[2:])
	
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	const defaultPort = "8080"
	if *port == "" {
		*port = defaultPort
	}

	server, err := web.NewServer(cfg.VideoDir)
	if err != nil {
		fmt.Println("启动Web服务失败:", err)
		return
	}
	server.StartServer(*port)
}

func handleCutCommand() {
	cutCmd := flag.NewFlagSet("cut", flag.ExitOnError)
	input := cutCmd.String("i", "", "输入视频文件")
	output := cutCmd.String("o", "", "输出视频文件")
	start := cutCmd.String("start", "", "开始时间 (HH:MM:SS)")
	end := cutCmd.String("end", "", "结束时间 (HH:MM:SS)")

	cutCmd.Parse(os.Args[2:])

	if *input == "" || *output == "" || *start == "" || *end == "" {
		cutCmd.Usage()
		return
	}

	processor, err := ffmpeg.NewVideoProcessor()
	if err != nil {
		fmt.Println("错误:", err)
		return
	}

	taskID := currentTaskID.Add(1)
	err = processor.Cut(*input, *output, *start, *end, taskID)
	if err != nil {
		fmt.Println("视频剪切失败:", err)
		return
	}

	fmt.Println("视频剪切成功:", *output)
}

func handleMergeCommand() {
	mergeCmd := flag.NewFlagSet("merge", flag.ExitOnError)
	output := mergeCmd.String("o", "", "输出视频文件")
	mergeCmd.Parse(os.Args[2:])

	if len(mergeCmd.Args()) < 1 || *output == "" {
		mergeCmd.Usage()
		return
	}

	processor, err := ffmpeg.NewVideoProcessor()
	if err != nil {
		fmt.Println("错误:", err)
		return
	}

	taskID := currentTaskID.Add(1)
	err = processor.Merge(mergeCmd.Args(), *output, taskID)
	if err != nil {
		fmt.Println("视频合并失败:", err)
		return
	}

	fmt.Println("视频合并成功:", *output)
}

func handleConvertCommand() {
	convertCmd := flag.NewFlagSet("convert", flag.ExitOnError)
	input := convertCmd.String("i", "", "输入视频文件")
	output := convertCmd.String("o", "", "输出视频文件 (默认mp4格式)")

	convertCmd.Parse(os.Args[2:])

	if *input == "" || *output == "" {
		convertCmd.Usage()
		return
	}

	processor, err := ffmpeg.NewVideoProcessor()
	if err != nil {
		fmt.Println("错误:", err)
		return
	}

	err = processor.Convert(*input, *output, currentTaskID.Add(1))
	if err != nil {
		fmt.Println("视频转换失败:", err)
		return
	}

	fmt.Println("视频转换成功:", *output)
}

func handleConvertDirCommand() {
	convertDirCmd := flag.NewFlagSet("convertdir", flag.ExitOnError)
	dir := convertDirCmd.String("d", "", "要转换的视频目录")
	convertDirCmd.Parse(os.Args[2:])

	if *dir == "" {
		convertDirCmd.Usage()
		return
	}

	processor, err := ffmpeg.NewVideoProcessor()
	if err != nil {
		fmt.Println("错误:", err)
		return
	}

	// 遍历目录查找视频文件
	files, err := os.ReadDir(*dir)
	if err != nil {
		fmt.Println("读取目录失败:", err)
		return
	}

	var toConvert []string
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		ext := filepath.Ext(file.Name())
		if ext == ".mp4" {
			continue
		}

		// 检查常见视频格式
		switch ext {
		case ".avi", ".mov", ".mkv", ".flv", ".wmv":
			// 检查是否已存在同名mp4文件
			mp4Path := filepath.Join(*dir, file.Name()[:len(file.Name())-len(ext)]+".mp4")
			if _, err := os.Stat(mp4Path); os.IsNotExist(err) {
				toConvert = append(toConvert, file.Name())
			}
		}
	}

	if len(toConvert) == 0 {
		fmt.Println("没有需要转换的视频文件")
		return
	}

	fmt.Println("待转换的视频文件:")
	for _, file := range toConvert {
		fmt.Println("-", file)
	}

	fmt.Print("\n开始转换...\n\n")
	for _, file := range toConvert {
		inputPath := filepath.Join(*dir, file)
		outputPath := filepath.Join(*dir, file[:len(file)-len(filepath.Ext(file))]+".mp4")
		
		fmt.Printf("正在转换: %s -> %s\n", file, outputPath)
		err := processor.Convert(inputPath, outputPath, currentTaskID.Add(1))	
		if err != nil {
			fmt.Printf("转换失败: %s (%v)\n", file, err)
			continue
		}
		fmt.Printf("转换成功: %s\n", outputPath)
	}
}

func printUsage() {
	fmt.Println("视频剪辑工具")
	fmt.Println("使用方法:")
	fmt.Println("  cut       - 剪切视频")
	fmt.Println("  merge     - 合并视频")
	fmt.Println("  web       - 启动Web服务")
	fmt.Println("  convert   - 转换视频格式")
	fmt.Println("  convertdir - 批量转换目录中的视频")
	fmt.Println("\n使用 'clipper [命令] -h' 查看具体命令帮助")
}
