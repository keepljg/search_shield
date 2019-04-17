package config

type SearchShieldConfig struct {
	DevMode   string `json:"devmode"`  // 开发模式 dev/prod
	RunMode   string `json:"runmode"`  // 运行模式 http/rpc
	Open      string `json:"open"`
	Reload    string `json:"reload"`
	ThreadNum int    `json:"threadnum"`
	ServerNum int    `json:"servernum"`
}

func GetSearchShieldConfig() SearchShieldConfig {
	return SearchShieldConfig{
		DevMode:   "dev",
		RunMode:   "http",
		Open:      "yes",
		Reload:    "",
		ThreadNum: 10,
		ServerNum: 1,
	}
}

var SearchListConfig SearchShieldConfig

func init() {
	SearchListConfig = GetSearchShieldConfig()
}
