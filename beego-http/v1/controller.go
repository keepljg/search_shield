package v1

import (
	"github.com/astaxie/beego"
	"time"
	"search_shield/bleakService"
	"search_shield/config"
)

type JSON map[string]interface{}

type BaseController struct {
	beego.Controller
}

func (this *BaseController) Success(data interface{}) {
	this.Data["json"] = JSON{
		"status": JSON{
			"code":    0,
			"message": "success",
			"time":    time.Now().Format("2006-01-02 15:04:05"),
		},
		"data": data,
	}
	this.ServeJSON()
}

func (this *BaseController) Error(errCode int, errMessage string, data interface{}) {

	if errCode == 404 {
		this.Ctx.ResponseWriter.WriteHeader(404)
	}
	this.Data["json"] = JSON{
		"status": JSON{
			"code":    errCode,
			"message": errMessage,
			"time":    time.Now().Format("2006-01-02 15:04:05"),
		},
		"data": data,
	}
	this.ServeJSON()
}



type SearchBleakController struct {
	BaseController
}

type UpdateShieldDateRequest struct {
	Id        int    `json:"id"        form:"id"`
	Match     string `json:"match"     form:"match"`
	Operation string `json:"operation" form:"operation"`
}

//func (c *SearchBleakController) URLMapping() {
//	c.Mapping("UpdateShieldData", c.UpdateShieldData)
//	c.Mapping("CheckKeyword", c.CheckKeyword)
//}

// 搜索屏蔽词更新
// @routers /shield/update [post]
func (this *SearchBleakController) UpdateShieldData() {
	inputs := UpdateShieldDateRequest{}
	this.ParseForm(&inputs)
	shieldServices := bleakService.InstanceShieldServices()
	shied := bleakService.ShieldSearchServiceData{
		Id:        int(inputs.Id),
		Match:     inputs.Match,
		Operation: inputs.Operation,
	}
	for _, shieldService := range shieldServices {
		shieldService.ReceiveShield(shied)
	}

	if config.SearchListConfig.DevMode == "dev" {
		bleakService.InstanceShieldService().WriteShieldToFile()
	}
	this.Success(struct{}{})
}


// 搜索屏蔽词检查
// @routers /shield/check [get]
func (this *SearchBleakController) CheckKeyword() {
	keyword := this.Input().Get("keyword")
	s := bleakService.InstanceShieldService()
	s.ReceiveKeyword(keyword)
	status := s.GetShieldStatus()
	if status {
		this.Success("yes")
	} else {
		this.Success("no")
	}
}