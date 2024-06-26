package model

import (
	"github.com/tidwall/gjson"
	"gorm.io/gorm"
)

type Upstream struct {
	ID        uint64  `json:"id,omitempty,string" gorm:"primaryKey;autoIncrement:false"`
	Name      *string `json:"name" gorm:"size:50;comment:名称"`
	TargetUrl *string `json:"targetUrl" gorm:"size:200;comment:目标URL"`
	// 健康监测地址
	HealthCheckUrl *string        `json:"healthCheckUrl" gorm:"size:200;comment:健康监测地址"`
	Status         int            `json:"status" gorm:"comment:状态,1-正常 2-健康检测失败"`
	LastCheckTime  int64          `json:"lastCheckTime" gorm:"default:0;comment:最后一次健康检测时间"`
	CreateTime     int64          `json:"createTime" gorm:"autoCreateTime:milli"`
	DeleteTime     gorm.DeletedAt `json:"deleteTime,omitempty" gorm:"index"`
}

func (*Upstream) TableComment() string {
	return "上游表"
}

func (u *Upstream) UnmarshalJSON(b []byte) error {
	j := gjson.ParseBytes(b)
	u.ID = j.Get("id").Uint()
	if nj := j.Get("name"); nj.Exists() {
		name := nj.String()
		u.Name = &name
	}
	if nj := j.Get("targetUrl"); nj.Exists() {
		targetUrl := nj.String()
		u.TargetUrl = &targetUrl
	}
	if nj := j.Get("healthCheckUrl"); nj.Exists() {
		healthCheckUrl := nj.String()
		u.HealthCheckUrl = &healthCheckUrl
	}
	u.Status = int(j.Get("status").Int())
	u.LastCheckTime = j.Get("lastCheckTime").Int()

	u.CreateTime = j.Get("createTime").Int()

	return nil
}

func init() {
	Models = append(Models, &Upstream{})
}
