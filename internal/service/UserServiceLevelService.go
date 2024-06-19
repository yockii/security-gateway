package service

import (
	logger "github.com/sirupsen/logrus"
	"security-gateway/internal/domain"
	"security-gateway/internal/model"
	"security-gateway/pkg/database"
	"security-gateway/pkg/util"
)

var UserServiceLevelService = &userServiceLevelService{}

type userServiceLevelService struct{}

func (u *userServiceLevelService) Add(instance *model.UserServiceLevel) (duplicated, success bool, err error) {
	if instance.ServiceID == 0 || instance.UserID == 0 {
		return
	}
	// 检查是否有名称或者url重复
	var c int64
	err = database.DB.Model(&model.UserServiceLevel{}).Where(&model.UserServiceLevel{UserID: instance.UserID, ServiceID: instance.ServiceID}).Count(&c).Error
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

func (u *userServiceLevelService) Update(instance *model.UserServiceLevel) (duplicated, success bool, err error) {
	if instance.ID == 0 {
		logger.Error("ID is required")
		return
	}
	// 检查是否有名称或者url重复
	var c int64
	err = database.DB.Model(&model.UserServiceLevel{}).Where(&model.UserServiceLevel{UserID: instance.UserID, ServiceID: instance.ServiceID}).Where("id <> ?", instance.ID).Count(&c).Error
	if err != nil {
		logger.Errorln(err)
		return
	}
	if c > 0 {
		duplicated = true
		return
	}

	if err = database.DB.Model(&model.UserServiceLevel{ID: instance.ID}).Updates(instance).Error; err != nil {
		logger.Errorln(err)
		return
	}
	success = true
	return
}

func (u *userServiceLevelService) Delete(id uint64) (success bool, err error) {
	if id == 0 {
		logger.Error("ID is required")
		return
	}
	if err = database.DB.Delete(&model.UserServiceLevel{ID: id}).Error; err != nil {
		logger.Errorln(err)
		return
	}
	success = true
	return
}

func (u *userServiceLevelService) Get(id uint64) (instance *model.UserServiceLevel, err error) {
	if id == 0 {
		logger.Error("ID is required")
		return
	}
	instance = new(model.UserServiceLevel)
	if err = database.DB.Where(&model.UserServiceLevel{ID: id}).First(instance).Error; err != nil {
		logger.Errorln(err)
		return
	}
	return
}

func (u *userServiceLevelService) List(page, pageSize int, condition *model.UserServiceLevel) (instances []*model.UserServiceLevel, total int64, err error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	sess := database.DB.Model(&model.UserServiceLevel{})
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

func (u *userServiceLevelService) ListWithService(page, pageSize int, condition *model.UserServiceLevel) (instances []*domain.UserLevelWithService, total int64, err error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	sess := database.DB.Model(&model.UserServiceLevel{})
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
	for _, v := range instances {
		v.Service, err = ServiceService.Get(v.ServiceID)
		if err != nil {
			logger.Errorln(err)
			continue
		}
	}
	return
}

func (u *userServiceLevelService) GetByUserAndServiceID(userID, serviceID uint64) *model.UserServiceLevel {
	instance := new(model.UserServiceLevel)
	err := database.DB.Where(&model.UserServiceLevel{UserID: userID, ServiceID: serviceID}).First(instance).Error
	if err != nil {
		logger.Errorln(err)
		return nil
	}
	return instance
}
