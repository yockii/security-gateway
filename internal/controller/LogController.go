package controller

import (
	"github.com/gofiber/fiber/v2"
	logger "github.com/sirupsen/logrus"
	"security-gateway/internal/domain"
	"time"
)

var LogController = &logController{}

type logController struct{}

func (c *logController) CountProxyTraceLog(ctx *fiber.Ctx) error {
	condition := new(domain.AccessLog)
	if err := ctx.QueryParser(condition); err != nil {
		return ctx.JSON(&CommonResponse{
			Code: ResponseCodeParamParseError,
			Msg:  ResponseMsgParamParseError,
		})
	}

	st := time.UnixMilli(condition.StartTime)
	et := time.UnixMilli(condition.EndTime)

	count, err := domain.CountProxyTraceLogs(st, et, condition)
	if err != nil {
		logger.Error(err)
		return ctx.JSON(&CommonResponse{
			Code: ResponseCodeUnknownError,
			Msg:  ResponseMsgUnknownError,
		})
	}

	return ctx.JSON(&CommonResponse{
		Data: count,
	})
}
