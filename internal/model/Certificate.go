package model

import "github.com/tidwall/gjson"

type Certificate struct {
	ID          uint64 `json:"id,string" gorm:"primaryKey:autoIncrement:false"`
	CertName    string `json:"certName" gorm:"size:50;comment:证书名称"`
	ServeDomain string `json:"serveDomain" gorm:"size:50;comment:服务域名"`
	CertDesc    string `json:"certDesc" gorm:"size:200;comment:证书描述"`
	CertPem     string `json:"certPem,omitempty" gorm:"type:text;comment:证书内容"`
	KeyPem      string `json:"keyPem,omitempty" gorm:"type:text;comment:私钥内容"`
	SignCertPem string `json:"signCertPem,omitempty" gorm:"type:text;comment:签名证书内容(国密signCert)"`
	SignKeyPem  string `json:"signKeyPem,omitempty" gorm:"type:text;comment:签名私钥内容(国密signKey)"`
	EncCertPem  string `json:"encCertPem,omitempty" gorm:"type:text;comment:加密证书内容(国密encCert)"`
	EncKeyPem   string `json:"encKeyPem,omitempty" gorm:"type:text;comment:加密私钥内容(国密encKey)"`
	CreateTime  int64  `json:"createTime" gorm:"autoCreateTime:milli"`
}

func (*Certificate) TableComment() string {
	return "证书表"
}

func (s *Certificate) UnmarshalJSON(b []byte) error {
	j := gjson.ParseBytes(b)

	s.ID = j.Get("id").Uint()
	s.CertName = j.Get("certName").String()
	s.ServeDomain = j.Get("serveDomain").String()
	s.CertDesc = j.Get("certDesc").String()
	s.CertPem = j.Get("certPem").String()
	s.KeyPem = j.Get("keyPem").String()
	s.SignCertPem = j.Get("signCertPem").String()
	s.SignKeyPem = j.Get("signKeyPem").String()
	s.EncCertPem = j.Get("encCertPem").String()
	s.EncKeyPem = j.Get("encKeyPem").String()
	s.CreateTime = j.Get("createTime").Int()

	return nil
}

type ServiceCertificate struct {
	ID         uint64 `json:"id,string" gorm:"primaryKey:autoIncrement:false"`
	ServiceID  uint64 `json:"serviceId,string" gorm:"comment:服务ID"`
	CertID     uint64 `json:"certId,string" gorm:"comment:证书ID"`
	CreateTime int64  `json:"createTime" gorm:"autoCreateTime:milli"`
}

func (*ServiceCertificate) TableComment() string {
	return "服务证书表"
}

func (s *ServiceCertificate) UnmarshalJSON(b []byte) error {
	j := gjson.ParseBytes(b)

	s.ID = j.Get("id").Uint()
	s.ServiceID = j.Get("serviceId").Uint()
	s.CertID = j.Get("certId").Uint()
	s.CreateTime = j.Get("createTime").Int()

	return nil
}

func init() {
	Models = append(Models, &Certificate{}, &ServiceCertificate{})
}
