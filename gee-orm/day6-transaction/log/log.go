package log

import (
	"io/ioutil"
	"log"
	"os"
	"sync"
)

var (
	// Lshortfile----final file name element and line number: d.go:23. overrides
	errorLog = log.New(os.Stdout, "\033[31m[error]\033[0m ", log.LstdFlags|log.Lshortfile)
	infoLog  = log.New(os.Stdout, "\033[34m[info]\033[0m ", log.LstdFlags|log.Lshortfile)
	loggers  = []*log.Logger{errorLog, infoLog}
	mu       sync.Mutex
)

// 定义方法型变量
var (
	Error  = errorLog.Println
	Errorf = errorLog.Printf
	Info   = infoLog.Println
	Infof  = infoLog.Printf
)

// 支持日志分级
// 不同层级日志用不同颜色显示
const (
	InfoLevel = iota
	// 1
	ErrorLevel
	// 2
	Disabled
)

// SetLevel controls log level
func SetLevel(level int) {
	mu.Lock()
	defer mu.Unlock()

	for _, logger := range loggers {
		logger.SetOutput(os.Stdout)
	}

	// 如果设置为ErrorLevel infoLog
	// 输出重定向为ioutil.Discard，即不打印日志
	if ErrorLevel < level {
		errorLog.SetOutput(ioutil.Discard)
	}
	if InfoLevel < level {
		infoLog.SetOutput(ioutil.Discard)
	}
}
