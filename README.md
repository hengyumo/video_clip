# 视频剪辑工具 Video Clipper Tool

这是一个基于 VSCode、Cline 和 DeepSeek 开发的视频剪辑工具。作为一个对视频剪辑领域完全陌生的开发者，我在 AI 的帮助下快速构建了一个可用的 DEMO，总耗时约 12 小时（利用零碎时间完成）。

This is a video editing tool developed using VSCode, Cline, and DeepSeek. As a developer with no prior experience in video editing, I was able to quickly build a functional demo with the help of AI, completing it in approximately 12 hours (using spare time).

### 开发过程 Development Process

AI 辅助开发：借助 DeepSeek 的智能支持，我快速实现了核心功能。春节期间，由于 DeepSeek API 不稳定，我切换到了 GPT-4o mini，但效果明显不如 DeepSeek，最终通过手动调整代码完成了所有功能。

AI-Assisted Development: With the support of DeepSeek's intelligence, I rapidly implemented core functionalities. During the Spring Festival, due to instability in the DeepSeek API, I switched to GPT-4o mini, but the results were noticeably inferior. Ultimately, I manually adjusted the code to complete all features.

### 功能亮点 Key Features:

- 可视化页面：支持视频的剪切和合并操作。
**Visual Interface**: Supports video cutting and merging.

- 命令行支持：提供更灵活的操作方式。
**Command-Line Support**: Provides a more flexible operation mode.

### 体验与思考 Experience and Reflection
这次开发经历让我深刻感受到 AI 对开发效率的提升。AI 不仅降低了技术门槛，还预示着一个趋势：**未来，低层次的开发工作可能会被 AI 取代，而优秀的产品经理和市场营销人员将变得更加重要**。

This development experience gave me a deep appreciation for how AI enhances productivity. AI not only lowers technical barriers but also signals a trend: **in the future, low-level development tasks may be replaced by AI, while excellent product managers and marketing professionals will become increasingly important.**

## 项目简介 Project Overview
这是一个基于Go语言的视频剪辑工具，支持剪切、合并和转换视频文件。旨在简化视频处理流程，使用户能够轻松操作视频。

This is a video cutting tool based on Go language that supports cutting, merging, and converting video files. It aims to simplify the video processing workflow, allowing users to operate videos easily.

## 功能描述 Features
- 剪切视频 Cut Videos
- 合并视频 Merge Videos
- 转换视频格式 Convert Video Formats
- 查看任务进度 View Task Progress 

## 安装和运行说明 Installation and Usage
1. 安装 ffmpeg Install ffmpeg 
   https://www.ffmpeg.org/download.html
2. 安装Go环境 Install Go environment
2. 克隆项目 Clone the project
   ```bash
   git clone https://github.com/hengyumo/video_clip.git
   ```
3. 进入项目目录 Navigate to the project directory
   ```bash
   cd video_clip
   ```
4. 运行项目 Run the project
   ```bash
   go run cmd/clipper/main.go web
   ``` 
5. 访问页面 
   http://127.0.0.1:808/static/

## Web页面功能 Web Interface Features

- **启动Web服务 Start Web Server**: 通过运行命令 `clipper web -p PORT`，用户可以指定启动Web界面的端口（默认端口为8080）。

- **任务进度查看 Task Progress Viewing**: Web界面提供任务列表，用户可以实时查看当前所有正在处理的任务及其进度。

- **用户友好的操作 User-Friendly Operations**: 用户可以通过简单的表单填写输入视频及输出视频信息来执行剪切、合并和转换操作，无需手动输入命令行。 

## 命令 Command Usage
### 剪切视频 Cut Video
运行以下命令剪切视频文件：
```bash
clipper cut -i input.mp4 -o output.mp4 -start 00:00:10 -end 00:00:30
```

### 合并视频 Merge Videos
运行以下命令合并视频文件：
```bash
clipper merge -o output.mp4 video1.mp4 video2.mp4
```

### 启动Web服务 Start Web Server
运行Web服务以便于通过界面操作：
```bash
clipper web -p 8080
```

### 转换视频格式 Convert Video Format
运行以下命令转换视频格式：
```bash
clipper convert -i input.avi -o output.mp4
```

### 批量转换目录中的视频 Batch Convert Videos in Directory
运行以下命令批量转换目录中的视频文件：
```bash
clipper convertdir -d /path/to/video/directory
```

## 贡献 Contributing
欢迎贡献，您的任何改进、bug修复、建议都是欢迎的！

Contributions are welcome! Any improvements, bug fixes, and suggestions are appreciated!

## 许可证 License
该项目采用 MIT 许可证。

This project is licensed under the MIT License.
