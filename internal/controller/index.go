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
	page := 1
	routeTargetList, total, err := service.RouteTargetService.List(page, 100, nil)
	if err != nil {
		logger.Errorln("初始化反向代理失败: ", err)
		return
	}
	for int(total) > len(routeTargetList) {
		var list []*model.RouteTarget
		page++
		list, total, err = service.RouteTargetService.List(page, 100, nil)
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
		proxy.Manager.AddRoute(serv, route, upstream, routeTarget)
	}

	// 加载所有获取用户信息的路由到反向代理管理器中
	page = 1
	userInfoRouteList, total, err := service.UserInfoRouteService.List(page, 100, "")
	if err != nil {
		logger.Errorln("初始化反向代理失败: ", err)
		return
	}
	for int(total) > len(userInfoRouteList) {
		var list []*model.UserInfoRoute
		page++
		list, total, err = service.UserInfoRouteService.List(page, 100, "")
		if err != nil {
			logger.Errorln("初始化反向代理失败: ", err)
			return
		}
		userInfoRouteList = append(userInfoRouteList, list...)
	}
	for _, userInfoRoute := range userInfoRouteList {
		serv, ok := services[userInfoRoute.ServiceID]
		if !ok {
			serv, err = service.ServiceService.Get(userInfoRoute.ServiceID)
			if err != nil {
				logger.Errorln("初始化反向代理失败: ", err)
				return
			}
			services[userInfoRoute.ServiceID] = serv
		}
		proxy.Manager.AddUserRoute(serv, userInfoRoute)
	}

	// 所有脱敏字段添加到反向代理管理器中
	page = 1
	secretFieldList, total, err := service.ServiceFieldService.List(page, 100, &model.ServiceField{})
	if err != nil {
		logger.Errorln("初始化反向代理失败: ", err)
		return
	}
	for int(total) > len(secretFieldList) {
		var list []*model.ServiceField
		page++
		list, total, err = service.ServiceFieldService.List(page, 100, &model.ServiceField{})
		if err != nil {
			logger.Errorln("初始化反向代理失败: ", err)
			return
		}
		secretFieldList = append(secretFieldList, list...)
	}
	for _, secretField := range secretFieldList {
		serv, ok := services[secretField.ServiceID]
		if !ok {
			serv, err = service.ServiceService.Get(secretField.ServiceID)
			if err != nil {
				logger.Errorln("初始化反向代理失败: ", err)
				return
			}
			services[secretField.ServiceID] = serv
		}
		proxy.Manager.UpdateServiceField(serv, secretField)
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
	serv := apiV1.Group("/service")
	serv.Post("/add", ServiceController.Add)
	serv.Post("/update", ServiceController.Update)
	serv.Post("/delete/:id", ServiceController.Delete)
	serv.Get("/instance/:id", ServiceController.Get)
	serv.Get("/list", ServiceController.List)
	serv.Get("/ports", ServiceController.Ports)

	// Route
	route := apiV1.Group("/route")
	route.Post("/add", RouteController.Add)
	route.Post("/update", RouteController.Update)
	route.Post("/delete/:id", RouteController.Delete)
	route.Get("/instance/:id", RouteController.Get)
	route.Get("/list", RouteController.List)
	route.Get("/listWithTarget", RouteController.ListWithTarget)

	// RouteTarget
	routeTarget := apiV1.Group("/routeTarget")
	routeTarget.Post("/save", RouteTargetController.Save)
	routeTarget.Post("/add", RouteTargetController.Add)
	//routeTarget.Post("/update", RouteTargetController.Update)
	routeTarget.Post("/delete/:id", RouteTargetController.Delete)
	routeTarget.Get("/instance/:id", RouteTargetController.Get)
	routeTarget.Get("/list", RouteTargetController.List)

	// User
	user := apiV1.Group("/user")
	user.Post("/add", UserController.Add)
	user.Post("/update", UserController.Update)
	user.Post("/delete/:id", UserController.Delete)
	user.Get("/instance/:id", UserController.Get)
	user.Get("/list", UserController.List)

	// UserInfoRoute
	userInfoRoute := apiV1.Group("/userInfoRoute")
	userInfoRoute.Post("/add", UserInfoRouteController.Add)
	userInfoRoute.Post("/update", UserInfoRouteController.Update)
	userInfoRoute.Post("/delete/:id", UserInfoRouteController.Delete)
	userInfoRoute.Get("/instance/:id", UserInfoRouteController.Get)
	userInfoRoute.Get("/list", UserInfoRouteController.List)

	// UserServiceLevel
	userServiceLevel := apiV1.Group("/userServiceLevel")
	userServiceLevel.Post("/add", UserServiceLevelController.Add)
	userServiceLevel.Post("/update", UserServiceLevelController.Update)
	userServiceLevel.Post("/delete/:id", UserServiceLevelController.Delete)
	userServiceLevel.Get("/instance/:id", UserServiceLevelController.Get)
	userServiceLevel.Get("/list", UserServiceLevelController.List)

	// ServiceField
	secretField := apiV1.Group("/serviceField")
	secretField.Post("/add", ServiceFieldController.Add)
	secretField.Post("/update", ServiceFieldController.Update)
	secretField.Post("/delete/:id", ServiceFieldController.Delete)
	secretField.Get("/instance/:id", ServiceFieldController.Get)
	secretField.Get("/list", ServiceFieldController.List)

	// RouteField
	routeField := apiV1.Group("/routeField")
	routeField.Post("/add", RouteFieldController.Add)
	routeField.Post("/update", RouteFieldController.Update)
	routeField.Post("/delete/:id", RouteFieldController.Delete)
	routeField.Get("/instance/:id", RouteFieldController.Get)
	routeField.Get("/list", RouteFieldController.List)

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
