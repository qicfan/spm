package helpers

import (
	"log"
	"os"
	//"fmt"
)

type SmLog struct {
	Logger *log.Logger
	File   *os.File
}

var Log SmLog

func (smLog *SmLog) Info(format string, v ...interface{}) {
	//fmt.Printf(format+"\n\n", v...)
	smLog.Logger.SetPrefix("[info] ")
	smLog.Logger.Printf(format+"\n\n", v...)
}

func (smLog *SmLog) Error(format string, v ...interface{}) {
	//fmt.Printf(format + "\n\n", v...)
	smLog.Logger.SetPrefix("[error] ")
	smLog.Logger.Printf(format+"\n\n", v...)
}

func (smLog *SmLog) Warning(format string, v ...interface{}) {
	//fmt.Printf(format + "\n\n", v...)
	smLog.Logger.SetPrefix("[warning] ")
	smLog.Logger.Printf(format+"\n\n", v...)
}

func CreateLogger(file string) SmLog {
	var logErr error
	Log.File, logErr = os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if logErr != nil {
		// 打开日志文件失败
		panic(logErr)
	}
	Log.Logger = log.New(Log.File, "info", log.LstdFlags)
	return Log
}

func CloseLogger() {
	Log.File.Close()
}
