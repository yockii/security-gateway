package controller

import (
	"github.com/gofiber/fiber/v2"
	"security-gateway/internal/model"
	"security-gateway/internal/proxy"
	"security-gateway/internal/service"
	"strconv"
)

var SecretFieldController = &secretFieldController{}

type secretFieldController struct {
}

func (c *secretFieldController) Add(ctx *fiber.Ctx) error {
	instance := new(model.SecretField)
	if err := ctx.BodyParser(instance); err != nil {
		return ctx.JSON(&CommonResponse{
			Code: ResponseCodeParamParseError,
			Msg:  ResponseMsgParamParseError,
		})
	}

	duplicated, success, err := service.SecretFieldService.Add(instance)
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

func (c *secretFieldController) Update(ctx *fiber.Ctx) error {
	instance := new(model.SecretField)
	if err := ctx.BodyParser(instance); err != nil {
		return ctx.JSON(&CommonResponse{
			Code: ResponseCodeParamParseError,
			Msg:  ResponseMsgParamParseError,
		})
	}

	duplicated, success, err := service.SecretFieldService.Update(instance)
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

func (c *secretFieldController) Delete(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	if idStr == "" {
		return ctx.JSON(&CommonResponse{
			Code: ResponseCodeParamNotEnough,
			Msg:  ResponseMsgParamNotEnough + " id",
		})
	}

	id, err := strconv.ParseUint(idStr, 10, 64)

	success, err := service.SecretFieldService.Delete(id)
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

func (c *secretFieldController) Get(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	if idStr == "" {
		return ctx.JSON(&CommonResponse{
			Code: ResponseCodeParamNotEnough,
			Msg:  ResponseMsgParamNotEnough + " id",
		})
	}

	id, err := strconv.ParseUint(idStr, 10, 64)

	instance, err := service.SecretFieldService.Get(id)
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

func (c *secretFieldController) List(ctx *fiber.Ctx) error {
	pageStr := ctx.Query("page")
	pageSizeStr := ctx.Query("pageSize")
	fieldName := ctx.Query("fieldName")

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		page = 1
	}
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		pageSize = 10
	}

	instances, total, err := service.SecretFieldService.List(page, pageSize, fieldName)
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

func (c *secretFieldController) fieldDeleted(id uint64) {
	// 获取字段信息
	field, err := service.SecretFieldService.Get(id)
	if err != nil {
		return
	}
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

	proxy.Manager.RemoveField(*serv.Port, *serv.Domain, field.FieldName)
}

func (c *secretFieldController) fieldUpdated(id uint64) {
	// 获取字段信息
	instance, err := service.SecretFieldService.Get(id)
	if err != nil || instance == nil {
		return
	}
	// 获取服务
	serv, err := service.ServiceService.Get(instance.ServiceID)
	if err != nil || serv == nil {
		return
	}

	proxy.Manager.UpdateField(serv, instance)
}

func (c *secretFieldController) fieldAdded(instance *model.SecretField) {
	// 获取服务
	serv, err := service.ServiceService.Get(instance.ServiceID)
	if err != nil || serv == nil {
		return
	}

	proxy.Manager.AddField(serv, instance)
}
