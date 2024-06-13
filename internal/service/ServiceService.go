package service

import (
	logger "github.com/sirupsen/logrus"
	"security-gateway/internal/model"
	"security-gateway/pkg/database"
	"security-gateway/pkg/util"
)

var ServiceService = &serviceService{}

type serviceService struct{}

func (u *serviceService) Add(instance *model.Service) (duplicated, success bool, err error) {
	if instance.Name == nil || instance.Domain == nil || instance.Port == nil || *(instance.Name) == "" || *(instance.Port) <= 0 || *(instance.Port) > 65535 {
		return
	}
	// 检查是否有名称或者url重复
	var c int64
	err = database.DB.Model(&model.Service{}).Where("name = ? or (domain = ? and port = ?)", instance.Name, instance.Domain, instance.Port).Count(&c).Error
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

func (u *serviceService) Update(instance *model.Service) (duplicated, success bool, err error) {
	if instance.ID == 0 {
		logger.Error("ID is required")
		return
	}
	if instance.Name != nil && *(instance.Name) == "" {
		instance.Name = nil
	}
	if instance.Domain != nil && *(instance.Domain) == "" {
		instance.Domain = nil
	}
	if instance.Port != nil && (*(instance.Port) <= 0 || *(instance.Port) > 65535) {
		instance.Port = nil
	}

	// 检查是否有名称或者url重复
	var c int64
	err = database.DB.Model(&model.Service{}).Where("name = ? or (domain = ? and port = ?)", instance.Name, instance.Domain, instance.Port).Count(&c).Error
	if err != nil {
		logger.Errorln(err)
		return

	}
	if err = database.DB.Model(&model.Service{ID: instance.ID}).Updates(instance).Error; err != nil {
		logger.Errorln(err)
		return
	}
	success = true
	return
}

func (u *serviceService) Delete(id uint64) (success bool, err error) {
	if id == 0 {
		logger.Error("ID is required")
		return
	}
	if err = database.DB.Delete(&model.Service{ID: id}).Error; err != nil {
		logger.Errorln(err)
		return
	}
	success = true
	return
}

func (u *serviceService) Get(id uint64) (instance *model.Service, err error) {
	if id == 0 {
		logger.Error("ID is required")
		return
	}
	instance = new(model.Service)
	if err = database.DB.Where(&model.Service{ID: id}).First(instance).Error; err != nil {
		logger.Errorln(err)
		return
	}
	return
}

func (u *serviceService) List(page, pageSize int, name string) (instances []*model.Service, total int64, err error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if name == "" {
		err = database.DB.Model(&model.Service{}).Count(&total).Error
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

	err = database.DB.Model(&model.Service{}).Where("name like ?", "%"+name+"%").Count(&total).Error
	if err != nil {
		logger.Errorln(err)
		return
	}
	if total == 0 {
		return
	}
	err = database.DB.Where("name like ?", "%"+name+"%").Offset((page - 1) * pageSize).Limit(pageSize).Find(&instances).Error
	if err != nil {
		logger.Errorln(err)
		return
	}
	return
}
