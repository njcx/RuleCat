package log

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"rulecat/utils"
)

var (
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
)

func init() {
	logFilePath := filepath.Join(utils.GetCurrentPath(), "logs", "rulecat.log")
	logDir := filepath.Dir(logFilePath)

	// 检查并创建目录
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		if err := os.MkdirAll(logDir, 0755); err != nil {
			log.Fatalf("创建日志目录失败: %v", err)
		}
	}

	errFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("打开日志文件失败: %v", err)
	}
	defer errFile.Close()

	Info = log.New(io.MultiWriter(os.Stdout, errFile), "Info:", log.Ldate|log.Ltime|log.Lshortfile)
	Warning = log.New(io.MultiWriter(os.Stdout, errFile), "Warning:", log.Ldate|log.Ltime|log.Lshortfile)
	Error = log.New(io.MultiWriter(os.Stderr, errFile), "Error:", log.Ldate|log.Ltime|log.Lshortfile)

}
