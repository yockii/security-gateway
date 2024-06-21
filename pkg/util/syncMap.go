package util

import "sync"

func ToSyncMap(m map[string]interface{}) *sync.Map {
	sm := &sync.Map{}
	for k, v := range m {
		sm.Store(k, v)
	}
	return sm
}

func ToMap(sm *sync.Map) map[string]interface{} {
	m := make(map[string]interface{})
	sm.Range(func(k, v interface{}) bool {
		m[k.(string)] = v
		return true
	})
	return m
}
