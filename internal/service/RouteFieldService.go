package service

import (
	logger "github.com/sirupsen/logrus"
	"security-gateway/internal/model"
	"security-gateway/pkg/database"
	"security-gateway/pkg/util"
)

var RouteFieldService = &routeFieldService{}

type routeFieldService struct{}

func (u *routeFieldService) Add(instance *model.RouteField) (duplicated, success bool, err error) {
	if instance.RouteID == 0 || instance.FieldName == "" {
		return
	}
	// 检查是否有名称或者url重复
	var c int64
	err = database.DB.Model(&model.RouteField{}).Where(&model.RouteField{RouteID: instance.RouteID, FieldName: instance.FieldName}).Count(&c).Error
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

func (u *routeFieldService) Update(instance *model.RouteField) (duplicated, success bool, err error) {
	if instance.ID == 0 {
		logger.Error("ID is required")
		return
	}

	// 检查是否有名称或者url重复
	var c int64
	err = database.DB.Model(&model.RouteField{}).Where("id <> ?", instance.ID).Where(&model.RouteField{RouteID: instance.RouteID, FieldName: instance.FieldName}).Count(&c).Error
	if err != nil {
		logger.Errorln(err)
		return
	}
	if c > 0 {
		duplicated = true
		return
	}

	if err = database.DB.Model(&model.RouteField{ID: instance.ID}).Updates(instance).Error; err != nil {
		logger.Errorln(err)
		return
	}
	success = true
	return
}

func (u *routeFieldService) Delete(id uint64) (success bool, err error) {
	if id == 0 {
		logger.Error("ID is required")
		return
	}
	if err = database.DB.Delete(&model.RouteField{ID: id}).Error; err != nil {
		logger.Errorln(err)
		return
	}
	success = true
	return
}

func (u *routeFieldService) Get(id uint64) (instance *model.RouteField, err error) {
	if id == 0 {
		logger.Error("ID is required")
		return
	}
	instance = new(model.RouteField)
	if err = database.DB.Where(&model.RouteField{ID: id}).First(instance).Error; err != nil {
		logger.Errorln(err)
		return
	}
	return
}

func (u *routeFieldService) List(page, pageSize int, condition *model.RouteField) (instances []*model.RouteField, total int64, err error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	sess := database.DB.Model(&model.RouteField{})
	if condition.FieldName != "" {
		sess = sess.Where("field_name like ?", "%"+condition.FieldName+"%")
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

func (u *routeFieldService) GetByRouteID(routeID uint64) (instances []*model.RouteField, err error) {
	if routeID == 0 {
		logger.Error("routeID is required")
		return
	}
	err = database.DB.Where(&model.RouteField{RouteID: routeID}).Find(&instances).Error
	if err != nil {
		logger.Errorln(err)
		return
	}
	return
}
