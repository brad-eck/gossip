package hosts

import (
	"strconv"
	"strings"
)

func ExpandBrace(args []string) []string {
	var result []string

	for _, arg := range args {
		if strings.Contains(arg, "{") && strings.Contains(arg, "..") && strings.Contains(arg, "}") {
			expanded := expandSingle(arg)
			result = append(result, expanded...)
		} else {
			result = append(result, arg)
		}
	}

	return result
}

func expandSingle(pattern string) []string {
	startBrace := strings.Index(pattern, "{")
	endBrace := strings.Index(pattern, "}")
	if startBrace == -1 || endBrace == -1 || startBrace >= endBrace {
		return []string{pattern}
	}

	rangeStr := pattern[startBrace+1 : endBrace]
	parts := strings.SplitN(rangeStr, "..", 2)
	if len(parts) != 2 {
		return []string{pattern}
	}

	prefix := pattern[:startBrace]
	suffix := pattern[endBrace+1:]

	start, err1 := strconv.Atoi(parts[0])
	end, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil {
		return []string{pattern}
	}

	if start > end {
		start, end = end, start
	}

	var expanded []string
	for i := start; i <= end; i++ {
		numStr := strconv.Itoa(i)
		// Preserve zero-padding if original had it
		if len(parts[0]) > len(numStr) && strings.HasPrefix(parts[0], "0") {
			numStr = strings.Repeat("0", len(parts[0])-len(numStr)) + numStr
		}
		expanded = append(expanded, prefix+numStr+suffix)
	}

	return expanded
}