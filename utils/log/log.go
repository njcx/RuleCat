package log

import (
	"io"
	"log"
	"os"
	"rule_engine_by_go/utils"
)

var (
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
)

func init() {

	errFile, err := os.OpenFile(utils.GetCurrentPath()+"/logs/nids.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("打开日志文件失败：", err)
	}

	Info = log.New(io.MultiWriter(os.Stdout, errFile), "Info:", log.Ldate|log.Ltime|log.Lshortfile)
	Warning = log.New(io.MultiWriter(os.Stdout, errFile), "Warning:", log.Ldate|log.Ltime|log.Lshortfile)
	Error = log.New(io.MultiWriter(os.Stderr, errFile), "Error:", log.Ldate|log.Ltime|log.Lshortfile)

}
