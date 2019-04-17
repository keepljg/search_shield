package main

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/micro/go-log"
	"github.com/micro/go-micro"
	"runtime"
	_ "search_shield/beego-http/routers"
	"search_shield/bleakService"
	"search_shield/config"
	"search_shield/handler"
	searchBleak "search_shield/proto/search-bleak"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU()) // 添加程序的多核执行代码
	if config.SearchListConfig.DevMode == "dev" {
		logs.SetLogger(logs.AdapterConsole)
	} else {
		logs.SetLogger(logs.AdapterMultiFile, `{"filename":"./logs/project.log","separate":["error", "warning", "info", "debug"]}`)
	}
	logs.Async()

	// New Service
	bleakService.InitShieldService()

	if config.SearchListConfig.RunMode == "rpc" {
		service := micro.NewService(
			micro.Name("go.micro.srv.search_shield"),
			micro.Version("latest"),
		)

		// Initialise bleakService
		service.Init()

		// Register Handler
		searchBleak.RegisterSearchBleakHandler(service.Server(), new(handler.SearchBleak))

		// Run bleakService
		if err := service.Run(); err != nil {
			log.Fatal(err)
		}
	} else {
		beego.Run()
	}

}
