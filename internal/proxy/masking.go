package proxy

import (
	"encoding/json"
	logger "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"net/http"
	"security-gateway/internal/model"
	"security-gateway/internal/service"
	"security-gateway/pkg/server"
	"security-gateway/pkg/util"
	"strings"
	"sync"
)

//func (m *manager) modifyResponse(req *fasthttp.Request, resp *fasthttp.Response, port uint16, domain string, fields []*server.DesensitizeField) (maskingLevel int, username string) {
//	resp.Header.Del(fiber.HeaderServer)
//	maskingLevel = 0
//	// 确保本方法不会panic
//	defer func() {
//		if r := recover(); r != nil {
//			logger.Error("modifyResponse panic: ", r)
//			maskingLevel = 0
//		}
//	}()
//
//	body := string(resp.Body())
//	if !gjson.Valid(body) {
//		// 不是json格式，不做处理
//		return
//	}
//	bodyJson := gjson.Parse(body)
//
//	token := ""
//	if d, has := m.domainToUserRoute[port]; has {
//		if uir, ok := d[domain]; ok {
//			// 获取token
//			tokenPosition := strings.Split(uir.TokenPosition, ":")
//			if len(tokenPosition) < 3 {
//				// token获取条件不满足
//				return
//			}
//			w := tokenPosition[1]
//			p := tokenPosition[2]
//			switch tokenPosition[0] {
//			case "request":
//				switch w {
//				case "header":
//					token = string(req.Header.Peek(p))
//				case "query":
//					token = string(req.URI().QueryArgs().Peek(p))
//				case "body":
//					// 判断req是否是form表单提交
//					if strings.Contains(string(req.Header.Peek(fiber.HeaderContentType)), "application/x-www-form-urlencoded") {
//						token = string(req.PostArgs().Peek(p))
//					} else {
//						token = gjson.ParseBytes(req.Body()).Get(p).String()
//					}
//				case "cookies":
//					token = string(req.Header.Cookie(p))
//				}
//			case "response":
//				switch w {
//				case "body":
//					token = bodyJson.Get(p).String()
//				}
//			}
//			// token为空，则所有密级都为1
//
//			// 判断是否是用户信息路由
//			if uir.Path == string(req.URI().Path()) && uir.Method == string(req.Header.Method()) {
//				// 获取用户信息，并缓存token和密级关系
//
//				// 2、获取用户信息
//				uniKey := bodyJson.Get(uir.UniKeyPath).String()
//				if uniKey == "" {
//					// uniKey获取失败
//					return
//				}
//
//				// 用户名
//				username = bodyJson.Get(uir.UsernamePath).String()
//
//				// 3、根据uniKey存储位置查找user
//				var user *model.User
//				if uir.MatchKey == "-" {
//					user = service.UserService.GetByUniKey(username, uniKey)
//				} else {
//					matchKey := bodyJson.Get(uir.MatchKey).String()
//					if matchKey == "" {
//						// matchKey获取失败
//						return
//					}
//					user = service.UserService.GetByUniKeyJson(username, uniKey, matchKey)
//				}
//				if user == nil {
//					// 用户信息获取失败
//					if username != "" {
//						// 保存用户信息
//						user = &model.User{
//							Username: username,
//							UniKey:   uniKey,
//						}
//						if _, _, err := service.UserService.Add(user); err != nil {
//							logger.Error("保存用户信息失败: ", err)
//							return
//						}
//					}
//					return
//				}
//
//				secLevel := user.SecLevel
//				// 还要看user在服务下的密级，如果存在，则以此为准
//				{
//					// 获取userServiceLevel
//					usl := service.UserServiceLevelService.GetByUserAndServiceID(user.ID, uir.ServiceID)
//					if usl != nil {
//						secLevel = usl.SecLevel
//					}
//				}
//
//				// 4、保存token和密级关系
//				cacheToken(port, domain, token, secLevel, user.Username)
//			}
//		}
//	}
//
//	// 其他路由，根据token获取密级
//	var secLevel = 1
//	if token != "" {
//		l, u := getTokenSecretLevel(port, domain, token)
//		if l > 0 {
//			secLevel = l
//		}
//		if u != "" {
//			username = u
//		}
//	}
//
//	if len(fields) > 0 {
//		// 对字段进行脱敏处理
//		modifiedBody := m.modifyFields(bodyJson, fields, secLevel)
//		resp.SetBody([]byte(modifiedBody))
//		maskingLevel = secLevel
//	}
//
//	return
//}

