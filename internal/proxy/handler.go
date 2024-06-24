package proxy

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/proxy"
	logger "github.com/sirupsen/logrus"
	"math/rand"
	"security-gateway/internal/model"
	"security-gateway/pkg/server"
	"security-gateway/pkg/util"
	"strings"
)

func (m *manager) generateHandler(routeProxy *RouteProxy, route *model.Route, port uint16, domain string) func(c *fiber.Ctx) error {
	handler := func(c *fiber.Ctx) error {
		if len(routeProxy.TargetUpstreams) == 0 {
			return fiber.ErrNotFound
		}
		customIp := c.IP()
		// 反向代理
		loadBalanceType := route.LoadBalance
		var realTargetUrl string
		switch loadBalanceType {
		case model.LoadBalanceRoundRobin:
			// 轮询
			realTargetUrl = routeProxy.TargetUpstreams[routeProxy.nextIndex].TargetUrl
			routeProxy.nextIndex = (routeProxy.nextIndex + 1) % len(routeProxy.TargetUpstreams)
		case model.LoadBalanceWeight:
			// 权重
			if routeProxy.WeightTotal == 0 {
				return fiber.ErrNotFound
			}
			weight := rand.Intn(routeProxy.WeightTotal)
			for _, tu := range routeProxy.TargetUpstreams {
				weight -= tu.Weight
				if weight <= 0 {
					realTargetUrl = tu.TargetUrl
					break
				}
			}
		case model.LoadBalanceIPHash:
			// IP哈希
			ipHash := util.IpHash(customIp)
			index := ipHash % len(routeProxy.TargetUpstreams)
			realTargetUrl = routeProxy.TargetUpstreams[index].TargetUrl
		}

		c.Request().Header.Add("X-Real-IP", c.IP())
		originalURL := c.OriginalURL()
		if !strings.HasSuffix(realTargetUrl, "/") && !strings.HasPrefix(originalURL, "/") {
			realTargetUrl += "/"
		} else if strings.HasSuffix(realTargetUrl, "/") && strings.HasPrefix(originalURL, "/") {
			originalURL = originalURL[1:]
		}
		trueTargetUrl := fmt.Sprintf("%s%s", realTargetUrl, originalURL)
		err := proxy.Do(c, trueTargetUrl)
		if err != nil {
			logger.Error("反向代理失败: ", err)
			return err
		}

		// 直接检查header来判断返回的数据是否是json格式
		if !strings.Contains(string(c.Response().Header.ContentType())+string(c.Request().Header.Header()), "application/json") {
			// 不是json格式，不做处理
			return nil
		}

		// 脱敏处理
		fieldsInterface := c.Locals("fields")
		var fields []*server.DesensitizeField
		if fieldsInterface != nil {
			fields = fieldsInterface.([]*server.DesensitizeField)
		}
		maskingLevel, username := m.modifyResponse(c.Request(), c.Response(), port, domain, fields)
		// 记录日志
		logger.WithFields(logger.Fields{
			"domain":       domain,
			"port":         port,
			"path":         c.Path(),
			"customIp":     customIp,
			"maskingLevel": maskingLevel,
			"username":     username,
		}).Info("requesting record")

		return nil
	}
	return handler
}
