package log

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
)

func init() {

	dir, _ := os.Getwd()
	currentPath := strings.Replace(dir, "\\", "/", -1)
	logFilePath := filepath.Join(currentPath, "logs", "rulecat.log")
	logDir := filepath.Dir(logFilePath)

	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		if err := os.MkdirAll(logDir, 0755); err != nil {
			log.Fatalf("Create log dir err: %v", err)
		}
	}

	errFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Open log file err: %v", err)
	}

	Info = log.New(io.MultiWriter(os.Stdout, errFile), "Info:", log.Ldate|log.Ltime|log.Lshortfile)
	Warning = log.New(io.MultiWriter(os.Stdout, errFile), "Warning:", log.Ldate|log.Ltime|log.Lshortfile)
	Error = log.New(io.MultiWriter(os.Stderr, errFile), "Error:", log.Ldate|log.Ltime|log.Lshortfile)

}
