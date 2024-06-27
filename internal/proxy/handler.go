package proxy

import (
	"bytes"
	"context"
	"fmt"
	logger "github.com/sirupsen/logrus"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"io"
	"math/rand"
	"net"
	"net/http"
	"security-gateway/internal/model"
	"security-gateway/pkg/server"
	"security-gateway/pkg/util"
	"strings"
)

//func (m *manager) generateHandler(routeProxy *RouteProxy, route *model.Route, port uint16, domain string) func(c *fiber.Ctx) error {
//	handler := func(c *fiber.Ctx) error {
//		if len(routeProxy.TargetUpstreams) == 0 {
//			return fiber.ErrNotFound
//		}
//		customIp := c.IP()
//		// 反向代理
//		loadBalanceType := route.LoadBalance
//		var realTargetUrl string
//		switch loadBalanceType {
//		case model.LoadBalanceRoundRobin:
//			// 轮询
//			// 先确保nextIndex不会越界
//			if routeProxy.nextIndex >= len(routeProxy.TargetUpstreams) {
//				routeProxy.nextIndex = 0
//			}
//			realTargetUrl = routeProxy.TargetUpstreams[routeProxy.nextIndex].TargetUrl
//			routeProxy.nextIndex = (routeProxy.nextIndex + 1) % len(routeProxy.TargetUpstreams)
//		case model.LoadBalanceWeight:
//			// 权重
//			if routeProxy.WeightTotal == 0 {
//				return fiber.ErrNotFound
//			}
//			weight := rand.Intn(routeProxy.WeightTotal)
//			for _, tu := range routeProxy.TargetUpstreams {
//				weight -= tu.Weight
//				if weight <= 0 {
//					realTargetUrl = tu.TargetUrl
//					break
//				}
//			}
//		case model.LoadBalanceIPHash:
//			// IP哈希
//			ipHash := util.IpHash(customIp)
//			index := ipHash % len(routeProxy.TargetUpstreams)
//			realTargetUrl = routeProxy.TargetUpstreams[index].TargetUrl
//		}
//
//		c.Request().Header.Add("X-Real-IP", c.IP())
//		originalURL := c.OriginalURL()
//
//		// 去除route中的uri
//		if route.Uri != nil {
//			originalURL = strings.Replace(originalURL, *(route.Uri), "", 1)
//		}
//
//		if !strings.HasSuffix(realTargetUrl, "/") && !strings.HasPrefix(originalURL, "/") {
//			realTargetUrl += "/"
//		} else if strings.HasSuffix(realTargetUrl, "/") && strings.HasPrefix(originalURL, "/") {
//			originalURL = originalURL[1:]
//		}
//		trueTargetUrl := fmt.Sprintf("%s%s", realTargetUrl, originalURL)
//
//		proxyId := util.GenerateXid()
//		logger.WithField("proxyId", proxyId).Debug("准备请求真实目标地址: ", trueTargetUrl)
//
//		err := proxy.DoTimeout(c, trueTargetUrl, 60*time.Second)
//		if err != nil {
//			logger.Error("反向代理失败: ", err)
//			return err
//		}
//
//		logger.WithField("proxyId", proxyId).Debug("真实目标地址请求成功: ", trueTargetUrl)
//
//		// 直接检查header来判断返回的数据是否是json格式
//		if !strings.Contains(string(c.Response().Header.ContentType())+string(c.Request().Header.Header()), "application/json") {
//			// 不是json格式，不做处理
//			return nil
//		}
//
//		// 脱敏处理
//		fieldsInterface := c.Locals("fields")
//		var fields []*server.DesensitizeField
//		if fieldsInterface != nil {
//			fields = fieldsInterface.([]*server.DesensitizeField)
//		}
//		maskingLevel, username := m.modifyResponse(c.Request(), c.Response(), port, domain, fields)
//		// 记录日志
//		log.WithFields(logger.Fields{
//			"domain":       domain,
//			"port":         port,
//			"path":         c.OriginalURL(),
//			"target":       trueTargetUrl,
//			"customIp":     customIp,
//			"maskingLevel": maskingLevel,
//			"username":     username,
//		}).Info("requesting record")
//
//		return nil
//	}
//	return handler
//}

