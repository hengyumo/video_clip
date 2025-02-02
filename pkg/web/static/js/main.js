// 任务队列
let taskQueue = []; 

// 创建新任务
function createTask(title, tid) {
    const taskId = tid || getCurrentTimestamp();
    const task = {
        id: taskId,
        title: title,
        progress: 0,
        status: 'processing'
    };
    
    // 创建进度条
    const template = document.getElementById('task-template');
    const clone = document.importNode(template.content, true);
    const taskElement = clone.querySelector('.task-progress');
    taskElement.id = `task-${taskId}`;
    
    // 设置任务信息
    taskElement.querySelector('.task-title').textContent = title;
    taskElement.querySelector('.task-percent').textContent = '0%';
    taskElement.querySelector('.progress-bar').classList.add('progress-bar-animated');
    
    // 添加到任务列表
    document.getElementById('tasks-container').appendChild(taskElement);
    
    // 添加到队列
    taskQueue.push(task);
    
    return taskId;
}

// 更新任务进度
function updateTaskProgress(taskId, progress) {
    const task = taskQueue.find(t => t.id === taskId);
    if (!task) return;

    task.progress = progress;
    
    const taskElement = document.getElementById(`task-${taskId}`);
    if (!taskElement) return;

    // 更新进度条
    const progressBar = taskElement.querySelector('.progress-bar');
    progressBar.style.width = `${progress}%`;
    progressBar.setAttribute('aria-valuenow', progress);
    
    // 更新百分比
    taskElement.querySelector('.task-percent').textContent = `${progress}%`;
    
    // 处理完成状态
    if (progress >= 100) {
        task.status = 'completed';
        progressBar.classList.remove('progress-bar-striped', 'progress-bar-animated');
        progressBar.classList.add('bg-success');
        
        // 显示查看按钮
        const viewBtn = taskElement.querySelector('.task-view-btn');
        viewBtn.style.display = 'inline-block';
        viewBtn.onclick = () => playTaskResult(taskId);
        
        // 更新状态
        taskElement.querySelector('.task-status').textContent = '已完成';
    }
}

// 播放任务结果
function playTaskResult(taskId) {
    const task = taskQueue.find(t => t.id === taskId);
    if (!task) return;

    // 根据任务类型播放结果
    if (task.type === 'cut') {
        playVideo('video-select', player, task.output);
    } else if (task.type === 'merge') {
        playVideo('merge-video-select', mergeVideoPlayer, task.output);
    }
}

// 设置开始时间为当前播放时间
function setStartTime() {
    const currentTime = player.currentTime();
    document.getElementById('start-time').value = formatTime(currentTime);
}

// 设置结束时间为当前播放时间
function setEndTime() {
    const currentTime = player.currentTime();
    document.getElementById('end-time').value = formatTime(currentTime);
}

// 获取视频列表
async function loadVideos() {
    try {
        const response = await fetch('/api/videos');
        const data = await response.json();
        const videos = data.videos;
        const select = document.getElementById('video-select');

        // 更新视频目录路径
        const videoDir = document.getElementById('video-directory');
        videoDir.value = data.videoDir;

        // 清空现有选项
        select.innerHTML = '<option value="">请选择视频...</option>';

        // 添加视频选项
        videos.forEach(video => {
            const option = document.createElement('option');
            option.value = video;
            option.textContent = video;
            select.appendChild(option);
        });
    } catch (error) {
        console.error('加载视频列表失败:', error);
    }
}

// 播放选中的视频
function playSelectedVideo() {
    playVideo('video-select', player);
}

// 播放视频
function playVideo(selectId, player, videoPath) {
    const select = document.getElementById(selectId);
    const video = videoPath || select.value;

    if (video) {
        const videoUrl = `/api/videos/${encodeURIComponent(video)}`;
        const ext = video.split('.').pop().toLowerCase();

        player.src({
            src: videoUrl,
            type: ext === 'mp4' ? 'video/mp4' : 
                  ext === 'mkv' ? 'video/x-matroska' : 
                  ext === 'avi' ? 'video/x-msvideo' : 
                  ext === 'mov' ? 'video/quicktime' : 'video/mp4'
        });
        player.play();
    }
}

