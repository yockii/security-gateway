package service

import (
	logger "github.com/sirupsen/logrus"
	"security-gateway/internal/model"
	"security-gateway/pkg/database"
	"security-gateway/pkg/util"
)

var RouteService = &routeService{}

type routeService struct{}

func (u *routeService) Add(instance *model.Route) (duplicated, success bool, err error) {
	if instance.ServiceID == nil || instance.Uri == nil || *(instance.ServiceID) == 0 || *(instance.Uri) == "" {
		return
	}
	// 检查是否有名称或者url重复
	var c int64
	err = database.DB.Model(&model.Route{}).Where("service_id = ? or uri = ?", instance.ServiceID, instance.Uri).Count(&c).Error
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

func (u *routeService) Update(instance *model.Route) (duplicated, success bool, err error) {
	if instance.ID == 0 {
		logger.Error("ID is required")
		return
	}
	if instance.Uri != nil && *(instance.Uri) == "" {
		instance.Uri = nil
	}
	if instance.ServiceID != nil && *(instance.ServiceID) == 0 {
		instance.ServiceID = nil
	}

	// 检查是否有名称或者url重复
	var c int64
	err = database.DB.Model(&model.Route{}).Where("id <> ?  and (service_id = ? or uri = ?)", instance.ID, instance.ServiceID, instance.Uri).Count(&c).Error
	if err != nil {
		logger.Errorln(err)
		return
	}

	if err = database.DB.Model(&model.Route{ID: instance.ID}).Updates(instance).Error; err != nil {
		logger.Errorln(err)
		return
	}
	success = true
	return
}

func (u *routeService) Delete(id uint64) (success bool, err error) {
	if id == 0 {
		logger.Error("ID is required")
		return
	}
	if err = database.DB.Delete(&model.Route{ID: id}).Error; err != nil {
		logger.Errorln(err)
		return
	}
	success = true
	return
}

func (u *routeService) Get(id uint64) (instance *model.Route, err error) {
	if id == 0 {
		logger.Error("ID is required")
		return
	}
	instance = new(model.Route)
	if err = database.DB.Where(&model.Route{ID: id}).First(instance).Error; err != nil {
		logger.Errorln(err)
		return
	}
	return
}

func (u *routeService) List(page, pageSize int, uri string) (instances []*model.Route, total int64, err error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if uri == "" {
		err = database.DB.Model(&model.Route{}).Count(&total).Error
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

	err = database.DB.Model(&model.Route{}).Where("uri like ?", "%"+uri+"%").Count(&total).Error
	if err != nil {
		logger.Errorln(err)
		return
	}
	if total == 0 {
		return
	}
	err = database.DB.Where("uri like ?", "%"+uri+"%").Offset((page - 1) * pageSize).Limit(pageSize).Find(&instances).Error
	if err != nil {
		logger.Errorln(err)
		return
	}
	return
}
