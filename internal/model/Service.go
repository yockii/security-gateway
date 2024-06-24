package model

import (
	"github.com/tidwall/gjson"
	"gorm.io/gorm"
)

type Service struct {
	ID            uint64         `json:"id,string" gorm:"primaryKey:autoIncrement:false"`
	Name          *string        `json:"name" gorm:"size:50;comment:名称"`
	Domain        *string        `json:"domain" gorm:"size:200;comment:监听的域名"`
	Port          *uint16        `json:"port" gorm:"comment:监听的端口"`
	CertificateID *uint64        `json:"certificateId,omitempty" gorm:"comment:证书ID"`
	CreateTime    int64          `json:"createTime" gorm:"autoCreateTime:milli"`
	DeleteTime    gorm.DeletedAt `json:"deleteTime,omitempty" gorm:"index"`
}

func (*Service) TableComment() string {
	return "服务表"
}

func (s *Service) UnmarshalJSON(b []byte) error {
	j := gjson.ParseBytes(b)

	s.ID = j.Get("id").Uint()
	if nj := j.Get("name"); nj.Exists() {
		name := nj.String()
		s.Name = &name
	}
	if nj := j.Get("domain"); nj.Exists() {
		domain := nj.String()
		s.Domain = &domain
	}
	if nj := j.Get("port"); nj.Exists() {
		port := uint16(nj.Uint())
		s.Port = &port
	}
	if nj := j.Get("certificateId"); nj.Exists() {
		certificateId := nj.Uint()
		s.CertificateID = &certificateId
	}
	s.CreateTime = j.Get("createTime").Int()

	return nil
}

func init() {
	Models = append(Models, &Service{})
}
