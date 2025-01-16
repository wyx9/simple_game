package pkg

import (
	"fmt"
	"io"
	"log"
	"os"
)

type Logger struct {
	Logs chan string
}

var logger Logger

const (
	flag       = log.Ldate | log.Ltime | log.Lshortfile
	preDebug   = "[DEBUG]"
	preInfo    = "[INFO]"
	preWarning = "[WARNING]"
	preError   = "[ERROR]"
)

var (
	logFile       io.Writer
	debugLogger   *log.Logger
	infoLogger    *log.Logger
	warningLogger *log.Logger
	errorLogger   *log.Logger
)

func init() {
	logger = Logger{
		Logs: make(chan string),
	}

	var err error

	logFile, err = os.OpenFile("./logs/game.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		logFile, err = os.OpenFile("./game.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
		if err != nil {
			log.Fatalf("create log file err %+v", err)
		}
	}

	debugLogger = log.New(logFile, preDebug, flag)
	infoLogger = log.New(logFile, preInfo, flag)
	warningLogger = log.New(logFile, preWarning, flag)
	errorLogger = log.New(logFile, preError, flag)
}

func DEBUG(v ...interface{}) {
	debugLogger.Print(v)
	sprintf := fmt.Sprintf("\033[33m[DEBUG] %v\033[0m\n", fmt.Sprint(v...))
	defer func() {
		logger.Logs <- sprintf
	}()
}

func INFO(v ...interface{}) {
	infoLogger.Print(v)
	sprintf := fmt.Sprintf("\033[32m[INFO] %v \033[0m", fmt.Sprint(v...))
	defer func() {
		logger.Logs <- sprintf
	}()
}

func WARNING(v ...interface{}) {
	warningLogger.Print(v...)
	sprintf := fmt.Sprintf("\033[34m[WARNING] %v \033[0m\n", fmt.Sprint(v...))
	defer func() {
		logger.Logs <- sprintf
	}()
}

func ERROR(v ...interface{}) {
	errorLogger.Print(v...)
	sprintf := fmt.Sprintf("\033[31m[ERROR] %v \033[0m\n", fmt.Sprint(v...))
	defer func() {
		logger.Logs <- sprintf
	}()
}

func (l *Logger) appendLog(log string) {
	l.Logs <- log
}

func StartLog() {
	go func() {
		for {
			select {
			case s := <-logger.Logs:
				log.Print(s)
			}
		}
	}()

}

func SetOutputPath(path string) {
	var err error
	logFile, err = os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		log.Fatalf("create log file err %+v", err)
	}
	debugLogger.SetOutput(logFile)
	infoLogger.SetOutput(logFile)
	warningLogger.SetOutput(logFile)
	errorLogger.SetOutput(logFile)
}
