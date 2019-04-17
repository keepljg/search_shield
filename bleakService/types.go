package bleakService

import (
	"regexp"
	"sync"
	"search_shield/config"
)

type BleakModel struct {
	Id    int    `orm:"pk;column(id)"`
	Match string `orm:"size(255);column(match)"`
}

type ShieldSearchServiceData struct {
	Id        int
	Match     string
	Operation string
}

// 屏蔽词任务服务
type ShieldTaskService struct {
	closing chan bool

	// 业务数据
	Keyword      chan string
	ShieldStatus chan bool
	Shield       chan ShieldSearchServiceData

	// 屏蔽词字典
	ShieldDict []*BleakModel

	// 屏蔽词正则表达式
	ShieldRegexp []*regexp.Regexp
}

type ShieldService struct {
	closing   chan bool
	waitGroup *sync.WaitGroup

	// 配置信息
	C config.SearchShieldConfig

	// 屏蔽词容器
	m []*BleakModel

	// 业务数据
	Keyword      chan string
	ShieldStatus chan bool
	Shield       chan ShieldSearchServiceData

	// 子任务
	n       int
	TaskMap map[int]*ShieldTaskService
}
