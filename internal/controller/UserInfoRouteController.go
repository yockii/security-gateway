package controller

import (
	"github.com/gofiber/fiber/v2"
	"security-gateway/internal/model"
	"security-gateway/internal/proxy"
	"security-gateway/internal/service"
	"strconv"
)

var UserInfoRouteController = &userInfoRouteController{}

type userInfoRouteController struct {
}

func (c *userInfoRouteController) Add(ctx *fiber.Ctx) error {
	instance := new(model.UserInfoRoute)
	if err := ctx.BodyParser(instance); err != nil {
		return ctx.JSON(&CommonResponse{
			Code: ResponseCodeParamParseError,
			Msg:  ResponseMsgParamParseError,
		})
	}

	duplicated, success, err := service.UserInfoRouteService.Add(instance)
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

	go c.infoAdded(instance)

	return ctx.JSON(&CommonResponse{
		Data: instance,
	})
}

func (c *userInfoRouteController) Update(ctx *fiber.Ctx) error {
	instance := new(model.UserInfoRoute)
	if err := ctx.BodyParser(instance); err != nil {
		return ctx.JSON(&CommonResponse{
			Code: ResponseCodeParamParseError,
			Msg:  ResponseMsgParamParseError,
		})
	}

	duplicated, success, err := service.UserInfoRouteService.Update(instance)
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

	go c.infoUpdated(instance.ID)

	return ctx.JSON(&CommonResponse{
		Data: instance,
	})
}

func (c *userInfoRouteController) Delete(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	if idStr == "" {
		return ctx.JSON(&CommonResponse{
			Code: ResponseCodeParamNotEnough,
			Msg:  ResponseMsgParamNotEnough + " id",
		})
	}

	id, err := strconv.ParseUint(idStr, 10, 64)

	oldInstance, err := service.UserInfoRouteService.Get(id)
	if err != nil {
		return ctx.JSON(&CommonResponse{
			Code: ResponseCodeDatabase,
			Msg:  ResponseMsgDatabase + err.Error(),
		})

	}

	success, err := service.UserInfoRouteService.Delete(id)
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

	go c.infoDeleted(oldInstance)

	return ctx.JSON(&CommonResponse{
		Data: success,
	})
}

func (c *userInfoRouteController) Get(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	if idStr == "" {
		return ctx.JSON(&CommonResponse{
			Code: ResponseCodeParamNotEnough,
			Msg:  ResponseMsgParamNotEnough + " id",
		})
	}

	id, err := strconv.ParseUint(idStr, 10, 64)

	instance, err := service.UserInfoRouteService.Get(id)
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

func (c *userInfoRouteController) List(ctx *fiber.Ctx) error {
	pageStr := ctx.Query("page")
	pageSizeStr := ctx.Query("pageSize")
	name := ctx.Query("name")

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		page = 1
	}
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		pageSize = 10
	}

	instances, total, err := service.UserInfoRouteService.List(page, pageSize, name)
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

func (c *userInfoRouteController) infoAdded(instance *model.UserInfoRoute) {
	// 获取服务
	serv, err := service.ServiceService.Get(instance.ServiceID)
	if err != nil || serv == nil {
		return
	}
	// 添加到反向代理
	proxy.Manager.AddUserRoute(serv, instance)
}

func (c *userInfoRouteController) infoUpdated(id uint64) {
	// 获取服务
	instance, err := service.UserInfoRouteService.Get(id)
	if err != nil || instance == nil {
		return
	}
	serv, err := service.ServiceService.Get(instance.ServiceID)
	if err != nil || serv == nil {
		return
	}
	// 更新反向代理
	proxy.Manager.UpdateUserRoute(serv, instance)
}

func (c *userInfoRouteController) infoDeleted(instance *model.UserInfoRoute) {
	if instance == nil {
		return
	}
	serv, err := service.ServiceService.Get(instance.ServiceID)
	if err != nil || serv == nil {
		return
	}
	// 删除反向代理
	proxy.Manager.RemoveUserRoute(*serv.Port, *serv.Domain)
}

func (c *userInfoRouteController) GetByService(ctx *fiber.Ctx) error {
	serviceIdStr := ctx.Params("serviceId")
	if serviceIdStr == "" {
		return ctx.JSON(&CommonResponse{
			Code: ResponseCodeParamNotEnough,
			Msg:  ResponseMsgParamNotEnough + " serviceId",
		})
	}

	serviceId, err := strconv.ParseUint(serviceIdStr, 10, 64)

	instances, err := service.UserInfoRouteService.GetByServiceID(serviceId)
	if err != nil {
		return ctx.JSON(&CommonResponse{
			Code: ResponseCodeDatabase,
			Msg:  ResponseMsgDatabase + err.Error(),
		})
	}
	return ctx.JSON(&CommonResponse{
		Data: instances,
	})
}
