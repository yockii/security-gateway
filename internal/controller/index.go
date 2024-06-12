package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	logger "github.com/sirupsen/logrus"

	"security-gateway/pkg/config"
)

var ServerApp *fiber.App

func init() {
	ServerApp = fiber.New(fiber.Config{
		DisableStartupMessage: true,
		BodyLimit:             config.GetInt("server.bodyLimit", -1),
	})
	ServerApp.Use(recover.New(recover.Config{
		EnableStackTrace: true,
		StackTraceHandler: func(c *fiber.Ctx, e interface{}) {
			logger.Errorln(e)
		},
	}))
	ServerApp.Use(cors.New())
}

func InitRouter() {
	apiV1 := ServerApp.Group("/api/v1")

	// Upstream
	upstream := apiV1.Group("/upstream")
	upstream.Post("/add", UpstreamController.Add)
	upstream.Post("/update", UpstreamController.Update)
	upstream.Post("/delete/:id", UpstreamController.Delete)
	upstream.Get("/instance/:id", UpstreamController.Get)
	upstream.Get("/list", UpstreamController.List)

	// Service
	service := apiV1.Group("/service")
	service.Post("/add", ServiceController.Add)
	service.Post("/update", ServiceController.Update)
	service.Post("/delete/:id", ServiceController.Delete)
	service.Get("/instance/:id", ServiceController.Get)
	service.Get("/list", ServiceController.List)

	// Route
	route := apiV1.Group("/route")
	route.Post("/add", RouteController.Add)
	route.Post("/update", RouteController.Update)
	route.Post("/delete/:id", RouteController.Delete)
	route.Get("/instance/:id", RouteController.Get)
	route.Get("/list", RouteController.List)

}

const (
	ResponseCodeParamParseError = -iota - 1000
	ResponseCodeParamNotEnough
	ResponseCodeDatabase
	ResponseCodeDuplicated
	ResponseCodeDataNotExists

	ResponseCodeUnknownError
)

var (
	ResponseMsgParamParseError = "Param parse error"
	ResponseMsgParamNotEnough  = "Param not enough"
	ResponseMsgDatabase        = "Database error"
	ResponseMsgDuplicated      = "Duplicated"
	ResponseMsgDataNotExists   = "Data not exists"

	ResponseMsgUnknownError = "Unknown error"
)

type CommonResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}
