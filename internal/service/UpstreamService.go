package service

import (
	logger "github.com/sirupsen/logrus"
	"security-gateway/internal/model"
	"security-gateway/pkg/database"
	"security-gateway/pkg/util"
)

type UpstreamService struct{}

func (u *UpstreamService) Add(instance *model.Upstream) (duplicated, success bool, err error) {
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

func (u *UpstreamService) Update(instance *model.Upstream) (duplicated, success bool, err error) {
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
	err = database.DB.Model(&model.Upstream{}).Where("name = ? or target_url = ?", instance.Name, instance.TargetUrl).Count(&c).Error
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

func (u *UpstreamService) Delete(id uint64) (success bool, err error) {
	if id == 0 {
		logger.Error("ID is required")
		return
	}
	if err = database.DB.Delete(&model.Upstream{ID: id}).Error; err != nil {
		logger.Errorln(err)
		return
	}
	success = true
	return
}

func (u *UpstreamService) Get(id uint64) (instance *model.Upstream, err error) {
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

func (u *UpstreamService) List(page, pageSize int, name string) (instances []*model.Upstream, total int64, err error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if name == "" {
		err = database.DB.Model(&model.Upstream{}).Count(&total).Error
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

	err = database.DB.Model(&model.Upstream{}).Where("name like ?", "%"+name+"%").Count(&total).Error
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
