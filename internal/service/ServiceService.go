package service

import (
	logger "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"security-gateway/internal/model"
	"security-gateway/pkg/database"
	"security-gateway/pkg/util"
)

var ServiceService = &serviceService{}

type serviceService struct{}

func (u *serviceService) Add(instance *model.Service) (duplicated, success bool, err error) {
	if instance.Name == nil || instance.Port == nil || *(instance.Name) == "" || *(instance.Port) <= 0 || *(instance.Port) > 65535 {
		return
	}
	if instance.Domain == nil {
		instance.Domain = new(string)
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
	if instance.Port != nil && (*(instance.Port) <= 0 || *(instance.Port) > 65535) {
		instance.Port = nil
	}

	// 检查是否有名称或者url重复
	var c int64
	err = database.DB.Model(&model.Service{}).Where("id <> ? and (name = ? or (domain = ? and port = ?))", instance.ID, instance.Name, instance.Domain, instance.Port).Count(&c).Error
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
	// 开启事务，删除服务的同时删除服务下的路由、路由目标、路由脱敏规则、服务脱敏规则
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		// 1、查询所有服务下的路由
		var routes []*model.Route
		if err = tx.Where(&model.Route{ServiceID: &id}).Find(&routes).Error; err != nil {
			logger.Errorln(err)
			return err
		}
		// 2、删除路由及路由下的路由目标、路由脱敏规则
		for _, route := range routes {
			// 删除路由下的路由目标
			if err = tx.Where(&model.RouteTarget{RouteID: &route.ID}).Delete(&model.RouteTarget{}).Error; err != nil {
				logger.Errorln(err)
				return err
			}
			// 删除路由下的路由脱敏规则
			if err = tx.Where(&model.RouteField{RouteID: route.ID}).Delete(&model.RouteField{}).Error; err != nil {
				logger.Errorln(err)
				return err
			}
			// 删除路由
			if err = tx.Delete(route).Error; err != nil {
				logger.Errorln(err)
				return err
			}
		}
		// 3、删除服务下的服务脱敏规则
		if err = tx.Where(&model.ServiceField{ServiceID: id}).Delete(&model.ServiceField{}).Error; err != nil {
			logger.Errorln(err)
			return err
		}

		// 4、删除服务
		if err = tx.Delete(&model.Service{ID: id}).Error; err != nil {
			logger.Errorln(err)
			return err
		}

		return nil
	})

	success = err == nil
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

func (u *serviceService) List(page, pageSize int, condition *model.Service) (instances []*model.Service, total int64, err error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	sess := database.DB.Model(&model.Service{})
	if condition.Name != nil && *(condition.Name) != "" {
		sess = sess.Where("name like ?", "%"+*(condition.Name)+"%")
		condition.Name = nil
	}
	if condition.Domain != nil && *(condition.Domain) != "" {
		sess = sess.Where("domain like ?", "%"+*(condition.Domain)+"%")
		condition.Domain = nil
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

func (u *serviceService) GetByDomainAndPort(domain string, port uint16) (*model.Service, error) {
	instance := new(model.Service)
	err := database.DB.Where(&model.Service{Domain: &domain, Port: &port}).First(instance).Error
	if err != nil {
		logger.Errorln(err)
		return nil, err
	}
	return instance, nil
}

func (u *serviceService) GetAllPorts() (ports []uint16, err error) {
	err = database.DB.Model(&model.Service{}).Distinct().Pluck("port", &ports).Error
	if err != nil {
		logger.Errorln(err)
		return nil, err
	}
	return
}
