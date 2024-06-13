package proxy

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/proxy"
	logger "github.com/sirupsen/logrus"
	"security-gateway/pkg/server"
	"strings"
)

// 反向代理管理器
// 管理所有配置的服务与反向代理的关系，并可以动态修改
type manager struct {
	// 端口 -> 域名 -> (路由 -> 目标)
	portToRoutes map[int]map[string][]*RouteProxy
	portToRouter map[int]map[string]*server.Router
	portToServer map[int]*fiber.App
}

type RouteProxy struct {
	Path      string // 路由路径
	TargetUrl string // 目标URL
	Weight    int    // 权重
}

var Manager = &manager{
	portToRoutes: make(map[int]map[string][]*RouteProxy),
	portToRouter: make(map[int]map[string]*server.Router),
	portToServer: make(map[int]*fiber.App),
}

func (m *manager) AddRoute(port int, domain, path, targetUrl string, weight int) {
	if _, ok := m.portToRoutes[port]; !ok {
		m.portToRoutes[port] = make(map[string][]*RouteProxy)
	}
	if _, ok := m.portToRouter[port]; !ok {
		m.portToRouter[port] = make(map[string]*server.Router)
	}
	if _, ok := m.portToServer[port]; !ok {
		app := fiber.New(fiber.Config{
			DisableStartupMessage: true,
		})
		m.portToServer[port] = app
		// 建立一个channel，用于监听服务启动是否成功
		go func() {
			m.initFiberAppHandler(app, port)
			err := app.Listen(fmt.Sprintf(":%d", port))
			if err != nil {
				logger.Error("启动服务失败: ", err)
				delete(m.portToServer, port)
				return
			}
		}()
	}

	m.portToRoutes[port][domain] = append(m.portToRoutes[port][domain], &RouteProxy{
		Path:      path,
		TargetUrl: targetUrl,
		Weight:    weight,
	})
	if _, ok := m.portToRouter[port][domain]; !ok {
		m.portToRouter[port][domain] = &server.Router{}
	}
	m.portToRouter[port][domain].AddRoute(path, func(c *fiber.Ctx) error {
		// 反向代理
		c.Request().Header.Add("X-Real-IP", c.IP())
		originalURL := c.OriginalURL()
		if !strings.HasSuffix(targetUrl, "/") && !strings.HasPrefix(originalURL, "/") {
			targetUrl += "/"
		} else if strings.HasSuffix(targetUrl, "/") && strings.HasPrefix(originalURL, "/") {
			originalURL = originalURL[1:]
		}
		realTargetUrl := fmt.Sprintf("%s%s", targetUrl, originalURL)
		err := proxy.Do(c, realTargetUrl)
		if err != nil {
			logger.Error("反向代理失败: ", err)
			return err
		}
		resp := c.Response()
		resp.Header.Del(fiber.HeaderServer)
		// TODO 记录日志，脱敏等操作
		str := string(resp.Body())
		rt := strings.ReplaceAll(str, "a", "1")
		resp.SetBody([]byte(rt))
		// TODO 以上仅为示例

		return nil
	})
}

func (m *manager) RemoveRoute(port int, domain, path string) {
	if routes, ok := m.portToRoutes[port][domain]; ok {
		for i, route := range routes {
			if route.Path == path {
				m.portToRoutes[port][domain] = append(routes[:i], routes[i+1:]...)
				break
			}
		}
	}
	if router, ok := m.portToRouter[port][domain]; ok {
		if router.RemoveRoute(path) {
			// 返回true表示router下的路由全部移除完毕，将router从map中删除
			delete(m.portToRouter[port], domain)
		}
	}
	// 检查，如果该端口下没有任何路由了，关闭服务
	if len(m.portToRoutes[port]) == 0 {
		if app, ok := m.portToServer[port]; ok {
			_ = app.Shutdown()
			delete(m.portToServer, port)
			delete(m.portToRoutes, port)
			delete(m.portToRouter, port)
		}
	}
}

func (m *manager) initFiberAppHandler(app *fiber.App, port int) {
	// 对app所有请求进行处理
	app.Use(func(c *fiber.Ctx) error {
		if allRouter, ok := m.portToRouter[port]; ok {
			var router *server.Router
			router, ok = allRouter[c.Hostname()]
			if !ok {
				router = allRouter[""]
			}
			if router != nil {
				route := router.FindRoute(c.Path())
				if route != nil {
					handler := route.Handler
					return handler(c)
				}
			}
		}
		return fiber.ErrNotFound
	})
}
