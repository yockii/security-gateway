package controller

import (
	"github.com/gofiber/fiber/v2"
	"security-gateway/internal/model"
	"security-gateway/internal/proxy"
	"security-gateway/internal/service"
	"strconv"
)

var ServiceFieldController = &serviceFieldController{}

type serviceFieldController struct {
}

func (c *serviceFieldController) Add(ctx *fiber.Ctx) error {
	instance := new(model.ServiceField)
	if err := ctx.BodyParser(instance); err != nil {
		return ctx.JSON(&CommonResponse{
			Code: ResponseCodeParamParseError,
			Msg:  ResponseMsgParamParseError,
		})
	}

	duplicated, success, err := service.ServiceFieldService.Add(instance)
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

func (c *serviceFieldController) Update(ctx *fiber.Ctx) error {
	instance := new(model.ServiceField)
	if err := ctx.BodyParser(instance); err != nil {
		return ctx.JSON(&CommonResponse{
			Code: ResponseCodeParamParseError,
			Msg:  ResponseMsgParamParseError,
		})
	}

	duplicated, success, err := service.ServiceFieldService.Update(instance)
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

func (c *serviceFieldController) Delete(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	if idStr == "" {
		return ctx.JSON(&CommonResponse{
			Code: ResponseCodeParamNotEnough,
			Msg:  ResponseMsgParamNotEnough + " id",
		})
	}

	id, err := strconv.ParseUint(idStr, 10, 64)

	oldInstance, err := service.ServiceFieldService.Get(id)
	if err != nil {
		return ctx.JSON(&CommonResponse{
			Code: ResponseCodeDatabase,
			Msg:  ResponseMsgDatabase + err.Error(),
		})

	}

	success, err := service.ServiceFieldService.Delete(id)
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

	go c.fieldDeleted(oldInstance)

	return ctx.JSON(&CommonResponse{
		Data: success,
	})
}

func (c *serviceFieldController) Get(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	if idStr == "" {
		return ctx.JSON(&CommonResponse{
			Code: ResponseCodeParamNotEnough,
			Msg:  ResponseMsgParamNotEnough + " id",
		})
	}

	id, err := strconv.ParseUint(idStr, 10, 64)

	instance, err := service.ServiceFieldService.Get(id)
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

func (c *serviceFieldController) List(ctx *fiber.Ctx) error {
	pageStr := ctx.Query("page")
	pageSizeStr := ctx.Query("pageSize")
	condition := new(model.ServiceField)
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

	instances, total, err := service.ServiceFieldService.List(page, pageSize, condition)
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

func (c *serviceFieldController) fieldDeleted(field *model.ServiceField) {
	if field == nil {
		return
	}
	// 获取服务
	serv, err := service.ServiceService.Get(field.ServiceID)
	if err != nil {
		return
	}
	if serv == nil {
		return
	}

	proxy.Manager.RemoveServiceField(*serv.Port, *serv.Domain, field.FieldName)
}

func (c *serviceFieldController) fieldUpdated(id uint64) {
	// 获取字段信息
	instance, err := service.ServiceFieldService.Get(id)
	if err != nil || instance == nil {
		return
	}
	// 获取服务
	serv, err := service.ServiceService.Get(instance.ServiceID)
	if err != nil || serv == nil {
		return
	}

	proxy.Manager.UpdateServiceField(serv, instance)
}

func (c *serviceFieldController) fieldAdded(instance *model.ServiceField) {
	// 获取服务
	serv, err := service.ServiceService.Get(instance.ServiceID)
	if err != nil || serv == nil {
		return
	}

	proxy.Manager.UpdateServiceField(serv, instance)
}