// 剪切视频函数
async function cutVideo() {
    const select = document.getElementById('video-select');
    const start = document.getElementById('start-time').value;
    const end = document.getElementById('end-time').value;
    const input = select.value;

    if (!input || !start || !end) {
        alert('请选择视频并填写开始和结束时间');
        return;
    }

    const output = prompt('请输入输出文件名:', `cut_${input}`);
    if (!output) return;

    // 创建任务
    const taskId = createTask(`剪切 ${input} 从 ${start} 到 ${end}`);
    const task = taskQueue.find(t => t.id === taskId);
    task.type = 'cut';
    task.output = output;
        
    let interval = setInterval(function(){
        loadTaskProgress(taskId, interval);
    }, 3000); 

    try {
        const response = await fetch('/api/videos/cut', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                taskId,
                input,
                output,
                start,
                end
            })
        });
        
        if (response.ok) {  
            loadVideos(); // 刷新视频列表
        } else {
            throw new Error('剪切失败');
        }
    } catch (error) {
        console.error('视频剪切失败:', error); 
    }
}

// 合并视频相关功能
let mergeVideosList = [];
let mergeVideosIndex = 0;

function addMergeVideo() {
    mergeVideosIndex++;
    const select = document.createElement('select');
    select.className = 'form-select mb-2 w-auto';
    select.id = `merge-video-select-${mergeVideosIndex}`;
    select.innerHTML = document.getElementById('video-select').innerHTML;

    select.onchange = function() {
        playVideo(select.id, mergeVideoPlayer);
    }

    const removeBtn = document.createElement('button');
    removeBtn.className = 'btn btn-danger btn-sm ms-2';
    removeBtn.textContent = '移除';
    removeBtn.onclick = function () {
        this.parentElement.remove();
    };

    const container = document.createElement('div');
    container.className = 'd-flex align-items-center mb-2';
    container.appendChild(select);
    container.appendChild(removeBtn);

    document.getElementById('merge-video-list').appendChild(container);
}

// 合并视频函数
async function mergeVideos() {
    const output = document.getElementById('output-name').value;
    if (!output) {
        alert('请输入输出文件名');
        return;
    }

    const videos = Array.from(
        document.querySelectorAll('#merge-video-list select')
    ).map(select => select.value);

    if (videos.length < 2) {
        alert('请选择至少两个视频进行合并');
        return;
    }

    // 创建任务
    const taskId = createTask(`合并视频: ${videos}`);
    const task = taskQueue.find(t => t.id === taskId);
    task.type = 'merge';
    task.output = output;
    let interval = setInterval(function(){
        loadTaskProgress(taskId, interval);
    }, 3000); 

    try {
        const response = await fetch('/api/videos/merge', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                taskId,
                videos,
                output
            })
        });

        if (response.ok) {  
            loadVideos(); // 刷新视频列表
        } else {
            throw new Error('合并失败');
        }
    } catch (error) {
        console.error('视频合并失败:', error); 
    }
}

// 更新视频目录
async function updateVideoDir() {
    const inputDir = document.getElementById('video-directory').value;
    if (!inputDir) {
        alert('请输入有效的目录路径!');
        return;
    }

    try {
        const response = await fetch('/api/videos/dir', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                videoDir: inputDir
            })
        });

        if (response.ok) {
            const data = await response.json();
            console.log('成功:', data); 
            loadVideos(); // 刷新视频列表
        } else {
            throw new Error('路径检查失败');
        }
    } catch (error) {
        console.error('错误:', error);
        alert('更新视频目录失败');
    }
}

// 获取任务列表的 API 调用
function loadTaskList() {
    fetch('/api/tasks')
        .then(response => response.json())
        .then(data => {
            if (data.tasks) {
                data.tasks.forEach(task => {
                    createTask(task.Description, task.TaskID);
                    const tt = taskQueue.find(t => t.id === task.TaskID);
                    tt.type = task.TaskType;
                    tt.output = task.OutputFile.split('\\').pop();
                    updateTaskProgress(task.TaskID, task.Progress); 
                    if (task.Status === "running") {
                        let interval = setInterval(function(){
                            loadTaskProgress(task.TaskID, interval);
                        }, 3000);
                    }
                });
            } else {
                console.error('Failed to load tasks:', data);
            }
        })
        .catch(error => console.error('Error fetching tasks:', error));
}

// 获取某个任务ID的当前进度并更新
function loadTaskProgress(taskID, interval) {
    fetch(`/api/tasks/${taskID}`)
        .then(response => response.json())
        .then(data => {
            if (data.task) {
                const task = taskQueue.find(t => t.id === data.task.TaskID);
                if (!task) return;
                task.output = data.task.OutputFile.split('\\').pop();
                updateTaskProgress(data.task.TaskID, data.task.Progress);
                if (interval && data.task.Progress >= 100) {
                    clearInterval(interval);
                }
            } else {
                console.error('Failed to load task progress:', data);
            }
        })
        .catch(error => console.error('Error fetching task progress:', error));
}

// 初始化加载视频列表
document.addEventListener('DOMContentLoaded', function() {
    loadVideos();
    loadTaskList();
});
