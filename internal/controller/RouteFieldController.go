package controller

import (
	"github.com/gofiber/fiber/v2"
	"security-gateway/internal/model"
	"security-gateway/internal/proxy"
	"security-gateway/internal/service"
	"strconv"
)

var RouteFieldController = &routeFieldController{}

type routeFieldController struct {
}

func (c *routeFieldController) Add(ctx *fiber.Ctx) error {
	instance := new(model.RouteField)
	if err := ctx.BodyParser(instance); err != nil {
		return ctx.JSON(&CommonResponse{
			Code: ResponseCodeParamParseError,
			Msg:  ResponseMsgParamParseError,
		})
	}

	duplicated, success, err := service.RouteFieldService.Add(instance)
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

	go c.fieldAdded(instance)

	return ctx.JSON(&CommonResponse{
		Data: instance,
	})
}

func (c *routeFieldController) Update(ctx *fiber.Ctx) error {
	instance := new(model.RouteField)
	if err := ctx.BodyParser(instance); err != nil {
		return ctx.JSON(&CommonResponse{
			Code: ResponseCodeParamParseError,
			Msg:  ResponseMsgParamParseError,
		})
	}

	duplicated, success, err := service.RouteFieldService.Update(instance)
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

	go c.fieldUpdated(instance.ID)

	return ctx.JSON(&CommonResponse{
		Data: instance,
	})
}

func (c *routeFieldController) Delete(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	if idStr == "" {
		return ctx.JSON(&CommonResponse{
			Code: ResponseCodeParamNotEnough,
			Msg:  ResponseMsgParamNotEnough + " id",
		})
	}

	id, err := strconv.ParseUint(idStr, 10, 64)

	success, err := service.RouteFieldService.Delete(id)
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

	go c.fieldDeleted(id)

	return ctx.JSON(&CommonResponse{
		Data: success,
	})
}

func (c *routeFieldController) Get(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	if idStr == "" {
		return ctx.JSON(&CommonResponse{
			Code: ResponseCodeParamNotEnough,
			Msg:  ResponseMsgParamNotEnough + " id",
		})
	}

	id, err := strconv.ParseUint(idStr, 10, 64)

	instance, err := service.RouteFieldService.Get(id)
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

func (c *routeFieldController) List(ctx *fiber.Ctx) error {
	pageStr := ctx.Query("page")
	pageSizeStr := ctx.Query("pageSize")
	condition := new(model.RouteField)
	if err := ctx.QueryParser(condition); err != nil {
		return ctx.JSON(&CommonResponse{
			Code: ResponseCodeParamParseError,
			Msg:  ResponseMsgParamParseError,
		})
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		page = 1
	}
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		pageSize = 10
	}

	instances, total, err := service.RouteFieldService.List(page, pageSize, condition)
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

func (c *routeFieldController) fieldDeleted(id uint64) {
	// 获取字段信息
	field, err := service.RouteFieldService.Get(id)
	if err != nil {
		return
	}
	if field == nil {
		return
	}
	// 获取路由
	route, err := service.RouteService.Get(field.RouteID)
	if err != nil {
		return
	}
	if route == nil {
		return
	}
	// 获取服务
	serv, err := service.ServiceService.Get(*route.ServiceID)
	if err != nil {
		return
	}
	if serv == nil {
		return
	}

	proxy.Manager.RemoveRouteField(serv, route, field.FieldName)
}

func (c *routeFieldController) fieldUpdated(id uint64) {
	// 获取字段信息
	instance, err := service.RouteFieldService.Get(id)
	if err != nil || instance == nil {
		return
	}
	// 获取路由
	route, err := service.RouteService.Get(instance.RouteID)
	if err != nil || route == nil {
		return
	}
	// 获取服务
	serv, err := service.ServiceService.Get(*route.ServiceID)
	if err != nil || serv == nil {
		return
	}

	proxy.Manager.UpdateRouteField(serv, route, instance)
}

func (c *routeFieldController) fieldAdded(instance *model.RouteField) {
	// 获取路由
	route, err := service.RouteService.Get(instance.RouteID)
	if err != nil || route == nil {
		return
	}
	// 获取服务
	serv, err := service.ServiceService.Get(*route.ServiceID)
	if err != nil || serv == nil {
		return
	}

	proxy.Manager.UpdateRouteField(serv, route, instance)
}
