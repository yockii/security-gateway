package service

import (
	logger "github.com/sirupsen/logrus"
	"security-gateway/internal/model"
	"security-gateway/pkg/database"
	"security-gateway/pkg/util"
)

var RouteTargetService = &routeTargetService{}

type routeTargetService struct{}

func (u *routeTargetService) Add(instance *model.RouteTarget) (duplicated, success bool, err error) {
	if instance.RouteID == nil || instance.UpstreamID == nil {
		return
	}
	// 检查是否有名称或者url重复
	var c int64
	err = database.DB.Model(&model.RouteTarget{}).Where("route_id = ? or upstream_id = ?", instance.RouteID, instance.UpstreamID).Count(&c).Error
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

func (u *routeTargetService) Update(instance *model.RouteTarget) (duplicated, success bool, err error) {
	if instance.ID == 0 {
		logger.Error("ID is required")
		return
	}

	// 检查是否有名称或者url重复
	var c int64
	err = database.DB.Model(&model.RouteTarget{}).Where("id <> ? and (route_id = ? or upstream_id = ?)", instance.ID, instance.RouteID, instance.UpstreamID).Count(&c).Error
	if err != nil {
		logger.Errorln(err)
		return
	}
	if c > 0 {
		duplicated = true
		return
	}

	if err = database.DB.Model(&model.RouteTarget{ID: instance.ID}).Updates(instance).Error; err != nil {
		logger.Errorln(err)
		return
	}
	success = true
	return
}

func (u *routeTargetService) Delete(id uint64) (success bool, err error) {
	if id == 0 {
		logger.Error("ID is required")
		return
	}
	if err = database.DB.Delete(&model.RouteTarget{ID: id}).Error; err != nil {
		logger.Errorln(err)
		return
	}
	success = true
	return
}

func (u *routeTargetService) Get(id uint64) (instance *model.RouteTarget, err error) {
	if id == 0 {
		logger.Error("ID is required")
		return
	}
	instance = new(model.RouteTarget)
	if err = database.DB.Where(&model.RouteTarget{ID: id}).First(instance).Error; err != nil {
		logger.Errorln(err)
		return
	}
	return
}

func (u *routeTargetService) List(page, pageSize int, condition *model.RouteTarget) (instances []*model.RouteTarget, total int64, err error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	err = database.DB.Model(&model.RouteTarget{}).Where(condition).Count(&total).Error
	if err != nil {
		logger.Errorln(err)
		return
	}
	if total == 0 {
		return
	}
	err = database.DB.Where(condition).Offset((page - 1) * pageSize).Limit(pageSize).Find(&instances).Error
	if err != nil {
		logger.Errorln(err)
		return
	}
	return
}

func (u *routeTargetService) Save(instance *model.RouteTarget) (success, isAdd bool, err error) {
	if instance.RouteID == nil || instance.UpstreamID == nil {
		return
	}
	// 检查是否有routeId重复
	var rtList []*model.RouteTarget
	err = database.DB.Model(&model.RouteTarget{}).Where(&model.RouteTarget{RouteID: instance.RouteID, UpstreamID: instance.UpstreamID}).Find(&rtList).Error
	if err != nil {
		logger.Errorln(err)
		return
	}
	if len(rtList) > 0 {
		// 更新
		if instance.ID == 0 {
			instance.ID = rtList[0].ID
		}
		isAdd = false
		if err = database.DB.Model(&model.RouteTarget{ID: instance.ID}).Updates(instance).Error; err != nil {
			logger.Errorln(err)
			return
		}
	} else {
		instance.ID = util.SnowflakeId()
		isAdd = true
		// 新增
		if err = database.DB.Create(instance).Error; err != nil {
			logger.Errorln(err)
			return
		}
	}

	success = true
	return
}

func (u *routeTargetService) GetByRouteIDAndUpstreamID(routeID uint64, upstreamID uint64) (*model.RouteTarget, error) {
	var instanceList []*model.RouteTarget
	err := database.DB.Model(&model.RouteTarget{}).Where(&model.RouteTarget{RouteID: &routeID, UpstreamID: &upstreamID}).Find(&instanceList).Error
	if err != nil {
		logger.Errorln(err)
		return nil, err
	}
	if len(instanceList) == 0 {
		return nil, nil
	}
	return instanceList[0], nil
}
