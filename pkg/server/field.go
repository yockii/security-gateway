package server

import "security-gateway/pkg/util"

// DesensitizeField 脱敏字段
type DesensitizeField struct {
	// 字段名
	Name string `json:"name"`
	// 是否是服务字段
	IsServiceField bool `json:"isServiceField"`
	// 一级脱敏规则
	Level1DesensitizeRule string `json:"level1DesensitizeRule"`
	// 二级脱敏规则
	Level2DesensitizeRule string `json:"level2DesensitizeRule"`
	// 三级脱敏规则
	Level3DesensitizeRule string `json:"level3DesensitizeRule"`
	// 四级脱敏规则
	Level4DesensitizeRule string `json:"level4DesensitizeRule"`
}

func (f *DesensitizeField) Mask(value string, level int) string {
	rule := ""
	switch level {
	case 1:
		rule = f.Level1DesensitizeRule
	case 2:
		rule = f.Level2DesensitizeRule
	case 3:
		rule = f.Level3DesensitizeRule
	case 4:
		rule = f.Level4DesensitizeRule
	default:
		rule = f.Level1DesensitizeRule
	}
	return f.mask(value, rule)
}

func (f *DesensitizeField) mask(value string, rule string) string {
	if rule == "-" || rule == "" {
		return value
	}
	mv, err := util.AdvanceMask(value, rule)
	if err != nil {
		return value
	}
	return mv
}
