package controller

import (
	"github.com/gofiber/fiber/v2"
	logger "github.com/sirupsen/logrus"
	"security-gateway/internal/model"
	"security-gateway/internal/proxy"
	"security-gateway/internal/service"
	"strconv"
)

var RouteTargetController = &routeTargetController{}

type routeTargetController struct {
}

func (c *routeTargetController) Add(ctx *fiber.Ctx) error {
	instance := new(model.RouteTarget)
	if err := ctx.BodyParser(instance); err != nil {
		return ctx.JSON(&CommonResponse{
			Code: ResponseCodeParamParseError,
			Msg:  ResponseMsgParamParseError,
		})
	}

	duplicated, success, err := service.RouteTargetService.Add(instance)
	if err != nil {
		return ctx.JSON(&CommonResponse{
			Code: ResponseCodeDatabase,
			Msg:  ResponseMsgDatabase + err.Error(),
		})
	}
	if duplicated {
		return ctx.JSON(&CommonResponse{
			Code: ResponseCodeDuplicated,
			Msg:  ResponseMsgDuplicated,
		})
	}
	if !success {
		return ctx.JSON(&CommonResponse{
			Code: ResponseCodeUnknownError,
			Msg:  ResponseMsgUnknownError,
		})
	}

	// 增加成功后，将路由添加到反向代理
	go c.addProxy(instance)

	return ctx.JSON(&CommonResponse{
		Data: instance,
	})
}

func (c *routeTargetController) Update(ctx *fiber.Ctx) error {
	instance := new(model.RouteTarget)
	if err := ctx.BodyParser(instance); err != nil {
		return ctx.JSON(&CommonResponse{
			Code: ResponseCodeParamParseError,
			Msg:  ResponseMsgParamParseError,
		})
	}

	duplicated, success, err := service.RouteTargetService.Update(instance)
	if err != nil {
		return ctx.JSON(&CommonResponse{
			Code: ResponseCodeDatabase,
			Msg:  ResponseMsgDatabase + err.Error(),
		})
	}
	if duplicated {
		return ctx.JSON(&CommonResponse{
			Code: ResponseCodeDuplicated,
			Msg:  ResponseMsgDuplicated,
		})
	}
	if !success {
		return ctx.JSON(&CommonResponse{
			Code: ResponseCodeUnknownError,
			Msg:  ResponseMsgUnknownError,
		})
	}
	return ctx.JSON(&CommonResponse{
		Data: instance,
	})
}

func (c *routeTargetController) Delete(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	if idStr == "" {
		return ctx.JSON(&CommonResponse{
			Code: ResponseCodeParamNotEnough,
			Msg:  ResponseMsgParamNotEnough + " id",
		})
	}

	id, err := strconv.ParseUint(idStr, 10, 64)

	success, err := service.RouteTargetService.Delete(id)
	if err != nil {
		return ctx.JSON(&CommonResponse{
			Code: ResponseCodeDatabase,
			Msg:  ResponseMsgDatabase + err.Error(),
		})
	}
	if !success {
		return ctx.JSON(&CommonResponse{
			Code: ResponseCodeUnknownError,
			Msg:  ResponseMsgUnknownError,
		})
	}

	// 删除成功后，将路由从反向代理中删除
	go c.removeProxy(id)

	return ctx.JSON(&CommonResponse{
		Data: success,
	})
}

func (c *routeTargetController) DeleteByRouteIDAndUpstreamID(ctx *fiber.Ctx) error {
	ru := new(model.RouteTarget)
	if err := ctx.BodyParser(ru); err != nil {
		return ctx.JSON(&CommonResponse{
			Code: ResponseCodeParamParseError,
			Msg:  ResponseMsgParamParseError,
		})

	}
	instance, err := service.RouteTargetService.GetByRouteIDAndUpstreamID(*ru.RouteID, *ru.UpstreamID)
	if err != nil {
		return ctx.JSON(&CommonResponse{
			Code: ResponseCodeDatabase,
			Msg:  ResponseMsgDatabase + err.Error(),
		})
	}
	if instance == nil {
		return ctx.JSON(&CommonResponse{
			Data: true,
		})
	}

	success, err := service.RouteTargetService.Delete(instance.ID)
	if err != nil {
		return ctx.JSON(&CommonResponse{
			Code: ResponseCodeDatabase,
			Msg:  ResponseMsgDatabase + err.Error(),
		})
	}
	if !success {
		return ctx.JSON(&CommonResponse{
			Code: ResponseCodeUnknownError,
			Msg:  ResponseMsgUnknownError,
		})
	}

	// 删除成功后，将路由从反向代理中删除
	go c.removeProxy(instance.ID)

	return ctx.JSON(&CommonResponse{
		Data: success,
	})

}

func (c *routeTargetController) Get(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	if idStr == "" {
		return ctx.JSON(&CommonResponse{
			Code: ResponseCodeParamNotEnough,
			Msg:  ResponseMsgParamNotEnough + " id",
		})
	}

	id, err := strconv.ParseUint(idStr, 10, 64)

	instance, err := service.RouteTargetService.Get(id)
	if err != nil {
		return ctx.JSON(&CommonResponse{
			Code: ResponseCodeDatabase,
			Msg:  ResponseMsgDatabase + err.Error(),
		})
	}
	if instance == nil {
		return ctx.JSON(&CommonResponse{
			Code: ResponseCodeDataNotExists,
			Msg:  ResponseMsgDataNotExists,
		})
	}
	return ctx.JSON(&CommonResponse{
		Data: instance,
	})
}

