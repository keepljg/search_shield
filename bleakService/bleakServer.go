package bleakService

import (
	"bufio"
	"fmt"
	"github.com/astaxie/beego/logs"
	"io"
	"math/rand"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"
	"search_shield/config"
	"sync"
)

const (
	default_task_num = 10
	search_shield_file = "search-shield.log"
)

var (
	default_server_num int
	shieldServices []*ShieldService
	mu sync.Mutex
)


// 初始化屏蔽词服务
func InitShieldService() {
	if config.SearchListConfig.ServerNum == 0 {
		default_server_num = 1
	} else {
		default_server_num = config.SearchListConfig.ServerNum
	}
	for i :=0; i < config.SearchListConfig.ServerNum; i++ {
		server := NewShieldService(config.GetSearchShieldConfig())
		shieldServices = append(shieldServices, server)
		go SearchShieldService(server)
	}
	//go shieldInfoFileSync()
	go StopShieldService(shieldServices)

}


func StopShieldService(shieldServices []*ShieldService) {
	// 阻塞等待接受程序退出信息
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch

	for _, shieldService := range shieldServices {
		shieldService.Stop()
	}
	s := InstanceShieldService()
	logs.Info("开始文件信息同步")
	WriteShieldToFile(s.m)
	logs.Info("文件同步完成")
}


// 获取屏蔽词服务实例
func InstanceShieldService() *ShieldService {
	randInt := RangeInt(default_server_num)
	shieldService := shieldServices[randInt]
	if shieldService != nil {
		return shieldService
	} else {
		if shieldService != nil {
			return shieldService
		}
		mu.Lock()
		defer mu.Unlock()
		shieldService = NewShieldService(config.SearchListConfig)
		shieldServices[randInt] = shieldService
		return shieldService
	}
}

// 获取所有屏蔽词实例
func InstanceShieldServices() []*ShieldService{
	return shieldServices
}


func NewShieldService(config config.SearchShieldConfig) *ShieldService {
	taskNum := config.ThreadNum
	if taskNum == 0 {
		taskNum = default_task_num
	}
	shieldService := &ShieldService{
		closing:      make(chan bool),
		waitGroup:    &sync.WaitGroup{},
		C:            config,
		Shield:       make(chan ShieldSearchServiceData),
		Keyword:      make(chan string),
		ShieldStatus: make(chan bool),
		n:            taskNum,
		TaskMap:      make(map[int]*ShieldTaskService),
	}
	return shieldService
}


func NewShieldTaskService(m []*BleakModel) *ShieldTaskService {
	shieldService := &ShieldTaskService{
		closing:      make(chan bool),
		Keyword:      make(chan string),
		ShieldStatus: make(chan bool),
		Shield:       make(chan ShieldSearchServiceData),
	}
	// 深拷贝
	shieldService.ShieldDict = make([]*BleakModel, len(m))
	shieldService.ShieldRegexp = make([]*regexp.Regexp, len(m))
	for i := 0; i < len(m); i++ {
		shieldService.ShieldDict[i] = new(BleakModel)
		shieldService.ShieldDict[i].Id = m[i].Id
		shieldService.ShieldDict[i].Match = m[i].Match
		shieldService.ShieldRegexp[i] = regexp.MustCompile(m[i].Match)
	}

	return shieldService
}

// 搜索屏蔽服务携程
func SearchShieldService(service *ShieldService) {
	go service.Serve()

	//// 阻塞等待接受程序退出信息
	//ch := make(chan os.Signal)
	//signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	//<-ch
	//
	//// 携程退出
	//service.Stop()
}


func (s *ShieldService) Serve() {
	s.waitGroup.Add(1)
	defer s.waitGroup.Done()

	// 检查是否开启
	if s.C.Open == "no" {
		return
	}

	// 读取屏蔽词信息
	s.m = GetSheildFromFile()
	//if s.C.Reload == "yes" {
	//	// 数据库读取数据
	//	var err error
	//	if s.m, err = blacklist.GetAll(); err != nil {
	//		return
	//	}
	//
	//	// 写入文件
	//	WriteShieldToFile(s.m)
	//} else {
	//	// 从文件读取数据
	//	s.m = GetSheildFromFile()
	//}

	// 当没有搜索屏蔽词词库时，自动关闭服务
	if len(s.m) == 0 {
		s.C.Open = "no"
		return
	}

	// 初始化屏蔽词任务
	if len(s.m) < s.n {
		s.n = len(s.m)
	}

	// 对数据进行分块
	shieldDict := make([][]*BleakModel, s.n)
	for i := 0; i < s.n; i++ {
		shieldDict[i] = make([]*BleakModel, 0)
	}
	for i := 0; i < len(s.m); i++ {
		index := s.m[i].Id % s.n
		shieldDict[index] = append(shieldDict[index], s.m[i])
	}
	for i := 0; i < s.n; i++ {
		s.TaskMap[i] = NewShieldTaskService(shieldDict[i])
	}

	// 开启任务
	for i := 0; i < s.n; i++ {
		s.waitGroup.Add(1)
		defer s.waitGroup.Done()
		go s.TaskMap[i].Serve()
	}

	for {
		select {
		case <-s.closing:
			// 给子任务发出退出信号
			for i := 0; i < s.n; i++ {
				s.TaskMap[i].Stop()
			}
			return

			// 监听搜索词
		case v, ok := <-s.Keyword:
			if ok {
				// 任务判断
				for i := 0; i < s.n; i++ {
					s.TaskMap[i].ReceiveKeyword(v)
				}

				// 汇总结果
				result := make([]bool, s.n)
				for i := 0; i < s.n; i++ {
					v, _ := <-s.TaskMap[i].ShieldStatus
					result[i] = v
				}

				// 回复报告
				var status bool
				for i := 0; i < s.n; i++ {
					if result[i] {
						status = true
						break
					}
				}
				s.SetShieldStatus(status)
			}

			// 监听屏蔽词更新：动态扩展中不扩展子任务数
		case v, ok := <-s.Shield:
			if ok {
				s.RefreshShieldData(v)
			}
		}
	}
}

