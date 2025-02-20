package logger

import (
	"fmt"
	"os"
)

func (l Logger) Println(level LogLevel, format string, args ...any) {
	if l.Level <= level {
		fmt.Printf("["+level.Name()+"] "+format+"\n", args...)
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
