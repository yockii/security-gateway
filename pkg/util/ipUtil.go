package util

import (
	"hash/fnv"
	"net/http"
	"strings"
)

// IpHash IP哈希
func IpHash(ip string) int {
	if ip == "" {
		return 0
	}
	h := fnv.New32a()
	_, _ = h.Write([]byte(ip))
	return int(h.Sum32())
}

// GetUserIP 获取用户IP
func GetUserIP(r *http.Request) string {
	userIP := r.Header.Get("X-Forwarded-For")
	if userIP == "" {
		userIP = r.Header.Get("X-Real-IP")
	}
	if userIP == "" {
		// 如果没有X-Forwarded-For和X-Real-IP，尝试从RemoteAddr获取IP
		remoteAddr := r.RemoteAddr
		if i := strings.LastIndex(remoteAddr, ":"); i >= 0 {
			userIP = remoteAddr[:i]
		}
	}
	return userIP
}
