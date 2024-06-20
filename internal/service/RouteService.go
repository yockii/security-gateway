package service

import (
	logger "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"security-gateway/internal/domain"
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
	err = database.DB.Model(&model.Route{}).Where(&model.Route{
		ServiceID: instance.ServiceID,
		Uri:       instance.Uri,
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
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		// 删除路由目标、路由规则
		if err = tx.Where(&model.RouteTarget{RouteID: &id}).Delete(&model.RouteTarget{}).Error; err != nil {
			logger.Errorln(err)
			return err
		}
		if err = tx.Where(&model.RouteField{RouteID: id}).Delete(&model.RouteField{}).Error; err != nil {
			logger.Errorln(err)
			return err
		}
		if err = tx.Where(&model.Route{ID: id}).Delete(&model.Route{}).Error; err != nil {
			logger.Errorln(err)
			return err
		}

		return nil
	})
	success = err == nil
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

func (u *routeService) List(page, pageSize int, condition *model.Route) (instances []*model.Route, total int64, err error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	sess := database.DB.Model(&model.Route{})
	if condition.Uri != nil && *(condition.Uri) != "" {
		sess = sess.Where("uri like ?", "%"+*(condition.Uri)+"%")
		condition.Uri = nil
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

func (u *routeService) ListWithTargets(page, pageSize int, condition *model.Route) (result []*domain.RouteWithTargets, total int64, err error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	sess := database.DB.Model(&model.Route{})
	if condition.Uri != nil && *(condition.Uri) != "" {
		sess = sess.Where("uri like ?", "%"+*(condition.Uri)+"%")
		condition.Uri = nil
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

	var instances []*model.Route
	err = sess.Offset((page - 1) * pageSize).Limit(pageSize).Find(&instances).Error
	if err != nil {
		logger.Errorln(err)
		return
	}

	for _, v := range instances {
		rt := &domain.RouteWithTargets{
			Route: v,
		}
		result = append(result, rt)

		// 获取关联的目标
		var targets []*model.Upstream
		// 关联查询，通过RouteTarget中的routeId（关联v.id)，去Upstream中查找关联的upstreamId的所有数据
		subQuery := database.DB.Model(&model.RouteTarget{}).Where(&model.RouteTarget{RouteID: &v.ID}).Select("upstream_id")
		if err = database.DB.Model(&model.Upstream{}).Where("id in (?)", subQuery).Find(&targets).Error; err != nil {
			logger.Errorln(err)
			continue
		}

		rt.Targets = targets
	}
	err = nil
	return
}
