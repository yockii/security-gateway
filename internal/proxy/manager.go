package proxy

import (
	"context"
	logger "github.com/sirupsen/logrus"
	"net/http"
	"net/http/httputil"
	"security-gateway/internal/domain"
	"security-gateway/internal/model"
	"security-gateway/internal/service"
	"security-gateway/pkg/server"
)

// 反向代理管理器
// 管理所有配置的服务与反向代理的关系，并可以动态修改
type manager struct {
	initialized bool

	// 端口 -> 域名 -> (路由 -> 目标)
	portToRoutes map[uint16]map[string][]*RouteProxy
	portToRouter map[uint16]map[string]*server.Router
	//portToServer map[uint16]*fiber.App
	portToServer map[uint16]*http.Server

	// 服务的用户信息接口
	domainToUserRoute map[uint16]map[string]*model.UserInfoRoute
	// 对应服务的token和密级关系 port -> domain -> token -> secret -> level
	//serviceTokenToSecret map[uint16]map[string]map[string]int

	// 服务证书管理
	certManager *certificateManager

	// 反向代理服务
	proxyServices map[string]*httputil.ReverseProxy
}

type RouteProxy struct {
	Path string // 路由路径
	//TargetUrl string // 目标URL
	//Weight    int    // 权重
	nextIndex       int               // 下一个目标的索引
	TargetUpstreams []*TargetUpstream // 目标列表，内部负载均衡
	WeightTotal     int               // 权重总和
}

type TargetUpstream struct {
	TargetUrl string
	Weight    int
}

var Manager = &manager{
	portToRoutes: make(map[uint16]map[string][]*RouteProxy),
	portToRouter: make(map[uint16]map[string]*server.Router),
	//portToServer:      make(map[uint16]*fiber.App),
	portToServer:      make(map[uint16]*http.Server),
	domainToUserRoute: make(map[uint16]map[string]*model.UserInfoRoute),
	//serviceTokenToSecret: make(map[uint16]map[string]map[string]int),

	certManager: &certificateManager{
		certificates: make(map[uint16]map[string]*serviceCertificate),
	},

	proxyServices: make(map[string]*httputil.ReverseProxy),
}

func (m *manager) GetUsedPorts() (ports []uint16) {
	for port := range m.portToServer {
		ports = append(ports, port)
	}
	return
}

func (m *manager) UpdateRouteField(serv *model.Service, route *model.Route, field *model.RouteField) {
	port := *serv.Port
	domainName := *serv.Domain
	path := *route.Uri
	if router, ok := m.portToRouter[port][domainName]; ok {
		router.UpdateRouteField(path, &server.DesensitizeField{
			Name:                  field.FieldName,
			IsServiceField:        false,
			Level1DesensitizeRule: field.Level1,
			Level2DesensitizeRule: field.Level2,
			Level3DesensitizeRule: field.Level3,
			Level4DesensitizeRule: field.Level4,
		})
	}
}

func (m *manager) RemoveRouteField(serv *model.Service, route *model.Route, fieldName string) {
	port := *serv.Port
	domainName := *serv.Domain
	path := *route.Uri
	if router, ok := m.portToRouter[port][domainName]; ok {
		// 查找同名的服务字段
		var serviceField *server.DesensitizeField
		serviceFields, total, err := service.ServiceFieldService.List(1, 10, &model.ServiceField{ServiceID: serv.ID, FieldName: fieldName})
		if err != nil {
			logger.Error("获取服务字段失败: ", err)
		} else if total > 0 {
			serviceField = &server.DesensitizeField{
				Name:                  serviceFields[0].FieldName,
				IsServiceField:        true,
				Level1DesensitizeRule: serviceFields[0].Level1,
				Level2DesensitizeRule: serviceFields[0].Level2,
				Level3DesensitizeRule: serviceFields[0].Level3,
				Level4DesensitizeRule: serviceFields[0].Level4,
			}
		}

		router.RemoveRouteFieldWithServiceFieldUpdate(path, serviceField)
	}
}

func (m *manager) RemoveServiceField(port uint16, domain, fieldName string) {
	if router, ok := m.portToRouter[port][domain]; ok {
		router.RemoveServiceField(fieldName)
	}
}

func (m *manager) UpdateServiceField(serv *model.Service, field *model.ServiceField) {
	port := *serv.Port
	domainName := *serv.Domain
	if router, ok := m.portToRouter[port][domainName]; ok {
		router.UpdateServiceField(&server.DesensitizeField{
			Name:                  field.FieldName,
			IsServiceField:        true,
			Level1DesensitizeRule: field.Level1,
			Level2DesensitizeRule: field.Level2,
			Level3DesensitizeRule: field.Level3,
			Level4DesensitizeRule: field.Level4,
		})
	}
}

