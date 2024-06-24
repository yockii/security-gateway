package service

import (
	logger "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"security-gateway/internal/model"
	"security-gateway/pkg/database"
	"security-gateway/pkg/util"
	"strings"
)

var CertificateService = &certificateService{}

type certificateService struct{}

func (u *certificateService) Add(instance *model.Certificate) (duplicated, success bool, err error) {
	if instance.CertName == "" || instance.ServeDomain == "" || instance.CertPem == "" || instance.KeyPem == "" {
		return
	}
	// 检查是否有重复
	var c int64
	err = database.DB.Model(&model.Certificate{}).Where(&model.Certificate{
		CertName:    instance.CertName,
		ServeDomain: instance.ServeDomain,
	}).Count(&c).Error
	if err != nil {
		logger.Errorln(err)
		return
	}
	if c > 0 {
		duplicated = true
		return
	}

	instance.ID = util.SnowflakeId()
	if err = database.DB.Create(instance).Error; err != nil {
		logger.Errorln(err)
		return
	}
	success = true
	return
}

func (u *certificateService) Update(instance *model.Certificate) (duplicated, success bool, err error) {
	if instance.ID == 0 {
		logger.Error("ID is required")
		return
	}

	// 检查是否有名称或者url重复
	var c int64
	err = database.DB.Model(&model.Certificate{}).Where("id <> ?  and (cert_name = ? or serve_domain = ?)", instance.ID, instance.CertName, instance.ServeDomain).Count(&c).Error
	if err != nil {
		logger.Errorln(err)
		return
	}

	if err = database.DB.Model(&model.Certificate{ID: instance.ID}).Updates(instance).Error; err != nil {
		logger.Errorln(err)
		return
	}
	success = true
	return
}

func (u *certificateService) Delete(id uint64) (success bool, err error) {
	if id == 0 {
		logger.Error("ID is required")
		return
	}
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		if err = tx.Where(&model.ServiceCertificate{CertID: id}).Delete(&model.ServiceCertificate{}).Error; err != nil {
			logger.Errorln(err)
			return err
		}
		if err = tx.Where(&model.Certificate{ID: id}).Delete(&model.Certificate{}).Error; err != nil {
			logger.Errorln(err)
			return err
		}

		return nil
	})
	success = err == nil
	return
}

func (u *certificateService) Get(id uint64) (instance *model.Certificate, err error) {
	if id == 0 {
		logger.Error("ID is required")
		return
	}
	instance = new(model.Certificate)
	if err = database.DB.Where(&model.Certificate{ID: id}).First(instance).Error; err != nil {
		logger.Errorln(err)
		return
	}
	return
}

func (u *certificateService) List(page, pageSize int, condition *model.Certificate) (instances []*model.Certificate, total int64, err error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	sess := database.DB.Model(&model.Certificate{})
	if condition.CertName != "" {
		sess = sess.Where("cert_name like ?", "%"+condition.CertName+"%")
		condition.CertName = ""
	}
	if condition.ServeDomain != "" {
		sess = sess.Where("serve_domain like ?", "%"+condition.ServeDomain+"%")
		condition.ServeDomain = ""
	}
	sess = sess.Where(condition)

	err = sess.Count(&total).Error
	if err != nil {
		logger.Errorln(err)
		return
	}
	if total == 0 {
		return
	}
	err = sess.Omit("cert_pem,key_pem").
		Offset((page - 1) * pageSize).Limit(pageSize).Find(&instances).Error
	if err != nil {
		logger.Errorln(err)
		return
	}
	return
}

func (u *certificateService) ListByDomain(domain string) (list []*model.Certificate, err error) {
	if domain == "" {
		logger.Error("domain is required")
		return
	}

	// 证书的域名可以是完全一致的域名；如果domain为非顶级域名，则可以是*.(domain的上级域名)的证书
	// 例如：domain为www.example.com，则可以是*.example.com的证书

	// 1. 查找完全一致的域名
	// 2. 查找*的证书, 必须是domain除去第一个.之前的部分加上*
	fanDomain := "*." + domain[strings.Index(domain, ".")+1:]

	sess := database.DB.Model(&model.Certificate{}).Omit("cert_pem,key_pem").
		Where("serve_domain in (?)", []string{domain, fanDomain})

	if err = sess.Find(&list).Error; err != nil {
		logger.Errorln(err)
		return
	}

	return
}

func (u *certificateService) AddServiceCertificate(instance *model.ServiceCertificate) (duplicate, success bool, err error) {
	if instance.ServiceID == 0 || instance.CertID == 0 {
		logger.Error("ServiceID and CertID is required")
		return
	}

	// 检查是否有重复
	var c int64
	err = database.DB.Model(&model.ServiceCertificate{}).Where(&model.ServiceCertificate{
		ServiceID: instance.ServiceID,
	}).Count(&c).Error
	if err != nil {
		logger.Errorln(err)
		return
	}
	if c > 0 {
		duplicate = true
		return
	}

	instance.ID = util.SnowflakeId()
	if err = database.DB.Create(instance).Error; err != nil {
		logger.Errorln(err)
		return
	}
	success = true
	return
}

func (u *certificateService) DeleteServiceCertificate(id uint64) (success bool, err error) {
	if id == 0 {
		logger.Error("ID is required")
		return
	}
	err = database.DB.Where(&model.ServiceCertificate{ID: id}).Delete(&model.ServiceCertificate{}).Error
	success = err == nil
	return
}
