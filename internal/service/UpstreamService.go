package service

import (
	logger "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"security-gateway/internal/domain"
	"security-gateway/internal/model"
	"security-gateway/pkg/database"
	"security-gateway/pkg/util"
)

var UpstreamService = &upstreamService{}

type upstreamService struct{}

func (u *upstreamService) Add(instance *model.Upstream) (duplicated, success bool, err error) {
	if instance.Name == nil || instance.TargetUrl == nil || *(instance.Name) == "" || *(instance.TargetUrl) == "" {
		return
	}
	// 检查是否有名称或者url重复
	var c int64
	err = database.DB.Model(&model.Upstream{}).Where("name = ? or target_url = ?", instance.Name, instance.TargetUrl).Count(&c).Error
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

func (u *upstreamService) Update(instance *model.Upstream) (duplicated, success bool, err error) {
	if instance.ID == 0 {
		logger.Error("ID is required")
		return
	}
	if instance.Name != nil && *(instance.Name) == "" {
		instance.Name = nil
	}
	if instance.TargetUrl != nil && *(instance.TargetUrl) == "" {
		instance.TargetUrl = nil
	}

	// 检查是否有名称或者url重复
	var c int64
	err = database.DB.Model(&model.Upstream{}).Where("id <> ? and (name = ? or target_url = ?)", instance.ID, instance.Name, instance.TargetUrl).Count(&c).Error
	if err != nil {
		logger.Errorln(err)
		return
	}
	if c > 0 {
		duplicated = true
		return
	}

	if err = database.DB.Model(&model.Upstream{ID: instance.ID}).Updates(instance).Error; err != nil {
		logger.Errorln(err)
		return
	}
	success = true
	return
}

func (u *upstreamService) Delete(id uint64) (success bool, err error) {
	if id == 0 {
		logger.Error("ID is required")
		return
	}

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		// 删除路由目标关联
		if err = tx.Where(&model.RouteTarget{UpstreamID: &id}).Delete(&model.RouteTarget{}).Error; err != nil {
			logger.Errorln(err)
			return err
		}
		if err = tx.Delete(&model.Upstream{ID: id}).Error; err != nil {
			logger.Errorln(err)
			return err
		}
		return nil
	})
	success = err == nil
	return
}

func (u *upstreamService) Get(id uint64) (instance *model.Upstream, err error) {
	if id == 0 {
		logger.Error("ID is required")
		return
	}
	instance = new(model.Upstream)
	if err = database.DB.Where(&model.Upstream{ID: id}).First(instance).Error; err != nil {
		logger.Errorln(err)
		return
	}
	return
}

func (u *upstreamService) List(page, pageSize int, condition *model.Upstream) (instances []*model.Upstream, total int64, err error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	if condition.Name != nil && *(condition.Name) == "" {
		condition.Name = nil
	}
	if condition.TargetUrl != nil && *(condition.TargetUrl) == "" {
		condition.TargetUrl = nil
	}

	sess := database.DB.Model(&model.Upstream{})
	if condition.Name != nil && *(condition.Name) != "" {
		sess = sess.Where("name like ?", "%"+*(condition.Name)+"%")
		condition.Name = nil
	}
	if condition.TargetUrl != nil && *(condition.TargetUrl) != "" {
		sess = sess.Where("target_url like ?", "%"+*(condition.TargetUrl)+"%")
		condition.TargetUrl = nil
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

func (u *upstreamService) ListByRoute(page, pageSize int, routeId uint64) (instances []*domain.TargetWithUpstream, total int64, err error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	err = database.DB.Model(&model.RouteTarget{}).Where(&model.RouteTarget{RouteID: &routeId}).Count(&total).Error
	if err != nil {
		logger.Errorln(err)
		return
	}
	if total == 0 {
		return
	}

	var rtList []*model.RouteTarget
	if err = database.DB.Model(&model.RouteTarget{}).Where(&model.RouteTarget{RouteID: &routeId}).Offset((page - 1) * pageSize).Limit(pageSize).Find(&rtList).Error; err != nil {
		logger.Errorln(err)
		return
	}

	var upstreamIdList []uint64
	for _, rt := range rtList {
		upstreamIdList = append(upstreamIdList, *(rt.UpstreamID))
	}

	var upstreamList []*model.Upstream
	if err = database.DB.Model(&model.Upstream{}).Where("id in ?", upstreamIdList).Find(&upstreamList).Error; err != nil {
		logger.Errorln(err)
		return
	}

	for _, rt := range rtList {
		tu := &domain.TargetWithUpstream{
			RouteTarget: *rt,
		}
		for _, upstream := range upstreamList {
			if *(rt.UpstreamID) == upstream.ID {
				tu.Upstream = upstream
				break
			}
		}
		instances = append(instances, tu)
	}

	return
}