func (c *routeTargetController) List(ctx *fiber.Ctx) error {
	condition := new(model.RouteTarget)
	if err := ctx.BodyParser(condition); err != nil {
		return ctx.JSON(&CommonResponse{
			Code: ResponseCodeParamParseError,
			Msg:  ResponseMsgParamParseError,
		})

	}
	pageStr := ctx.Query("page")
	pageSizeStr := ctx.Query("pageSize")
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		page = 1
	}
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		pageSize = 10
	}

	instances, total, err := service.RouteTargetService.List(page, pageSize, condition)
	if err != nil {
		return ctx.JSON(&CommonResponse{
			Code: ResponseCodeDatabase,
			Msg:  ResponseMsgDatabase + err.Error(),
		})
	}
	if total == 0 {
		return ctx.JSON(&CommonResponse{
			Data: instances,
		})
	}
	return ctx.JSON(&CommonResponse{
		Data: map[string]interface{}{
			"total": total,
			"items": instances,
		},
	})
}

func (c *routeTargetController) addProxy(instance *model.RouteTarget) {
	// 1、获取路由信息
	route, err := service.RouteService.Get(*instance.RouteID)
	if err != nil {
		logger.Error("获取路由信息失败", err)
		return
	}
	if route == nil {
		logger.Error("路由信息不存在")
		return
	}
	// 2、获取路由对应的服务信息
	if route.ServiceID == nil {
		logger.Error("路由未绑定服务")
		return
	}
	serv, err := service.ServiceService.Get(*route.ServiceID)
	if err != nil {
		logger.Error("获取服务信息失败", err)
		return
	}
	if serv == nil {
		logger.Error("服务信息不存在")
		return
	}
	// 3、获取服务对应的上游信息
	upstream, err := service.UpstreamService.Get(*instance.UpstreamID)
	if err != nil {
		logger.Error("获取上游信息失败", err)
		return
	}
	if upstream == nil {
		logger.Error("上游信息不存在")
		return
	}

	// 4、添加到反向代理
	proxy.Manager.AddRoute(serv, route, upstream, instance.Weight)
}

func (c *routeTargetController) removeProxy(id uint64) {
	instance, err := service.RouteTargetService.Get(id)
	if err != nil {
		logger.Error("获取路由目标信息失败", err)
		return
	}
	if instance == nil {
		logger.Error("路由目标信息不存在")
		return
	}

	// 1、获取路由信息
	route, err := service.RouteService.Get(*instance.RouteID)
	if err != nil {
		logger.Error("获取路由信息失败", err)
		return
	}
	if route == nil {
		logger.Error("路由信息不存在")
		return
	}
	// 2、获取路由对应的服务信息
	if route.ServiceID == nil {
		logger.Error("路由未绑定服务")
		return
	}
	serv, err := service.ServiceService.Get(*route.ServiceID)
	if err != nil {
		logger.Error("获取服务信息失败", err)
		return
	}
	if serv == nil {
		logger.Error("服务信息不存在")
		return
	}

	// 3、获取上游信息
	upstream, err := service.UpstreamService.Get(*instance.UpstreamID)
	if err != nil {
		logger.Error("获取上游信息失败", err)
		return
	}

	// 4、删除反向代理
	proxy.Manager.RemoveRoute(*serv.Port, *serv.Domain, *route.Uri, *upstream.TargetUrl)
}

func (c *routeTargetController) Save(ctx *fiber.Ctx) error {
	instance := new(model.RouteTarget)
	if err := ctx.BodyParser(instance); err != nil {
		return ctx.JSON(&CommonResponse{
			Code: ResponseCodeParamParseError,
			Msg:  ResponseMsgParamParseError,
		})
	}

	success, isAdd, err := service.RouteTargetService.Save(instance)
	if err != nil {
		return ctx.JSON(&CommonResponse{
			Code: ResponseCodeDatabase,
			Msg:  ResponseMsgDatabase + err.Error(),
		})
	}
	if !success {
		return ctx.JSON(&CommonResponse{
			Code: ResponseCodeUnknownError,
			Msg:  ResponseMsgUnknownError,
		})
	}

	if isAdd {
		// 增加成功后，将路由添加到反向代理
		go c.addProxy(instance)
	} else {
		// 更新成功后，将路由从反向代理中更新
		go c.updateProxy(instance)
	}
	return ctx.JSON(&CommonResponse{
		Data: instance,
	})
}

func (c *routeTargetController) updateProxy(instance *model.RouteTarget) {
	// 1、获取路由信息
	route, err := service.RouteService.Get(*instance.RouteID)
	if err != nil {
		logger.Error("获取路由信息失败", err)
		return
	}
	if route == nil {
		logger.Error("路由信息不存在")
		return
	}
	// 2、获取路由对应的服务信息
	if route.ServiceID == nil {
		logger.Error("路由未绑定服务")
		return
	}
	serv, err := service.ServiceService.Get(*route.ServiceID)
	if err != nil {
		logger.Error("获取服务信息失败", err)
		return
	}
	if serv == nil {
		logger.Error("服务信息不存在")
		return
	}
	// 3、获取服务对应的上游信息
	upstream, err := service.UpstreamService.Get(*instance.UpstreamID)
	if err != nil {
		logger.Error("获取上游信息失败", err)
		return
	}
	if upstream == nil {
		logger.Error("上游信息不存在")
		return
	}

	// 4、删除旧的反向代理
	proxy.Manager.RemoveRoute(*serv.Port, *serv.Domain, *route.Uri, *upstream.TargetUrl)
	// 5、添加新的反向代理
	proxy.Manager.AddRoute(serv, route, upstream, instance.Weight)
}
