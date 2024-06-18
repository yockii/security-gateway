package proxy

import (
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/proxy"
	logger "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"github.com/valyala/fasthttp"
	"security-gateway/internal/model"
	"security-gateway/internal/service"
	"security-gateway/pkg/server"
	"security-gateway/pkg/util"
	"strings"
	"sync"
)

// 反向代理管理器
// 管理所有配置的服务与反向代理的关系，并可以动态修改
type manager struct {
	// 端口 -> 域名 -> (路由 -> 目标)
	portToRoutes map[uint16]map[string][]*RouteProxy
	portToRouter map[uint16]map[string]*server.Router
	portToServer map[uint16]*fiber.App

	// 服务的用户信息接口
	domainToUserRoute map[uint16]map[string]*model.UserInfoRoute
	// 对应服务的token和密级关系 port -> domain -> token -> secret -> level
	serviceTokenToSecret map[uint16]map[string]map[string]int
	// 服务对应的要处理的字段
	serviceToField map[uint16]map[string][]*model.SecretField
}

type RouteProxy struct {
	Path      string // 路由路径
	TargetUrl string // 目标URL
	Weight    int    // 权重
}

var Manager = &manager{
	portToRoutes:         make(map[uint16]map[string][]*RouteProxy),
	portToRouter:         make(map[uint16]map[string]*server.Router),
	portToServer:         make(map[uint16]*fiber.App),
	domainToUserRoute:    make(map[uint16]map[string]*model.UserInfoRoute),
	serviceTokenToSecret: make(map[uint16]map[string]map[string]int),
	serviceToField:       make(map[uint16]map[string][]*model.SecretField),
}

func (m *manager) GetUsedPorts() (ports []uint16) {
	for port := range m.portToServer {
		ports = append(ports, port)
	}
	return
}

func (m *manager) AddField(serv *model.Service, field *model.SecretField) {
	port := *serv.Port
	domain := *serv.Domain
	if _, ok := m.serviceToField[port]; !ok {
		m.serviceToField[port] = make(map[string][]*model.SecretField)
	}
	if _, ok := m.serviceToField[port][domain]; !ok {
		m.serviceToField[port][domain] = make([]*model.SecretField, 0)
	}
	m.serviceToField[port][domain] = append(m.serviceToField[port][domain], field)
}

func (m *manager) RemoveField(port uint16, domain, fieldName string) {
	if fields, ok := m.serviceToField[port][domain]; ok {
		for i, field := range fields {
			if field.FieldName == fieldName {
				m.serviceToField[port][domain] = append(fields[:i], fields[i+1:]...)
				break
			}
		}
	}
}

func (m *manager) UpdateField(serv *model.Service, field *model.SecretField) {
	port := *serv.Port
	domain := *serv.Domain
	if fields, ok := m.serviceToField[port][domain]; ok {
		for i, f := range fields {
			if f.FieldName == field.FieldName {
				m.serviceToField[port][domain][i] = field
				break
			}
		}
	}
}

func (m *manager) AddUserRoute(serv *model.Service, uir *model.UserInfoRoute) {
	port := *serv.Port
	domain := *serv.Domain

	if _, ok := m.domainToUserRoute[port]; !ok {
		m.domainToUserRoute[port] = make(map[string]*model.UserInfoRoute)
	}
	if _, ok := m.domainToUserRoute[port][domain]; !ok {
		m.domainToUserRoute[port][domain] = uir
	}
}

func (m *manager) RemoveUserRoute(port uint16, domain string) {
	if _, ok := m.domainToUserRoute[port]; ok {
		delete(m.domainToUserRoute[port], domain)
	}
}

func (m *manager) UpdateUserRoute(serv *model.Service, uir *model.UserInfoRoute) {
	port := *serv.Port
	domain := *serv.Domain
	if _, ok := m.domainToUserRoute[port]; ok {
		m.domainToUserRoute[port][domain] = uir
	}
}

func (m *manager) AddRoute(serv *model.Service, route *model.Route, upstream *model.Upstream, rt *model.RouteTarget) {
	port := *serv.Port
	domain := *serv.Domain
	path := *route.Uri
	targetUrl := *upstream.TargetUrl

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
		Weight:    rt.Weight,
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

		// 脱敏处理
		m.modifyResponse(c.Request(), c.Response(), port, domain)
		// TODO 记录日志

		return nil
	})
}

