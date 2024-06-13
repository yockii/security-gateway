package server

import (
	"github.com/gofiber/fiber/v2"
)

type Router struct {
	routes map[string]*Route
	// 优化后的路由树
	tree *TreeRoute
}

func (r *Router) AddRoute(path string, handler fiber.Handler) {
	if r.routes == nil {
		r.routes = make(map[string]*Route)
	}
	route := &Route{path, handler}
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
