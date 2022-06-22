package comprise

// Methods for string slices
func Contains(ls []string, el string) bool {
	for _, value := range ls {
		if value == el {
			return true
		}
	}
	return false
}

func Map(ls []string, f func(string) string) (ret []string) {
	for _, value := range ls {
		ret = append(ret, f(value))
	}
	return
}

func Unique(ls []string) (ret []string) {
	keys := make(map[string]bool)
	for _, value := range ls {
		if _, dup := keys[value]; !dup {
			keys[value] = true
			ret = append(ret, value)
		}
	}
	return
}

func CreateEmptyMap(ls []string) map[string][]string {
	ret := make(map[string][]string)
	for _, value := range ls {
		ret[value] = []string{}
	}
	return ret
}

// Methods for maps
func Keys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, len(m))

	j := 0
	for k := range m {
		keys[j] = k
		j++
	}

	return keys
}

func Values[K comparable, V any](m map[K]V) []V {
	values := make([]V, len(m))

	j := 0
	for _, v := range m {
		values[j] = v
		j++
	}

	return values
}