func (m *manager) modifyResponse(req *fasthttp.Request, resp *fasthttp.Response, port uint16, domain string) {
	resp.Header.Del(fiber.HeaderServer)
	// 确保本方法不会panic
	defer func() {
		if r := recover(); r != nil {
			logger.Error("modifyResponse panic: ", r)
		}
	}()

	body := string(resp.Body())
	if !gjson.Valid(body) {
		// 不是json格式，不做处理
		return
	}
	bodyJson := gjson.Parse(body)

	token := ""
	if d, has := m.domainToUserRoute[port]; has {
		if uir, ok := d[domain]; ok {
			// 获取token
			tokenPosition := strings.Split(uir.TokenPosition, ":")
			if len(tokenPosition) < 3 {
				// token获取条件不满足
				return
			}
			w := tokenPosition[1]
			p := tokenPosition[2]
			switch tokenPosition[0] {
			case "request":
				switch w {
				case "header":
					token = string(req.Header.Peek(p))
				case "query":
					token = string(req.URI().QueryArgs().Peek(p))
				case "body":
					// 判断req是否是form表单提交
					if strings.Contains(string(req.Header.Peek(fiber.HeaderContentType)), "application/x-www-form-urlencoded") {
						token = string(req.PostArgs().Peek(p))
					} else {
						token = gjson.ParseBytes(req.Body()).Get(p).String()
					}
				case "cookies":
					token = string(req.Header.Cookie(p))
				}
			case "response":
				switch w {
				case "body":
					token = bodyJson.Get(p).String()
				}
			}
			// token为空，则所有密级都为1

			// 判断是否是用户信息路由
			if uir.Path == string(req.URI().Path()) {
				// 获取用户信息，并缓存token和密级关系

				// 2、获取用户信息
				uniKey := bodyJson.Get(uir.UniKeyPath).String()
				if uniKey == "" {
					// uniKey获取失败
					return
				}

				// 用户名
				username := bodyJson.Get(uir.UsernamePath).String()

				// 3、根据uniKey存储位置查找user
				var user *model.User
				if uir.MatchKey == "-" {
					user = service.UserService.GetByUniKey(username, uniKey)
				} else {
					matchKey := bodyJson.Get(uir.MatchKey).String()
					if matchKey == "" {
						// matchKey获取失败
						return
					}
					user = service.UserService.GetByUniKeyJson(username, uniKey, matchKey)
				}
				if user == nil {
					// 用户信息获取失败
					if username != "" {
						// 保存用户信息
						user = &model.User{
							Username: username,
							UniKey:   uniKey,
						}
						if _, _, err := service.UserService.Add(user); err != nil {
							logger.Error("保存用户信息失败: ", err)
							return
						}
					}
					return
				}

				secLevel := user.SecLevel
				// 还要看user在服务下的密级，如果存在，则以此为准
				{
					// 获取userServiceLevel
					usl := service.UserServiceLevelService.GetByUserAndServiceID(user.ID, uir.ServiceID)
					if usl != nil {
						secLevel = usl.SecLevel
					}
				}

				// 4、保存token和密级关系
				if _, ok = m.serviceTokenToSecret[port]; !ok {
					m.serviceTokenToSecret[port] = make(map[string]map[string]int)
				}
				if _, ok = m.serviceTokenToSecret[port][domain]; !ok {
					m.serviceTokenToSecret[port][domain] = make(map[string]int)
				}
				m.serviceTokenToSecret[port][domain][token] = secLevel
				return
			}
		}
	}

	// 其他路由，根据token获取密级
	var secLevel = 1
	if dm, okp := m.serviceTokenToSecret[port]; okp {
		if d, okd := dm[domain]; okd {
			if s, okt := d[token]; okt {
				secLevel = s
			}
		}
	}

	// 获取所有该服务对应的字段
	var fields []*model.SecretField
	if fs, okp := m.serviceToField[port]; okp {
		if f, okd := fs[domain]; okd {
			fields = f
		}
	}
	if len(fields) > 0 {
		// 对字段进行脱敏处理
		modifiedBody := m.modifyFields(bodyJson, fields, secLevel)
		resp.SetBody([]byte(modifiedBody))
	}
}

