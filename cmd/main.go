package main

import (
	"embed"
	"fmt"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/panjf2000/ants/v2"
	logger "github.com/sirupsen/logrus"
	"net/http"
	"security-gateway/internal/controller"
	"security-gateway/internal/model"
	"security-gateway/internal/task"
	"security-gateway/pkg/config"
	"security-gateway/pkg/database"
	"security-gateway/pkg/util"
)

//go:embed dist/*
var embedPage embed.FS

func main() {
	startApp()
}

func startApp() {
	defer ants.Release()
	config.InitialLogger()

	err := util.InitNode(config.GetUint64("server.node"))
	if err != nil {
		logger.Errorln("初始化节点失败: ", err)
		return
	}

	err = database.Initial()
	if err != nil {
		logger.Errorln("数据库初始化失败: ", err)
		return
	}
	defer database.Close()

	err = database.AutoMigrate(model.Models...)
	if err != nil {
		logger.Errorln("数据库迁移失败: ", err)
		return
	}

	if config.GetBool("task.checkHealth") {
		err = task.StartHealthCheckTask()
		if err != nil {
			logger.Errorln(err)
		}
	}
	task.Start()
	defer task.Stop()

	controller.InitProxyManager()
	controller.InitRouter()

	controller.ServerApp.Use("/", filesystem.New(filesystem.Config{
		Root:       http.FS(embedPage),
		PathPrefix: "dist",
		Index:      "index.html",
		Browse:     true,
	}))

	err = controller.ServerApp.Listen(fmt.Sprintf(":%d", config.GetInt("server.port", 8080)))
	if err != nil {
		return
	}
}
