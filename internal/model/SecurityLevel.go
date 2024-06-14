package model

import "github.com/tidwall/gjson"

type UserServiceLevel struct {
	ID         uint64 `json:"id,string" gorm:"primaryKey:autoIncrement:false"`
	UserID     uint64 `json:"userId,string" gorm:"comment:用户ID"`
	ServiceID  uint64 `json:"serviceId,string" gorm:"comment:服务ID"`
	SecLevel   int    `json:"secLevel" gorm:"comment:安全等级"`
	CreateTime int64  `json:"createTime" gorm:"autoCreateTime:milli"`
}

func (*UserServiceLevel) TableComment() string {
	return "用户服务密级表"
}

func (u *UserServiceLevel) UnmarshalJSON(b []byte) error {
	j := gjson.ParseBytes(b)
	u.ID = j.Get("id").Uint()
	u.UserID = j.Get("userId").Uint()
	u.ServiceID = j.Get("serviceId").Uint()
	u.SecLevel = int(j.Get("secLevel").Int())
	u.CreateTime = j.Get("createTime").Int()

	return nil
}

func init() {
	Models = append(Models, &UserServiceLevel{})
}
