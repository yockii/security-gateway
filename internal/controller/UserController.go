package controller

import (
	"github.com/gofiber/fiber/v2"
	logger "github.com/sirupsen/logrus"
	"security-gateway/internal/domain"
	"security-gateway/internal/model"
	"security-gateway/internal/proxy"
	"security-gateway/internal/service"
	"strconv"
)

var UserController = &userController{}

type userController struct {
}

func (c *userController) Add(ctx *fiber.Ctx) error {
	instance := new(model.User)
	if err := ctx.BodyParser(instance); err != nil {
		return ctx.JSON(&CommonResponse{
			Code: ResponseCodeParamParseError,
			Msg:  ResponseMsgParamParseError,
		})
	}

	duplicated, success, err := service.UserService.Add(instance)
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

func (c *userController) Update(ctx *fiber.Ctx) error {
	instance := new(model.User)
	if err := ctx.BodyParser(instance); err != nil {
		return ctx.JSON(&CommonResponse{
			Code: ResponseCodeParamParseError,
			Msg:  ResponseMsgParamParseError,
		})
	}

	oldInstance, err := service.UserService.Get(instance.ID)
	if err != nil {
		return ctx.JSON(&CommonResponse{
			Code: ResponseCodeDatabase,
			Msg:  ResponseMsgDatabase + err.Error(),
		})
	}

	duplicated, success, err := service.UserService.Update(instance)
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

	if instance.SecLevel != 0 && oldInstance.SecLevel != instance.SecLevel {
		go c.updateUserLevel(instance.ID)
	}

	return ctx.JSON(&CommonResponse{
		Data: instance,
	})
}

func (c *userController) Delete(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	if idStr == "" {
		return ctx.JSON(&CommonResponse{
			Code: ResponseCodeParamNotEnough,
			Msg:  ResponseMsgParamNotEnough + " id",
		})
	}

	id, err := strconv.ParseUint(idStr, 10, 64)

	success, err := service.UserService.Delete(id)
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
	return ctx.JSON(&CommonResponse{
		Data: success,
	})
}

func (c *userController) Get(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	if idStr == "" {
		return ctx.JSON(&CommonResponse{
			Code: ResponseCodeParamNotEnough,
			Msg:  ResponseMsgParamNotEnough + " id",
		})
	}

	id, err := strconv.ParseUint(idStr, 10, 64)

	instance, err := service.UserService.Get(id)
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

func (c *userController) List(ctx *fiber.Ctx) error {
	pageStr := ctx.Query("page")
	pageSizeStr := ctx.Query("pageSize")
	condition := new(model.User)
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

	instances, total, err := service.UserService.List(page, pageSize, condition)
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

func (c *userController) updateUserLevel(id uint64) {
	// 获取用户密级
	user, err := service.UserService.Get(id)
	if err != nil {
		logger.Error(err)
		return
	}
	if user == nil {
		return
	}

	// 更新用户所有密级为该用户密级
	proxy.Manager.UpdateUserAllSecretLevel(user.Username, user.SecLevel)

	// 查找用户独立的服务密级，再次更新
	page := 1
	serviceLevels, total, err := service.UserServiceLevelService.ListWithService(page, 1000, &model.UserServiceLevel{UserID: id})
	if err != nil {
		logger.Error(err)
		return
	}
	for len(serviceLevels) < int(total) {
		page++
		var sl []*domain.UserLevelWithService
		sl, total, err = service.UserServiceLevelService.ListWithService(page, 1000, &model.UserServiceLevel{UserID: id})
		if err != nil {
			logger.Error(err)
			return
		}
		serviceLevels = append(serviceLevels, sl...)
	}

	// 更新服务密级
	for _, serviceLevel := range serviceLevels {
		serv := serviceLevel.Service
		proxy.Manager.UpdateServiceSecretLevel(*serv.Port, *serv.Domain, user.Username, serviceLevel.SecLevel)
	}
}
