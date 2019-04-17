package handler

import (
	"context"
	"time"
	"search_shield/config"
	pb "search_shield/proto/search-bleak"
	"search_shield/bleakService"
)

type SearchBleak struct{}

// 更新屏蔽词数据
func (this *SearchBleak) UpdateShieldData(ctx context.Context, req *pb.ShieldDateRequest, resp *pb.Response) error {
	shieldServices := bleakService.InstanceShieldServices()
	shied := bleakService.ShieldSearchServiceData{
		Id:        int(req.Id),
		Match:     req.Match,
		Operation: req.Operation,
	}
	for _, shieldService := range shieldServices {
		shieldService.ReceiveShield(shied)
	}
	if config.SearchListConfig.DevMode == "dev" {
		bleakService.InstanceShieldService().WriteShieldToFile()
	}
	resp.Status = &pb.Status{}
	resp.Status.Message = "success"
	resp.Status.Time = time.Now().String()
	resp.Status.Code = 0
	return nil
}


// 检查是否是屏蔽词
func (this *SearchBleak) CheckKeyword (ctx context.Context, keyword *pb.CheckWord, resp *pb.Response) error {
	s := bleakService.InstanceShieldService()
	s.ReceiveKeyword(keyword.Keyword)
	status := s.GetShieldStatus()
	if status {
		resp.Status = &pb.Status{}
		resp.Status.Message = "屏蔽词"
		resp.Status.Time = time.Now().String()
		resp.Status.Code = 1
		resp.Data = "yes"
	}
	return nil
}
