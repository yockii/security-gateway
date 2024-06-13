package server

import (
	"testing"
)

func TestTreeRoute(t *testing.T) {
	type args struct {
		route *Route
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "路由: /a",
			args: args{
				route: NewRoute("/a", nil),
			},
		},
		{
			name: "路由: /a/b",
			args: args{
				route: NewRoute("/a/b", nil),
			},
		},
		{
			name: "路由: /a/b/c",
			args: args{
				route: NewRoute("/a/b/c", nil),
			},
		},
		{
			name: "路由: /a/*/c",
			args: args{
				route: NewRoute("/a/*/c", nil),
			},
		},
		{
			name: "路由: /a/{\\d+}/c",
			args: args{
				route: NewRoute("/a/{\\d+}/c", nil),
			},
		},
	}
	r := &TreeRoute{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r.AddRoute(tt.args.route)
		})
	}

	route := r.FindRoute("/a/b/c")
	if route == nil {
		t.Errorf("路由 /a/b/c 未找到")
	}

	route = r.FindRoute("/a/b")
	if route == nil {
		t.Errorf("路由 /a/b 未找到")
	}

	route = r.FindRoute("/a")
	if route == nil {
		t.Errorf("路由 /a 未找到")
	}

	route = r.FindRoute("/a/1/c")
	if route == nil {
		t.Errorf("路由 /a/1/c 未找到")
	} else if route.path != "/a/{\\d+}/c" {
		t.Errorf("路由 /a/1/c 未找到对应的处理函数")
	}

	route = r.FindRoute("/a/1")
	if route != nil {
		t.Errorf("路由 /a/1 不应该找到")
	}

	route = r.FindRoute("/a/1/c/1")
	if route == nil {
		t.Errorf("路由 /a/1/c/1 未找到")
	}

	r.RemoveRoute("/a/b/c")

	route = r.FindRoute("/a/b/c")
	if route != nil && route.path != "/a/b" {
		t.Errorf("路由 /a/b/c 查找错误")
	}

	r.RemoveRoute("/a/b")

	route = r.FindRoute("/a/b/c")
	if route != nil && route.path != "/a/*/c" {
		t.Errorf("路由 /a/b/c 未删除")

	}

	route = r.FindRoute("/a/b")
	if route != nil {
		t.Errorf("路由 /a/b 未删除")
	}

}
