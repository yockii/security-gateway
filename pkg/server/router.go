package server

import (
	"net/http"
)

type Router struct {
	routes map[string]*Route
	// 优化后的路由树
	tree *TreeRoute
}

//func (r *Router) AddRoute(path string, handler fiber.Handler, fields []*DesensitizeField) {
//	if r.routes == nil {
//		r.routes = make(map[string]*Route)
//	}
//
//	route := &Route{path, handler, fields}
//	r.routes[path] = route
//	// 优化路由树
//	if r.tree == nil {
//		r.tree = &TreeRoute{}
//	}
//	r.tree.AddRoute(route)
//}

func (r *Router) AddRoute(path string, handler http.HandlerFunc, fields []*DesensitizeField) {
	if r.routes == nil {
		r.routes = make(map[string]*Route)
	}

	route := &Route{path, handler, fields}
	r.routes[path] = route
	// 优化路由树
	if r.tree == nil {
		r.tree = &TreeRoute{}
	}
	r.tree.AddRoute(route)
}

func (r *Router) RemoveRoute(path string) (noUsed bool) {
	if r.tree.RemoveRoute(path) {
		delete(r.routes, path)
		return true
	}
	return false
}

func (r *Router) FindRoute(path string) *Route {
	if r.tree == nil {
		return nil
	}
	return r.tree.FindRoute(path)
}

// UpdateServiceField 更新服务字段，如果有则替换，如果没有则添加
func (r *Router) UpdateServiceField(field *DesensitizeField) {
	// 遍历所有路由，更新字段
	for _, route := range r.routes {
		route.UpdateServiceField(field)
	}
}

// RemoveServiceField 删除服务字段
func (r *Router) RemoveServiceField(name string) {
	// 遍历所有路由，删除字段
	for _, route := range r.routes {
		route.RemoveServiceField(name)
	}
}

// UpdateRouteField 更新路由字段，如果有则替换，如果没有则添加
func (r *Router) UpdateRouteField(path string, field *DesensitizeField) {
	route, has := r.routes[path]
	if !has {
		return
	}
	route.UpdateRouteField(field)
}

// RemoveRouteFieldWithServiceFieldUpdate 删除路由字段并更新服务字段
func (r *Router) RemoveRouteFieldWithServiceFieldUpdate(path string, name string, serviceField *DesensitizeField) {
	route, has := r.routes[path]
	if !has {
		return
	}
	route.RemoveRouteFieldAndUpdateServiceField(name, serviceField)
}
