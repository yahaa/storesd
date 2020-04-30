package utils

import "sort"

// StringSet 去重
func StringSet(as []string) []string {
	mp := make(map[string]bool)
	for _, item := range as {
		mp[item] = true
	}

	res := make([]string, 0)
	for k := range mp {
		res = append(res, k)
	}
	return res
}

// StringSliceEqual 比较数组内容是否相同
func StringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	sort.Strings(a)
	sort.Strings(b)

	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
