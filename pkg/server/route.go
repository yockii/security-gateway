package server

import (
	"github.com/gofiber/fiber/v2"
	logger "github.com/sirupsen/logrus"
	"regexp"
)

type Route struct {
	path    string
	Handler fiber.Handler
}

func NewRoute(path string, handler fiber.Handler) *Route {
	return &Route{path: path, Handler: handler}
}

type TreeRoute struct {
	segment          string                // 当前节点的路径片段，可能是正则表达式， 暂时只考虑*的情况
	route            *Route                // 当前节点直接匹配到的路由，可能为nil
	regexp           *regexp.Regexp        // 当前节点的正则表达式，可能为nil
	children         map[string]*TreeRoute // 子节点
	childrenSegments []string              // 子节点的路径片段，用于排序
}

func (r *TreeRoute) AddRoute(route *Route) {
	// 如果只有一个 / ，直接设置route
	if route.path == "/" {
		r.route = route
		return
	}

	segments := splitPath(route.path)
	if len(segments) == 0 {
		r.route = route
		return
	}
	segment := segments[0]
	if segment == "" {
		if len(segments) == 1 {
			r.route = route
			return
		}
		segment = segments[1]
	}
	if r.children == nil {
		r.children = make(map[string]*TreeRoute)
	}
	if child, ok := r.children[segment]; ok {
		child.addRoute(segments[1:], route)
	} else {
		child = &TreeRoute{segment: segment}
		child.addRoute(segments[1:], route)
		r.children[segment] = child
		r.childrenSegments = append(r.childrenSegments, segment)
	}
	// 排序，按照字典序，*放在最后，{}包裹的正则表达式次之
	sortSegments(r.childrenSegments)
}

func (r *TreeRoute) addRoute(segments []string, route *Route) {
	if len(segments) == 0 {
		r.route = route
		return
	}
	segment := segments[0]
	if segment != "" && segment[0] == '{' && segment[len(segment)-1] == '}' { // 判断segment是否是正则表达式
		// 正则表达式，去掉{}
		regexpStr := segment[1 : len(segment)-1]
		reg, err := regexp.Compile(regexpStr)
		if err != nil {
			logger.Warn("compile regexp failed: ", err)
			return
		}
		r.regexp = reg
	}

	if len(segments) == 1 {
		r.route = route
		return
	}
	// 如果还有segment，继续往下走

	segment = segments[1]

	if r.children == nil {
		r.children = make(map[string]*TreeRoute)
	}
	if child, ok := r.children[segment]; ok {
		child.addRoute(segments[1:], route)
	} else {
		child = &TreeRoute{segment: segment}
		child.addRoute(segments[1:], route)
		r.children[segment] = child
		r.childrenSegments = append(r.childrenSegments, segment)
	}
	// 排序，按照字典序，*放在最后，{}包裹的正则表达式次之
	sortSegments(r.childrenSegments)
}

func (r *TreeRoute) RemoveRoute(path string) bool {
	segments := splitPath(path)
	if len(segments) == 0 {
		// 删除当前节点全部路由
		r.route = nil
		r.children = nil
		r.childrenSegments = nil
		return true
	}
	segment := segments[0]
	if segment == "" {
		if len(segments) == 1 {
			r.route = nil
			r.children = nil
			r.childrenSegments = nil
			return true
		}
		segment = segments[1]
	}
	if child, ok := r.children[segment]; ok {
		canDel := child.removeRoute(segments[1:])
		if canDel {
			// childrenSegments中删除segment
			for i, s := range r.childrenSegments {
				if s == segment {
					r.childrenSegments = append(r.childrenSegments[:i], r.childrenSegments[i+1:]...)
					break
				}
			}
			// 删除子节点
			delete(r.children, segment)
		}
	}
	// 删除子节点后，如果子节点为空，删除子节点
	if len(r.children) == 0 {
		r.children = nil
		r.childrenSegments = nil
		return true
	}
	return false
}

func (r *TreeRoute) removeRoute(segments []string) bool {
	if len(segments) == 1 {
		r.route = nil
		r.children = nil
		r.childrenSegments = nil
		return true
	}
	segment := segments[1]
	if child, ok := r.children[segment]; ok {
		canDel := child.removeRoute(segments[1:])
		if canDel {
			// childrenSegments中删除segment
			for i, s := range r.childrenSegments {
				if s == segment {
					r.childrenSegments = append(r.childrenSegments[:i], r.childrenSegments[i+1:]...)
					break
				}
			}
			// 删除子节点
			delete(r.children, segment)
		}
	}
	// 删除子节点后，如果子节点为空，删除子节点
	if len(r.children) == 0 {
		r.children = nil
		r.childrenSegments = nil
		return r.route == nil
	}
	return false
}

func (r *TreeRoute) FindRoute(path string) *Route {
	segments := splitPath(path)
	if len(segments) == 0 {
		return r.route
	}
	segment := segments[0]
	if segment == "" {
		if len(segments) == 1 {
			return r.route
		}
		segment = segments[1]
	}
	// 按照r.childrenSegments的顺序查找
	for _, s := range r.childrenSegments {
		if s == segment {
			result := r.children[s].findRoute(segments[1:])
			if result != nil {
				return result
			}
			break
		}
	}
	return r.route
}

func (r *TreeRoute) findRoute(segments []string) *Route {
	if len(segments) == 1 {
		return r.route
	}
	segment := segments[1]
	// 按照r.childrenSegments的顺序查找
	for _, s := range r.childrenSegments {
		child := r.children[s]
		if matchSegment(s, segment, child.regexp) {
			return r.children[s].findRoute(segments[1:])
		}
	}
	return r.route
}
