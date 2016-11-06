package logger

import (
	"fmt"
	"path"
	"runtime"
	"strings"
)

type Logger struct {
	Debug bool
}

func Init(debug bool) *Logger {
	return &Logger{
		Debug: debug,
	}
}

func (l *Logger) Error(m error) {
	if l.Debug {
		if _, file, line, ok := runtime.Caller(1); ok {
			fmt.Printf("DEBUG:%s:%d [ERROR] %s\n", truncateDir(file), line, m)
		}
	}
}

func (l *Logger) Info(m string) {
	if l.Debug {
		if _, file, line, ok := runtime.Caller(1); ok {
			fmt.Printf("DEBUG:%s:%d [INFO] %s\n", truncateDir(file), line, m)
		}
	}
}

func truncateDir(file string) string {
	return strings.Replace(file, path.Dir(file), "", 1)
}
