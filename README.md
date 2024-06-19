# security-gateway
安全网关

本项目主要为安全反向代理，对字段进行分类分级，并依据用户数据进行密级评定，对需要的字段进行脱敏后再返回

## 项目预期

- 项目主要为安全网关，对请求的字段进行分类分级，对用户进行密级评定，对字段进行脱敏后再返回
- 能够对服务进行反向代理，基本实现Nginx的功能
- 能够拦截服务的用户信息，进行用户信息的密级评定
- 项目主要为实验性项目，不保证安全性，不保证生产可用性，不保证功能完整性，不保证功能正确性，不保证性能，不保证稳定性，不保证可维护性，不保证可扩展性
- 不排除项目实践良好，转为生产项目的可能性

## 使用说明

- 先增加好服务、路由、上游路径
- 关联路由和上游路径，这将会启动一个端口监听并执行对应的代理

## 进度目标

- [x] 上游管理(Upstream)
- [x] 服务管理(Service)
- [x] 路由管理(Route)：路由在服务下
- [x] 路由与上游关联
- [x] 代理服务
- [x] 自动启动服务监听
- [x] 服务级别字段脱敏
- [ ] 路由级别字段脱敏
- [ ] 服务的用户认证请求配置：用于拦截对应服务的用户信息，进行用户信息的密级评定
- [ ] 用户管理
- [ ] 用户密级评定
- [ ] 用户在服务下的密级评定
- [ ] 支持均衡负载配置：多个上游，支持轮询、随机、权重等
- [ ] 支持TLS配置，支持HTTPS
- [ ] 支持分布式部署*
- [ ] 请求统计分析：按服务、路由、用户、时间等维度，以及脱敏次数
- [ ] 服务监控：服务的健康检查，服务的状态