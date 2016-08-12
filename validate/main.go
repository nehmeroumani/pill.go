package validate

import (
	"regexp"
	"unicode/utf8"
)

func Email(email string) bool {
	Re := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	return Re.MatchString(email)
}

func StringLength(str string, minLength int, maxLength int) bool {
	strLength := utf8.RuneCountInString(str)
	if strLength < minLength || strLength > maxLength {
		return false
	}
	return true
}
