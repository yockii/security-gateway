package task

import (
	"github.com/gofiber/fiber/v2"
	logger "github.com/sirupsen/logrus"
	"security-gateway/internal/model"
	"security-gateway/pkg/database"
	"time"
)

func StartHealthCheckTask() error {
	// 每1分钟执行一次
	_, err := c.AddFunc("0 * * * * *", func() {
		t := time.Now().UnixMilli()
		t5 := t - 5*60*1000
		t1 := t - 1*60*1000
		// 查询所有上次检查在5分钟前的正常状态的Upstream以及1分钟前的异常状态的Upstream
		var ups []*model.Upstream
		if err := database.DB.Model(&model.Upstream{}).Where("health_check_url like 'http%'").Where("status = 0 OR (status = 1 AND last_check_time < ?) OR (status = 2 AND last_check_time < ?)", t5, t1).Find(&ups).Error; err != nil {
			logger.Error(err)
			return
		}
		for _, up := range ups {
			checkHealth(up)
		}
	})
	return err
}

func checkHealth(upstream *model.Upstream) {
	if upstream.HealthCheckUrl == nil || *(upstream.HealthCheckUrl) == "" {
		return
	}
	healthUrl := *(upstream.HealthCheckUrl)
	// 尝试访问（3s超时）并获取返回状态，如果为200则正常，否则异常
	a := fiber.Get(healthUrl)
	statusCode, _, _ := a.Bytes()
	if statusCode == 200 {
		database.DB.Model(&model.Upstream{
			ID: upstream.ID,
		}).Updates(&model.Upstream{
			LastCheckTime: time.Now().UnixMilli(),
			Status:        1,
		})
	} else {
		database.DB.Model(&model.Upstream{
			ID: upstream.ID,
		}).Updates(&model.Upstream{
			LastCheckTime: time.Now().UnixMilli(),
			Status:        2,
		})
	}
}
