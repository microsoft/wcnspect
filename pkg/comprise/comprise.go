package comprise

func Contains(s []string, el string) bool {
	for _, value := range s {
		if value == el {
			return true
		}
	}
	return false
}

func GetKeys(m map[string]int32) []string {
	keys := make([]string, len(m))

	j := 0
	for k := range m {
		keys[j] = k
		j++
	}

	return keys
}
