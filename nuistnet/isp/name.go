package isp

func Name(isp Type) string {
	switch isp {
	case INTERNAL:
		return "internal"
	case TELECOM:
		return "telecom"
	case UNICOM:
		return "unicom"
	case MOBILE:
		return "mobile"
	default:
		return ""
	}
}
