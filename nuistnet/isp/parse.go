package isp

func Parse(name string) Type {
	switch name {
	case "internal":
		fallthrough
	case "校园网":
		return INTERNAL
	case "telecom":
		fallthrough
	case "中国电信":
		return TELECOM
	case "unicom":
		fallthrough
	case "中国联通":
		return UNICOM
	case "mobile":
		fallthrough
	case "中国移动":
		return MOBILE
	}
	return UNKNOWN
}