func (m *manager) modifyResponse(req *http.Request, resp *ModifiableResponseWriter, port uint16, domain string, fields []*server.DesensitizeField) (maskingLevel int, username string) {
	maskingLevel = 0
	// 确保本方法不会panic
	defer func() {
		if r := recover(); r != nil {
			logger.Error("modifyResponse panic: ", r)
			maskingLevel = 0
		}
	}()

	body := string(resp.body.Bytes())
	if !gjson.Valid(body) {
		// 不是json格式，不做处理
		return
	}
	bodyJson := gjson.Parse(body)

	token := ""
	if d, has := m.domainToUserRoute[port]; has {
		if uir, ok := d[domain]; ok {
			// 获取token
			tokenPosition := strings.Split(uir.TokenPosition, ":")
			if len(tokenPosition) < 3 {
				// token获取条件不满足
				return
			}
			w := tokenPosition[1]
			p := tokenPosition[2]
			switch tokenPosition[0] {
			case "request":
				switch w {
				case "header":
					token = req.Header.Get(p)
				case "query":
					token = req.URL.Query().Get(p)
				case "body":
					// 判断req是否是form表单提交
					if strings.Contains(req.Header.Get(http.CanonicalHeaderKey("Content-Type")), "application/x-www-form-urlencoded") {
						token = req.PostFormValue(p)
					} else {
						reqBody := make([]byte, req.ContentLength)
						_, _ = req.Body.Read(reqBody)
						token = gjson.ParseBytes(reqBody).Get(p).String()
					}
				case "cookies":
					tokenCookie, err := req.Cookie(p)
					if err != nil {
						logger.Warn(err)
					}
					if tokenCookie != nil {
						token = tokenCookie.Value
					}
				}
			case "response":
				switch w {
				case "body":
					token = bodyJson.Get(p).String()
				}
			}
			// token为空，则所有密级都为1

			// 判断是否是用户信息路由
			if uir.Path == req.URL.Path && uir.Method == req.Method {
				// 获取用户信息，并缓存token和密级关系

				// 2、获取用户信息
				uniKey := bodyJson.Get(uir.UniKeyPath).String()
				if uniKey == "" {
					// uniKey获取失败
					return
				}

				// 用户名
				username = bodyJson.Get(uir.UsernamePath).String()

				// 3、根据uniKey存储位置查找user
				var user *model.User
				if uir.MatchKey == "-" {
					user = service.UserService.GetByUniKey(username, uniKey)
				} else {
					matchKey := bodyJson.Get(uir.MatchKey).String()
					if matchKey == "" {
						// matchKey获取失败
						return
					}
					user = service.UserService.GetByUniKeyJson(username, uniKey, matchKey)
				}
				if user == nil {
					// 用户信息获取失败
					if username != "" {
						// 保存用户信息
						user = &model.User{
							Username: username,
							UniKey:   uniKey,
						}
						if _, _, err := service.UserService.Add(user); err != nil {
							logger.Error("保存用户信息失败: ", err)
							return
						}
					}
					return
				}

				secLevel := user.SecLevel
				// 还要看user在服务下的密级，如果存在，则以此为准
				{
					// 获取userServiceLevel
					usl := service.UserServiceLevelService.GetByUserAndServiceID(user.ID, uir.ServiceID)
					if usl != nil {
						secLevel = usl.SecLevel
					}
				}

				// 4、保存token和密级关系
				cacheToken(port, domain, token, secLevel, user.Username)
			}
		}
	}

	// 其他路由，根据token获取密级
	var secLevel = 1
	if token != "" {
		l, u := getTokenSecretLevel(port, domain, token)
		if l > 0 {
			secLevel = l
		}
		if u != "" {
			username = u
		}
	}

	if len(fields) > 0 {
		// 对字段进行脱敏处理
		modifiedBody := m.modifyFields(bodyJson, fields, secLevel)

		resp.body.Reset()
		_, _ = resp.Write([]byte(modifiedBody))

		maskingLevel = secLevel
	}

	return
}

