package controller

import (
	"crypto/tls"
	"github.com/gofiber/fiber/v2"
	logger "github.com/sirupsen/logrus"
	"github.com/tjfoc/gmsm/gmtls"
	"security-gateway/internal/model"
	"security-gateway/internal/proxy"
	"security-gateway/internal/service"
	"strconv"
)

var CertificateController = &certificateController{}

type certificateController struct {
}

func (c *certificateController) Add(ctx *fiber.Ctx) error {
	instance := new(model.Certificate)
	if err := ctx.BodyParser(instance); err != nil {
		return ctx.JSON(&CommonResponse{
			Code: ResponseCodeParamParseError,
			Msg:  ResponseMsgParamParseError,
		})
	}

	// 检查证书是否有效
	if instance.CertPem != "" || instance.KeyPem != "" {
		pass := false
		if instance.KeyPem != "" {
			_, err := tls.X509KeyPair([]byte(instance.CertPem), []byte(instance.KeyPem))
			pass = err == nil
		}
		if !pass {
			return ctx.JSON(&CommonResponse{
				Code: ResponseCodeParamParseError,
				Msg:  ResponseMsgParamParseError + " RSA证书无效",
			})
		}
	}
	if instance.SignKeyPem != "" || instance.SignCertPem != "" || instance.EncCertPem != "" || instance.EncKeyPem != "" {
		pass := false
		if instance.SignCertPem != "" && instance.EncCertPem != "" && instance.EncKeyPem != "" {
			_, err := gmtls.X509KeyPair([]byte(instance.SignCertPem), []byte(instance.SignKeyPem))
			pass = err == nil
			if pass {
				_, err = gmtls.X509KeyPair([]byte(instance.EncCertPem), []byte(instance.EncKeyPem))
				pass = err == nil
			}
		}
		if !pass {
			return ctx.JSON(&CommonResponse{
				Code: ResponseCodeParamParseError,
				Msg:  ResponseMsgParamParseError + " SM证书无效",
			})
		}
	}

	duplicated, success, err := service.CertificateService.Add(instance)
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

func (c *certificateController) Update(ctx *fiber.Ctx) error {
	instance := new(model.Certificate)
	if err := ctx.BodyParser(instance); err != nil {
		return ctx.JSON(&CommonResponse{
			Code: ResponseCodeParamParseError,
			Msg:  ResponseMsgParamParseError,
		})
	}

	// 检查证书是否有效
	if instance.CertPem != "" || instance.KeyPem != "" {
		pass := false
		if instance.KeyPem != "" {
			_, err := tls.X509KeyPair([]byte(instance.CertPem), []byte(instance.KeyPem))
			pass = err == nil
		}
		if !pass {
			return ctx.JSON(&CommonResponse{
				Code: ResponseCodeParamParseError,
				Msg:  ResponseMsgParamParseError + " RSA证书无效",
			})
		}
	}
	if instance.SignKeyPem != "" || instance.SignCertPem != "" || instance.EncCertPem != "" || instance.EncKeyPem != "" {
		pass := false
		if instance.SignCertPem != "" && instance.EncCertPem != "" && instance.EncKeyPem != "" {
			_, err := gmtls.X509KeyPair([]byte(instance.SignCertPem), []byte(instance.SignKeyPem))
			pass = err == nil
			if pass {
				_, err = gmtls.X509KeyPair([]byte(instance.EncCertPem), []byte(instance.EncKeyPem))
				pass = err == nil
			}
		}
		if !pass {
			return ctx.JSON(&CommonResponse{
				Code: ResponseCodeParamParseError,
				Msg:  ResponseMsgParamParseError + " SM证书无效",
			})
		}
	}

	duplicated, success, err := service.CertificateService.Update(instance)
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

	go c.certificateUpdated(instance.ID)

	return ctx.JSON(&CommonResponse{
		Data: instance,
	})
}

func (c *certificateController) Delete(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	if idStr == "" {
		return ctx.JSON(&CommonResponse{
			Code: ResponseCodeParamNotEnough,
			Msg:  ResponseMsgParamNotEnough + " id",
		})
	}

	id, err := strconv.ParseUint(idStr, 10, 64)

	success, err := service.CertificateService.Delete(id)
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

func (c *certificateController) Get(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	if idStr == "" {
		return ctx.JSON(&CommonResponse{
			Code: ResponseCodeParamNotEnough,
			Msg:  ResponseMsgParamNotEnough + " id",
		})
	}

	id, err := strconv.ParseUint(idStr, 10, 64)

	instance, err := service.CertificateService.Get(id)
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

func (c *certificateController) List(ctx *fiber.Ctx) error {
	pageStr := ctx.Query("page")
	pageSizeStr := ctx.Query("pageSize")
	condition := new(model.Certificate)
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

	instances, total, err := service.CertificateService.List(page, pageSize, condition)
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

func (c *certificateController) AddServiceCertificate(ctx *fiber.Ctx) error {
	instance := new(model.ServiceCertificate)
	if err := ctx.BodyParser(instance); err != nil {
		return ctx.JSON(&CommonResponse{
			Code: ResponseCodeParamParseError,
			Msg:  ResponseMsgParamParseError,
		})
	}

	duplicated, success, err := service.CertificateService.AddServiceCertificate(instance)
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

func (c *certificateController) DeleteServiceCertificate(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	if idStr == "" {
		return ctx.JSON(&CommonResponse{
			Code: ResponseCodeParamNotEnough,
			Msg:  ResponseMsgParamNotEnough + " id",
		})
	}

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return ctx.JSON(&CommonResponse{
			Code: ResponseCodeParamParseError,
			Msg:  ResponseMsgParamParseError,
		})
	}

	var success bool
	success, err = service.CertificateService.DeleteServiceCertificate(id)
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

func (c *certificateController) ListByDomain(ctx *fiber.Ctx) error {
	domain := ctx.Query("domain")
	if domain == "" {
		return ctx.JSON(&CommonResponse{
			Code: ResponseCodeParamNotEnough,
			Msg:  ResponseMsgParamNotEnough + " domain",
		})
	}

	instances, err := service.CertificateService.ListByDomain(domain)
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

func (c *certificateController) certificateUpdated(id uint64) {
	if id == 0 {
		return
	}
	// 找到所有服务
	services, _, err := service.ServiceService.List(1, 100, &model.Service{CertificateID: &id})
	if err != nil {
		logger.Error(err)
		return
	}
	for _, serv := range services {
		proxy.Manager.UpdateServiceCertificate(serv.ID)
	}
}