func (m *manager) generateHandler(routeProxy *RouteProxy, route *model.Route, port uint16, domain string) http.HandlerFunc {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if len(routeProxy.TargetUpstreams) == 0 {
			// 返回404
			http.NotFound(w, r)
			return
		}
		customIp := util.GetUserIP(r)
		// 反向代理
		loadBalanceType := route.LoadBalance
		var realTargetUrl string
		switch loadBalanceType {
		case model.LoadBalanceRoundRobin:
			// 轮询
			// 先确保nextIndex不会越界
			if routeProxy.nextIndex >= len(routeProxy.TargetUpstreams) {
				routeProxy.nextIndex = 0
			}
			realTargetUrl = routeProxy.TargetUpstreams[routeProxy.nextIndex].TargetUrl
			routeProxy.nextIndex = (routeProxy.nextIndex + 1) % len(routeProxy.TargetUpstreams)
		case model.LoadBalanceWeight:
			// 权重
			if routeProxy.WeightTotal == 0 {
				http.NotFound(w, r)
				return
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

		// 增加X-Real-IP头
		r.Header.Add("X-Real-IP", customIp)

		// 获取带参数的原始URL
		originalURL := r.URL.String()

		// 去除route中的uri
		if route.Uri != nil {
			originalURL = strings.Replace(originalURL, *(route.Uri), "", 1)
		}

		if !strings.HasSuffix(realTargetUrl, "/") && !strings.HasPrefix(originalURL, "/") {
			realTargetUrl += "/"
		} else if strings.HasSuffix(realTargetUrl, "/") && strings.HasPrefix(originalURL, "/") {
			originalURL = originalURL[1:]
		}
		trueTargetUrl := fmt.Sprintf("%s%s", realTargetUrl, originalURL)

		proxyId := util.GenerateXid()
		logger.WithField("proxyId", proxyId).Debug("准备请求真实目标地址: ", trueTargetUrl)

		// 反向代理
		proxy, err := m.getProxyService(realTargetUrl)
		if err != nil {
			logger.Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		mrw := &ModifiableResponseWriter{
			ResponseWriter: w,
			status:         http.StatusOK,
			body:           new(bytes.Buffer),
		}

		proxy.ServeHTTP(mrw, r)

		logger.WithField("proxyId", proxyId).Debug("真实目标地址请求成功: ", trueTargetUrl)

		// 检查返回的数据是否是json格式
		if !strings.Contains(w.Header().Get("Content-Type"), "application/json") {
			// 不是json格式，直接写入body
			_, err = io.Copy(w, mrw.body)
			if err != nil {
				logger.Error(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			return
		}

		// 脱敏处理
		fieldsInterface := r.Context().Value("fields")
		var fields []*server.DesensitizeField
		if fieldsInterface != nil {
			fields = fieldsInterface.([]*server.DesensitizeField)
		}
		maskingLevel, username := m.modifyResponse(r, mrw, port, domain, fields)
		// 记录日志
		log.WithFields(logger.Fields{
			"domain":       domain,
			"port":         port,
			"path":         r.URL.String(),
			"target":       trueTargetUrl,
			"customIp":     customIp,
			"maskingLevel": maskingLevel,
			"username":     username,
		}).Info("requesting record")

		_, err = io.Copy(w, mrw.body)
		if err != nil {
			logger.Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}
	return handler
}

//func (m *manager) initFiberAppHandler(app *fiber.App, port uint16) {
//	// 对app所有请求进行处理
//	app.Use(func(c *fiber.Ctx) error {
//		// 确保不会被异常终止
//		defer func() {
//			if r := recover(); r != nil {
//				logger.Error("port service panic: ", r)
//			}
//		}()
//
//		if allRouter, ok := m.portToRouter[port]; ok {
//			var router *server.Router
//			domainName := strings.Split(c.Hostname(), ":")[0]
//			router, ok = allRouter[domainName]
//			if !ok {
//				router = allRouter[""]
//			}
//			if router != nil {
//				route := router.FindRoute(c.Path())
//				if route != nil {
//					handler := route.Handler
//					c.Locals("fields", route.DesensitizeFields)
//					return handler(c)
//				}
//			}
//		}
//		return fiber.ErrNotFound
//	})
//}

func (m *manager) initFiberAppHandler(ps *http.Server, port uint16) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 确保不会被异常终止
		defer func() {
			if e := recover(); e != nil {
				logger.Error("port service panic: ", e)
			}
		}()

		if allRouter, ok := m.portToRouter[port]; ok {
			var router *server.Router
			domainName := strings.Split(r.Host, ":")[0]
			router, ok = allRouter[domainName]
			if !ok {
				router = allRouter[""]
			}
			if router != nil {
				route := router.FindRoute(r.URL.Path)
				if route != nil {
					handler := route.Handler
					r = r.WithContext(context.WithValue(r.Context(), "fields", route.DesensitizeFields))
					handler(w, r)
					return
				}
			}
		}
		http.NotFound(w, r)
	})
	ps.Handler = h2c.NewHandler(handler, &http2.Server{})
}

//func (m *manager) handleProxyServer(port uint16) error {
//	app := fiber.New(fiber.Config{
//		DisableStartupMessage: true,
//	})
//	err := m.listenAndServeApp(port, app)
//	if err != nil {
//		return err
//	}
//
//	m.portToServer[port] = app
//	return nil
//}

func (m *manager) handleProxyServer(port uint16) error {
	app := &http.Server{}
	err := m.listenAndServeApp(port, app)
	if err != nil {
		return err
	}

	m.portToServer[port] = app
	return nil
}

//func (m *manager) listenAndServeApp(port uint16, app *fiber.App) error {
//	m.initFiberAppHandler(app, port)
//	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
//
//	if err != nil {
//		logger.Error(err)
//		return err
//	}
//
//	appLn := NewAppListener(ln, m.certManager.generateDynamicTLSConfig(port))
//
//	//ln, err := gmtls.Listen("tcp", fmt.Sprintf(":%d", port), m.certManager.generateDynamicTLSConfig(port))
//	//if err != nil {
//	//	logger.Error(err)
//	//	return err
//	//}
//
//	go func() {
//		e := app.Listener(appLn)
//		//err := app.Listen(fmt.Sprintf(":%d", port))
//		if e != nil {
//			logger.Error("服务异常停止: ", e)
//			delete(m.portToServer, port)
//
//			// 重新启动服务
//			e = m.handleProxyServer(port)
//			if e != nil {
//				logger.Error("重新启动服务失败: ", e)
//			}
//			return
//		}
//	}()
//	//err := app.Listener(ln)
//	//if err != nil {
//	//	logger.Error("启动服务失败: ", err)
//	//	delete(m.portToServer, port)
//	//	return
//	//}
//
//	return nil
//}

func (m *manager) listenAndServeApp(port uint16, ps *http.Server) error {
	m.initFiberAppHandler(ps, port)
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))

	if err != nil {
		logger.Error(err)
		return err
	}

	appLn := NewAppListener(ln, m.certManager.generateDynamicTLSConfig(port))

	go func() {
		e := ps.Serve(appLn)
		//err := app.Listen(fmt.Sprintf(":%d", port))
		if e != nil {
			logger.Error("服务异常停止: ", e)
			delete(m.portToServer, port)

			// 重新启动服务
			e = m.handleProxyServer(port)
			if e != nil {
				logger.Error("重新启动服务失败: ", e)
			}
			return
		}
	}()
	//err := app.Listener(ln)
	//if err != nil {
	//	logger.Error("启动服务失败: ", err)
	//	delete(m.portToServer, port)
	//	return
	//}

	return nil
}
