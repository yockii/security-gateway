package server

import (
	"regexp"
	"sort"
	"strings"
)

var regex = regexp.MustCompile(`\{[^}]*\}`)

// splitPath 分割路径，注意路径中可能存在正则表达式，约定：正则表达式以{}包裹。
func splitPath(path string) []string {
	var segments []string
	// 正则表达式匹配{}包裹的部分
	// 使用正则表达式替换路径中的{}包裹的部分，替换为特殊字符，这里使用了'\x00'，这是一个不会在路径中出现的字符
	tempPath := regex.ReplaceAllString(path, "\x00")
	parts := strings.Split(tempPath, "/")
	for _, part := range parts {
		if part == "\x00" {
			// 如果部分是特殊字符，那么这个部分是一个正则表达式，从原始路径中提取出来
			match := regex.FindString(path)
			segments = append(segments, match)
			// 从原始路径中删除已经处理过的正则表达式部分
			path = strings.Replace(path, match, "", 1)
		} else {
			segments = append(segments, part)
		}
	}
	return segments
}

// sortSegments 对子节点的路径片段进行排序，按照字典序，*放在最后，{}包裹的正则表达式次之
func sortSegments(segments []string) {
	sort.Slice(segments, func(i, j int) bool {
		iContainsStar := strings.Contains(segments[i], "*")
		jContainsStar := strings.Contains(segments[j], "*")
		iContainsRegex := strings.Contains(segments[i], "{") && strings.Contains(segments[i], "}")
		jContainsRegex := strings.Contains(segments[j], "{") && strings.Contains(segments[j], "}")

		if iContainsStar {
			return false
		}
		if jContainsStar {
			return true
		}
		if iContainsRegex && jContainsRegex {
			return segments[i] < segments[j]
		}
		if iContainsRegex {
			return false
		}
		if jContainsRegex {
			return true
		}
		return segments[i] < segments[j]
	})
}

func matchSegment(segment, path string, reg *regexp.Regexp) bool {
	if segment == "*" {
		return true
	}
	if strings.Contains(segment, "{") && strings.Contains(segment, "}") {
		// 正则表达式匹配
		return reg.MatchString(path)
	}
	return segment == path
}
