package logger

type Logger struct {
	Level LogLevel
}

type LogLevel int

const (
	LOG LogLevel = iota
	INFO
	WARNING
	EXCEPTION
	ERROR
	UNKNOWN
)
