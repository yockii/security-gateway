package model

import "github.com/tidwall/gjson"

type User struct {
	ID          uint64 `json:"id,string" gorm:"primaryKey:autoIncrement:false"`
	Username    string `json:"username" gorm:"size:50;comment:用户名"`
	UniKey      string `json:"uniKey" gorm:"size:50;comment:唯一标识"`
	UniKeysJson string `json:"uniKeysJson" gorm:"size:500;comment:唯一标识json,可用于存储多个唯一标识"`
	SecLevel    int    `json:"secLevel" gorm:"comment:安全等级"`
	CreateTime  int64  `json:"createTime" gorm:"autoCreateTime:milli"`
}

func (*User) TableComment() string {
	return "用户表"
}

func (u *User) UnmarshalJSON(b []byte) error {
	j := gjson.ParseBytes(b)
	u.ID = j.Get("id").Uint()
	u.Username = j.Get("username").String()
	u.UniKey = j.Get("uniKey").String()
	u.SecLevel = int(j.Get("secLevel").Int())
	u.CreateTime = j.Get("createTime").Int()

	return nil
}

func init() {
	Models = append(Models, &User{})
}
