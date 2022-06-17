package comprise

// Methods for string slices
func Contains(s []string, el string) bool {
	for _, value := range s {
		if value == el {
			return true
		}
	}
	return false
}

func Map(ls []string, f func(string) string) (ret []string) {
	for _, s := range ls {
		ret = append(ret, f(s))
	}
	return
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