func (m *manager) modifyFields(bodyJson gjson.Result, fields []*server.DesensitizeField, level int) (modifiedBody string) {
	modifiedBody = bodyJson.Raw
	if bodyJson.IsArray() {
		var modifiedArray []interface{}
		// 如果是数组，遍历每个元素
		for _, element := range bodyJson.Array() {
			modifiedArray = append(modifiedArray, m.modifyFields(element, fields, level))
		}
		modifiedBodyBytes, err := json.Marshal(modifiedArray)
		if err != nil {
			logger.Error("json.Marshal failed: ", err)
			return
		}
		modifiedBody = string(modifiedBodyBytes)
		return
	}

	var bodyMap map[string]interface{}
	if err := json.Unmarshal([]byte(bodyJson.Raw), &bodyMap); err != nil {
		logger.Error("json.Unmarshal failed: ", err)
		return
	}

	bodyMap = m.doModifyFields(bodyJson, bodyMap, fields, level)
	modifiedBodyBytes, err := json.Marshal(bodyMap)
	if err != nil {
		logger.Error("json.Marshal failed: ", err)
		return
	}
	modifiedBody = string(modifiedBodyBytes)
	return
}

func (m *manager) doModifyFields(bodyJson gjson.Result, bodyMap map[string]interface{}, fields []*server.DesensitizeField, level int) map[string]interface{} {
	// 遍历所有字段，每个字段并发独立处理
	syncMap := util.ToSyncMap(bodyMap)
	var wg sync.WaitGroup
	for _, field := range fields {
		fieldName := field.Name
		maskPattern := field.Level4DesensitizeRule
		switch level {
		case 1:
			maskPattern = field.Level1DesensitizeRule
		case 2:
			maskPattern = field.Level2DesensitizeRule
		case 3:
			maskPattern = field.Level3DesensitizeRule
		case 4:
			maskPattern = field.Level4DesensitizeRule
		}
		wg.Add(1)
		go func(bodyJson gjson.Result, bMap *sync.Map, field *server.DesensitizeField, level int) {
			defer wg.Done()
			m.doModifyField(bodyJson, bMap, fieldName, maskPattern)
		}(bodyJson, syncMap, field, level)
	}
	wg.Wait()
	return util.ToMap(syncMap)
}

// doModifyField 对字段进行脱敏处理, 确保j和m是同级别的
func (m *manager) doModifyField(j gjson.Result, obj *sync.Map, fieldName string, maskPattern string) map[string]interface{} {
	if maskPattern == "-" || maskPattern == "" {
		return util.ToMap(obj)
	}

	if _, ok := obj.Load(fieldName); ok {
		val := j.Get(fieldName).String()
		modifiedValue, err := util.AdvanceMask(val, maskPattern)
		if err != nil {
			logger.Error("AdvanceMask failed: ", err)
			return util.ToMap(obj)
		}
		obj.Store(fieldName, modifiedValue)
	}

	// 找到嵌套的字段进行递归处理
	obj.Range(func(key, value interface{}) bool {
		if key == fieldName {
			return true
		}
		k, ok := key.(string)
		if !ok {
			return true
		}
		jk := j.Get(k)
		if jk.IsArray() {
			jArr := jk.Array()
			vArr, ok := value.([]interface{})
			if !ok {
				return true
			}
			for i, element := range vArr {
				eleMap, ok := element.(map[string]interface{})
				if !ok {
					return true
				}
				syncMap := util.ToSyncMap(eleMap)
				om := m.doModifyField(jArr[i], syncMap, fieldName, maskPattern)
				vArr[i] = om
			}

			obj.Store(k, vArr)
		} else if jk.IsObject() {
			vMap, ok := value.(map[string]interface{})
			if !ok {
				return true
			}
			syncMap := util.ToSyncMap(vMap)
			om := m.doModifyField(jk, syncMap, fieldName, maskPattern)
			obj.Store(k, om)
		}
		return true
	})

	return util.ToMap(obj)
}
