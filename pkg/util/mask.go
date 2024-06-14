package util

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// AdvanceMask 对origin进行maskPattern的掩码处理
// maskPattern的格式为：{type}-[^]{replacementWithN}, 如: all-* 表示替换所有字符为一个*，each-^** 表示每2个字符替换为一个*，middle-****表示中间4个字符替换为4个*

func AdvanceMask(origin string, maskPattern string) (string, error) {
	parts := strings.SplitN(maskPattern, "-", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid mask pattern: %s", maskPattern)
	}
	maskType := parts[0]
	maskReplacement := parts[1]

	if len(maskReplacement) == 0 {
		return "", fmt.Errorf("invalid mask replacement: %s", maskReplacement)
	}

	isCharCount := false

	if maskReplacement[0] == '^' {
		maskReplacement = maskReplacement[1:]
		isCharCount = true
	}

	count := len(maskReplacement)
	if count == 0 {
		return "", fmt.Errorf("invalid mask replacement: %s", maskReplacement)
	}

	originLength := utf8.RuneCountInString(origin)

	switch maskType {
	case "all":
		return maskReplacement, nil
	case "each":
		if isCharCount {
			return strings.Repeat(maskReplacement[0:1], originLength/count), nil
		} else {
			return strings.Repeat(maskReplacement, originLength), nil
		}
	case "start":
		return maskStart(origin, maskReplacement, count, originLength)
	case "middle":
		return maskMiddle(origin, maskReplacement, count, originLength)
	case "end":
		return maskEnd(origin, maskReplacement, count, originLength)
	}

	return "", fmt.Errorf("invalid mask type: %s", maskType)
}

func maskStart(origin string, maskReplacement string, count int, originLength int) (string, error) {
	if originLength < count {
		return strings.Repeat(maskReplacement, originLength), nil
	}
	return strings.Repeat(maskReplacement, count) + string([]rune(origin)[count:]), nil
}

func maskMiddle(origin string, maskReplacement string, count int, originLength int) (string, error) {
	if originLength < count {
		return strings.Repeat(maskReplacement, originLength), nil
	}
	start := (originLength - count) / 2
	return string([]rune(origin)[:start]) + strings.Repeat(maskReplacement, count) + string([]rune(origin)[start+count:]), nil
}

func maskEnd(origin string, maskReplacement string, count int, originLength int) (string, error) {
	if originLength < count {
		return strings.Repeat(maskReplacement, originLength), nil
	}
	return string([]rune(origin)[:originLength-count]) + strings.Repeat(maskReplacement, count), nil
}
