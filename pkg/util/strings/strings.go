package strings

func OrDefault(s, def string) string {
	if s != "" {
		return s
	}
	return def
}
