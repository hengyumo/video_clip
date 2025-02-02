
// 将秒数格式化为 HH:MM:SS
function formatTime(seconds) {
    const h = Math.floor(seconds / 3600);
    const m = Math.floor((seconds % 3600) / 60);
    const s = Math.floor(seconds % 60);
    return `${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`;
}

// 获取当前时间戳，精确到秒
function getCurrentTimestamp() {
    return Math.floor(Date.now() / 1000);
} 