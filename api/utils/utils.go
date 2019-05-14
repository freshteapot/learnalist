package utils

func StringArrayContains(items []string, find string) bool {
	for _, item := range items {
		if item == find {
			return true
		}
	}
	return false
}

func StringArrayIndexOf(items []string, search string) int {
	for i, v := range items {
		if v == search {
			return i
		}
	}
	return -1
}
