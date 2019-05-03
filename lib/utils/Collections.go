package utils

import (
	"sort"
	"strings"
)

func Index(vs []string, t string) int {
	for i, v := range vs {
		if v == t {
			return i
		}
	}
	return -1
}

func Contains(vs []string, t string) bool {
	return Index(vs, t) >= 0
}

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

func Keys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k, _ := range m {
		keys = append(keys, k)
	}
	return keys
}

func Reverse(arr []string) []string {
	size := len(arr)
	for i:=0; i<size/2; i++ {
		tmp := arr[i]
		arr[i] = arr[size -1 - i]
		arr[size -1 - i] = tmp
	}
	return arr
}

func SortDesc(arr []string) []string {
	sort.Strings(arr)
	return Reverse(arr)
}

func FirstHasPrefix(arr []string, prefix string, cropped bool) (string, int) {
	for i, str := range arr {
		if strings.HasPrefix(str, prefix) {
			if cropped {
				return strings.TrimPrefix(str, prefix), i
			}
			return str, i
		}
	}
	return "", -1
}

func FirstHasSuffix(arr []string, suffix string, cropped bool) (string, int) {
	for i, str := range arr {
		if strings.HasSuffix(str, suffix) {
			if cropped {
				return strings.TrimSuffix(str, suffix), i
			}
			return str, i
		}
	}
	return "", -1
}

func FilterByPrefix(arr []string, prefix string) ([]string) {
	result := make([]string, 0)
	for _, str := range arr {
		if strings.HasPrefix(str, prefix) {
			result = append(result, str)
		}
	}
	return result
}

func FilterBySuffix(arr []string, suffix string) ([]string) {
	result := make([]string, 0)
	for _, str := range arr {
		if strings.HasSuffix(str, suffix) {
			result = append(result, str)
		}
	}
	return result
}

func SubSlice(arr []string, from int, to int) []string {
	if arr == nil {
		return nil
	}
	size := len(arr)
	if size <= from {
		return []string{}
	}
	if size < to {
		return arr[from:]
	}
	return arr[from:to]
}

func SliceHead(arr []string, before int) []string {
	if arr == nil {
		return nil
	}
	if before < len(arr) {
		return arr[:before]
	}
	return arr[:]
}

func SliceTail(arr []string, from int) []string {
	if arr == nil {
		return nil
	}
	if len(arr) <= from {
		return []string{}
	}
	return arr[from:]
}
