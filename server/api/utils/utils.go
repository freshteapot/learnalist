package utils

func IntArrayContains(items []int, find int) bool {
	for _, item := range items {
		if item == find {
			return true
		}
	}
	return false
}
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

func StringArrayRemoveAtIndex(s []string, i int) []string {
	s[len(s)-1], s[i] = s[i], s[len(s)-1]
	return s[:len(s)-1]
}