func (m *manager) AddUserRoute(serv *model.Service, uir *model.UserInfoRoute) {
	port := *serv.Port
	domainName := *serv.Domain

	if _, ok := m.domainToUserRoute[port]; !ok {
		m.domainToUserRoute[port] = make(map[string]*model.UserInfoRoute)
	}
	if _, ok := m.domainToUserRoute[port][domainName]; !ok {
		m.domainToUserRoute[port][domainName] = uir
	}
}

func (m *manager) RemoveUserRoute(port uint16, domain string) {
	if _, ok := m.domainToUserRoute[port]; ok {
		delete(m.domainToUserRoute[port], domain)
	}
}

func (m *manager) UpdateUserRoute(serv *model.Service, uir *model.UserInfoRoute) {
	port := *serv.Port
	domainName := *serv.Domain
	if _, ok := m.domainToUserRoute[port]; ok {
		m.domainToUserRoute[port][domainName] = uir
	}
}

func (m *manager) AddRoute(serv *model.Service, route *model.Route, upstream *model.Upstream, weight int) {
	port := *serv.Port
	domainName := *serv.Domain
	path := *route.Uri
	targetUrl := *upstream.TargetUrl

	if _, ok := m.portToRoutes[port]; !ok {
		m.portToRoutes[port] = make(map[string][]*RouteProxy)
	}
	if _, ok := m.portToRouter[port]; !ok {
		m.portToRouter[port] = make(map[string]*server.Router)
	}
	if _, ok := m.portToServer[port]; !ok {
		_ = m.handleProxyServer(port)
	}

	// 获取已有的路由
	var routeProxy *RouteProxy
	if routes, ok := m.portToRoutes[port][domainName]; ok {
		// 如果有，则找到已有的路由，path相同
		for _, r := range routes {
			if r.Path == path {
				routeProxy = r
				break
			}
		}
	}
	if routeProxy == nil {
		routeProxy = &RouteProxy{
			Path: path,
		}
		m.portToRoutes[port][domainName] = append(m.portToRoutes[port][domainName], routeProxy)
	}

	// 检查目标是否存在
	hasTargetUpstream := false
	for _, tu := range routeProxy.TargetUpstreams {
		if tu.TargetUrl == targetUrl {
			hasTargetUpstream = true
			break
		}
	}
	if !hasTargetUpstream {
		routeProxy.TargetUpstreams = append(routeProxy.TargetUpstreams, &TargetUpstream{
			TargetUrl: targetUrl,
			Weight:    weight,
		})
		routeProxy.WeightTotal += weight
	}

	if _, ok := m.portToRouter[port][domainName]; !ok {
		m.portToRouter[port][domainName] = &server.Router{}
	}

	handler := m.generateHandler(routeProxy, route, port, domainName)

	// 整理要脱敏的字段
	fieldMap := make(map[string]*server.DesensitizeField)
	// 1、获取服务对应的字段
	serviceFields, err := service.ServiceFieldService.GetByServiceID(serv.ID)
	if err != nil {
		logger.Error("获取服务字段失败: ", err)
	}
	for _, field := range serviceFields {
		fieldMap[field.FieldName] = &server.DesensitizeField{
			Name:                  field.FieldName,
			IsServiceField:        true,
			Level1DesensitizeRule: field.Level1,
			Level2DesensitizeRule: field.Level2,
			Level3DesensitizeRule: field.Level3,
			Level4DesensitizeRule: field.Level4,
		}
	}
	// 2、获取路由对应的字段
	routeFields, err := service.RouteFieldService.GetByRouteID(route.ID)
	if err != nil {
		logger.Error("获取路由字段失败: ", err)
	}
	for _, field := range routeFields {
		fieldMap[field.FieldName] = &server.DesensitizeField{
			Name:                  field.FieldName,
			IsServiceField:        false,
			Level1DesensitizeRule: field.Level1,
			Level2DesensitizeRule: field.Level2,
			Level3DesensitizeRule: field.Level3,
			Level4DesensitizeRule: field.Level4,
		}
	}

	m.portToRouter[port][domainName].AddRoute(path, handler, fieldMap)
}

