package domain

import "security-gateway/internal/model"

type UpstreamWithWeight struct {
	model.Upstream
	Weight int `json:"weight"`
}
