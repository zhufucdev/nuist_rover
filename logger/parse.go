package logger

func ParseLevel(name string) LogLevel {
	switch name {
	case "log":
		return LOG
	case "info":
		return INFO
	case "warning":
		return WARNING
	case "exception":
		return EXCEPTION
	default:
		return UNKNOWN
	}
}