func (m *manager) RemoveRoute(port uint16, domain, path string) {
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
			delete(m.domainToUserRoute[port], domain)
			delete(m.portToRoutes[port], domain)
		}
	}
	// 检查，如果该端口下没有任何路由了，关闭服务
	if len(m.portToRoutes[port]) == 0 {
		if app, ok := m.portToServer[port]; ok {
			_ = app.Shutdown()
			delete(m.portToServer, port)
			delete(m.portToRoutes, port)
			delete(m.portToRouter, port)
			delete(m.domainToUserRoute, port)
		}
	}
}

func (m *manager) initFiberAppHandler(app *fiber.App, port uint16) {
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

func (m *manager) modifyFields(bodyJson gjson.Result, fields []*model.SecretField, level int) (modifiedBody string) {
	modifiedBody = bodyJson.Raw
	if bodyJson.IsArray() {
		var modifiedArray []interface{}
		// 如果是数组，遍历每个元素
		for _, element := range bodyJson.Array() {
			modifiedArray = append(modifiedArray, m.modifyFields(element, fields, level))
		}
		modifiedBodyBytes, err := json.Marshal(modifiedArray)
		if err != nil {
			logger.Error("json.Marshal failed: ", err)
			return
		}
		modifiedBody = string(modifiedBodyBytes)
		return
	}

	var bodyMap map[string]interface{}
	if err := json.Unmarshal([]byte(bodyJson.Raw), &bodyMap); err != nil {
		logger.Error("json.Unmarshal failed: ", err)
		return
	}

	m.doModifyFields(bodyJson, bodyMap, fields, level)
	modifiedBodyBytes, err := json.Marshal(bodyMap)
	if err != nil {
		logger.Error("json.Marshal failed: ", err)
		return
	}
	modifiedBody = string(modifiedBodyBytes)
	return
}

func (m *manager) doModifyFields(bodyJson gjson.Result, bodyMap map[string]interface{}, fields []*model.SecretField, level int) {
	// 遍历所有字段，每个字段并发独立处理
	var wg sync.WaitGroup
	for _, field := range fields {
		fieldName := field.FieldName
		maskPattern := field.Level4
		switch level {
		case 1:
			maskPattern = field.Level1
		case 2:
			maskPattern = field.Level2
		case 3:
			maskPattern = field.Level3
		case 4:
			maskPattern = field.Level4
		}
		wg.Add(1)
		go func(bodyJson gjson.Result, bodyMap map[string]interface{}, field *model.SecretField, level int) {
			defer wg.Done()
			m.doModifyField(bodyJson, bodyMap, fieldName, maskPattern)
		}(bodyJson, bodyMap, field, level)
	}
	wg.Wait()
	return
}

// doModifyField 对字段进行脱敏处理, 确保j和m是同级别的
func (m *manager) doModifyField(j gjson.Result, obj map[string]interface{}, fieldName string, maskPattern string) {
	if maskPattern == "-" || maskPattern == "" {
		return
	}

	if _, ok := obj[fieldName]; ok {
		val := j.Get(fieldName).String()
		modifiedValue, err := util.AdvanceMask(val, maskPattern)
		if err != nil {
			logger.Error("AdvanceMask failed: ", err)
			return
		}
		obj[fieldName] = modifiedValue
	}

	// 找到嵌套的字段进行递归处理
	for k, v := range obj {
		if k == fieldName {
			continue
		}
		jk := j.Get(k)
		if jk.IsArray() {
			// 如果是数组，那么v也必定为数组
			jArr := jk.Array()
			vArr, ok := v.([]interface{})
			if !ok {
				continue // 不是数组，不做处理
			}
			for i, element := range vArr {
				eleMap, ok := element.(map[string]interface{})
				if !ok {
					continue // 不是map，不做处理
				}
				m.doModifyField(jArr[i], eleMap, fieldName, maskPattern)
			}
		} else if jk.IsObject() {
			// 如果是对象，那么v也必定为对象
			vMap, ok := v.(map[string]interface{})
			if !ok {
				continue // 不是map，不做处理
			}
			m.doModifyField(jk, vMap, fieldName, maskPattern)
		}
		// 其他情况不做处理
	}
}