func (m *manager) RemoveRoute(port uint16, domain, path, targetUrl string) {
	if routes, ok := m.portToRoutes[port][domain]; ok {
		for i, route := range routes {
			if route.Path == path {
				if targetUrl == "" {
					// 移除所有目标
					route.TargetUpstreams = nil
					route.WeightTotal = 0
				}
				for j, tu := range route.TargetUpstreams {
					if tu.TargetUrl == targetUrl {
						route.TargetUpstreams = append(route.TargetUpstreams[:j], route.TargetUpstreams[j+1:]...)
						route.WeightTotal -= tu.Weight
						break
					}
				}
				if len(route.TargetUpstreams) == 0 {
					m.portToRoutes[port][domain] = append(routes[:i], routes[i+1:]...)
				} else if route.nextIndex >= len(route.TargetUpstreams) {
					route.nextIndex = 0
				}

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
	// 检查，如果该端口下没有任何路由，关闭服务
	if len(m.portToRoutes[port]) == 0 {
		if app, ok := m.portToServer[port]; ok {
			//_ = app.Shutdown()
			_ = app.Shutdown(context.Background())
			delete(m.portToServer, port)
			delete(m.portToRoutes, port)
			delete(m.portToRouter, port)
			delete(m.domainToUserRoute, port)
		}
	}
}

func (m *manager) UpdateService(oldService *model.Service, newService *model.Service) {
	if newService.ID == 0 || oldService.ID == 0 {
		return
	}
	if oldService.Port != newService.Port || oldService.Domain != newService.Domain {
		// 端口或域名发生变化，先增加新的服务，再删除旧的服务
		m.AddService(newService)
		m.RemoveService(oldService)
	}
}

func (m *manager) AddService(serv *model.Service) {
	// 获取服务下的所有路由
	page := 1
	pageSize := 100
	serviceRouteList, total, err := service.RouteService.List(1, 100, &model.Route{ServiceID: &serv.ID})
	if err != nil {
		logger.Error("获取服务下的路由失败: ", err)
		return
	}
	for len(serviceRouteList) < int(total) {
		page++
		routes, _, err := service.RouteService.List(page, pageSize, &model.Route{ServiceID: &serv.ID})
		if err != nil {
			logger.Error("获取服务下的路由失败: ", err)
			return
		}
		serviceRouteList = append(serviceRouteList, routes...)
	}

	for _, route := range serviceRouteList {
		// 获取路由下的所有目标
		page = 1
		var targetWithUpstreams []*domain.TargetWithUpstream
		targetWithUpstreams, total, err = service.UpstreamService.ListByRoute(page, pageSize, route.ID)
		if err != nil {
			logger.Error(err)
			continue
		}
		for len(targetWithUpstreams) < int(total) {
			var uww []*domain.TargetWithUpstream
			page++
			uww, total, err = service.UpstreamService.ListByRoute(page, pageSize, route.ID)
			if err != nil {
				logger.Error(err)
				break
			}
			targetWithUpstreams = append(targetWithUpstreams, uww...)
		}

		for _, upstreamWithWeight := range targetWithUpstreams {
			m.AddRoute(serv, route, upstreamWithWeight.Upstream, upstreamWithWeight.Weight)
		}
	}
}

func (m *manager) RemoveService(serv *model.Service) {
	port := *serv.Port
	domainName := *serv.Domain
	if routes, ok := m.portToRoutes[port][domainName]; ok {
		for _, route := range routes {
			for _, tu := range route.TargetUpstreams {
				m.RemoveRoute(port, domainName, route.Path, tu.TargetUrl)
			}
		}
	}
}

func (m *manager) UpdateUserAllSecretLevel(username string, level int) {
	modifyUserAllSecretLevel(username, level)
}

func (m *manager) UpdateServiceSecretLevel(port uint16, domain string, username string, level int) {
	modifyTokenSecretLevel(port, domain, username, level)
}

func (m *manager) UpdateServiceCertificate(serviceID uint64) {
	if serviceID == 0 {
		return
	}
	// 获取服务信息
	serv, err := service.ServiceService.Get(serviceID)
	if err != nil {
		logger.Error("获取服务信息失败: ", err)
		return
	}
	if serv.CertificateID == nil {
		return
	}

	if *(serv.CertificateID) == 0 {
		// 证书ID为空，删除证书
		m.certManager.deleteServiceCertificate(*serv.Port, *serv.Domain)
	}

	// 获取对应的证书
	cert, err := service.CertificateService.Get(*serv.CertificateID)
	if err != nil {
		logger.Error("获取证书信息失败: ", err)
		return
	}
	if cert == nil {
		return
	}

	// 更新证书
	err = m.certManager.UpdateServiceCertificate(*serv.Port, *serv.Domain, cert)
	if err != nil {
		logger.Error(err)
		return
	}
}
