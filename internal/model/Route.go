package model

import (
	"github.com/tidwall/gjson"
	"gorm.io/gorm"
)

// 均衡算法
const (
	//LoadBalanceRoundRobin 轮询
	LoadBalanceRoundRobin = 1
	//LoadBalanceWeight 权重
	LoadBalanceWeight = 2
	//LoadBalanceIPHash IP哈希
	LoadBalanceIPHash = 3
)

type Route struct {
	ID        uint64  `json:"id,omitempty,string" gorm:"primaryKey;autoIncrement:false"`
	ServiceID *uint64 `json:"serviceId,omitempty,string" gorm:"index;comment:服务ID"`
	Uri       *string `json:"uri,omitempty" gorm:"size:200;comment:URI"`
	// 负载均衡算法类型
	LoadBalance int `json:"loadBalance" gorm:"default:1;comment:负载均衡算法类型,1=轮询,2=权重,3=IP哈希"`

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
	p.LoadBalance = int(j.Get("loadBalance").Int())
	p.CreateTime = j.Get("createTime").Int()

	return nil
}

type RouteTarget struct {
	ID         uint64  `json:"id,omitempty,string" gorm:"primaryKey;autoIncrement:false"`
	RouteID    *uint64 `json:"routeId,omitempty,string" gorm:"index;comment:位置ID"`
	UpstreamID *uint64 `json:"upstreamId,omitempty,string" gorm:"index;comment:上游ID"`
	Weight     int     `json:"weight" gorm:"comment:权重"`
	CreateTime int64   `json:"createTime" gorm:"autoCreateTime:milli"`
}

func (*RouteTarget) TableComment() string {
	return "路径目标表"
}

func (pt *RouteTarget) UnmarshalJSON(b []byte) error {
	j := gjson.ParseBytes(b)
	pt.ID = j.Get("id").Uint()
	if locationID := j.Get("routeId"); locationID.Exists() {
		locationIDUint := locationID.Uint()
		pt.RouteID = &locationIDUint
	}
	if upstreamID := j.Get("upstreamId"); upstreamID.Exists() {
		upstreamIDUint := upstreamID.Uint()
		pt.UpstreamID = &upstreamIDUint
	}
	pt.Weight = int(j.Get("weight").Int())
	pt.CreateTime = j.Get("createTime").Int()

	return nil

}

func init() {
	Models = append(Models, &Route{}, &RouteTarget{})
}
