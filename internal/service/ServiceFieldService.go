package service

import (
	logger "github.com/sirupsen/logrus"
	"security-gateway/internal/model"
	"security-gateway/pkg/database"
	"security-gateway/pkg/util"
	"strings"
)

var ServiceFieldService = &serviceFieldService{}

type serviceFieldService struct{}

func (u *serviceFieldService) Add(instance *model.ServiceField) (duplicated, success bool, err error) {
	if instance.ServiceID == 0 || instance.FieldName == "" {
		return
	}
	// 检查是否有名称或者url重复
	var c int64
	err = database.DB.Model(&model.ServiceField{}).Where(&model.ServiceField{ServiceID: instance.ServiceID, FieldName: instance.FieldName}).Count(&c).Error
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

func (u *serviceFieldService) Update(instance *model.ServiceField) (duplicated, success bool, err error) {
	if instance.ID == 0 {
		logger.Error("ID is required")
		return
	}

	// 检查是否有名称或者url重复
	var c int64
	err = database.DB.Model(&model.ServiceField{}).Where("id <> ?", instance.ID).Where(&model.ServiceField{ServiceID: instance.ServiceID, FieldName: instance.FieldName}).Count(&c).Error
	if err != nil {
		logger.Errorln(err)
		return
	}
	if c > 0 {
		duplicated = true
		return
	}

	if err = database.DB.Model(&model.ServiceField{ID: instance.ID}).Updates(instance).Error; err != nil {
		logger.Errorln(err)
		return
	}
	success = true
	return
}

func (u *serviceFieldService) Delete(id uint64) (success bool, err error) {
	if id == 0 {
		logger.Error("ID is required")
		return
	}
	if err = database.DB.Delete(&model.ServiceField{ID: id}).Error; err != nil {
		logger.Errorln(err)
		return
	}
	success = true
	return
}

func (u *serviceFieldService) Get(id uint64) (instance *model.ServiceField, err error) {
	if id == 0 {
		logger.Error("ID is required")
		return
	}
	instance = new(model.ServiceField)
	if err = database.DB.Where(&model.ServiceField{ID: id}).First(instance).Error; err != nil {
		logger.Errorln(err)
		return
	}
	return
}

func (u *serviceFieldService) List(page, pageSize int, condition *model.ServiceField) (instances []*model.ServiceField, total int64, err error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	sess := database.DB.Model(&model.ServiceField{})
	if condition.FieldName != "" {
		sess = sess.Where("field_name like ?", "%"+strings.TrimSpace(condition.FieldName)+"%")
		condition.FieldName = ""
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
	err = sess.Offset((page - 1) * pageSize).Limit(pageSize).Find(&instances).Error
	if err != nil {
		logger.Errorln(err)
		return
	}
	return
}

func (u *serviceFieldService) GetByServiceID(serviceID uint64) (instances []*model.ServiceField, err error) {
	if serviceID == 0 {
		logger.Error("serviceID is required")
		return
	}
	err = database.DB.Where(&model.ServiceField{ServiceID: serviceID}).Find(&instances).Error
	if err != nil {
		logger.Errorln(err)
		return
	}
	return
}
