package util

import "hash/fnv"

// IpHash IP哈希
func IpHash(ip string) int {
	if ip == "" {
		return 0
	}
	h := fnv.New32a()
	_, _ = h.Write([]byte(ip))
	return int(h.Sum32())
}
