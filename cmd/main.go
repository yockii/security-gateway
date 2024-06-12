package main

import (
	"fmt"
	"github.com/panjf2000/ants/v2"
	logger "github.com/sirupsen/logrus"
	"security-gateway/internal/controller"
	"security-gateway/internal/model"
	"security-gateway/pkg/config"
	"security-gateway/pkg/database"
	"security-gateway/pkg/util"
)

func main() {
	defer ants.Release()
	config.InitialLogger()

	err := util.InitNode(config.GetUint64("server.node"))
	if err != nil {
		return
	}

	database.Initial()
	defer database.Close()

	err = database.AutoMigrate(model.Models...)
	if err != nil {
		logger.Errorln("数据库迁移失败: ", err)
		return
	}

	controller.InitRouter()

	err = controller.ServerApp.Listen(fmt.Sprintf(":%d", config.GetInt("server.port", 8080)))
	if err != nil {
		return
	}
}
