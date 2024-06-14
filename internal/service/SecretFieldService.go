package service

import (
	logger "github.com/sirupsen/logrus"
	"security-gateway/internal/model"
	"security-gateway/pkg/database"
	"security-gateway/pkg/util"
)

var SecretFieldService = &secretFieldService{}

type secretFieldService struct{}

func (u *secretFieldService) Add(instance *model.SecretField) (duplicated, success bool, err error) {
	if instance.ServiceID == 0 || instance.FieldName == "" {
		return
	}
	// 检查是否有名称或者url重复
	var c int64
	err = database.DB.Model(&model.SecretField{}).Where(&model.SecretField{ServiceID: instance.ServiceID, FieldName: instance.FieldName}).Count(&c).Error
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

func (u *secretFieldService) Update(instance *model.SecretField) (duplicated, success bool, err error) {
	if instance.ID == 0 {
		logger.Error("ID is required")
		return
	}

	// 检查是否有名称或者url重复
	var c int64
	err = database.DB.Model(&model.SecretField{}).Where("id <> ?", instance.ID).Where(&model.SecretField{ServiceID: instance.ServiceID, FieldName: instance.FieldName}).Count(&c).Error
	if err != nil {
		logger.Errorln(err)
		return
	}
	if c > 0 {
		duplicated = true
		return
	}

	if err = database.DB.Model(&model.SecretField{ID: instance.ID}).Updates(instance).Error; err != nil {
		logger.Errorln(err)
		return
	}
	success = true
	return
}

func (u *secretFieldService) Delete(id uint64) (success bool, err error) {
	if id == 0 {
		logger.Error("ID is required")
		return
	}
	if err = database.DB.Delete(&model.SecretField{ID: id}).Error; err != nil {
		logger.Errorln(err)
		return
	}
	success = true
	return
}

func (u *secretFieldService) Get(id uint64) (instance *model.SecretField, err error) {
	if id == 0 {
		logger.Error("ID is required")
		return
	}
	instance = new(model.SecretField)
	if err = database.DB.Where(&model.SecretField{ID: id}).First(instance).Error; err != nil {
		logger.Errorln(err)
		return
	}
	return
}

func (u *secretFieldService) List(page, pageSize int, fieldName string) (instances []*model.SecretField, total int64, err error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if fieldName == "" {
		err = database.DB.Model(&model.SecretField{}).Count(&total).Error
		if err != nil {
			logger.Errorln(err)
			return
		}
		if total == 0 {
			return
		}
		err = database.DB.Offset((page - 1) * pageSize).Limit(pageSize).Find(&instances).Error
		if err != nil {
			logger.Errorln(err)
			return
		}
		return
	}

	err = database.DB.Model(&model.SecretField{}).Where("field_name like ?", "%"+fieldName+"%").Count(&total).Error
	if err != nil {
		logger.Errorln(err)
		return
	}
	if total == 0 {
		return
	}
	err = database.DB.Where("field_name like ?", "%"+fieldName+"%").Offset((page - 1) * pageSize).Limit(pageSize).Find(&instances).Error
	if err != nil {
		logger.Errorln(err)
		return
	}
	return
}
