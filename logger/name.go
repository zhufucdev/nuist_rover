package logger

func (l LogLevel) Name() string {
	switch l {
	case LOG:
		return "log"
	case INFO:
		return "info"
	case WARNING:
		return "warning"
	case EXCEPTION:
		return "exception"
	case ERROR:
		return "error"
	default:
		return "unknown"
	}
}
