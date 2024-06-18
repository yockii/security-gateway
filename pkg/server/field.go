package server

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
