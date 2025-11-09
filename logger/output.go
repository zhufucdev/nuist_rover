package logger

import (
	"fmt"
	"os"
	"time"
)

func (l Logger) Println(level LogLevel, format string, args ...any) {
	if l.Level <= level {
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		fmt.Printf("[%s] [%s] ", timestamp, level.Name())
		fmt.Printf(format+"\n", args...)
	}
}

func (l Logger) Log(format string, args ...any) {
	l.Println(LOG, format, args...)
}

func (l Logger) Info(format string, args ...any) {
	l.Println(INFO, format, args...)
}

func (l Logger) Warning(format string, args ...any) {
	l.Println(WARNING, format, args...)
}

func (l Logger) Exception(format string, args ...any) {
	l.Println(EXCEPTION, format, args)
}

func (l Logger) Error(format string, args ...any) {
	l.Println(ERROR, format, args...)
	os.Exit(128)
}
