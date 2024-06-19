package domain

import "security-gateway/internal/model"

type Port struct {
	Port  uint16 `json:"port"`
	InUse bool   `json:"inUse"`
}

type UserLevelWithService struct {
	*model.UserServiceLevel
	Service *model.Service `json:"service"`
}
