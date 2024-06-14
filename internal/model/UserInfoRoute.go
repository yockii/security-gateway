package model

import "github.com/tidwall/gjson"

type UserInfoRoute struct {
	ID            uint64 `json:"id,string" gorm:"primaryKey:autoIncrement:false"`
	ServiceID     uint64 `json:"serviceId,string" gorm:"comment:服务ID"`
	Path          string `json:"path" gorm:"size:50;comment:路径"`
	UsernamePath  string `json:"usernamePath" gorm:"size:50;comment:用户名路径"`
	UniKeyPath    string `json:"uniKeyPath" gorm:"size:50;comment:唯一标识路径"`
	MatchKey      string `json:"matchKey" gorm:"size:50;comment:匹配键,-表示直接匹配User.UniKey,否则匹配User.UniKeysJson中的值"`
	TokenPosition string `json:"tokenPosition" gorm:"size:100;comment:token位置"` // request:header:Authorization, request:query:token, request:body:auth.token, request:cookies:token, request:cookies:sessionId, response:body:data.token 之类的
	CreateTime    int64  `json:"createTime" gorm:"autoCreateTime:milli"`
}

func (*UserInfoRoute) TableComment() string {
	return "用户信息路由表"
}

func (u *UserInfoRoute) UnmarshalJSON(b []byte) error {
	u.ID = gjson.GetBytes(b, "id").Uint()
	u.ServiceID = gjson.GetBytes(b, "serviceId").Uint()
	u.Path = gjson.GetBytes(b, "path").String()
	u.UsernamePath = gjson.GetBytes(b, "usernamePath").String()
	u.UniKeyPath = gjson.GetBytes(b, "uniKeyPath").String()
	u.MatchKey = gjson.GetBytes(b, "matchKey").String()
	u.TokenPosition = gjson.GetBytes(b, "tokenPosition").String()
	u.CreateTime = gjson.GetBytes(b, "createTime").Int()

	return nil
}

func init() {
	Models = append(Models, &UserInfoRoute{})
}
