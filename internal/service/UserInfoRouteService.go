package service

import (
	logger "github.com/sirupsen/logrus"
	"security-gateway/internal/model"
	"security-gateway/pkg/database"
	"security-gateway/pkg/util"
	"strings"
)

var UserInfoRouteService = &userInfoRouteService{}

type userInfoRouteService struct{}

func (u *userInfoRouteService) Add(instance *model.UserInfoRoute) (duplicated, success bool, err error) {
	if instance.ServiceID == 0 || instance.Path == "" || instance.UsernamePath == "" || instance.UniKeyPath == "" || instance.MatchKey == "" {
		return
	}
	// 检查是否有名称或者url重复
	var c int64
	err = database.DB.Model(&model.UserInfoRoute{}).Where(&model.UserInfoRoute{ServiceID: instance.ServiceID}).Count(&c).Error
	if err != nil {
		logger.Errorln(err)
		return
	}
	if c > 0 {
		duplicated = true
		return
	}
	if instance.Method == "" {
		instance.Method = "GET"
	}

	instance.ID = util.SnowflakeId()
	instance.Method = strings.ToUpper(instance.Method)
	if err = database.DB.Create(instance).Error; err != nil {
		logger.Errorln(err)
		return
	}
	success = true
	return
}

func (u *userInfoRouteService) Update(instance *model.UserInfoRoute) (duplicated, success bool, err error) {
	if instance.ID == 0 {
		logger.Error("ID is required")
		return
	}
	// 检查是否有名称或者url重复
	var c int64
	err = database.DB.Model(&model.UserInfoRoute{}).Where(&model.UserInfoRoute{ServiceID: instance.ServiceID}).Where("id <> ?", instance.ID).Count(&c).Error
	if err != nil {
		logger.Errorln(err)
		return
	}
	if c > 0 {
		duplicated = true
		return
	}

	instance.Method = strings.ToUpper(instance.Method)

	if err = database.DB.Model(&model.UserInfoRoute{ID: instance.ID}).Updates(instance).Error; err != nil {
		logger.Errorln(err)
		return
	}
	success = true
	return
}

func (u *userInfoRouteService) Delete(id uint64) (success bool, err error) {
	if id == 0 {
		logger.Error("ID is required")
		return
	}
	if err = database.DB.Delete(&model.UserInfoRoute{ID: id}).Error; err != nil {
		logger.Errorln(err)
		return
	}
	success = true
	return
}

func (u *userInfoRouteService) Get(id uint64) (instance *model.UserInfoRoute, err error) {
	if id == 0 {
		logger.Error("ID is required")
		return
	}
	instance = new(model.UserInfoRoute)
	if err = database.DB.Where(&model.UserInfoRoute{ID: id}).First(instance).Error; err != nil {
		logger.Errorln(err)
		return
	}
	return
}

func (u *userInfoRouteService) List(page, pageSize int, name string) (instances []*model.UserInfoRoute, total int64, err error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if name == "" {
		err = database.DB.Model(&model.UserInfoRoute{}).Count(&total).Error
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

	err = database.DB.Model(&model.UserInfoRoute{}).Where("service_id = ? or path like ?", name, "%"+name+"%").Count(&total).Error
	if err != nil {
		logger.Errorln(err)
		return
	}
	if total == 0 {
		return
	}
	err = database.DB.Where("service_id = ? or path like ?", name, "%"+name+"%").Offset((page - 1) * pageSize).Limit(pageSize).Find(&instances).Error
	if err != nil {
		logger.Errorln(err)
		return
	}
	return
}

func (u *userInfoRouteService) GetByServiceID(serviceID uint64) (*model.UserInfoRoute, error) {
	var result []*model.UserInfoRoute
	err := database.DB.Where(&model.UserInfoRoute{ServiceID: serviceID}).Find(&result).Error
	if err != nil {
		logger.Errorln(err)
		return nil, err
	}
	if len(result) == 0 {
		return nil, nil
	}
	return result[0], nil
}
