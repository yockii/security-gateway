package service

import (
	logger "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"security-gateway/internal/model"
	"security-gateway/pkg/database"
	"security-gateway/pkg/util"
)

var UserService = &userService{}

type userService struct{}

func (u *userService) Add(instance *model.User) (duplicated, success bool, err error) {
	if instance.UniKey == "" && instance.UniKeysJson == "" {
		return
	}
	// 检查是否有名称或者url重复
	var c int64
	err = database.DB.Model(&model.User{}).Where("uni_key = ? or uni_keys_json = ?", instance.UniKey, instance.UniKeysJson).Count(&c).Error
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

func (u *userService) Update(instance *model.User) (duplicated, success bool, err error) {
	if instance.ID == 0 {
		logger.Error("ID is required")
		return
	}
	// 检查是否有名称或者url重复
	var c int64
	err = database.DB.Model(&model.User{}).Where("id <> ? and (uni_key = ? or uni_keys_json = ?)", instance.ID, instance.UniKey, instance.UniKeysJson).Count(&c).Error
	if err != nil {
		logger.Errorln(err)
		return
	}
	if c > 0 {
		duplicated = true
		return
	}

	if err = database.DB.Model(&model.User{ID: instance.ID}).Updates(instance).Error; err != nil {
		logger.Errorln(err)
		return
	}
	success = true
	return
}

func (u *userService) Delete(id uint64) (success bool, err error) {
	if id == 0 {
		logger.Error("ID is required")
		return
	}
	if err = database.DB.Delete(&model.User{ID: id}).Error; err != nil {
		logger.Errorln(err)
		return
	}
	success = true
	return
}

func (u *userService) Get(id uint64) (instance *model.User, err error) {
	if id == 0 {
		logger.Error("ID is required")
		return
	}
	instance = new(model.User)
	if err = database.DB.Where(&model.User{ID: id}).First(instance).Error; err != nil {
		logger.Errorln(err)
		return
	}
	return
}

func (u *userService) List(page, pageSize int, username string) (instances []*model.User, total int64, err error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if username == "" {
		err = database.DB.Model(&model.User{}).Count(&total).Error
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

	err = database.DB.Model(&model.User{}).Where("username like ?", "%"+username+"%").Count(&total).Error
	if err != nil {
		logger.Errorln(err)
		return
	}
	if total == 0 {
		return
	}
	err = database.DB.Where("username like ?", "%"+username+"%").Offset((page - 1) * pageSize).Limit(pageSize).Find(&instances).Error
	if err != nil {
		logger.Errorln(err)
		return
	}
	return
}

func (u *userService) GetByUniKey(username, uniKey string) *model.User {
	user := new(model.User)
	if err := database.DB.Where(&model.User{Username: username, UniKey: uniKey}).First(user).Error; err != nil {
		logger.Errorln(err)
		return nil
	}
	return user
}

func (u *userService) GetByUniKeyJson(username, uniKey, matchKey string) *model.User {
	// 从json中获取匹配键的值, 这里的uniKey是uniKeysJson中的matchKey代表的键的值
	var users []*model.User
	sess := database.DB.Where("uni_keys_json like ?", "%"+uniKey+"%")
	if username != "" {
		sess = sess.Where("username = ?", username)
	}
	if err := sess.Find(&users).Error; err != nil {
		logger.Errorln(err)
		return nil
	}

	// 遍历这些数据，检查uni_keys_json字段是否确实符合要求
	for _, user := range users {
		matchKeyValue := gjson.Get(user.UniKeysJson, matchKey).String()
		if matchKeyValue == uniKey {
			return user
		}
	}

	// 如果没有找到符合要求的数据，返回nil
	return nil
}
