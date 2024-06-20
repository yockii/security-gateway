package domain

import "security-gateway/internal/model"

type RouteWithTargets struct {
	*model.Route
	Targets []*model.Upstream `json:"targets"`
}
