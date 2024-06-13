package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	logger "github.com/sirupsen/logrus"
	"security-gateway/internal/model"
	"security-gateway/internal/proxy"
	"security-gateway/internal/service"

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

func InitProxyManager() {
	routeTargetList, total, err := service.RouteTargetService.List(1, 100, nil)
	if err != nil {
		logger.Errorln("初始化反向代理失败: ", err)
		return
	}
	for int(total) > len(routeTargetList) {
		var list []*model.RouteTarget
		list, total, err = service.RouteTargetService.List(1, 100, nil)
		if err != nil {
			logger.Errorln("初始化反向代理失败: ", err)
			return
		}
		routeTargetList = append(routeTargetList, list...)
	}

	// 遍历所有的路由目标，添加到反向代理管理器中
	var routes = make(map[uint64]*model.Route)
	var upstreams = make(map[uint64]*model.Upstream)
	var services = make(map[uint64]*model.Service)

	for _, routeTarget := range routeTargetList {
		route, ok := routes[*routeTarget.RouteID]
		if !ok {
			route, err = service.RouteService.Get(*routeTarget.RouteID)
			if err != nil {
				logger.Errorln("初始化反向代理失败: ", err)
				return
			}
			routes[*routeTarget.RouteID] = route
		}
		serv, ok := services[*route.ServiceID]
		if !ok {
			serv, err = service.ServiceService.Get(*route.ServiceID)
			if err != nil {
				logger.Errorln("初始化反向代理失败: ", err)
				return
			}
			services[*route.ServiceID] = serv
		}
		upstream, ok := upstreams[*routeTarget.UpstreamID]
		if !ok {
			upstream, err = service.UpstreamService.Get(*routeTarget.UpstreamID)
			if err != nil {
				logger.Errorln("初始化反向代理失败: ", err)
				return
			}
			upstreams[*routeTarget.UpstreamID] = upstream
		}
		proxy.Manager.AddRoute(*serv.Port, *serv.Domain, *route.Uri, *upstream.TargetUrl, routeTarget.Weight)
	}
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

	// RouteTarget
	routeTarget := apiV1.Group("/routeTarget")
	routeTarget.Post("/add", RouteTargetController.Add)
	//routeTarget.Post("/update", RouteTargetController.Update)
	routeTarget.Post("/delete/:id", RouteTargetController.Delete)
	routeTarget.Get("/instance/:id", RouteTargetController.Get)
	routeTarget.Get("/list", RouteTargetController.List)
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
