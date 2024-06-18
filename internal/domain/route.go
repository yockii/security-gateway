package domain

import "security-gateway/internal/model"

type RouteWithTarget struct {
	*model.Route
	Target *model.Upstream `json:"target"`
}
