package metrics

import (
	"sort"
	"strings"
)

// MergeTags merge two tag maps into sorted tags string representation (separated by comma)
func MergeTags(a, b map[string]string) string {
	dst := make(map[string]string)
	for k, v := range a {
		dst[k] = v
	}
	for k, v := range b {
		if _, exist := dst[k]; !exist {
			dst[k] = v
		}
	}
	if len(dst) == 0 {
		return ""
	}
	tags := make([]string, 0, len(dst))
	for k, v := range dst {
		tags = append(tags, k+"="+v)
	}
	sort.Strings(tags)
	return ";" + strings.Join(tags, ";")
}
