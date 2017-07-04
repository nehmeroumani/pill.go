package helpers

func BytesToString(b []byte) string {
	if len(b) > 0 {
		return string(b)
	}
	return ""
}
