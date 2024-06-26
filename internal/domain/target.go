package domain

import "security-gateway/internal/model"

type TargetWithUpstream struct {
	model.RouteTarget
	Upstream *model.Upstream `json:"upstream"`
}
