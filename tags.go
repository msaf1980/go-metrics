package metrics

import (
	"sort"
	"strings"
)

// MergeTags merge two tag maps into sorted tags string representation (separated by comma)
func MergeTags(a, b map[string]string) (map[string]string, string) {
	var dst map[string]string
	if a == nil || len(a) == 0 {
		if b == nil || len(b) == 0 {
			return nil, ""
		}
		dst = b
	} else if b == nil || len(b) == 0 {
		dst = a
	} else {
		dst = make(map[string]string)
		for k, v := range a {
			dst[k] = v
		}
		for k, v := range b {
			if _, exist := dst[k]; !exist {
				dst[k] = v
			}
		}
	}
	tags := make([]string, 0, len(dst))
	for k, v := range dst {
		tags = append(tags, k+"="+v)
	}
	sort.Strings(tags)
	return dst, ";" + strings.Join(tags, ";")
}
