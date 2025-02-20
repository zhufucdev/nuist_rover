package helper

import "strings"

func GetCharset(contentType string) string {
	key := "charset="
	index := strings.Index(contentType, key)
	if index < 0 {
		return "utf-8"
	}
	return contentType[index+len(key):]
}
