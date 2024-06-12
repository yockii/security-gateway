package model

import (
	"github.com/tidwall/gjson"
	"gorm.io/gorm"
)

type Route struct {
	ID         uint64         `json:"id,omitempty,string" gorm:"primaryKey;autoIncrement:false"`
	ServiceID  *uint64        `json:"serviceId,omitempty,string" gorm:"index;comment:服务ID"`
	Uri        *string        `json:"uri,omitempty" gorm:"size:200;comment:URI"`
	CreateTime int64          `json:"createTime" gorm:"autoCreateTime:milli"`
	DeleteTime gorm.DeletedAt `json:"deleteTime,omitempty" gorm:"index"`
}

func (*Route) TableComment() string {
	return "路径表"
}

func (p *Route) UnmarshalJSON(b []byte) error {
	j := gjson.ParseBytes(b)
	p.ID = j.Get("id").Uint()
	if serviceID := j.Get("serviceId"); serviceID.Exists() {
		serviceIDUint := serviceID.Uint()
		p.ServiceID = &serviceIDUint
	}
	if uri := j.Get("uri"); uri.Exists() {
		uriStr := uri.String()
		p.Uri = &uriStr
	}
	p.CreateTime = j.Get("createTime").Int()

	return nil
}

type RouteTarget struct {
	ID         uint64  `json:"id,omitempty,string" gorm:"primaryKey;autoIncrement:false"`
	LocationID *uint64 `json:"locationId,omitempty,string" gorm:"index;comment:位置ID"`
	UpstreamID *uint64 `json:"upstreamId,omitempty,string" gorm:"index;comment:上游ID"`
	Weight     *int    `json:"weight" gorm:"comment:权重"`
	CreateTime int64   `json:"createTime" gorm:"autoCreateTime:milli"`
}

func (*RouteTarget) TableComment() string {
	return "路径目标表"
}

func (pt *RouteTarget) UnmarshalJSON(b []byte) error {
	j := gjson.ParseBytes(b)
	pt.ID = j.Get("id").Uint()
	if locationID := j.Get("locationId"); locationID.Exists() {
		locationIDUint := locationID.Uint()
		pt.LocationID = &locationIDUint
	}
	if upstreamID := j.Get("upstreamId"); upstreamID.Exists() {
		upstreamIDUint := upstreamID.Uint()
		pt.UpstreamID = &upstreamIDUint
	}
	if weight := j.Get("weight"); weight.Exists() {
		weightInt := int(weight.Int())
		pt.Weight = &weightInt
	}
	pt.CreateTime = j.Get("createTime").Int()

	return nil

}

func init() {
	Models = append(Models, &Route{}, &RouteTarget{})
}