func (s *ShieldService) SetShieldStatus (status bool) {
	s.ShieldStatus <- status
}

func (s *ShieldService) WriteShieldToFile() {
	WriteShieldToFile(s.m)
}


func (s *ShieldService) Stop() {
	close(s.closing)
	s.waitGroup.Wait()
}

// 结果的接口
func (s *ShieldService) GetShieldStatus() (status bool) {
	status, _ = <-s.ShieldStatus
	return
}

// 发送keyword的接口
func (s *ShieldService) ReceiveKeyword(keyword string) {
	s.Keyword <- keyword
}

func (s *ShieldService) ReceiveShield(shield ShieldSearchServiceData) {
	s.Shield <- shield
}

func (s *ShieldService) RefreshShieldData(v ShieldSearchServiceData) {
	var find bool
	var index int
	switch v.Operation {
	case "insert":
		m1 := &BleakModel{
			Id:    v.Id,
			Match: v.Match,
		}
		find = true
		s.m = append(s.m, m1)
		index = m1.Id % s.n
	case "update":
		for i := 0; i < len(s.m); i++ {
			if s.m[i].Id == v.Id {
				find = true
				s.m[i].Match = v.Match
			}
		}
		if !find {
			break
		}
		index = v.Id % s.n
	case "delete":
		var delindex int
		for i := 0; i < len(s.m); i++ {
			if s.m[i].Id == v.Id {
				find = true
				delindex = i
			}
		}
		if !find {
			break
		}
		s.m = append(s.m[:delindex], s.m[delindex+1:]...)
		index = v.Id % s.n
	default:
		break
	}
	if find {
		//WriteShieldToFile(s.m)
		if s.n == 1 {
			s.TaskMap[0].ReceiveShield(v)
		} else {
			s.TaskMap[index].ReceiveShield(v)
		}
	}
}

func (s *ShieldTaskService) ReceiveShield(shield ShieldSearchServiceData) {
	s.Shield <- shield
}


func (s *ShieldTaskService) Serve() {
	// 循环监听信号
	for {
		select {
		case <-s.closing:
			return

			// 屏蔽词检查请求
		case v, _ := <-s.Keyword:
			s.SendShieldStaus(v)

			// 数据更新请求
		case v, _ := <-s.Shield:
			s.RefreshShieldData(v)
		}
	}
}

func (s *ShieldTaskService) SendShieldStaus(keyWord string) {
	var status bool
	for k, reg := range s.ShieldRegexp {
		if reg.MatchString(keyWord) {
			logs.Info("keyword [%s] match shield keyword [%s] with id [%d]", keyWord, s.ShieldDict[k].Match, s.ShieldDict[k].Id)
			status = true
		}
	}
	s.ShieldStatus <- status
}

func (s *ShieldTaskService) ReceiveKeyword(keyword string) {
	s.Keyword <- keyword
}

// 关闭task任务
func (s *ShieldTaskService) Stop() {
	close(s.closing)
}


func (s *ShieldTaskService) RefreshShieldData(v ShieldSearchServiceData) {
	switch v.Operation {
	case "insert":
		m := &BleakModel{
			Id:    v.Id,
			Match: v.Match,
		}
		s.ShieldDict = append(s.ShieldDict, m)
		s.ShieldRegexp = append(s.ShieldRegexp, regexp.MustCompile(m.Match))
	case "update":
		for i := 0; i < len(s.ShieldDict); i++ {
			if s.ShieldDict[i].Id == v.Id {
				s.ShieldDict[i].Match = v.Match
				s.ShieldRegexp[i] = regexp.MustCompile(v.Match)
			}
		}
	case "delete":
		var index = -1
		for i := 0; i < len(s.ShieldDict); i++ {
			if s.ShieldDict[i].Id == v.Id {
				index = i
				break
			}
		}
		if index != -1 {
			s.ShieldDict = append(s.ShieldDict[:index], s.ShieldDict[index+1:]...)
			s.ShieldRegexp = append(s.ShieldRegexp[:index], s.ShieldRegexp[index+1:]...)
		}
	}
}

// 同步信息到屏蔽词文件
func shieldInfoFileSync() {
	t := time.NewTicker(time.Hour)
	select {
	case <-t.C:
		s := InstanceShieldService()
		WriteShieldToFile(s.m)
	}
}

// 将屏蔽词写入文件
func WriteShieldToFile(m []*BleakModel) {
	var buf strings.Builder
	for i := 0; i < len(m); i++ {
		buf.WriteString(fmt.Sprintf("%d %s\n", m[i].Id, m[i].Match))
	}
	file, _ := os.Create(search_shield_file)
	file.WriteString(buf.String())
	//file.Sync()
	file.Close()
}

// 从文件读取屏蔽词
func GetSheildFromFile() (m []*BleakModel) {
	m = make([]*BleakModel, 0)
	file, _ := os.Open(search_shield_file)
	buff := bufio.NewReader(file) //读入缓存
	for {
		line, err := buff.ReadString('\n')
		if err != nil || err == io.EOF {
			break
		}
		m1 := new(BleakModel)
		fmt.Sscanf(line, "%d %s\n", &m1.Id, &m1.Match)
		m = append(m, m1)
	}
	file.Close()
	return
}


// 0 到 end 的随机数
func RangeInt(end int) int {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Intn(end)
}
