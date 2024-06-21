package proxy

import (
	"errors"
	"fmt"
	"github.com/gomodule/redigo/redis"
	logger "github.com/sirupsen/logrus"
	"security-gateway/pkg/cache"
)

const cacheTime = 60 * 60 * 24 * 3

// RedisKeyTokenToSecretLevel 保存token和密级关系, %d为端口号, %s为域名和token为key，密级为值
var RedisKeyTokenToSecretLevel = cache.Prefix + ":token_to_secret_level:%d:%s:%s"

// RedisKeyTokenToUsername 保存token和用户名关系, %d为端口号, %s为域名和token为key，用户名为值
var RedisKeyTokenToUsername = cache.Prefix + ":token_to_username:%d:%s:%s"

// RedisKeyUsernameToTokens 保存用户名和token关系, %d为端口号, %s为域名和用户名为key，token列表为值, 并设定过期时间3天
var RedisKeyUsernameToTokens = cache.Prefix + ":username_to_tokens:%d:%s:%s"

func cacheToken(port uint16, domain, token string, secretLevel int, username string) {
	conn := cache.Get()
	defer func(conn redis.Conn) {
		err := conn.Close()
		if err != nil {
			logger.Error(err)
		}
	}(conn)
	// 存储token和密级关系, RedisKeyTokenToSecretLevel格式化后为key，密级为value，并设定过期时间3天
	_, err := conn.Do("SETEX", fmt.Sprintf(RedisKeyTokenToSecretLevel, port, domain, token), cacheTime, secretLevel)
	if err != nil {
		logger.Error(err)
	}
	// 存储token和用户名关系, RedisKeyTokenToUsername格式化后为key，用户名为value，并设定过期时间3天
	_, err = conn.Do("SETEX", fmt.Sprintf(RedisKeyTokenToUsername, port, domain, token), cacheTime, username)
	if err != nil {
		logger.Error(err)
	}
	// 存储用户名和token关系, RedisKeyUsernameToTokens格式化后为key，token列表为value，set，并设定过期时间3天
	_, err = conn.Do("SADD", fmt.Sprintf(RedisKeyUsernameToTokens, port, domain, username), token)
	if err != nil {
		logger.Error(err)
	}
	_, err = conn.Do("EXPIRE", fmt.Sprintf(RedisKeyUsernameToTokens, port, domain, username), cacheTime)
}

func getTokenSecretLevel(port uint16, domain, token string) (secretLevel int, username string) {
	var err error
	conn := cache.Get()
	defer func(conn redis.Conn) {
		err = conn.Close()
		if err != nil {
			logger.Error(err)
		}
	}(conn)
	// 获取token和密级关系, 并再次刷新过期时间
	secretLevel, err = redis.Int(conn.Do("GET", fmt.Sprintf(RedisKeyTokenToSecretLevel, port, domain, token)))
	if err != nil {
		logger.Error(err)
	}
	_, err = conn.Do("EXPIRE", fmt.Sprintf(RedisKeyTokenToSecretLevel, port, domain, token), cacheTime)
	if err != nil {
		logger.Error(err)
	}
	// 获取token和用户名关系, 并再次刷新过期时间
	username, err = redis.String(conn.Do("GET", fmt.Sprintf(RedisKeyTokenToUsername, port, domain, token)))
	if err != nil {
		logger.Error(err)
	}
	_, err = conn.Do("EXPIRE", fmt.Sprintf(RedisKeyTokenToUsername, port, domain, token), cacheTime)
	if err != nil {
		logger.Error(err)
	}
	_, err = conn.Do("EXPIRE", fmt.Sprintf(RedisKeyUsernameToTokens, port, domain, username), cacheTime)
	if err != nil {
		logger.Error(err)
	}
	return
}

func modifyTokenSecretLevel(port uint16, domain, username string, secretLevel int) {
	conn := cache.Get()
	defer func(conn redis.Conn) {
		err := conn.Close()
		if err != nil {
			logger.Error(err)
		}
	}(conn)

	// 取出所有token
	tokens, err := redis.Strings(conn.Do("SMEMBERS", fmt.Sprintf(RedisKeyUsernameToTokens, port, domain, username)))
	if err != nil {
		logger.Error(err)
		return
	}

	// 修改token和密级关系
	for _, token := range tokens {
		// 检查是否存在token和密级关系
		_, err = redis.Int(conn.Do("GET", fmt.Sprintf(RedisKeyTokenToSecretLevel, port, domain, token)))
		if err != nil && errors.Is(err, redis.ErrNil) {
			// 不存在则删除RedisKeyUsernameToTokens中的token
			_, _ = conn.Do("SREM", fmt.Sprintf(RedisKeyUsernameToTokens, port, domain, username), token)
			continue
		}

		_, err = conn.Do("SETEX", fmt.Sprintf(RedisKeyTokenToSecretLevel, port, domain, token), cacheTime, secretLevel)
		if err != nil {
			logger.Error(err)
		}
	}
}

func modifyUserAllSecretLevel(username string, secretLevel int) {
	conn := cache.Get()
	defer func(conn redis.Conn) {
		err := conn.Close()
		if err != nil {
			logger.Error(err)
		}
	}(conn)

	// 搜索所有带username的key （RedisKeyUsernameToTokens）
	keys, err := redis.Strings(conn.Do("KEYS", fmt.Sprintf(RedisKeyUsernameToTokens, "*", "*", username)))
	if err != nil {
		logger.Error(err)
		return
	}

	// 修改token和密级关系
	for _, key := range keys {
		// 取出所有token
		var tokens []string
		tokens, err = redis.Strings(conn.Do("SMEMBERS", key))
		if err != nil {
			logger.Error(err)
			continue
		}

		// 从key中提取port和domain
		var port uint16
		var domain string
		_, err = fmt.Sscanf(key, fmt.Sprintf(RedisKeyUsernameToTokens, "%d", "%s", username), &port, &domain)
		if err != nil {
			logger.Error(err)
			continue
		}
		if port == 0 || domain == "" {
			logger.Error("port or domain is empty")
			continue
		}

		for _, token := range tokens {
			// 检查是否存在token和密级关系
			_, err = redis.Int(conn.Do("GET", fmt.Sprintf(RedisKeyTokenToSecretLevel, port, domain, token)))
			if err != nil && errors.Is(err, redis.ErrNil) {
				// 不存在则删除RedisKeyUsernameToTokens中的token
				_, _ = conn.Do("SREM", fmt.Sprintf(RedisKeyUsernameToTokens, port, domain, username), token)
				continue
			}

			_, err = conn.Do("SETEX", fmt.Sprintf(RedisKeyTokenToSecretLevel, port, domain, token), cacheTime, secretLevel)
			if err != nil {
				logger.Error(err)
			}
		}
	}
}
