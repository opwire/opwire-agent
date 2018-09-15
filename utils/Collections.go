package utils

import "strings"

func Filter(vs []string, f func(string, int) bool) []string {
	vsf := make([]string, 0)
	for i, v := range vs {
		if f(v, i) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

func Map(vs []string, f func(string, int) string) []string {
	vsm := make([]string, len(vs))
	for i, v := range vs {
		vsm[i] = f(v, i)
	}
	return vsm
}

func Split(str string, sep string) []string {
	arr := strings.Split(str, sep)
	arr = Map(arr, func(s string, idx int) string {
		return strings.TrimSpace(s)
	})
	arr = Filter(arr, func(s string, idx int) bool {
		return len(s) > 0
	})
	return arr
}