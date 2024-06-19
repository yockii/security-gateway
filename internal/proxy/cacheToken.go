package proxy

import (
	"fmt"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gomodule/redigo/redis"
	"security-gateway/pkg/cache"
)

const cacheTime = 60 * 60 * 24 * 3

// RedisKeyTokenToSecretLevel 保存token和密级关系, %d为端口号, %s为域名和token为key，密级为值
var RedisKeyTokenToSecretLevel = cache.Prefix + ":token_to_secret_level:%d:%s:%s"

func cacheToken(port uint16, domain, token string, secretLevel int) {
	conn := cache.Get()
	defer func(conn redis.Conn) {
		err := conn.Close()
		if err != nil {
			log.Error(err)
		}
	}(conn)
	// 存储token和密级关系, RedisKeyTokenToSecretLevel格式化后为key，密级为value，并设定过期时间3天
	_, err := conn.Do("SETEX", fmt.Sprintf(RedisKeyTokenToSecretLevel, port, domain, token), cacheTime, secretLevel)
	if err != nil {
		log.Error(err)
	}
}

func getTokenSecretLevel(port uint16, domain, token string) (int, error) {
	conn := cache.Get()
	defer func(conn redis.Conn) {
		err := conn.Close()
		if err != nil {
			log.Error(err)
		}
	}(conn)
	// 获取token和密级关系, 并再次刷新过期时间
	secretLevel, err := redis.Int(conn.Do("GET", fmt.Sprintf(RedisKeyTokenToSecretLevel, port, domain, token)))
	if err != nil {
		log.Error(err)
	}
	_, err = conn.Do("EXPIRE", fmt.Sprintf(RedisKeyTokenToSecretLevel, port, domain, token), cacheTime)
	if err != nil {
		log.Error(err)
	}
	return secretLevel, err
}
