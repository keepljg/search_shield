package main

import (
	"context"
	"fmt"
	"github.com/micro/go-micro"
	pb "search_shield/proto/search-bleak"
)
func main() {
	service := micro.NewService(
	)
	service.Init()
	cli := pb.NewSearchBleakService("go.micro.srv.search_shield", service.Client())
	resp, err := cli.CheckKeyword(context.TODO(), &pb.CheckWord{Keyword:"毛泽东"})
	cli.UpdateShieldData(context.TODO(), &pb.ShieldDateRequest{
		Id:                   20,
		Match:                "tutu",
		Operation:            "insert",
	})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(resp)
}


