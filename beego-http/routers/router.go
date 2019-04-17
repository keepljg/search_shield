package routers

import (
	"github.com/astaxie/beego"
	"search_shield/beego-http/v1"
)

func init() {
	ns := beego.NewNamespace("/v1",
		beego.NSNamespace("/search",
			//beego.NSInclude(
			//	&v1.SearchBleakController{},
			//),
			beego.NSRouter("/shield/update", &v1.SearchBleakController{}, "post:UpdateShieldData"),
			beego.NSRouter("/shield/check", &v1.SearchBleakController{}, "get:CheckKeyword"),
		),
	)
	beego.AddNamespace(ns)
}