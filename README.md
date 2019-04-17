# 屏蔽词服务
- 提供http，rpc 2中模式的屏蔽词服务
  rpc
  go-micro
```
micro new search_bleak/ --namespace=go.micro --alias=search_bleak --type=srv
```
# 配置
  DevMode   string `json:"devmode"`  // 选择开发模式 dev/prod
	RunMode   string `json:"runmode"`  // 选择运行模式 http/rpc
	Open      string `json:"open"`     // 是否开启屏蔽词服务
	Reload    string `json:"reload"`    // 配置是否重数据库读取屏蔽词数据（数据库方面代码没放上去）
	ThreadNum int    `json:"threadnum"`  // 开启并发数
	ServerNum int    `json:"servernum"`  // 数据分段数
  
# 测试
- 1000并发 2秒内 （屏蔽词大概4000个）  
  Concurrency Level:      1000
  Time taken for tests:   1.717 seconds
  Complete requests:      1000
  Failed requests:        0
  Total transferred:      206000 bytes
  HTML transferred:       83000 bytes
  Requests per second:    582.34 [#/sec] (mean)
  Time per request:       1717.197 [ms] (mean)
  Time per request:       1.717 [ms] (mean, across all concurrent requests)
  Transfer rate:          117.15 [Kbytes/sec] received

  
