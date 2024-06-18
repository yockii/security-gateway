package model

import "github.com/tidwall/gjson"

type ServiceField struct {
	ID         uint64 `json:"id,string" gorm:"primaryKey:autoIncrement:false"`
	ServiceID  uint64 `json:"serviceId,string" gorm:"comment:服务ID"`
	FieldName  string `json:"fieldName" gorm:"size:50;comment:字段名"`
	Comment    string `json:"comment" gorm:"size:200;comment:注释"`
	Level1     string `json:"level1" gorm:"comment:密级1规则, all-****, start-**, middle-^**, end-**, each-**"`
	Level2     string `json:"level2" gorm:"comment:密级2规则, all-****, start-**, middle-^**, end-**, each-**"`
	Level3     string `json:"level3" gorm:"comment:密级3规则, all-****, start-**, middle-^**, end-**, each-**"`
	Level4     string `json:"level4" gorm:"comment:密级4规则, all-****, start-**, middle-^**, end-**, each-**"`
	CreateTime int64  `json:"createTime" gorm:"autoCreateTime:milli"`
}

func (*ServiceField) TableComment() string {
	return "服务脱敏字段表"
}

func (s *ServiceField) UnmarshalJSON(b []byte) error {
	j := gjson.ParseBytes(b)

	s.ID = j.Get("id").Uint()
	s.ServiceID = j.Get("serviceId").Uint()
	s.FieldName = j.Get("fieldName").String()
	s.Comment = j.Get("comment").String()
	s.Level1 = j.Get("level1").String()
	s.Level2 = j.Get("level2").String()
	s.Level3 = j.Get("level3").String()
	s.Level4 = j.Get("level4").String()
	s.CreateTime = j.Get("createTime").Int()

	return nil
}

type RouteField struct {
	ID         uint64 `json:"id,string" gorm:"primaryKey:autoIncrement:false"`
	RouteID    uint64 `json:"routeId,string" gorm:"comment:路由ID"`
	FieldName  string `json:"fieldName" gorm:"size:50;comment:字段名"`
	Comment    string `json:"comment" gorm:"size:200;comment:注释"`
	Level1     string `json:"level1" gorm:"comment:密级1规则, all-****, start-**, middle-^**, end-**, each-**"`
	Level2     string `json:"level2" gorm:"comment:密级2规则, all-****, start-**, middle-^**, end-**, each-**"`
	Level3     string `json:"level3" gorm:"comment:密级3规则, all-****, start-**, middle-^**, end-**, each-**"`
	Level4     string `json:"level4" gorm:"comment:密级4规则, all-****, start-**, middle-^**, end-**, each-**"`
	CreateTime int64  `json:"createTime" gorm:"autoCreateTime:milli"`
}

func (*RouteField) TableComment() string {
	return "路由脱敏字段表"
}

func (s *RouteField) UnmarshalJSON(b []byte) error {
	j := gjson.ParseBytes(b)

	s.ID = j.Get("id").Uint()
	s.RouteID = j.Get("routeId").Uint()
	s.FieldName = j.Get("fieldName").String()
	s.Comment = j.Get("comment").String()
	s.Level1 = j.Get("level1").String()
	s.Level2 = j.Get("level2").String()
	s.Level3 = j.Get("level3").String()
	s.Level4 = j.Get("level4").String()
	s.CreateTime = j.Get("createTime").Int()

	return nil

}

func init() {
	Models = append(Models, &ServiceField{}, &RouteField{})
}
