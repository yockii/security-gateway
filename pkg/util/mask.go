package util

import (
	"fmt"
	"strings"
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

	// 标识*数量是否表示要替换的字符数量
	isCharCount := false

	if maskReplacement[0] == '^' {
		maskReplacement = maskReplacement[1:]
		isCharCount = true
	}

	count := len(maskReplacement)
	if count == 0 {
		return "", fmt.Errorf("invalid mask replacement: %s", maskReplacement)
	}

	switch maskType {
	case "all":
		return maskReplacement, nil
	case "each":
		if isCharCount {
			return strings.Repeat(maskReplacement[0:1], len(origin)/count), nil
		} else {
			return strings.Repeat(maskReplacement, len(origin)), nil
		}
	case "start":
		if len(origin) < count {
			// 如果origin长度小于掩码长度，则全部掩码
			return strings.Repeat(maskReplacement, len(origin)), nil
		}
		return strings.Repeat(maskReplacement, count) + origin[count:], nil
	case "middle":
		if len(origin) < count {
			// 如果origin长度小于掩码长度，则全部掩码
			return strings.Repeat(maskReplacement, len(origin)), nil
		}
		start := (len(origin) - count) / 2
		return origin[:start] + strings.Repeat(maskReplacement, count) + origin[start+count:], nil
	case "end":
		if len(origin) < count {
			// 如果origin长度小于掩码长度，则全部掩码
			return strings.Repeat(maskReplacement, len(origin)), nil
		}
		return origin[:len(origin)-count] + strings.Repeat(maskReplacement, count), nil
	}

	return "", fmt.Errorf("invalid mask type: %s", maskType)
}
