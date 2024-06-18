package domain

type Port struct {
	Port  uint16 `json:"port"`
	InUse bool   `json:"inUse"`
}
